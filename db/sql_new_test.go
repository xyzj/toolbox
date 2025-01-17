package db

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/xyzj/toolbox/cache"
	"github.com/xyzj/toolbox/coord"
	"github.com/xyzj/toolbox/logger"
)

var cn *Conn

// func TestInit(t *testing.T) {
// 	b, _ := os.ReadFile("dbinit.sql")
// 	opt := &Opt{
// 		Server:      "192.168.50.83:13306",
// 		User:        "root",
// 		Passwd:      "lp1234xy",
// 		DBNames:     []string{"v5db_test_mrg", "dba2"},
// 		InitScripts: []string{string(b)},
// 		DriverType:  DriveMySQL,
// 		Logger:      logger.NewConsoleLogger(),
// 	}
// 	a, err := New(opt)
// 	if err != nil {
// 		t.Fatal(err)
// 		return
// 	}
// 	println(a.IsReady())
// }

// func TestUpg(t *testing.T) {
// 	b, _ := os.ReadFile("dbupg.sql")
// 	opt := &Opt{
// 		Server:     "192.168.50.83:13306",
// 		User:       "root",
// 		Passwd:     "lp1234xy",
// 		DBNames:    []string{"v5db_test_mrg", "dba2"},
// 		DriverType: DriveMySQL,
// 		Logger:     logger.NewConsoleLogger(),
// 	}
// 	a, err := New(opt)
// 	if err != nil {
// 		t.Fatal(err)
// 		return
// 	}
// 	for _, s := range strings.Split(string(b), ";|") {
// 		if strings.TrimSpace(s) == "" {
// 			continue
// 		}
// 		a.Exec(s)
// 	}
// }

// func TestMrg(t *testing.T) {
// 	opt := &Opt{
// 		Server:     "192.168.50.83:13306",
// 		User:       "root",
// 		Passwd:     "lp1234xy",
// 		DBNames:    []string{"v5db_test_mrg", "dba2"},
// 		DriverType: DriveMySQL,
// 		Logger:     logger.NewConsoleLogger(),
// 	}
// 	a, err := New(opt)
// 	if err != nil {
// 		t.Fatal(err)
// 		return
// 	}
// 	err = a.MergeTable("v5db_test_mrg", "event_record", 10, 10, 10)
// 	if err != nil {
// 		t.Fatal(err)
// 		return
// 	}
// }

func TestDup(t *testing.T) {
	opt := &Opt{
		Server:     "192.168.50.83:13306",
		User:       "root",
		Passwd:     "lp1234xy",
		DBNames:    []string{"v5db_facmanager_test"},
		DriverType: DriveMySQL,
		Logger:     logger.NewConsoleLogger(),
		QueryCache: cache.NewAnyCache[*QueryData](time.Minute * 30),
	}
	a, err := New(opt)
	if err != nil {
		t.Fatal(err)
		return
	}
	s := `insert into asset_info (aid,sys,name,gid,pid,phyid,imei,sim,loc,pole_code,region,road,geo,st,dev_attr,dev_type,lc,imgs,gids,dt_create,barcode,iccid,dev_id,grid,line_id,region_id,road_id,grid_id,sout,dt_setup,imsi,contractor,contractor_id,ip,dt_update) values
	 (?,?,?,?,?,?,?,?,?,?,?,?,st_geomfromtext(?),?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
	  ON DUPLICATE KEY UPDATE
			sys=values(sys),
			name=values(name),
			gid=values(gid),
			pid=values(pid),
			phyid=values(phyid),
			imei=values(imei),
			sim=values(sim),
			loc=values(loc),
			pole_code=values(pole_code),
			region=values(region),
			road=values(road),
			geo=values(geo),
			st=values(st),
			dev_attr=values(dev_attr),
			dev_type=values(dev_type),
			lc=values(lc),
			imgs=values(imgs),
			gids=values(gids),
			dt_create=values(dt_create),
			barcode=values(barcode),
			iccid=values(iccid),
			dev_id=values(dev_id),
			grid=values(grid),
			line_id=values(line_id),
			region_id=values(region_id),
			road_id=values(road_id),
			grid_id=values(grid_id),
			sout=values(sout),
			dt_setup=values(dt_setup),
			imsi=values(imsi),
			contractor=values(contractor),
			contractor_id=values(contractor_id),
			ip=values(ip),
			dt_update=values(dt_update);`
	println(a.MariadbDuplicate2Mysql(s))
}

