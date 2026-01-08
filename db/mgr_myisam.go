package db

import (
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"
)

// MergeTable 进行分表操作
func (d *Conn) MergeTable(dbname, tableName string, maxSubTables, maxTableSize, maxTableRows int) error {
	if d.cfg.DriverType != DriveMySQL {
		return errors.New("this function only support mysql driver")
	}
	dbidx := 1
	for k, v := range d.dbs {
		if v.name == dbname {
			dbidx = k
			break
		}
	}
	if !d.IsReady() {
		return errors.New("sql connection is not ready")
	}
	// 查询引擎
	strsql := "select engine from information_schema.tables where table_schema=? and table_name=?"
	ans, err := d.QueryByDB(dbidx, strsql, 1, dbname, tableName)
	if err != nil {
		return err
	}
	if ans.Total == 0 {
		return errors.New("table " + dbname + "." + tableName + " not found")
	}
	engine := ans.Rows[0].Cells[0]
	if strings.ToLower(engine) != "mrg_myisam" {
		return errors.New(dbname + "." + tableName + "'s engine is not 'mrg-myisam'")
	}
	// 找到所有以指定命名开头的所有表

	// strsql = "select table_name from information_schema.tables where table_schema=? and engine='MyISAM' and table_type='BASE TABLE' and table_name like '" + tableName + "_%' order by create_time desc,table_name desc limit ?"
	strsql = "select table_name from information_schema.tables where table_schema=? and engine='MyISAM' and table_type='BASE TABLE' and table_name like '" + tableName + "_%'" // 取消数据库查询，手动排序截取
	ans, err = d.QueryByDB(dbidx, strsql, 0, dbname)
	if err != nil {
		return err
	}
	subTablelist := make([]string, 0)
	for _, row := range ans.Rows {
		subTablelist = append(subTablelist, row.Cells[0])
	}
	if len(subTablelist) == 0 {
		return errors.New("no sub tables found")
	}
	sort.Slice(subTablelist, func(i, j int) bool {
		ic, _ := strconv.Atoi(strings.TrimPrefix(subTablelist[i], tableName+"_"))
		jc, _ := strconv.Atoi(strings.TrimPrefix(subTablelist[j], tableName+"_"))
		return ic > jc
	})
	subTablelist = subTablelist[:min(maxSubTables, len(subTablelist))]
	// 检查子表大小
	strsql = `select round(sum(DATA_LENGTH/1000000)),sum(table_rows) from information_schema.tables where table_schema=? and table_name=?`
	ans, err = d.QueryByDB(dbidx, strsql, 1, dbname, subTablelist[0])
	if err != nil {
		return err
	}
	if ans.Rows[0].VCells[0].TryInt() < maxTableSize && ans.Rows[0].VCells[1].TryInt() < maxTableRows {
		return nil
	}
	// 创建新子表
	subTableLatest := tableName + "_" + time.Now().Format("200601021504")
	strsql = "create table " + subTableLatest + " like " + subTablelist[0]
	_, _, err = d.ExecByDB(dbidx, strsql)
	if err != nil {
		return errors.New("create new table " + subTableLatest + " error: " + err.Error())
	}
	newsub := make([]string, 0)
	newsub = append(newsub, subTableLatest)
	newsub = append(newsub, subTablelist...)
	// 修改总表
	strsql = "ALTER TABLE " + tableName + " ENGINE = MRG_MyISAM DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci INSERT_METHOD=FIRST UNION=(" + strings.Join(newsub, ",") + ");"
	_, _, err = d.ExecByDB(dbidx, strsql)
	if err != nil {
		return errors.New("alter table " + tableName + " error: " + err.Error())
	}
	return nil
}

