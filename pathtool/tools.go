// Package pathtool some path-related methods
package pathtool

import (
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// SliceFlag 切片型参数，仅支持字符串格式
type SliceFlag []string

// String 返回参数
func (f *SliceFlag) String() string {
	return strings.Join(*f, ", ")
}

// Set 设置值
func (f *SliceFlag) Set(value string) error {
	*f = append(*f, value)
	return nil
}

// IsExist file is exist or not
func IsExist(p string) bool {
	if p == "" {
		return false
	}
	_, err := os.Stat(p)
	return err == nil || os.IsExist(err)
}

// GetExecFullpath get current file path
func GetExecFullpath() string {
	return JoinPathFromHere(GetExecName())
}

// GetExecDir get current file path
func GetExecDir() string {
	a, _ := os.Executable()
	execdir := filepath.Dir(a)
	if strings.Contains(execdir, "go-build") {
		execdir, _ = filepath.Abs(".")
	}
	return execdir
}

// GetExecName 获取可执行文件的名称
func GetExecName() string {
	exe, _ := os.Executable()
	if exe == "" {
		return ""
	}
	return filepath.Base(exe)
}

// GetExecNameWithoutExt 获取可执行文件的名称,去除扩展名
func GetExecNameWithoutExt() string {
	name := GetExecName()
	return strings.ReplaceAll(name, filepath.Ext(name), "")
}

// JoinPathFromHere 从程序执行目录开始拼接路径
func JoinPathFromHere(path ...string) string {
	s := []string{GetExecDir()}
	s = append(s, path...)
	sp := filepath.Join(s...)
	p, err := filepath.Abs(sp)
	if err != nil {
		return sp
	}
	return p
}

// AddPathEnvFromHere add `pwd` to PATH env
func AddPathEnvFromHere() error {
	p := GetExecDir()
	if strings.Contains(os.Getenv("PATH"), p) {
		return nil
	}
	u, err := user.Current()
	if err != nil {
		return err
	}
	b, err := os.OpenFile(filepath.Join(u.HomeDir, ".bashrc"), os.O_WRONLY|os.O_APPEND, 0o664)
	if err != nil {
		return err
	}
	b.WriteString(`export PATH="$PATH:` + p + "\"")
	b.Close()
	return nil
}

// MakeRuntimeDirs creates and returns the full paths for the configuration, log, and cache directories.
// If the rootpath is ".", it uses the directories relative to the current executable's location.
// If the rootpath is "..", it uses the directories relative to the parent directory of the current executable's location.
// For any other rootpath, it creates the directories under the specified rootpath.
//
// Parameters:
// rootpath (string): The root path for creating the directories.
//
// Returns:
// string: The full path of the configuration directory.
// string: The full path of the log directory.
// string: The full path of the cache directory.
func MakeRuntimeDirs(rootpath string) (string, string, string) {
	var sconf, slog, scache string
	switch rootpath {
	case ".":
		sconf = JoinPathFromHere("conf")
		slog = JoinPathFromHere("log")
		scache = JoinPathFromHere("cache")
	case "..":
		sconf = JoinPathFromHere("..", "conf")
		slog = JoinPathFromHere("..", "log")
		scache = JoinPathFromHere("..", "cache")
	default:
		sconf = filepath.Join(rootpath, "conf")
		slog = filepath.Join(rootpath, "log")
		scache = filepath.Join(rootpath, "cache")
	}
	os.MkdirAll(sconf, 0o775)
	os.MkdirAll(slog, 0o775)
	os.MkdirAll(scache, 0o775)
	return sconf, slog, scache
}
