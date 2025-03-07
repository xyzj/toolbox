package storage

import (
	"time"

	"github.com/tidwall/gjson"
	"github.com/xyzj/toolbox/db"
	"github.com/xyzj/toolbox/llms"
	"github.com/xyzj/toolbox/loopfunc"
)

type FileStorage struct {
	f  string
	db *db.BoltDB
}

func NewFileStorage(filename string) llms.Storage {
	return &FileStorage{
		f: filename,
	}
}

func (s *FileStorage) Init() error {
	db, err := db.NewBolt(s.f)
	if err != nil {
		return err
	}
	s.db = db
	go loopfunc.LoopFunc(func(params ...interface{}) {
		t := time.NewTicker(time.Hour)
		for range t.C {
			s.db.ForEach(func(k, v string) error {
				u := gjson.Parse(v).Get("last_update").Int()
				if time.Now().Unix()-u > 60*60*24*7 {
					s.db.Delete(k)
				}
				return nil
			})
		}
	}, "", nil)
	return nil
}

func (s *FileStorage) Clear(d time.Duration) {
	s.db.ForEach(func(k, v string) error {
		u := gjson.Parse(v).Get("last_update").Int()
		if time.Now().Unix()-u > int64(d.Seconds()) {
			s.db.Delete(k)
		}
		return nil
	})
}

func (s *FileStorage) Import() (map[string]*llms.ChatData, error) {
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

func (s *FileStorage) Update(d *llms.ChatData) error {
	s.db.Write(d.ID, d.ToJSON())
	return nil
}
