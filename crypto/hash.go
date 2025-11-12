package crypto

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"hash"

	"github.com/tjfoc/gmsm/sm3"
	gopool "github.com/xyzj/go-pool"
)

type HashType byte

const (
	// HashMD5 md5算法
	HashMD5 HashType = iota
	// HashSHA256 sha256算法
	HashSHA256
	// HashSHA512 sha512算法
	HashSHA512
	// HashHMACSHA1 hmacsha1摘要算法
	HashHMACSHA1
	// HashHMACSHA256 hmacsha256摘要算法
	HashHMACSHA256
	// HashSHA1 sha1算法
	HashSHA1
	// HashSM3 国密sm3
	HashSM3
)

type HashOpt struct {
	hmackey  []byte
	poolsize int
}
type HashOpts func(opt *HashOpt)

func HashOptPoolSize(t int) HashOpts {
	return func(o *HashOpt) {
		o.poolsize = t
	}
}
func HashOptHMacKey(b []byte) HashOpts {
	return func(o *HashOpt) {
		o.hmackey = b
	}
}

// HASH hash算法
type HASH struct {
	// locker   sync.Mutex
	// hash     hash.Hash
	// pool sync.Pool
	pool *gopool.GoPool[hash.Hash]
	// workType HashType
}

// SetHMACKey 设置hmac算法的key
// func (w *HASH) SetHMACKey(key []byte) {
// 	switch w.workType {
// 	case HashHMACSHA1:
// 		w.hash = hmac.New(sha1.New, key)
// 	case HashHMACSHA256:
// 		w.hash = hmac.New(sha256.New, key)
// 	}
// }

// Hash 计算哈希值
func (w *HASH) Hash(b []byte) CValue {
	h := w.pool.Get()
	defer w.pool.Put(h)
	h.Reset()
	h.Write(b)
	return CValue(h.Sum(nil))
	// w.locker.Lock()
	// defer w.locker.Unlock()
	// w.hash.Reset()
	// w.hash.Write(b)
	// return CValue(w.hash.Sum(nil))
}

// NewHash creates a new hash algorithm instance based on the provided hash type and HMAC key.
// It uses a sync.Pool to reuse hash instances, improving performance.
//
// Parameters:
// - t: The hash type, which can be one of the following: HashMD5, HashHMACSHA1, HashHMACSHA256, HashSHA1, HashSHA256, HashSHA512, HashSM3.
// - hmacKey: The HMAC key to be used for HashHMACSHA1 and HashHMACSHA256 hash types. If not required, an empty slice can be passed.
//
// Returns:
// - A pointer to the newly created HASH instance.
func NewHash(t HashType, opts ...HashOpts) *HASH {
	opt := &HashOpt{
		hmackey: []byte{},
	}
	for _, o := range opts {
		o(opt)
	}
	return &HASH{
		pool: gopool.New(func() hash.Hash {
			switch t {
			case HashMD5:
				return md5.New()
			case HashHMACSHA1:
				return hmac.New(sha1.New, opt.hmackey)
			case HashHMACSHA256:
				return hmac.New(sha256.New, opt.hmackey)
			case HashSHA1:
				return sha1.New()
			case HashSHA256:
				return sha256.New()
			case HashSHA512:
				return sha512.New()
			case HashSM3:
				return sm3.New()
			default:
				return nil
			}
		},
			gopool.WithMaxIdleSize(uint32(opt.poolsize)),
		),
	}
	// w := &HASH{
	// 	locker: sync.Mutex{},
	// 	workType: t,
	// }
	// switch t {
	// case HashMD5:
	// 	w.hash = md5.New()
	// case HashHMACSHA1:
	// 	w.hash = hmac.New(sha1.New, []byte{})
	// case HashHMACSHA256:
	// 	w.hash = hmac.New(sha256.New, []byte{})
	// case HashSHA1:
	// 	w.hash = sha1.New()
	// case HashSHA256:
	// 	w.hash = sha256.New()
	// case HashSHA512:
	// 	w.hash = sha512.New()
	// case HashSM3:
	// 	w.hash = sm3.New()
	// }
	// return w
}