type aaa struct {
	A64 float32
}

func TestQueryA(t *testing.T) {
	opt := &Opt{
		Server:     "192.168.50.83:13306",
		User:       "root",
		Passwd:     "lp1234xy",
		DBNames:    []string{"v5db_datarecord"},
		DriverType: DriveMySQL,
		Logger:     logger.NewConsoleLogger(),
		QueryCache: cache.NewAnyCache[*QueryData](time.Minute * 30),
	}
	a, err := New(opt)
	if err != nil {
		t.Fatal(err)
		return
	}
	ans, err := a.Query("select asset_id,voltage_a from rtu_record_all where voltage_a>0 limit 10", 0)
	if err != nil {
		t.Fatal(err)
		return
	}
	for _, row := range ans.Rows {
		println(fmt.Sprintf("%+v", &aaa{A64: row.VCells[1].TryFloat32(2)}))
	}
	// am := make(map[int]*assetManager)
	// for _, row := range ans.Rows {
	// 	for j, cell := range row.VCells {
	// 		println(fmt.Sprintf("%d %+v", j, cell.String()))
	// 	}
	// 	// am[k] = &assetManager{
	// 	// 	Aid:      row.Cells[0],
	// 	// 	Name:     row.Cells[2],
	// 	// 	Gid:      row.VCells[3].TryInt32(),
	// 	// 	Pid:      row.Cells[4],
	// 	// 	Phyid:    row.VCells[5].TryInt64(),
	// 	// 	Imei:     row.Cells[6],
	// 	// 	Sim:      row.Cells[7],
	// 	// 	Loc:      row.Cells[8],
	// 	// 	PoleCode: row.Cells[9],
	// 	// 	Region:   row.Cells[10],
	// 	// 	Road:     row.Cells[11],
	// 	// 	// Geo:       text2Geo(row.Cells[12]), // make([]*coord.Point, 0), // row.Cells[12]
	// 	// 	St:        row.VCells[13].TryInt32(),
	// 	// 	DevAttr:   row.VCells[14].TryInt32(),
	// 	// 	DevType:   row.Cells[15],
	// 	// 	Sys:       row.Cells[16],
	// 	// 	LampCount: row.VCells[17].TryInt32(),
	// 	// 	Ord:       fmt.Sprintf("%03s%s", row.Cells[18], row.Cells[0][:6]),
	// 	// 	Imgs:      row.Cells[19],
	// 	// 	Hash:      row.Cells[1],
	// 	// 	// Gids:      string2IntSlice(row.Cells[20]),
	// 	// 	DtCreate: row.VCells[21].TryInt64(),
	// 	// 	DtC: func(dtsetup string) string {
	// 	// 		if dtsetup == "" {
	// 	// 			return toolbox.Stamp2Time(row.VCells[21].TryInt64())[:10]
	// 	// 		}
	// 	// 		return dtsetup
	// 	// 	}(row.Cells[34]),
	// 	// 	BarCode:    row.Cells[22],
	// 	// 	MeshID:     row.Cells[23],
	// 	// 	ICCID:      row.Cells[24],
	// 	// 	Contractor: row.Cells[25],
	// 	// 	DevID:      row.Cells[26],
	// 	// 	MeshData:   row.Cells[27],
	// 	// 	LineID:     row.Cells[28],
	// 	// 	Grid:       row.Cells[29],
	// 	// 	RegionID:   row.VCells[30].TryInt32(),
	// 	// 	RoadID:     row.VCells[31].TryInt32(),
	// 	// 	GridID:     row.VCells[32].TryInt32(),
	// 	// 	Sout:       row.Cells[33],
	// 	// 	IconType: func(atype string) int32 {
	// 	// 		switch atype {
	// 	// 		case "010101", "030114": // 灯杆，控制柜用attr确定图标
	// 	// 			if a := row.VCells[14].TryInt32(); a > 0 {
	// 	// 				return a
	// 	// 			}
	// 	// 			return 1
	// 	// 		case "030111", "010109", "010111": // 终端/控制器用attr确定图标
	// 	// 			return row.VCells[14].TryInt32() + 1
	// 	// 		case "030117": // 上海路灯用厂家
	// 	// 			return row.VCells[35].TryInt32()
	// 	// 		default:
	// 	// 			return 0
	// 	// 		}
	// 	// 	}(row.Cells[0][:6]),
	// 	// 	ContractorID: row.VCells[35].TryInt32(),
	// 	// 	Imsi:         row.Cells[36],
	// 	// 	IP:           row.Cells[37],
	// 	// 	MaxAge:       row.VCells[38].TryInt32(),
	// 	// 	IsFac:        row.VCells[39].TryInt32(),
	// 	// 	Property:     row.Cells[40],
	// 	// }
	// }
	// println("done", len(am), ans.CacheTag)
	// time.Sleep(time.Second)
	// zz := a.QueryCache(ans.CacheTag, 3, 20)
	// println(fmt.Sprintf("%+v", zz))
}

