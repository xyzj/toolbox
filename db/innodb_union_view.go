package db

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrNoNeedReUnion = errors.New("no need to re-union view")
)

func (d *Conn) AlterUnionView(dbname, tableName, alterStr string, maxSubTables int) error {
	// 判断是否符合分表要求
	if d.cfg.DriverType != DriveMySQL {
		return errors.New("this function only support mysql driver")
	}
	dbidx := 1
	for k, v := range d.dbs {
		if v.name == dbname {
			dbidx = k
		}
	}
	if !d.IsReady() {
		return errors.New("sql connection is not ready")
	}
	var strsql string
	var err error
	var ans *QueryData
	subTablelist := make([]string, 0)
	subTablelist = append(subTablelist, tableName)
	// 找到所有以指定命名开头的所有表
	strsql = "select table_name from information_schema.tables where table_schema=? and table_type='BASE TABLE' and table_name like '" + strings.TrimSuffix(tableName, "_all") + "%' order by create_time desc,table_name desc"
	ans, err = d.QueryByDB(dbidx, strsql, 0, dbname)
	if err != nil {
		return err
	}
	for _, row := range ans.Rows {
		subTablelist = append(subTablelist, row.Cells[0])
	}
	if len(subTablelist) == 0 {
		return errors.New("no sub tables found")
	}
	for _, sub := range subTablelist {
		_, _, err = d.ExecByDB(dbidx, strings.Replace(alterStr, tableName, sub, 1))
		if err != nil {
			d.cfg.Logger.Error("[unview] alter table " + sub + " error:" + err.Error())
			continue
		}
	}
	return nil
}
func (d *Conn) UnionView(dbname, tableName string, maxSubTables, maxTableSize, maxTableRows int, forceReUnion bool) error {
	// 判断是否符合分表要求
	if d.cfg.DriverType != DriveMySQL {
		return errors.New("this function only support mysql driver")
	}
	dbidx := 1
	for k, v := range d.dbs {
		if v.name == dbname {
			dbidx = k
		}
	}
	if !d.IsReady() {
		return errors.New("sql connection is not ready")
	}
	var strsql string
	var err error
	var ans *QueryData
	if !forceReUnion {
		// 检查子表大小
		strsql = `select round(sum(DATA_LENGTH/1000000)),sum(table_rows) from information_schema.tables where table_schema=? and table_name=?`
		ans, err = d.QueryByDB(dbidx, strsql, 1, dbname, tableName)
		if err != nil {
			return err
		}
		if ans.Rows[0].VCells[0].TryInt() < maxTableSize && ans.Rows[0].VCells[1].TryInt() < maxTableRows {
			return ErrNoNeedReUnion
		}
	}
	var tmpTableName = strings.TrimSuffix(tableName, "_all")
	// 将主表重命名为日期后缀子表
	newTableName := tmpTableName + "_" + time.Now().Format("200601021504")
	strsql = fmt.Sprintf("rename table %s to %s", tableName, newTableName)
	_, _, err = d.ExecByDB(dbidx, strsql)
	if err != nil {
		return err
	}
	// 找到所有以指定命名开头的所有表
	strsql = "select table_name from information_schema.tables where table_schema=? and table_type='BASE TABLE' and table_name like '" + tmpTableName + "_%' order by table_name desc limit ?"
	ans, err = d.QueryByDB(dbidx, strsql, 0, dbname, maxSubTables)
	if err != nil {
		return err
	}
	subTablelist := make([]string, 0, maxSubTables)
	i := 0
	for _, row := range ans.Rows {
		subTablelist = append(subTablelist, row.Cells[0])
		i++
		if i >= maxSubTables {
			break
		}
	}
	// 创建新的空主表
	strsql = fmt.Sprintf("create table %s like %s", tableName, newTableName)
	_, _, err = d.ExecByDB(dbidx, strsql)
	if err != nil {
		return err
	}
	// 修改视图，加入新子表或以及判断删除旧子表
	strsql = fmt.Sprintf(`CREATE OR REPLACE ALGORITHM = MERGE SQL SECURITY INVOKER VIEW %s_view AS
	select * from %s `, tableName, tableName)
	for _, s := range subTablelist {
		strsql += " union select * from " + s
	}
	_, _, err = d.ExecByDB(dbidx, strsql)
	if err != nil {
		return err
	}
	return nil
}
