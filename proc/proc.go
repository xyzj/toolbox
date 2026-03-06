// Package proc watch process status
package proc

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
	"github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"
	"github.com/tidwall/sjson"
	"github.com/xyzj/toolbox"
	"github.com/xyzj/toolbox/json"
	"github.com/xyzj/toolbox/logger"
	"github.com/xyzj/toolbox/loopfunc"
)

// go-echats should be v2.4.3

//go:embed echarts.min.js
var echarts []byte

type procStatus struct {
	Memrss  uint64  `json:"rss"`
	Memvms  uint64  `json:"vms"`
	IORead  uint64  `json:"ior"`
	IOWrite uint64  `json:"iow"`
	Dt      int64   `json:"dt"`
	Cpup    float32 `json:"cpu"`
	Memp    float32 `json:"mem"`
	Ofd     int32   `json:"ofd"`
	Conns   int32   `json:"conn"`
}

func (ps *procStatus) String() string {
	return fmt.Sprintf("cpu: %.2f%%; mem: %.2f%%; rss: %s", ps.Cpup, ps.Memp, toolbox.FormatFileSize(ps.Memrss))
}

func (ps *procStatus) HTML() string {
	return fmt.Sprintf("cpu: %.2f%%\nmem: %.2f%%\nrss: %s\nofd: %d", ps.Cpup, ps.Memp, toolbox.FormatFileSize(ps.Memrss), ps.Ofd)
}

func (ps *procStatus) JSON() string {
	js, err := json.MarshalToString(ps)
	if err != nil {
		return ""
	}
	return js
}

type RecordOpt struct {
	Logg        logger.Logger
	Timer       time.Duration
	DataTimeout time.Duration
	Name        string
	SaveFile    string
}
type Recorder struct {
	buffer   []*procStatus
	opt      *RecordOpt
	locker   sync.Mutex
	filename string
}

func (r *Recorder) LastHTML() string {
	if len(r.buffer) == 0 {
		return ""
	}
	return r.buffer[len(r.buffer)-1].HTML()
}

func (r *Recorder) LastJSON() string {
	if len(r.buffer) == 0 {
		return ""
	}
	return r.buffer[len(r.buffer)-1].JSON()
}

func (r *Recorder) LastString() string {
	if len(r.buffer) == 0 {
		return ""
	}
	return r.buffer[len(r.buffer)-1].String()
}

