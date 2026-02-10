package db

import (
	"context"
	"crypto/tls"
	"fmt"
	"hash/crc32"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/redis/go-redis/v9"
)

const defaultBatchSize = 500

type redisOpt struct {
	tls   *tls.Config
	addr  string
	user  string
	pwd   string
	dbcnt int
}
type RedisOptions func(opt *redisOpt)

var defaultRedisOpt = redisOpt{
	addr:  "127.0.0.1:6379",
	user:  "",
	pwd:   "",
	dbcnt: 0,
	tls:   &tls.Config{InsecureSkipVerify: true},
}

// WithRedisAddr sets the redis address.
func WithRedisAddr(s string) RedisOptions {
	return func(o *redisOpt) {
		o.addr = s
	}
}

// WithRedisUser sets the redis username.
func WithRedisUser(s string) RedisOptions {
	return func(o *redisOpt) {
		o.user = s
	}
}

// WithRedisPwd sets the redis password.
func WithRedisPwd(s string) RedisOptions {
	return func(o *redisOpt) {
		o.pwd = s
	}
}

// WithRedisDB sets the redis database index.
func WithRedisDB(n int) RedisOptions {
	return func(o *redisOpt) {
		o.dbcnt = n
	}
}

// WithRedisTLS sets the TLS config for redis.
func WithRedisTLS(t *tls.Config) RedisOptions {
	return func(o *redisOpt) {
		if t != nil {
			o.tls = t
		}
	}
}

type RedisCli struct {
	cli     *redis.Client
	mainver int
}

// Close closes the redis client.
func (rdb *RedisCli) Close() error {
	if rdb.cli == nil {
		return nil
	}
	return rdb.cli.Close()
}

// Cli returns the underlying redis client.
func (rdb *RedisCli) Cli() *redis.Client {
	return rdb.cli
}

// MainVer returns the major version of redis server.
func (rdb *RedisCli) MainVer() int {
	return rdb.mainver
}

// todo: may be next version
// NewRedisClient creates a redis client with options.
func NewRedisClient(opts ...RedisOptions) *RedisCli {
	opt := defaultRedisOpt
	for _, o := range opts {
		o(&opt)
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:      opt.addr,
		Username:  opt.user,
		Password:  opt.pwd,
		DB:        opt.dbcnt,
		TLSConfig: opt.tls,
	})
	cli := &RedisCli{
		cli: rdb,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	a, err := rdb.Info(ctx, "Server").Result()
	if err == nil {
		for v := range strings.SplitSeq(a, "\r\n") {
			if strings.HasPrefix(v, "redis_version:") {
				cli.mainver, _ = strconv.Atoi(strings.Split(strings.Split(v, ":")[1], ".")[0])
				break
			}
		}
	}
	return cli
}

type RedisSharder struct {
	client       *redis.Client
	baseKey      string
	numShards    uint32
	maxBatchSize uint32
	mainVersion  uint32
}

// NewRedisSharder 创建一个优化的 Redis 分片器
// 参数：
// - client: 已初始化的 Redis 客户端
// - baseKey: 用于分片的基础 Key 前缀
// - shards: 分片数量，建议 32-512 之间
// - maxBatchSize: 每次批量操作的最大字段数，建议 100-3000 之间
func NewRedisSharder(client *redis.Client, baseKey string, shards, maxBatchSize, mainver uint32) *RedisSharder {
	maxBatchSize = min(max(100, maxBatchSize), 3000)
	shards = min(max(32, shards), 512)
	return &RedisSharder{
		client:       client,
		baseKey:      baseKey,
		numShards:    shards,
		maxBatchSize: maxBatchSize,
		mainVersion:  mainver,
	}
}

// getShardKey 使用 unsafe 零拷贝 + CRC32 硬件加速计算分片
func (s *RedisSharder) getShardKey(field string) string {
	// Go 1.20+ 推荐的零拷贝 string 转 []byte 方式
	// 避免了海量 field 转换时的内存分配和拷贝开销
	b := unsafe.Slice(unsafe.StringData(field), len(field))

	// ChecksumIEEE 在现代 CPU 上直接映射为硬件指令，极其高效
	shardID := crc32.ChecksumIEEE(b) % s.numShards
	return fmt.Sprintf("%s:%d", s.baseKey, shardID)
}

// Set writes a field to the shard.
func (s *RedisSharder) Set(ctx context.Context, field, value string) error {
	return s.client.HSet(ctx, s.getShardKey(field), field, value).Err()
}

