package base

import (
	"db"
)

type RedisList struct {
	client  *db.RedisClient
	listKey string
	max_len int
}

func NewRedisList(key string, max_len int32) *RedisList {
	return NewRedisListWith(key, max_len, db.Redis)
}

func NewRedisListWith(key string, max_len int32, client *db.RedisClient) *RedisList {
	return &RedisList{
		client:  client,
		listKey: key,
		max_len: int(max_len),
	}
}

func (this *RedisList) RedisKey() string {
	return this.listKey
}

func (this *RedisList) Push(value string) {

	if length, _ := this.client.ListPush(this.RedisKey(), value); length > this.max_len && this.max_len > 0 {
		this.Trim(0, int32(this.max_len))
	}
}

func (this *RedisList) Pop() string {
	value, _ := this.client.ListPop(this.RedisKey())
	return value
}

func (this *RedisList) Count() int32 {
	return int32(this.client.ListLen(this.RedisKey()))
}

func (this *RedisList) IndexOf(index int32) string {
	if value, ok := this.client.ListIndex(this.RedisKey(), index); ok {
		return value
	}
	return ""
}

func (this *RedisList) Range(start, stop int32) []string {
	list, _ := this.client.ListRange(this.RedisKey(), start, stop)
	return list
}

func (this *RedisList) Remove(member string) int {
	return this.client.ListRem(this.RedisKey(), member)
}

func (this *RedisList) Trim(start, stop int32) {
	this.client.ListTrim(this.RedisKey(), start, stop)
}

func (this *RedisList) Clear() {
	this.client.Del(this.RedisKey())
}

func (this *RedisList) CountBy(member string) (num int32) {
	list := this.Range(0, -1)
	for _, name := range list {
		if name == member {
			num += 1
		}
	}
	return
}
