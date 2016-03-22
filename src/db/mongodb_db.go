package db

import (
	"gopkg.in/mgo.v2"
)

type MongoDatabase struct {
	db *mgo.Database
}

func NewMongoDatabase(db *mgo.Database) *MongoDatabase {

	if nil == db {
		return nil
	}

	return &MongoDatabase{
		db: db,
	}
}

// 数据库名字
func (this *MongoDatabase) Name() string {
	return this.db.Name
}

// 注销数据库
func (this *MongoDatabase) Logout() {
	this.db.Logout()
	this.db = nil
}

// 设置索引
func (this *MongoDatabase) EnsureIndex(data IRecord) bool {
	return EnsureIndex(this.db, data)
}

// 查找数据
func (this *MongoDatabase) Find(cond C, result interface{}) bool {
	return Find(this.db, cond, result)
}

// 查找一行数据
func (this *MongoDatabase) FindOne(result interface{}) bool {
	return FindOne(this.db, result)
}

// 查找所有符合条件的结果
func (this *MongoDatabase) FindAll(table string, key C, result interface{}) bool {
	return FindAll(this.db, table, key, result)
}

// 插入一行数据
func (this *MongoDatabase) Insert(data IRecord) bool {
	return Insert(this.db, data)
}

func (this *MongoDatabase) UpdateBy(table string, cond, updater C) bool {
	return UpdateBy(this.db, table, cond, updater)
}

// 更新一行数据
func (this *MongoDatabase) Update(data IRecord) bool {
	return Update(this.db, data)
}

// 更新某一个字段
func (this *MongoDatabase) UpdateField(data IRecord, field string) bool {
	return UpdateField(this.db, data, field)
}

// 更新多个字段
func (this *MongoDatabase) UpdateFields(data IRecord, fields []string) bool {
	return UpdateFields(this.db, data, fields)
}

// 更新多个字段集合
func (this *MongoDatabase) UpdateFieldsSet(data IRecord, fields SaveFields) bool {
	return UpdateFieldsSet(this.db, data, fields)
}

// 更新多个字段集合
func (this *MongoDatabase) UpdateFieldsSetBy(cond C, data IRecord, fields SaveFields) bool {
	return UpdateFieldsSetBy(this.db, cond, data, fields)
}

// 删除一行数据
func (this *MongoDatabase) Delete(data IRecord) bool {
	return Delete(this.db, data)
}

// 删除符合条件的数据
func (this *MongoDatabase) DeleteBy(table string, cond C) bool {
	return DeleteBy(this.db, table, cond)
}

// 删除所有符合条件的数据
func (this *MongoDatabase) DeleteAll(table string, cond C) bool {
	return DeleteAll(this.db, table, cond)
}

// 统计符合的总行数
func (this *MongoDatabase) Count(table string, cond C) int {
	return Count(this.db, table, cond)
}

// 高级查询 排序 & 过滤字段 & 限制数量
func (this *MongoDatabase) Select(cond, filter C, sort []string, offset, limit int, result interface{}) bool {
	return Select(this.db, cond, filter, sort, offset, limit, result)
}

// 执行指定命令
func (this *MongoDatabase) Run(cmd interface{}, result interface{}) bool {
	return Run(this.db, cmd, result)
}

// 聚合
func (this *MongoDatabase) Aggregate(table string, pipeline []C, result interface{}) bool {
	return Pipe(this.db, table, pipeline, result)
}

// 去重
func (this *MongoDatabase) Distinct(table string, cond C, field string, result interface{}) bool {
	return Distinct(this.db, table, cond, field, result)
}

// 清空表
func (this *MongoDatabase) Truncate(table string) bool {
	return Truncate(this.db, table)
}

// 查找 & 修改
func (this *MongoDatabase) FindAndModify(table string, cond, updater C, result IRecord, upsert ...bool) bool {
	return FindAndModify(this.db, table, cond, updater, result, upsert...)
}

// 迭代数据库
func (this *MongoDatabase) ForEach(ins IRecord, cond, filter C, sort []string, batchCount int, callback func(page, totoalPage int, results interface{})) (taskNum int, taskFun func(taskNo int)) {

	return ForEach(this.db, ins, cond, filter, sort, batchCount, callback)
}
