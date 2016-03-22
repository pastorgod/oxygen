package db

import (
	"github.com/garyburd/redigo/redis"
	. "logger"
	"strconv"
)

func newRedisConn(addr string) (redis.Conn, error) {

	conn, err := redis.Dial("tcp", addr)

	if err != nil {
		LOG_ERROR("fail to dial redis: %s %s", addr, err.Error())
		return nil, err
	}

	return conn, err
}

//===============================================================================================

type RedisClient struct {
	pool *redis.Pool
	addr string
}

func NewRedisClient(addr string, max_idle int) *RedisClient {

	client := &RedisClient{
		pool: redis.NewPool(func() (redis.Conn, error) { return newRedisConn(addr) }, max_idle),
		addr: addr,
	}

	// 检查一下这个地址能不能连得上
	if err := client.Ping(); err != nil {
		client.Close()
		return nil
	}

	// 程序退出时关闭这个连接池
	onDestroy(client)

	return client
}

// ping - pong
func (this *RedisClient) Ping() error {
	conn := this.pool.Get()
	defer conn.Close()

	if _, err := conn.Do("PING"); err != nil {
		LOG_ERROR("redis ping fail. %s", err.Error())
		return err
	}

	return nil
}

func (this *RedisClient) RunCommand(handler func(conn redis.Conn)) {

	// pop redis conn from pool.
	conn := this.pool.Get()

	// push redis conn to pool.
	defer func() {
		if err := conn.Close(); err != nil {
			LOG_ERROR("redis close error: %s", err.Error())
		}
	}()

	// run redis command in handler.
	handler(conn)
}

func (this *RedisClient) Do(cmd string, args ...interface{}) (reply interface{}, err error) {

	this.RunCommand(func(conn redis.Conn) {
		if reply, err = conn.Do(cmd, args...); err != nil {
			LOG_ERROR("redis cmd: %s, args: %v, error: %s", cmd, args, err.Error())
		}
	})

	return
}

func (this *RedisClient) Close() {
	this.pool.Close()
	LOG_WARN("redis client close. %s", this.addr)
}

//查找所有符合给定模式 pattern 的 key 。
//
//KEYS * 匹配数据库中所有 key 。
//KEYS h?llo 匹配 hello ， hallo 和 hxllo 等。
//KEYS h*llo 匹配 hllo 和 heeeeello 等。
//KEYS h[ae]llo 匹配 hello 和 hallo ，但不匹配 hillo 。
//特殊符号用 \ 隔开
func (this *RedisClient) Keys(pattern string) ([]string, bool) {
	ret, err := redis.Strings(this.Do("KEYS", pattern))
	return ret, nil == err
}

// 设置一个键值对
func (this *RedisClient) Set(key, value string) bool {
	_, err := this.Do("SET", key, value)
	return nil == err
}

// 设置一个带有效期的键值对
func (this *RedisClient) SetEx(key, value string, sec int) bool {
	_, err := this.Do("SET", key, value, "EX", sec)
	return nil == err
}

// 修改一个已有的键值对的生存时间
func (this *RedisClient) SetExpire(key string, sec int) bool {
	ret, err := redis.Int(this.Do("EXPIRE", key, sec))
	return nil == err && 1 == ret
}

// 移除一个给定的key的生存时间
func (this *RedisClient) CancelExpire(key string) bool {
	ret, err := redis.Int(this.Do("PERSIST", key))
	return nil == err && 1 == ret
}

// 修改一个已经存在的key名
func (this *RedisClient) Rename(key, newKey string) bool {
	_, err := this.Do("RENAME", key, newKey)
	return nil == err
}

// 获得一个键的值
func (this *RedisClient) Get(key string) (string, bool) {
	ret, err := redis.String(this.Do("GET", key))
	return ret, nil == err
}

// 删除所有给的key的数据
func (this *RedisClient) Del(keys ...interface{}) int {
	ret, _ := redis.Int(this.Do("DEL", keys...))
	return ret
}

// 返回所有给定key的值
func (this *RedisClient) MGet(keys ...interface{}) ([]string, bool) {
	ret, err := redis.Strings(this.Do("MGET", keys...))
	return ret, nil == err
}