// flushToFile 将缓冲区数据追加到文件
func (r *Recorder) flushToFile() {
	f, err := os.OpenFile(r.filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	for _, v := range r.buffer {
		line, _ := json.Marshal(v)
		f.Write(append(line, '\n'))
	}
	r.buffer = r.buffer[:0]
}

// maintenance 精确清理过期数据
func (r *Recorder) maintenance() {
	r.locker.Lock()

	// 先把缓冲区剩下的写进去
	r.flushToFile()
	r.locker.Unlock()

	oldFile, err := os.Open(r.filename)
	if err != nil {
		return
	}

	tmpPath := r.filename + ".tmp"
	newFile, err := os.Create(tmpPath)
	if err != nil {
		oldFile.Close()
		return
	}
	writer := bufio.NewWriter(newFile)

	now := time.Now().Unix()
	limit := int64(r.opt.DataTimeout.Seconds())

	scanner := bufio.NewScanner(oldFile)
	var ps procStatus
	for scanner.Scan() {
		if err := json.Unmarshal(scanner.Bytes(), &ps); err == nil {
			// 精确比对：保留未过期的数据
			if now-ps.Dt <= limit {
				_, _ = writer.Write(scanner.Bytes())
				_ = writer.WriteByte('\n')
			}
		}
	}
	_ = writer.Flush()
	// 3. 核心修正：在 Rename 之前必须先关闭所有句柄
	oldFile.Close()
	newFile.Close()

	// 4. 加锁替换文件
	r.locker.Lock()
	os.Rename(tmpPath, r.filename)
	r.locker.Unlock()
}

// StartRecord 记录进程状态
func StartRecord(opt *RecordOpt) *Recorder {
	if opt == nil {
		opt = &RecordOpt{}
	}
	if opt.Logg == nil {
		opt.Logg = &logger.NilLogger{}
	}
	if opt.Timer < time.Second*5 {
		opt.Timer = time.Second * 60
	}
	if opt.DataTimeout < time.Minute {
		opt.DataTimeout = time.Minute
	}
	if opt.DataTimeout > time.Hour*24*366 {
		opt.DataTimeout = time.Hour * 24 * 366
	}
	if opt.SaveFile == "" {
		opt.SaveFile = "proc_records.process"
	}
	r := &Recorder{
		locker:   sync.Mutex{},
		opt:      opt,
		buffer:   make([]*procStatus, 0, 20),
		filename: opt.SaveFile,
	}
	go loopfunc.LoopFunc(func(params ...any) {
		var proce *process.Process
		var err error
		var memi *process.MemoryInfoStat
		var iost *process.IOCountersStat
		var connst []net.ConnectionStat
		var cp float64
		f := func() {
			if proce == nil {
				proce, err = process.NewProcess(int32(os.Getpid()))
				if err != nil {
					opt.Logg.Error("[PROC] " + err.Error())
					return
				}
			}
			cp, _ = proce.CPUPercent()
			p := &procStatus{}
			p.Cpup = float32(cp)
			p.Ofd, _ = proce.NumFDs()
			p.Memp, _ = proce.MemoryPercent()
			memi, _ = proce.MemoryInfo()
			if memi == nil {
				memi = &process.MemoryInfoStat{}
			}
			p.Memrss = memi.RSS
			p.Memvms = memi.VMS
			iost, _ = proce.IOCounters()
			if iost == nil {
				iost = &process.IOCountersStat{}
			}
			p.IORead = iost.ReadBytes
			p.IOWrite = iost.WriteBytes
			connst, _ = proce.Connections()
			p.Conns = int32(len(connst))
			p.Dt = time.Now().Unix()
			// 1. 存入缓冲区
			r.locker.Lock()
			r.buffer = append(r.buffer, p)

			// 2. 缓冲区满 10 条或达到某种条件写入文件
			if len(r.buffer) >= 20 {
				r.flushToFile()
			}
			r.locker.Unlock()
		}

		t := time.NewTimer(r.opt.Timer)
		tl := max(time.Hour, r.opt.Timer*2)
		tsave := time.NewTimer(tl)
		c := 0
		for {
			select {
			case <-t.C:
				f()
				c++
				if c%30 == 0 {
					c = 0
					opt.Logg.Info("[PROC] " + r.LastString())
				}
				t.Reset(r.opt.Timer)
			case <-tsave.C:
				go r.maintenance()
				tsave.Reset(tl)
			}
		}
	}, "proc", opt.Logg.DefaultWriter())
	return r
}

func (r *Recorder) Export() []byte {
	ps := r.allData()
	jsonData, err := json.Marshal(ps)
	if err != nil {
		return []byte{}
	}
	return jsonData
}

// allData 返回所有数据
func (r *Recorder) allData() []*procStatus {
	r.locker.Lock()
	r.flushToFile() // 确保当前内存中的数据也进入读取范围
	r.locker.Unlock()

	data := make([]*procStatus, 0)
	f, err := os.Open(r.filename)
	if err != nil {
		return data
	}
	defer f.Close()

	// 1. 先快速估算行数或获取文件大小
	// 如果文件很大，我们采用采样读取
	fi, _ := f.Stat()
	fileSize := fi.Size()

	// 假设每行 JSON 大约 200 字节，如果超过 2000 行则开启采样
	// 200 * 2000 = 400,000 字节 (~400KB)
	step := 1
	if fileSize > 400000 {
		// 粗略计算步长，目标是让返回给前端的点控制在 1000-2000 个左右
		approxRows := fileSize / 200
		step = max(int(approxRows/1500), 1)
	}

	scanner := bufio.NewScanner(f)
	count := 0
	for scanner.Scan() {
		count++
		// 采样逻辑：只处理符合步长的行
		if step > 1 && count%step != 0 {
			continue
		}

		var ps procStatus
		if err := json.Unmarshal(scanner.Bytes(), &ps); err == nil {
			data = append(data, &ps)
		}
	}

	// 依然保持你原有的排序逻辑
	sort.Slice(data, func(i, j int) bool {
		return data[i].Dt < data[j].Dt
	})
	return data
}

func (r *Recorder) BuildLines(width string, js []*procStatus) []byte {
	var nametail string
	if r.opt.Name != "" {
		nametail = " (" + r.opt.Name + ")"
	}
	l := len(js)
	x := make([]string, 0, l)
	cpu := make([]opts.LineData, 0, l)
	mem := make([]opts.LineData, 0, l)
	rss := make([]opts.LineData, 0, l)
	vms := make([]opts.LineData, 0, l)
	ofd := make([]opts.LineData, 0, l)
	ior := make([]opts.LineData, 0, l)
	iow := make([]opts.LineData, 0, l)
	con := make([]opts.LineData, 0, l)
	for _, v := range js {
		x = append(x, toolbox.Stamp2Time(v.Dt, "01-02 15:04:05"))
		cpu = append(cpu, opts.LineData{Name: fmt.Sprintf("%.2f%%", v.Cpup), Value: toolbox.FormatFloat64(float64(v.Cpup), 2), Symbol: "circle"})
		mem = append(mem, opts.LineData{Name: fmt.Sprintf("%.2f%%", v.Memp), Value: toolbox.FormatFloat64(float64(v.Memp), 2), Symbol: "circle"})
		rss = append(rss, opts.LineData{Name: toolbox.FormatFileSize(v.Memrss), Value: v.Memrss / 1024 / 1024, Symbol: "circle"})
		vms = append(vms, opts.LineData{Name: toolbox.FormatFileSize(v.Memvms), Value: v.Memvms / 1024 / 1024, Symbol: "circle"})
		ofd = append(ofd, opts.LineData{Name: fmt.Sprintf("%d", v.Ofd), Value: v.Ofd, Symbol: "circle"})
		ior = append(ior, opts.LineData{Name: toolbox.FormatFileSize(v.IORead), Value: v.IORead / 1024 / 1024, Symbol: "circle"})
		iow = append(iow, opts.LineData{Name: toolbox.FormatFileSize(v.IOWrite), Value: v.IOWrite / 1024 / 1024, Symbol: "circle"})
		con = append(con, opts.LineData{Name: fmt.Sprintf("%d", v.Conns), Value: v.Conns, Symbol: "circle"})
	}
	line1 := charts.NewLine()
	line1.SetGlobalOptions(SetupLineGOpts(&LineOpt{
		PageTitle:   "Process Records" + nametail,
		Name:        "CPU & MEM Use" + nametail,
		Total:       float32(l),
		TTFormatter: "{a0}: <b>{b0}</b><br>{a1}: <b>{b1}</b>",
		YFormatter:  "{value} %",
		Width:       width,
	})...)
	line1.YAxisList = SetupYAxisOpts(YAxisOpt{
		Name:        "cpu & mem",
		SplitNumber: 7,
		Formatter:   "{value} %",
		Align:       "left",
	})
	line1.SetXAxis(x)
	line1.SetSeriesOptions(SetupLineSOpts()...)
	line1.AddSeries("cpu", cpu)
	line1.AddSeries("mem", mem)

	lineMem := charts.NewLine()
	lineMem.SetGlobalOptions(SetupLineGOpts(&LineOpt{
		Name:        "Resident Set Size & Virtual Memory Size" + nametail,
		Total:       float32(l),
		TTFormatter: "{a0}: <b>{b0}</b><br>{a1}: <b>{b1}</b>",
		YFormatter:  "{value} MB",
		Width:       width,
	})...)
	lineMem.YAxisList = SetupYAxisOpts(YAxisOpt{
		Name:        "rss",
		SplitNumber: 7,
		SplitLine:   false,
		SplitArea:   true,
		Formatter:   "{value} MB",
		Align:       "left",
	}, YAxisOpt{
		Name:        "vms",
		SplitNumber: 7,
		SplitLine:   false,
		SplitArea:   false,
		Formatter:   "{value} MB",
		Align:       "right",
	})
	op := float32(0.3)
	lineMem.SetXAxis(x)
	lineMem.AddSeries("rss", rss, charts.WithLineChartOpts(
		opts.LineChart{
			Smooth:     opts.Bool(true),
			YAxisIndex: 0,
		}),
		charts.WithLineStyleOpts(opts.LineStyle{
			Width: 2,
		}), charts.WithAreaStyleOpts(opts.AreaStyle{
			Opacity: &op,
		}))
	lineMem.AddSeries("vms", vms, charts.WithLineChartOpts(
		opts.LineChart{
			Smooth:     opts.Bool(false),
			YAxisIndex: 1,
		}))

	lineOfd := charts.NewLine()
	lineOfd.SetGlobalOptions(SetupLineGOpts(&LineOpt{
		Name:        "Opened File Descriptors & Connections" + nametail,
		Total:       float32(l),
		TTFormatter: "{a0}: <b>{b0}</b><br>{a1}: <b>{b1}</b>",
		Width:       width,
	})...)
	lineOfd.YAxisList = SetupYAxisOpts(YAxisOpt{
		Name:        "ofd & conn",
		SplitNumber: 7,
		Formatter:   "{value}",
		Align:       "left",
	})
	op = float32(0.2)
	lineOfd.SetXAxis(x)
	lineOfd.SetSeriesOptions(SetupLineSOpts(charts.WithAreaStyleOpts(opts.AreaStyle{
		Opacity: &op,
	}))...)
	lineOfd.AddSeries("ofd", ofd)
	lineOfd.AddSeries("conn", con)

	lineIO := charts.NewLine()
	lineIO.SetGlobalOptions(SetupLineGOpts(&LineOpt{
		Name:        "IO Counts" + nametail,
		Total:       float32(l),
		TTFormatter: "{a0}: <b>{b0}</b><br>{a1}: <b>{b1}</b>",
		YFormatter:  "{value} MB",
		Width:       width,
	})...)
	lineIO.YAxisList = SetupYAxisOpts(YAxisOpt{
		Name:        "read",
		SplitNumber: 7,
		Formatter:   "{value} MB",
		Align:       "left",
	}, YAxisOpt{
		Name:        "write",
		SplitNumber: 7,
		SplitArea:   true,
		Formatter:   "{value} MB",
		Align:       "right",
	})
	op = float32(0.3)
	lineIO.SetXAxis(x)
	lineIO.SetSeriesOptions(SetupLineSOpts(charts.WithAreaStyleOpts(opts.AreaStyle{
		Opacity: &op,
	}))...)
	lineIO.AddSeries("read", ior, charts.WithLineChartOpts(
		opts.LineChart{
			Smooth:     opts.Bool(true),
			YAxisIndex: 0,
		}),
	)
	lineIO.AddSeries("write", iow, charts.WithLineChartOpts(
		opts.LineChart{
			Smooth:     opts.Bool(true),
			YAxisIndex: 1,
		}))
	a := components.NewPage()
	a.PageTitle = "Process Records" + nametail
	a.AddCharts(line1, lineOfd, lineMem, lineIO) //, line1,lineOfd, lineMem, lineIO)
	b := &bytes.Buffer{}
	a.Render(b)
	return b.Bytes()
}

func (r *Recorder) GinHandler(c *gin.Context) {
	js := r.allData()
	switch c.Request.Method {
	case "POST":
		c.Set("data", js)
		c.Set("status", 1)
		c.JSON(200, c.Keys)
	case "GET":
		x := LocalEchartsJS(r.BuildLines(c.Param("width"), js))
		os.WriteFile("a.html", x, 0o664)
		c.Writer.Write(LocalEchartsJS(r.BuildLines(c.Param("width"), js)))
	}
}

func (r *Recorder) HTTPHandler(w http.ResponseWriter, req *http.Request) {
	js := r.allData()
	switch req.Method {
	case "POST":
		s, _ := sjson.SetBytes([]byte{}, "status", 1)
		s, _ = sjson.SetBytes(s, "data", js)
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")
		w.Write(s)
	case "GET":
		values := req.URL.Query()
		w.Write(LocalEchartsJS(r.BuildLines(values.Get("width"), js)))
	}
}

type LineOpt struct {
	PageTitle   string
	Name        string
	TTFormatter string
	YFormatter  string
	Width       string
	Heigh       string
	Total       float32
	PageCount   float32
}

func SetupLineGOpts(lopt *LineOpt) []charts.GlobalOpts {
	if lopt == nil {
		lopt = &LineOpt{
			Name: "unknow",
		}
	}
	if lopt.Width == "" {
		lopt.Width = "2000px"
	}
	if lopt.Heigh == "0" {
		lopt.Heigh = "500px"
	}
	start := float32(0)
	if lopt.PageCount > 0 && lopt.Total > lopt.PageCount {
		start = (lopt.Total - lopt.PageCount) / lopt.Total * 100
	}
	return []charts.GlobalOpts{
		charts.WithDataZoomOpts(opts.DataZoom{
			Start: start,
		}),
		charts.WithAnimation(true),
		charts.WithTitleOpts(opts.Title{
			Title: lopt.Name,
			Left:  "10%",
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Trigger: "axis",
			Show:    opts.Bool(true),
			// Formatter: opts.FuncOpts(lopt.TTFormatter),
			AxisPointer: &opts.AxisPointer{
				Show: opts.Bool(true),
				Type: "line",
				Label: &opts.Label{
					Show: opts.Bool(true),
				},
			},
		}),
		charts.WithToolboxOpts(opts.Toolbox{
			Show: opts.Bool(true),
			Feature: &opts.ToolBoxFeature{
				SaveAsImage: &opts.ToolBoxFeatureSaveAsImage{
					Show: opts.Bool(true),
				},
			},
		}),
		charts.WithInitializationOpts(opts.Initialization{
			PageTitle: lopt.PageTitle,
			Theme:     types.ChartThemeRiver,
			Width:     lopt.Width,
			Height:    lopt.Heigh,
		}),
		charts.WithXAxisOpts(opts.XAxis{
			AxisLabel: &opts.AxisLabel{
				Show:         opts.Bool(true),
				Rotate:       30,
				ShowMinLabel: opts.Bool(true),
				ShowMaxLabel: opts.Bool(true),
			},
		}),
		charts.WithYAxisOpts(opts.YAxis{
			SplitLine: &opts.SplitLine{
				Show: opts.Bool(false),
			},
			SplitArea: &opts.SplitArea{
				Show: opts.Bool(true),
			},
			AxisLabel: &opts.AxisLabel{
				Show:         opts.Bool(true),
				ShowMinLabel: opts.Bool(true),
				ShowMaxLabel: opts.Bool(true),
				Formatter:    opts.FuncOpts(lopt.YFormatter),
			},
		}),
	}
}

func SetupLineSOpts(lopt ...charts.SeriesOpts) []charts.SeriesOpts {
	serOpt := []charts.SeriesOpts{
		charts.WithLineChartOpts(opts.LineChart{
			Smooth: opts.Bool(true),
		}),
		charts.WithLineStyleOpts(opts.LineStyle{
			Width: 2,
		}),
	}
	serOpt = append(serOpt, lopt...)
	return serOpt
}

func SetupBarSOpts(lopt ...charts.SeriesOpts) []charts.SeriesOpts {
	serOpt := []charts.SeriesOpts{
		charts.WithBarChartOpts(opts.BarChart{
			RoundCap: opts.Bool(true),
		}),
	}
	serOpt = append(serOpt, lopt...)
	return serOpt
}

type YAxisOpt struct {
	Name        string
	Formatter   string
	Align       string
	SplitNumber int
	SplitLine   bool
	SplitArea   bool
}

func SetupYAxisOpts(lopt ...YAxisOpt) []opts.YAxis {
	var yAxis []opts.YAxis
	for _, o := range lopt {
		yAxis = append(yAxis, opts.YAxis{
			Name:        o.Name,
			SplitNumber: o.SplitNumber,
			Show:        opts.Bool(true),
			SplitLine: &opts.SplitLine{
				Show: opts.Bool(o.SplitLine),
			},
			SplitArea: &opts.SplitArea{
				Show: opts.Bool(o.SplitArea),
			},
			AxisLabel: &opts.AxisLabel{
				Show:      opts.Bool(true),
				Interval:  "auto",
				Formatter: types.FuncStr(o.Formatter),
				Align:     o.Align,
			},
		})
	}
	return yAxis
}

func EchartsJS() []byte {
	return json.Bytes("<script type=\"text/javascript\">" + json.String(echarts) + "\n</script>")
}

func LocalEchartsJS(htmlpage []byte) []byte {
	if len(htmlpage) == 0 {
		return EchartsJS()
	}
	// return htmlpage
	s := bytes.ReplaceAll(htmlpage, json.Bytes(`<script src="https://go-echarts.github.io/go-echarts-assets/assets/echarts.min.js"></script>`), EchartsJS())
	s = bytes.ReplaceAll(s, json.Bytes(`<script src="https://go-echarts.github.io/go-echarts-assets/assets/themes/themeRiver.js"></script>`), []byte{})
	return s
}
