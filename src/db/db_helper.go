package db

import (
	. "logger"
)

var DB *MongoDatabase

func InitializeMongodb(url string, tables ...IRecord) bool {

	session := NewMongoSession(url, func(db *MongoDatabase) {
		DB = db
		LOG_DEBUG("mongodb selected: %s", db.Name())
	})

	if nil == session {
		LOG_FATAL("连接数据库失败: %s", url)
		return false
	}

	if DB != nil {

		for _, table := range tables {
			if !DB.EnsureIndex(table) {
				ERROR("table: %s, EnsureIndex: %v", table.Table(), table.IndexKey())
				return false
			}
		}
	}

	INFO("connect mongodb server: %s ok", url)
	return true
}

var Redis *RedisClient

func InitializeRedis(url string) bool {

	if Redis = NewRedisClient(url, 1024); Redis != nil {
		INFO("connect redis server: %s ok", url)
		return true
	}

	LOG_FATAL("初始化连接Redis失败: %s", url)
	return false
}