// 判断指定的key是否存在
func (this *RedisClient) Exists(key string) bool {
	exists, err := redis.Bool(this.Do("EXISTS", key))
	return exists && nil == err
}

// 添加一个有续集(主要用于做排行榜)
// 如果某个 member 已经是有序集的成员，那么更新这个 member 的 score 值，并通过重新插入这个 member 元素，来保证该 member 在正确的位置上
func (this *RedisClient) AddSortedSet(key string, sets map[string]uint64) bool {

	if 0 == len(sets) {
		return true
	}

	// ZADD key score member [[score member] [score member] ...]
	key_pair := make([]interface{}, 0, len(sets)*2+1)
	key_pair = append(key_pair, key)

	for k, score := range sets {
		key_pair = append(key_pair, score, k)
	}

	// 被成功添加的新成员的数量，不包括那些被更新的、已经存在的成员
	_, err := this.Do("ZADD", key_pair...)
	return nil == err
}

// 更新一个指定的score值
// 不存在则插入
func (this *RedisClient) UpdateSortedSet(key, member string, score uint64) bool {
	_, err := this.Do("ZADD", key, score, member)
	return nil == err
}

// 为指定key添加score值
func (this *RedisClient) IncSortedSet(key, member string, score uint64) (uint64, bool) {
	value, err := redis.Uint64(this.Do("ZINCRBY", key, score, member))
	return value, nil == err
}

// 移除一个或者多个有续集节点
func (this *RedisClient) RemoveSortedSet(key string, members ...string) int {

	args := make([]interface{}, 0, len(members)+1)
	args = append(args, key)

	for _, mem := range members {
		args = append(args, mem)
	}

	ret, _ := redis.Int(this.Do("ZREM", args...))
	return ret
}

// 统计一下这个有续集数量
func (this *RedisClient) CountSortedSet(key string) int32 {
	ret, _ := redis.Int(this.Do("ZCOUNT", key, "-inf", "+inf"))
	return int32(ret)
}

// 移除指定排名区间的节点 0 代表第一个元素 -1 代表最后一个元素
func (this *RedisClient) RemoveRangeSortedSet(key string, min_rank, max_rank int32) bool {
	_, err := this.Do("ZREMRANGEBYRANK", key, min_rank, max_rank)
	return nil == err
}

// 获取指定排名区间里面的key数据(升序)
func (this *RedisClient) RanksByAsc(key string, start, stop int32) ([]string, bool) {
	ret, err := redis.Strings(this.Do("ZRANGE", key, start, stop))
	return ret, nil == err
}

// 获取指定排名区间里面的key数据(降序)
func (this *RedisClient) RanksByDesc(key string, start, stop int32) ([]string, bool) {
	ret, err := redis.Strings(this.Do("ZREVRANGE", key, start, stop))
	return ret, nil == err
}

// 返回指定排名区间里面的key数据和score数据(升序)
func (this *RedisClient) RanksByAscWithScores(key string, start, stop int32) ([]string, bool) {
	ret, err := redis.Strings(this.Do("ZRANGE", key, start, stop, "WITHSCORES"))
	return ret, nil == err
}

// 返回指定排名区间里面的key数据和score数据(降序)
func (this *RedisClient) RanksByDescWithScores(key string, start, stop int32) ([]string, bool) {
	ret, err := redis.Strings(this.Do("ZREVRANGE", key, start, stop, "WITHSCORES"))
	return ret, nil == err
}

// 获取某个指定key的排名(升序排名)
func (this *RedisClient) RankByAsc(key, memeber string) int32 {
	ret, err := redis.Int(this.Do("ZRANK", key, memeber))

	if nil != err {
		return 0
	}

	// 下标转换为排名
	return int32(ret + 1)
}

// 获取某个指定key的排名(升序排名)及分数
func (this *RedisClient) GetScore(key, member string) (uint64, bool) {
	ret, err := redis.Uint64(this.Do("ZSCORE", key, member))
	return uint64(ret), nil == err
}

// 返回有序集 key 中，所有 score 值介于 min 和 max 之间(升序排名)
func (this *RedisClient) MembersByAsc(key string, min, max int32) ([]string, bool) {
	ret, err := redis.Strings(this.Do("ZRANGEBYSCORE", key, min, max))
	return ret, nil == err
}