// BatchSet 针对 10w+ 数据的极致优化写入
func (s *RedisSharder) BatchSet(ctx context.Context, data map[string]string) error {
	// 1. 将数据按分片 Key 进行预归类
	// 分片 ID -> {field: value, ...}
	groupedData := make(map[string]map[string]string)
	for f, v := range data {
		sk := s.getShardKey(f)
		if _, ok := groupedData[sk]; !ok {
			groupedData[sk] = make(map[string]string)
		}
		groupedData[sk][f] = v
	}

	// 2. 分批次执行 Pipeline
	// 我们不希望一次性塞 10w 个 HSET 到一个 Pipeline，这样会阻塞网络和 Redis 缓冲区
	pipe := s.client.Pipeline()
	if s.mainVersion > 3 {
		for sk, fields := range groupedData {
			// 根据 Redis 版本选择：
			// 如果是 Redis 4.0+，直接一条命令传 Map：pipe.HSet(ctx, sk, fields)
			// 如果是旧版，循环写入对：
			pipe.HSet(ctx, sk, fields)
			// for f, v := range fields {
			// 	pipe.HSet(ctx, sk, f, v)
			// 	opCount++

			// 	// 达到阈值即提交，释放内存并防止阻塞
			// 	if opCount >= s.maxBatchSize {
			// 		if _, err := pipe.Exec(ctx); err != nil {
			// 			return err
			// 		}
			// 		opCount = 0
			// 	}
			// }
		}
	} else {
		opCount := uint32(0)
		for sk, fields := range groupedData {
			// 根据 Redis 版本选择：
			// 如果是 Redis 4.0+，直接一条命令传 Map：pipe.HSet(ctx, sk, fields)
			// 如果是旧版，循环写入对：
			for f, v := range fields {
				pipe.HSet(ctx, sk, f, v)
				opCount++

				// 达到阈值即提交，释放内存并防止阻塞
				if opCount >= s.maxBatchSize {
					if _, err := pipe.Exec(ctx); err != nil {
						return err
					}
					opCount = 0
				}
			}
		}
		// 旧版，提交剩余的指令
		if opCount > 0 {
			_, err := pipe.Exec(ctx)
			return err
		}
	}

	_, err := pipe.Exec(ctx)
	return err
}

// Get 读取单个字段
func (s *RedisSharder) Get(ctx context.Context, field string) (string, error) {
	return s.client.HGet(ctx, s.getShardKey(field), field).Result()
}

// ScanAll iterates all shards and calls handler for each field/value.
func (s *RedisSharder) ScanAll(ctx context.Context, handler func(k, f, v string) bool) error {
	for i := uint32(0); i < s.numShards; i++ {
		sk := fmt.Sprintf("%s:%d", s.baseKey, i)
		// 使用 HSCAN 迭代每一个分片，避免阻塞
		iter := s.client.HScan(ctx, sk, 0, "", defaultBatchSize).Iterator()
		for iter.Next(ctx) {
			field := iter.Val()
			if iter.Next(ctx) { // HScan 返回的是 field, value, field, value...
				value := iter.Val()
				if !handler(sk, field, value) {
					return nil
				}
			}
		}
		if err := iter.Err(); err != nil {
			return err
		}
	}
	return nil
}

// GetAll returns all fields across all shards.
func (s *RedisSharder) GetAll(ctx context.Context) (map[string]string, error) {
	result := make(map[string]string)
	for i := uint32(0); i < s.numShards; i++ {
		sk := fmt.Sprintf("%s:%d", s.baseKey, i)
		data, err := s.client.HGetAll(ctx, sk).Result()
		if err != nil && err != redis.Nil {
			return nil, err
		}
		for k, v := range data {
			result[k] = v
		}
	}
	return result, nil
}

// GetByPrefix returns all fields with the given prefix.
func (s *RedisSharder) GetByPrefix(ctx context.Context, prefix string) (map[string]string, error) {
	result := make(map[string]string)
	err := s.ScanAll(ctx, func(key, field, value string) bool {
		if len(field) >= len(prefix) && field[:len(prefix)] == prefix {
			result[field] = value
		}
		return true
	})
	return result, err
}

// GetBySuffix returns all fields with the given suffix.
func (s *RedisSharder) GetBySuffix(ctx context.Context, suffix string) (map[string]string, error) {
	result := make(map[string]string)
	err := s.ScanAll(ctx, func(key, field, value string) bool {
		if len(field) >= len(suffix) && field[len(field)-len(suffix):] == suffix {
			result[field] = value
		}
		return true
	})
	return result, err
}

