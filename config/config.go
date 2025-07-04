// Package config 类ini的配置文件库，支持注释信息
package config

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/xyzj/toolbox/json"
	"github.com/xyzj/toolbox/mapfx"
	"gopkg.in/yaml.v3"
)

// File 配置文件
type File struct {
	items    *mapfx.StructMap[string, Item]
	data     *bytes.Buffer
	filepath string
}

// Item 配置内容，包含注释，key,value,是否加密value
type Item struct {
	Value   *Value `json:"value" yaml:"value"`
	Comment string `json:"comment" yaml:"comment"`
	Key     string `json:"-" yaml:"-"`
	// EncryptValue bool   `json:"-" yaml:"-"`
}

// String 把配置项格式化成字符串
func (i *Item) String() string {
	if i.Comment == "" {
		return "\n" + i.Key + "=" + i.Value.String() + "\n"
	}
	ss := strings.Split(i.Comment, "\n")
	xcom := ""
	for _, v := range ss {
		if strings.HasPrefix(v, "#") {
			xcom = v + "\n"
		} else {
			xcom = "# " + v + "\n"
		}
	}
	return "\n" + xcom + i.Key + "=" + i.Value.String() + "\n" // fmt.Sprintf("\n%s%s=%s\n", xcom, i.Key, i.Value)
}

// NewConfig 创建一个key:value格式的配置文件
//
//	依据文件的扩展名，支持yaml和json格式的文件
func NewConfig(filepath string) *File {
	f := &File{
		items: mapfx.NewStructMap[string, Item](),
		data:  &bytes.Buffer{},
	}
	f.FromFile(filepath)
	return f
}

// Keys 获取所有Key
func (f *File) Keys() []string {
	return f.items.Keys()
	// ss := make([]string, 0, f.items.Len())
	// f.items.ForEach(func(key string, value *Item) bool {
	// 	ss = append(ss, key)
	// 	return true
	// })
	// return ss
}

// Clear 清空配置项
func (f *File) Clear() {
	f.items.Clear()
	f.data.Reset()
}

// DelItem 删除配置项
func (f *File) DelItem(key string) {
	f.items.Delete(key)
}

// PutItem 添加配置项
func (f *File) PutItem(item *Item) {
	// if item.EncryptValue {
	// 	item.Value = NewValue(toolbox.CodeString(item.Value.String()))
	// }
	if v, ok := f.items.Load(item.Key); ok {
		if item.Comment == "" {
			item.Comment = v.Comment
		}
	}
	f.items.Store(item.Key, item)
}

// GetDefault 读取一个配置，若不存在，则添加这个配置
func (f *File) GetDefault(item *Item) *Value {
	if v, ok := f.items.Load(item.Key); ok {
		return v.Value
	}
	f.PutItem(item)
	return item.Value
}

// GetItem 获取一个配置值
func (f *File) GetItem(key string) *Value {
	if v, ok := f.items.Load(key); ok {
		return v.Value
	}
	return EmptyValue
}

// ForEach 遍历所有值
func (f *File) ForEach(do func(key string, value *Value) bool) {
	f.items.ForEach(func(key string, value *Item) bool {
		return do(key, value.Value)
	})
}

// Len 获取配置数量
func (f *File) Len() int {
	return f.items.Len()
}

// Has 判断key是否存在
func (f *File) Has(key string) bool {
	return f.items.Has(key)
}

// Print 返回所有配置项
func (f *File) Print() string {
	x := make([]*Item, 0, f.items.Len())
	f.items.ForEach(func(key string, value *Item) bool {
		x = append(x, value)
		return true
	})
	sort.Slice(x, func(i, j int) bool {
		return x[i].Key < x[j].Key
	})
	f.data.Reset()
	for _, v := range x {
		f.data.WriteString(v.String())
	}
	return f.data.String()
}

func (f *File) PrintJSON() string {
	var js string
	f.items.ForEach(func(key string, value *Item) bool {
		it, _ := sjson.Set("", "key", key)
		it, _ = sjson.Set(it, "value", value.Value.String())
		it, _ = sjson.Set(it, "comment", value.Comment)
		js, _ = sjson.Set(js, "data.-1", gjson.Parse(it).Value())
		return true
	})
	return js
}

// GetAll 返回所有配置项
func (f *File) GetAll() string {
	x := f.items.Clone()
	buf := make([]string, 0)
	for k, v := range x {
		buf = append(buf, "\""+k+"\":\""+v.Value.String()+"\"")
	}
	return "{" + strings.Join(buf, ",") + "}"
}

// FromFile 从文件载入配置
func (f *File) FromFile(configfile string) error {
	if configfile != "" {
		f.filepath = configfile
	}
	if f.filepath == "" {
		return nil
	}
	if f.data == nil {
		f.data = &bytes.Buffer{}
	} else {
		f.data.Reset()
	}
	if f.items == nil {
		f.items = mapfx.NewStructMap[string, Item]()
	} else {
		f.items.Clear()
	}
	b, err := os.ReadFile(f.filepath)
	if err != nil {
		return err
	}
	if len(b) == 0 {
		return nil
	}
	f.data.Write(b)
	if b[0] == '{' {
		if f.fromJSON(b) == nil {
			return nil
		}
	}
	if f.fromYAML(b) == nil {
		return nil
	}
	ss := strings.Split(f.data.String(), "\n")
	tip := make([]string, 0)
	for _, v := range ss {
		s := strings.TrimSpace(v)
		if strings.HasPrefix(s, "#") {
			if xt := strings.TrimSpace(s[1:]); xt != "" {
				tip = append(tip, xt)
			}
			continue
		}
		it := strings.Split(s, "=")
		if len(it) != 2 {
			continue
		}
		f.items.Store(it[0], &Item{Key: it[0], Value: NewValue(it[1]), Comment: strings.Join(tip, "\n")})
		tip = []string{}
	}
	return nil
}

// SaveTo 将配置写入指定文件，依据文件扩展名判断写入格式
func (f *File) SaveTo(filename string) error {
	f.filepath = filename
	return f.ToFile()
}

// Save 将配置写入文件，依据文件扩展名判断写入格式
func (f *File) Save() error {
	return f.ToFile()
}

// ToFile 将配置写入文件，依据文件扩展名判断写入格式
func (f *File) ToFile() error {
	switch strings.ToLower(filepath.Ext(f.filepath)) {
	case ".yaml":
		return f.ToYAML()
	case ".json":
		return f.ToJSON()
	}
	f.Print()
	return os.WriteFile(f.filepath, f.data.Bytes(), 0o644)
}

// ToYAML 保存为yaml格式文件
func (f *File) ToYAML() error {
	b, err := yaml.Marshal(f.items.Clone())
	if err != nil {
		return err
	}
	return os.WriteFile(f.filepath, b, 0o644)
}

// ToJSON 保存为json格式文件
func (f *File) ToJSON() error {
	b, err := json.MarshalIndent(f.items.Clone(), "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(f.filepath, b, 0o644)
}

func (f *File) fromYAML(b []byte) error {
	x := make(map[string]*Item)
	err := yaml.Unmarshal(b, &x)
	if err != nil {
		return err
	}
	for k, v := range x {
		f.items.Store(k, &Item{Key: k, Value: v.Value, Comment: v.Comment})
	}
	return nil
}

func (f *File) fromJSON(b []byte) error {
	x := make(map[string]*Item)
	err := json.Unmarshal(b, &x)
	if err != nil {
		return err
	}
	for k, v := range x {
		f.items.Store(k, &Item{Key: k, Value: v.Value, Comment: v.Comment})
	}
	return nil
}