// 返回有序集 key 中，所有 score 值介于 min 和 max 之间(降序排名)
func (this *RedisClient) MembersByDesc(key string, min, max int32) ([]string, bool) {
	ret, err := redis.Strings(this.Do("ZREVRANGEBYSCORE", key, max, min))
	return ret, nil == err
}

// 获取某个指定的key排名(降序排名)
func (this *RedisClient) RankByDesc(key, memeber string) int32 {
	ret, err := redis.Int(this.Do("ZREVRANK", key, memeber))

	if nil != err {
		return 0
	}

	// 下标转换为排名
	return int32(ret + 1)
}

func MinInt32(v1, v2 int32) int32 {

	if v1 < v2 {
		return v1
	}

	return v2
}

func MaxInt32(v1, v2 int32) int32 {

	if v1 > v2 {
		return v1
	}

	return v2
}

func (this *RedisClient) redisBatch(key string, min, max, batch int32, handler func(min, max, page, totalPage int32)) {

	if batch <= 0 {
		panic("use error, batch must > 0!")
	}

	// 总条目数
	totalNum := this.CountSortedSet(key)

	// 没有最低名次
	if min > totalNum {
		return
	}

	// 最高名次超出了, -1 代表最后一名
	if max > totalNum || max < 0 {
		max = totalNum
	}

	// 有多少条需要显示
	range_val := max - min

	// 总页数
	totalPage := range_val / batch

	// 又多余不足一页的算一页
	if 0 != (range_val % batch) {
		totalPage += 1
	}

	for page := int32(0); page < totalPage; page++ {
		// 计算起点
		r_min := min + page*batch
		// 计算终点
		r_max := MinInt32(r_min+batch, max) - 1

		handler(r_min, r_max, page, totalPage)
	}
}

// 升序迭代
func (this *RedisClient) ForEachRankAsc(key string, min, max, batch int32, callback func(page, totalPage int32, list []string)) {

	this.redisBatch(key, min, max, batch, func(r_min, r_max, page, totalPage int32) {
		if list, ok := this.RanksByAsc(key, r_min, r_max); ok && len(list) > 0 {
			callback(page, totalPage, list)
		}
	})
}

// 降序迭代
func (this *RedisClient) ForEachRankDesc(key string, min, max, batch int32, callback func(page, totalPage int32, list []string)) {

	this.redisBatch(key, min, max, batch, func(r_min, r_max, page, totalPage int32) {
		if list, ok := this.RanksByDesc(key, r_min, r_max); ok && len(list) > 0 {
			callback(page, totalPage, list)
		}
	})
}

// 升序迭代 with scores
func (this *RedisClient) ForEachRankAscWithScores(key string, min, max, batch int32, callback func(page, totalPage int32, list []string)) {

	this.redisBatch(key, min, max, batch, func(r_min, r_max, page, totalPage int32) {
		if list, ok := this.RanksByAscWithScores(key, r_min, r_max); ok && len(list) > 0 {
			callback(page, totalPage, list)
		}
	})
}

// 降序迭代 with scores
func (this *RedisClient) ForEachRankDescWithScores(key string, min, max, batch int32, callback func(page, totalPage int32, list []string)) {

	this.redisBatch(key, min, max, batch, func(r_min, r_max, page, totalPage int32) {
		if list, ok := this.RanksByDescWithScores(key, r_min, r_max); ok && len(list) > 0 {
			callback(page, totalPage, list)
		}
	})
}

///////////////////////////////////////////////////////////////////////////////////////////////

// 将一个或多个值 value 插入到列表 key 的表头
func (this *RedisClient) ListPush(key string, members ...string) (int, bool) {

	args := make([]interface{}, 0, len(members)+1)
	args = append(args, key)

	for _, member := range members {
		args = append(args, member)
	}

	length, err := redis.Int(this.Do("LPUSH", args...))
	return length, nil == err
}

// 弹出列表头部
func (this *RedisClient) ListPop(key string) (string, bool) {
	str, err := redis.String(this.Do("LPOP", key))
	return str, nil == err
}