// UnlinkAll 删除所有分片数据
func (s *RedisSharder) UnlinkAll(ctx context.Context) error {
	// 1. 构造所有的分片 Key
	shardKeys := make([]string, s.numShards)
	for i := uint32(0); i < s.numShards; i++ {
		shardKeys[i] = fmt.Sprintf("%s:%d", s.baseKey, i)
	}

	// 2. 使用 UNLINK 一次性删除
	// UNLINK 是非阻塞的，即使某些分片很大，也不会卡死 Redis
	// go-redis 的 Unlink 接受变长参数
	return s.client.Unlink(ctx, shardKeys...).Err()
}

// Delete 删除多个字段
func (s *RedisSharder) Delete(ctx context.Context, fields ...string) error {
	if len(fields) == 0 {
		return nil
	}
	// 1. 将字段按分片 Key 进行预归类
	groupedFields := make(map[string][]string)
	for _, f := range fields {
		sk := s.getShardKey(f)
		groupedFields[sk] = append(groupedFields[sk], f)
	}

	// 2. 分批次执行 Pipeline 删除
	pipe := s.client.Pipeline()
	// opCount := uint32(0) // 分片数量不会大于批量上限，不需要计数

	for sk, fs := range groupedFields {
		pipe.HDel(ctx, sk, fs...)
		// opCount++

		// // 达到阈值即提交
		// if opCount >= s.maxBatchSize {
		// 	if _, err := pipe.Exec(ctx); err != nil {
		// 		return err
		// 	}
		// 	opCount = 0
		// }
	}

	// 提交剩余的指令
	// if opCount > 0 {
	// 	if _, err := pipe.Exec(ctx); err != nil {
	// 		return err
	// 	}
	// }
	_, err := pipe.Exec(ctx)
	return err
}

// Exists checks whether a field exists.
func (s *RedisSharder) Exists(ctx context.Context, field string) (bool, error) {
	count, err := s.client.HExists(ctx, s.getShardKey(field), field).Result()
	return count, err
}

// DeleteBySuffix deletes all fields with the given suffix.
func (s *RedisSharder) DeleteBySuffix(ctx context.Context, suffix string) error {
	suffixLen := len(suffix)
	for i := uint32(0); i < s.numShards; i++ {
		sk := fmt.Sprintf("%s:%d", s.baseKey, i)
		// 使用 HSCAN 迭代每一个分片，避免阻塞
		iter := s.client.HScan(ctx, sk, 0, "*"+suffix, defaultBatchSize).Iterator()
		var fieldsToDelete []string
		for iter.Next(ctx) {
			field := iter.Val()
			if len(field) >= suffixLen && field[len(field)-suffixLen:] == suffix {
				fieldsToDelete = append(fieldsToDelete, field)
			}
			if iter.Next(ctx) { // HScan 返回的是 field, value, field, value...
				_ = iter.Val() // 忽略 value
			}
		}
		if err := iter.Err(); err != nil {
			return err
		}
		if len(fieldsToDelete) > 0 {
			if err := s.Delete(ctx, fieldsToDelete...); err != nil {
				return err
			}
		}
	}
	return nil
}

// DeleteByPrefix deletes all fields with the given prefix.
func (s *RedisSharder) DeleteByPrefix(ctx context.Context, prefix string) error {
	prefixLen := len(prefix)
	for i := uint32(0); i < s.numShards; i++ {
		sk := fmt.Sprintf("%s:%d", s.baseKey, i)
		// 使用 HSCAN 迭代每一个分片，避免阻塞
		iter := s.client.HScan(ctx, sk, 0, prefix+"*", defaultBatchSize).Iterator()
		var fieldsToDelete []string
		for iter.Next(ctx) {
			field := iter.Val()
			if len(field) >= prefixLen && field[:prefixLen] == prefix {
				fieldsToDelete = append(fieldsToDelete, field)
			}
			if iter.Next(ctx) { // HScan 返回的是 field, value, field, value...
				_ = iter.Val() // 忽略 value
			}
		}
		if err := iter.Err(); err != nil {
			return err
		}
		if len(fieldsToDelete) > 0 {
			if err := s.Delete(ctx, fieldsToDelete...); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *RedisSharder) FindKey(field string) string {
	return s.getShardKey(field)
}
