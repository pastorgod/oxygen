package base

import (
	"db"
)

type RedisHash struct {
	client  *db.RedisClient
	hashKey string
}

func NewRedisHash(key string) *RedisHash {
	return NewRedisHashWith(key, db.Redis)
}

func NewRedisHashWith(key string, client *db.RedisClient) *RedisHash {
	return &RedisHash{
		client:  client,
		hashKey: key,
	}
}

func (this *RedisHash) RedisKey() string {
	return this.hashKey
}

func (this *RedisHash) Exists(field string) bool {
	return this.client.HExists(this.RedisKey(), field)
}

func (this *RedisHash) Keys() []string {
	return this.client.HashKeys(this.RedisKey())
}

func (this *RedisHash) Values() []string {
	return this.client.HashVals(this.RedisKey())
}

func (this *RedisHash) Get(field string) (string, bool) {
	return this.client.HashGet(this.RedisKey(), field)
}

func (this *RedisHash) Set(field, value string) bool {
	return this.client.HashSet(this.RedisKey(), field, value)
}

func (this *RedisHash) Sets(sets map[string]string) bool {
	return this.client.HashMSet(this.RedisKey(), sets)
}

func (this *RedisHash) Remove(fields ...string) int {
	return this.client.HashDel(this.RedisKey(), fields...)
}

func (this *RedisHash) Count() int {
	return this.client.HashLen(this.RedisKey())
}

func (this *RedisHash) ForEach(handler func(field, value string)) {
	this.client.HashScan(this.RedisKey(), handler)
}

func (this *RedisHash) Clear() {
	this.client.Del(this.RedisKey())
}
