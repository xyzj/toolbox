package db

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/xyzj/toolbox/json"
	"go.etcd.io/bbolt"
)

// BoltDB bolt数据文件实例
type BoltDB struct {
	cli      *bbolt.DB
	filename string
}

func (c *BoltDB) Write(bucket, key, value string) error {
	if c.cli == nil {
		return fmt.Errorf("bolt client is not initialized")
	}
	if bucket == "" {
		bucket = "default"
	}
	return c.cli.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}
		return b.Put([]byte(key), []byte(value))
	})
}

func (c *BoltDB) Read(bucket, key string) (string, error) {
	if c.cli == nil {
		return "", fmt.Errorf("bolt client is not initialized")
	}
	if bucket == "" {
		bucket = "default"
	}
	var value string
	err := c.cli.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucket)
		}
		v := b.Get([]byte(key))
		if v == nil {
			return fmt.Errorf("key %s not found in bucket %s", key, bucket)
		}
		value = string(v)
		return nil
	})
	return value, err
}
func (c *BoltDB) Delete(bucket, key string) error {
	if c.cli == nil {
		return fmt.Errorf("bolt client is not initialized")
	}
	if bucket == "" {
		bucket = "default"
	}
	return c.cli.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucket)
		}
		return b.Delete([]byte(key))
	})
}
func (c *BoltDB) DeleteBucket(bucket string) error {
	if c.cli == nil {
		return fmt.Errorf("bolt client is not initialized")
	}
	if bucket == "" {
		bucket = "default"
	}
	return c.cli.Update(func(tx *bbolt.Tx) error {
		return tx.DeleteBucket([]byte(bucket))
	})
}

func (c *BoltDB) List(bucket string) (map[string]string, error) {
	if c.cli == nil {
		return nil, fmt.Errorf("bolt client is not initialized")
	}
	if bucket == "" {
		bucket = "default"
	}
	result := make(map[string]string)
	err := c.cli.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucket)
		}
		return b.ForEach(func(k, v []byte) error {
			result[string(k)] = string(v)
			return nil
		})
	})
	return result, err
}
func (c *BoltDB) ListBuckets() ([]string, error) {
	if c.cli == nil {
		return nil, fmt.Errorf("bolt client is not initialized")
	}
	var buckets []string
	err := c.cli.View(func(tx *bbolt.Tx) error {
		return tx.ForEach(func(name []byte, _ *bbolt.Bucket) error {
			buckets = append(buckets, string(name))
			return nil
		})
	})
	return buckets, err
}
func (c *BoltDB) Exists(bucket, key string) (bool, error) {
	if c.cli == nil {
		return false, fmt.Errorf("bolt client is not initialized")
	}
	if bucket == "" {
		bucket = "default"
	}
	var exists bool
	err := c.cli.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucket)
		}
		v := b.Get([]byte(key))
		exists = v != nil
		return nil
	})
	return exists, err
}

func (c *BoltDB) Health() error {
	if c.cli == nil {
		return fmt.Errorf("bolt client is not initialized")
	}
	err := c.cli.View(func(tx *bbolt.Tx) error {
		return nil
	})
	return err
}

func (c *BoltDB) Close() error {
	if c.cli == nil {
		return nil
	}
	return c.cli.Close()
}

// ForEach 遍历所有key,value
func (b *BoltDB) ForEach(bucket string, f func(k, v string) error) {
	if b.cli == nil {
		return
	}
	var buc []byte
	if bucket == "" {
		bucket = "default"
	}
	buc = json.Bytes(bucket)
	data := make(map[string]string)
	b.cli.View(func(tx *bbolt.Tx) error {
		t := tx.Bucket(buc)
		if t == nil {
			return nil
		}
		return t.ForEach(func(k, v []byte) error {
			data[json.String(k)] = json.String(v)
			return nil
			// defer func() {
			// 	recover()
			// }()
			// return f(json.String(k), json.String(v))
		})
	})
	defer func() {
		recover()
	}()
	for k, v := range data {
		f(k, v)
	}
}

// NewBolt 创建一个新的bolt数据文件
func NewBolt(f string) (*BoltDB, error) {
	dir := filepath.Dir(f)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}
	db, err := bbolt.Open(f, 0o640, &bbolt.Options{Timeout: time.Second * 2})
	if err != nil {
		return nil, err
	}

	return &BoltDB{
		cli:      db,
		filename: f,
	}, nil
}
