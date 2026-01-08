// Package pathtool some path-related methods
package pathtool

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strings"
	"time"
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

// 添加文件大小格式化函数
func FormatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// EnsureDirAndSplit 将输入路径分离为目录和文件名，并确保目录存在。
// 支持绝对路径和相对路径。如果输入以路径分隔符结尾或表示目录，则返回目录（绝对路径）和空文件名。
// 返回值：dir（绝对目录路径）, filename（文件名或空）, err（出错返回非 nil）。
func EnsureDirAndSplit(p string) (string, string, error) {
	if strings.TrimSpace(p) == "" {
		return "", "", fmt.Errorf("empty path")
	}

	// 规范化路径
	p = filepath.Clean(p)

	// 如果输入以路径分隔符结尾，视为目录
	if strings.HasSuffix(p, string(os.PathSeparator)) || p == "." || p == ".." {
		absDir, err := filepath.Abs(p)
		if err != nil {
			return "", "", err
		}
		if err := os.MkdirAll(absDir, 0o755); err != nil {
			return "", "", err
		}
		return absDir, "", nil
	}

	// 分离文件名和目录
	filename := filepath.Base(p)
	dir := filepath.Dir(p)
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", "", err
	}

	// 创建目录（如果不存在）
	if err := os.MkdirAll(absDir, 0o755); err != nil {
		return "", "", err
	}

	return absDir, filename, nil
}

// CopyFile 复制文件到指定路径，若目标目录不存在则自动创建，保持源文件权限。
func CopyFile(src, dst string) error {
	if strings.TrimSpace(src) == "" || strings.TrimSpace(dst) == "" {
		return fmt.Errorf("source or destination is empty")
	}

	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("stat source: %w", err)
	}
	if !srcInfo.Mode().IsRegular() {
		return fmt.Errorf("source is not a regular file")
	}

	in, err := os.OpenFile(src, os.O_RDONLY, 0)
	if err != nil {
		return fmt.Errorf("open source: %w", err)
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("make dest dir: %w", err)
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode().Perm())
	if err != nil {
		return fmt.Errorf("open dest: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copy data: %w", err)
	}
	return nil
}

type FileInfo struct {
	Path    string
	ModTime time.Time
}

// SearchFilesByTime 在指定目录 dir 下搜索文件名包含 name 的所有文件。
// 返回的文件列表按修改时间升序排列（旧 -> 新）。
// 仅搜索当前目录，不递归进入子目录。
func SearchFilesByTime(dir, name string) ([]FileInfo, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []FileInfo

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.Contains(entry.Name(), name) {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			files = append(files, FileInfo{
				Path:    filepath.Join(dir, entry.Name()),
				ModTime: info.ModTime(),
			})
		}
	}

	// 按修改时间排序
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime.Before(files[j].ModTime)
	})

	return files, nil
	// 提取路径
	// result := make([]string, len(files))
	// for i, f := range files {
	// 	result[i] = f.path
	// }
	// return result, nil
}
