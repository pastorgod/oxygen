package record

import (
	"db"
)

type IRecordObject interface {
	db.IRecord
	Mask() uint64
	Dirty(...uint64)
	Recover(uint64)
	ClearMask()
	MergeFrom(uint64, db.IRecord)
	CopyFrom(uint64, db.IRecord)
	ToFields(uint64) []string
	Flush() bool
}

func Insert(data db.IRecord) bool {
	return db.DB.Insert(data)
}

func UpdateBy(table string, cond, updater db.C) bool {
	return db.DB.UpdateBy(table, cond, updater)
}

func Update(data db.IRecord) bool {
	return db.DB.Update(data)
}

func UpdateField(data db.IRecord, field string) bool {
	return db.DB.UpdateField(data, field)
}

func UpdateFields(data db.IRecord, fields []string) bool {
	return db.DB.UpdateFields(data, fields)
}

func UpdateFieldsSet(data db.IRecord, fields db.SaveFields) bool {
	return db.DB.UpdateFieldsSet(data, fields)
}

func UpdateFieldsSetBy(cond db.C, data db.IRecord, fields db.SaveFields) bool {
	return db.DB.UpdateFieldsSetBy(cond, data, fields)
}

func Delete(data db.IRecord) bool {
	return db.DB.Delete(data)
}

func DeleteBy(table string, cond db.C) bool {
	return db.DB.DeleteBy(table, cond)
}

func DeleteAll(table string, cond db.C) bool {
	return db.DB.DeleteAll(table, cond)
}

func Find(cond db.C, result interface{}) bool {
	return db.DB.Find(cond, result)
}

// result := &Account{ Id : "linbo" }
func FindOne(result db.IRecord) bool {
	return db.DB.FindOne(result)
}

// result := []Account{}
func FindAll(table string, cond db.C, result interface{}) bool {
	return db.DB.FindAll(table, cond, result)
}

func Count(table string, cond db.C) int {
	return db.DB.Count(table, cond)
}

// ( []string{ "-age", "name" }, 10, &[]Character )
func Select(cond, filter db.C, sort []string, offset, limit int, result interface{}) bool {
	return db.DB.Select(cond, filter, sort, offset, limit, result)
}

// 执行指定命令
func Run(cmd interface{}, result interface{}) bool {
	return db.DB.Run(cmd, result)
}

// 聚合
func Aggregate(table string, pipeline []db.C, result interface{}) bool {
	return db.DB.Aggregate(table, pipeline, result)
}

// 去重
func Distinct(table string, cond db.C, field string, result interface{}) bool {
	return db.DB.Distinct(table, cond, field, result)
}

// 清空表
func Truncate(table string) bool {
	return db.DB.Truncate(table)
}

// 查找 & 修改
// update 如果不指定set( {$set : {...}} ) 默认为替换
func FindAndModify(table string, cond, updater db.C, result db.IRecord, upsert ...bool) bool {
	return db.DB.FindAndModify(table, cond, updater, result, upsert...)
}

// 迭代数据库
func ForEach(ins db.IRecord,
	cond, filter db.C,
	sort []string,
	batchCount int,
	callback func(page, totoalPage int, results interface{})) (taskNum int, taskFun func(taskNo int)) {
	return db.DB.ForEach(ins, cond, filter, sort, batchCount, callback)
}