// 返回列表长度
func (this *RedisClient) ListLen(key string) int {
	length, _ := redis.Int(this.Do("LLEN", key))
	return length
}

// 返回指定小标的元素
// 以 0 表示列表的第一个元素, 以 -1 表示列表的最后一个元素， -2 表示列表的倒数第二个元素，以此类推
func (this *RedisClient) ListIndex(key string, index int32) (string, bool) {
	reply, err := this.Do("LINDEX", key, index)

	if reply == nil {
		return "", false
	}

	str, _ := redis.String(reply, err)
	return str, true
}

// 返回指定范围的元素
func (this *RedisClient) ListRange(key string, start, stop int32) ([]string, bool) {
	list, err := redis.Strings(this.Do("LRANGE", key, start, stop))
	return list, nil == err
}

// 移除所有列表中的指定元素
func (this *RedisClient) ListRem(key string, member string) int {
	ret, _ := redis.Int(this.Do("LREM", key, 0, member))
	return ret
}

// 对一个列表进行修剪(trim)，就是说，让列表只保留指定区间内的元素，不在指定区间之内的元素都将被删除
func (this *RedisClient) ListTrim(key string, start, stop int32) {
	this.Do("LTRIM", key, start, stop)
}

// 将列表 key 下标为 index 的元素的值设置为 value
func (this *RedisClient) ListSet(key string, index int32, value string) bool {
	_, err := this.Do("LSET", key, index, value)
	return nil == err
}

///////////////////////////////////////////////////////////////////////////////////////////////
//

// 获取所有keys
func (this *RedisClient) HashKeys(key string) []string {
	rets, _ := redis.Strings(this.Do("HKEYS", key))
	return rets
}

// 获取所有的values
func (this *RedisClient) HashVals(key string) []string {
	rets, _ := redis.Strings(this.Do("HVALS", key))
	return rets
}

// 从hash里面获取一个key
func (this *RedisClient) HashGet(key, field string) (string, bool) {
	ret, err := redis.String(this.Do("HGET", key, field))
	return ret, nil == err
}

// 设置一个 key - value
func (this *RedisClient) HashSet(key, field string, value interface{}) bool {
	_, err := redis.Int(this.Do("HSET", key, field, value))
	return nil == err
}

// 一次性设置多个键到redis hash里面
func (this *RedisClient) HashMSet(key string, sets map[string]string) bool {

	if 0 == len(sets) {
		return true
	}

	// ZADD key score member [[score member] [score member] ...]
	key_pair := make([]interface{}, 0, len(sets)*2+1)
	key_pair = append(key_pair, key)

	for field, value := range sets {
		key_pair = append(key_pair, field, value)
	}

	_, err := this.Do("HMSET", key_pair...)
	return nil == err
}

// hash del.
func (this *RedisClient) HashDel(key string, fields ...string) int {
	args := make([]interface{}, 0, len(fields)+1)
	args = append(args, key)

	for _, field := range fields {
		args = append(args, field)
	}

	ret, _ := redis.Int(this.Do("HDEL", args...))
	return ret
}

// 统计个数
func (this *RedisClient) HashLen(key string) int {
	ret, _ := redis.Int(this.Do("HLEN", key))
	return ret
}

// 判断是否存在
func (this *RedisClient) HExists(key, field string) bool {
	ret, err := redis.Int(this.Do("HEXISTS", key, field))
	return 1 == ret && nil == err
}

// 迭代hash Key
func (this *RedisClient) HashScan(key string, handler func(field, value string)) {

	var cursor int64

	for {
		reply, err := this.Do("HSCAN", key, cursor, "COUNT", 1024)

		if err != nil {
			LOG_ERROR("redis.Strings: %v", err)
			break
		}

		reply_list := reply.([]interface{})

		cursor_str, _ := redis.String(reply_list[0], err)
		cursor, err = strconv.ParseInt(cursor_str, 10, 64)

		list, _ := redis.Strings(reply_list[1], err)

		for i, size := 0, len(list); i < size; i += 2 {
			handler(list[i], list[i+1])
		}

		if err != nil {
			LOG_ERROR("strconv.ParseInt: %v", err)
			break
		}

		// 木有更多数据了 需要跳出
		if cursor <= 0 {
			break
		}
	}
}