// AlterMergeTable 用于修改mrg-myasim表
// tableName为主表名称
// tableName和alter语句中的table name必须一致
func (d *Conn) AlterMergeTable(dbname, tableName, alterStr string, maxSubTables int) error {
	if d.cfg.DriverType != DriveMySQL {
		return errors.New("this function only support mysql driver")
	}
	if maxSubTables < 1 {
		maxSubTables = 1
	}
	dbidx := 1
	for k, v := range d.dbs {
		if v.name == dbname {
			dbidx = k
			break
		}
	}
	if !d.IsReady() {
		return errors.New("sql connection is not ready")
	}
	var engine string
	// 查询引擎
	strsql := "select engine from information_schema.tables where table_schema=? and table_name=?"
	ans, err := d.QueryByDB(dbidx, strsql, 1, dbname, tableName)
	if err != nil {
		return err
	}
	if ans.Total == 0 { // 表不在，可能之前操作失败，这里默认不做引擎判断，尝试从子表恢复
		goto TODO
	}
	engine = ans.Rows[0].Cells[0] // 表存在的情况下，检查引擎是否符合要求
	if strings.ToLower(engine) != "mrg_myisam" {
		return errors.New("engine " + engine + " is not support")
	}
TODO:
	// 找到所有以指定命名开头的所有表
	strsql = "select table_name from information_schema.tables where table_schema=? and engine='MyISAM' and table_type='BASE TABLE' and table_name like '" + tableName + "_%' order by create_time desc"
	ans, err = d.QueryByDB(dbidx, strsql, 0, dbname)
	if err != nil {
		return err
	}
	subTablelist := make([]string, 0)
	for _, row := range ans.Rows {
		subTablelist = append(subTablelist, row.Cells[0])
	}
	if len(subTablelist) == 0 {
		return errors.New("no sub tables found")
	}
	if maxSubTables > len(subTablelist) {
		maxSubTables = len(subTablelist)
	}
	// 获取主表语句，用于结尾恢复
	strMrgTail := ""
	idx := 0
	ans, err = d.QueryByDB(dbidx, "show create table "+tableName+";", 1)
	if err == nil && ans.Total > 0 {
		strMrgTail = ans.Rows[0].Cells[1]
		idx = strings.LastIndex(strMrgTail, "ENGINE=")                                              // 找到最后一个括号
		strMrgTail = " " + strMrgTail[idx:]                                                         // 去掉头部
		strMrgTail = strings.Replace(strMrgTail, "CHARSET=utf8 ", "CHARSET=utf8mb4_general_ci ", 1) // 修改默认字符编码
	} else {
		// 读取总表错误，可能之前修改过程中断，或被手动修改，尝试重建尾部
		strMrgTail = " ENGINE=MRG_MyISAM DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci INSERT_METHOD=FIRST UNION=(" + strings.Join(subTablelist[:maxSubTables], ",") + ");"
	}
	// 先删除主表
	_, _, err = d.ExecByDB(dbidx, "DROP TABLE IF EXISTS "+tableName+";")
	if err != nil {
		return err
	}
	// 逐一修改所有子表
	for _, sub := range subTablelist {
		_, _, err = d.ExecByDB(dbidx, strings.Replace(alterStr, tableName, sub, 1))
		if err != nil {
			continue
		}
	}
	// 获取最新子表的建表语句
	ans, err = d.QueryByDB(dbidx, "show create table "+subTablelist[0]+";", 1)
	if err != nil {
		return err
	}
	if ans.Total == 0 {
		return errors.New("show create table " + subTablelist[0] + " failed")
	}
	// 进行语句替换
	strsql = ans.Rows[0].Cells[1]
	idx = strings.LastIndex(strsql, "ENGINE=")                      // 找到尾巴
	strsql = strsql[:idx]                                           // 去掉尾巴
	strsql = strings.Replace(strsql, subTablelist[0], tableName, 1) // 替换名字
	strsql += strMrgTail                                            // 加回尾巴
	// gogogo
	// 重新创建
	_, _, err = d.ExecByDB(dbidx, strsql)
	if err != nil {
		return err
	}
	return nil
}
