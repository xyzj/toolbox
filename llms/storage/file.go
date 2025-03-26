package storage

import (
	"time"

	"github.com/tidwall/gjson"
	"github.com/xyzj/toolbox/db"
	"github.com/xyzj/toolbox/llms"
)

type FileStorage struct {
	f  string
	db *db.BoltDB
}

func NewFileStorage(filename string) (llms.Storage, error) {
	d, err := db.NewBolt(filename)
	if err != nil {
		return nil, err
	}
	// go func() {
	// 	t := time.NewTicker(time.Minute)
	// 	for range t.C {
	// 		d.ForEach(func(k, v string) error {
	// 			if time.Since(time.Unix(gjson.Get(v, "last_update").Int(), 0)) > lifetime {
	// 				d.Delete(k)
	// 			}
	// 			return nil
	// 		})
	// 	}
	// }()
	return &FileStorage{
		f:  filename,
		db: d,
	}, nil
}

func (s *FileStorage) Clear() {
	s.db.ForEach(func(k, v string) error {
		s.db.Delete(k)
		return nil
	})
}

func (s *FileStorage) Load() (map[string]*llms.ChatData, error) {
	data := make(map[string]*llms.ChatData)
	var err error
	s.db.ForEach(func(k, v string) error {
		x := &llms.ChatData{}
		err = x.FromJSON(v)
		if err != nil {
			return err
		}
		data[k] = x
		return nil
	})
	return data, nil
}

func (s *FileStorage) Store(d *llms.ChatData) error {
	return s.db.Write(d.ID, d.ToJSON())
}

func (s *FileStorage) RemoveDead(t time.Duration) {
	s.db.ForEach(func(k, v string) error {
		if time.Since(time.Unix(gjson.Get(v, "last_update").Int(), 0)) > t {
			s.db.Delete(k)
		}
		return nil
	})
}
