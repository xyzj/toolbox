package gocmd

import (
	"flag"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/xyzj/toolbox/pathtool"
)

// var pidfile = flag.String("pid-file", "", "set the pid file path (from gocmd)")

// DefaultProgram Create a default console program, this program contains commands: `start`, `stop`, `restart`, `run`, `status`, `version`, `help`
func DefaultProgram(info *Info) *Program {
	return NewProgram(info).AddCommand(CmdRun).AddCommand(CmdStart).AddCommand(CmdStop).AddCommand(CmdRestart).AddCommand(CmdStatus)
}

// NewProgram Create a empty console program, this program contains commands: `version`, `help`
func NewProgram(info *Info) *Program {
	if info == nil {
		info = &Info{}
	}
	if info.Ver == "" {
		info.Ver = "0.1.19"
	}
	params := os.Args
	pinfo := &ProcInfo{
		params:       make([]string, 0),
		Args:         make([]string, 0),
		sigc:         NewSignalQuit(),
		beforeStart:  func() {},
		onSignalQuit: func() {},
	}
	// 获取程序信息
	pinfo.params = params[1:]
	if exec, err := filepath.Abs(params[0]); err == nil {
		pinfo.exec = exec
	} else {
		pinfo.exec = params[0]
	}
	pinfo.name = pathtool.GetExecName()
	pinfo.dir = pathtool.GetExecDir()
	if info.Title == "" {
		info.Title = pinfo.name
	}
	notparseflag, _ := strconv.ParseBool(os.Getenv(strings.ToUpper(pathtool.GetExecNameWithoutExt()) + "_NOT_PARSE_FLAG"))
	// 处理参数
	idx := 0
	for k, v := range params {
		if !strings.HasPrefix(v, "-") || v == "--help" {
			continue
		}
		idx = k
		break
	}
	if idx == 0 {
		idx = len(params)
	}
	pinfo.Args = params[idx:]
	if !notparseflag {
		if !flag.CommandLine.Parsed() {
			flag.CommandLine.Parse(params[idx:])
		}
	}
	// 设置pid文件
	pinfo.Pfile = os.Getenv("pid_file") // os.Getenv(fmt.Sprintf("%s_PID_FILE", strings.ToUpper(pathtool.GetExecNameWithoutExt())))
	if pinfo.Pfile == "" {
		pinfo.Pfile = pathtool.JoinPathFromHere(pinfo.name + ".pid")
	}
	return &Program{
		info:  info,
		cmds:  &commandList{data: make(map[string]*Command)},
		pinfo: pinfo,
	}
}