func BenchmarkQueryB(t *testing.B) {
	t.StopTimer()
	z := make([]*aaa, 0)
	var err error
	if cn == nil {
		opt := &Opt{
			Server:     "192.168.50.83:13306",
			User:       "root",
			Passwd:     "lp1234xy",
			DBNames:    []string{"v5db_datarecord"},
			DriverType: DriveMySQL,
			Logger:     logger.NewConsoleLogger(),
			QueryCache: cache.NewAnyCache[*QueryData](time.Minute * 30),
		}
		cn, err = New(opt)
		if err != nil {
			t.Fatal(err)
			return
		}
	}
	t.StartTimer()

	ans, err := cn.Query("select asset_id,voltage_a from rtu_record_all where voltage_a>0", 0)
	if err != nil {
		t.Fatal(err)
		return
	}
	for _, row := range ans.Rows {
		z = append(z, &aaa{A64: row.VCells[1].TryFloat32(2)})
	}
	// println("done", len(am))
}

type assetManager struct {
	Geo          []*coord.Point `json:"geo,omitempty"`           // 坐标数据
	upgFields    []string       `json:"-"`                       // 需要更新的字段
	SwitchOutSt  []int          `json:"sout_st,omitempty"`       // 开关量输出状态，0-关，1-开
	Gids         []int32        `json:"gids,omitempty"`          // 资产属于的分组列表
	hashBuff     *bytes.Buffer  `json:"-"`                       // hash缓存
	DtCreate     int64          `json:"dt_create,omitempty"`     // 安装时间，unix时间戳
	StUptime     int64          `json:"st_uptime,omitempty"`     // 状态数据最后更新时间
	Phyid        int64          `json:"phyid,omitempty"`         // 资产物理地址
	Property     string         `json:"property,omitempty"`      // 产权，宁波特有
	Aid          string         `json:"aid,omitempty"`           // 资产id
	Name         string         `json:"name,omitempty"`          // 资产名称
	Pid          string         `json:"pid,omitempty"`           // 上级资产id
	LineID       string         `json:"line_id,omitempty"`       // 资产对应的线缆id
	Imei         string         `json:"imei,omitempty"`          // imei编号
	Imsi         string         `json:"imsi,omitempty"`          // imsi编号
	Sim          string         `json:"sim,omitempty"`           // sim卡号
	DevType      string         `json:"dev_type,omitempty"`      // 8位资产型号
	DevTName     string         `json:"dev_tname,omitempty"`     // 资产型号名称
	Loc          string         `json:"loc,omitempty"`           // 资产安装位置
	PoleCode     string         `json:"pc,omitempty"`            // 灯杆编码
	Region       string         `json:"region,omitempty"`        // 资产所属区域
	Road         string         `json:"road,omitempty"`          // 资产所属道路
	Grid         string         `json:"grid,omitempty"`          // 网格名称
	Sys          string         `json:"sys,omitempty"`           // 资产所属子系统
	Auth         string         `json:"auth,omitempty"`          // 用户具备的权限
	GName        string         `json:"gname,omitempty"`         // 分组名称
	TName        string         `json:"tname,omitempty"`         // 分类名称
	DtC          string         `json:"dtc,omitempty"`           // 安装时间，仅年月日
	BarCode      string         `json:"bc,omitempty"`            // 控制器条码
	Imgs         string         `json:"img,omitempty"`           // 照片路径
	MeshID       string         `json:"mesh_id,omitempty"`       // 资产对应的模型id
	MeshData     string         `json:"mesh_data,omitempty"`     // 模型信息
	ICCID        string         `json:"iccid,omitempty"`         // 资产使用sim卡的iccid
	Contractor   string         `json:"contractor,omitempty"`    // 厂家名称
	DevID        string         `json:"dev_id,omitempty"`        // 设备编号
	IP           string         `json:"ip,omitempty"`            // 设备ip地址
	Pinyin       string         `json:"-"`                       // 拼音
	PinyinFL     string         `json:"-"`                       // 拼音首字母
	Hash         string         `json:"-"`                       // 数据hash，用于比对数据是否一致
	Ord          string         `json:"-"`                       // 排序
	Geostr       string         `json:"-"`                       // mysql geo string
	IconPath     string         `json:"icont,omitempty"`         // 最终图标状态
	sord         string         `json:"-"`                       // 总体排序字段
	RoadPath     string         `json:"road_path,omitempty"`     // 4.13用匹配id
	Sout         string         `json:"sout,omitempty"`          // 对应的接触器序号
	Signal       int32          `json:"signal,omitempty"`        // 信号强度，终端，集中器，cat1有
	Snr          int32          `json:"snr,omitempty"`           // 信噪比，cat1有
	NetType      int32          `json:"net_type,omitempty"`      // 模块类型，4-4g,其他2g
	IsFac        int32          `json:"-"`                       // 是否设施类型
	Gid          int32          `json:"gid,omitempty"`           // 分组id
	St           int32          `json:"st,omitempty"`            // 资产状态
	DevAttr      int32          `json:"dev_attr,omitempty"`      // 资产附加属性
	LampCount    int32          `json:"lc,omitempty"`            // 灯头数量
	Icon         int32          `json:"icon,omitempty"`          // 图标状态
	IconType     int32          `json:"-"`                       // 图标类型，0-默认图标，1...依据定义的图标，前端找不到对应资源时使用默认图标
	Online       int32          `json:"online,omitempty"`        // 是否在线，0-离线，1-在线
	ErrC         int32          `json:"err,omitempty"`           // 故障数量
	Useage       int32          `json:"useage,omitempty"`        // 使用时长
	Pgid         int32          `json:"pgid,omitempty"`          // 载波控制器所属集中器的分组id
	ContractorID int32          `json:"contractor_id,omitempty"` // 厂家id
	MaxAge       int32          `json:"max_age,omitempty"`       // 资产类型的建议年限
	RemainingAge int32          `json:"rem_age,omitempty"`       // 剩余时间
	RegionID     int32          `json:"region_id,omitempty"`     // 区域id
	RoadID       int32          `json:"road_id,omitempty"`       // 道路id
	GridID       int32          `json:"grid_id,omitempty"`       // 网格id
}
