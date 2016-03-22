package base

import (
	"db"
)

type RankMode int32

const (
	RankMode_Desc RankMode = 1 // 降序排行
	RankMode_Asc  RankMode = 2 // 升序排行
)

type RedisRank struct {
	client   *db.RedisClient
	rankKey  string
	rankMode RankMode
}

func NewRedisRank(key string, mode RankMode) *RedisRank {
	return NewRedisRankWith(key, mode, db.Redis)
}

func NewRedisRankWith(key string, mode RankMode, client *db.RedisClient) *RedisRank {
	return &RedisRank{
		client:   client,
		rankKey:  key,
		rankMode: mode,
	}
}

func (this *RedisRank) RedisKey() string {
	return this.rankKey
}

func (this *RedisRank) CheckMode(mode RankMode) bool {
	return mode == this.rankMode
}

// 查询排名 0 表示不再排行榜里面或者出错
func (this *RedisRank) RankOf(member string) int32 {

	if this.CheckMode(RankMode_Desc) {
		return this.client.RankByDesc(this.RedisKey(), member)
	}

	return this.client.RankByAsc(this.RedisKey(), member)
}

// 查询指定排名的member
func (this *RedisRank) RankBy(rank int32) string {
	if ns := this.RanksBy(rank, rank); len(ns) > 0 {
		return ns[0]
	}
	return ""
}

// 获得改玩家的Score值
func (this *RedisRank) Score(member string) uint64 {
	score, _ := this.client.GetScore(this.RedisKey(), member)
	return score
}

// 批量添加或者更新排行榜
func (this *RedisRank) Updates(sets map[string]uint64) {
	this.client.AddSortedSet(this.RedisKey(), sets)
}

// 添加或者更新到排行榜
func (this *RedisRank) Update(memeber string, score uint64) {
	this.client.UpdateSortedSet(this.RedisKey(), memeber, score)
}

// 添加key的score值并且返回最新的值
func (this *RedisRank) AddScore(memeber string, score uint64) uint64 {
	value, _ := this.client.IncSortedSet(this.RedisKey(), memeber, score)
	return value
}

// 从排行榜中移除一个或者多个
func (this *RedisRank) Remove(memebers ...string) int {
	return this.client.RemoveSortedSet(this.RedisKey(), memebers...)
}

// 删除一段范围
func (this *RedisRank) RemoveRange(begin, end int32) {
	this.client.RemoveRangeSortedSet(this.RedisKey(), begin, end)
}

// 获得指定排行区间的数据(升序)
func (this *RedisRank) RanksByAsc(begin, end int32) []string {
	keys, _ := this.client.RanksByAsc(this.RedisKey(), begin, end)
	return keys
}

// 获得指定排名区间的数据(降序)
func (this *RedisRank) RanksByDesc(begin, end int32) []string {
	keys, _ := this.client.RanksByDesc(this.RedisKey(), begin, end)
	return keys
}

// 获得排名区间的数据
func (this *RedisRank) RanksBy(begin, end int32) []string {

	if this.CheckMode(RankMode_Asc) {
		return this.RanksByAsc(begin, end)
	}

	return this.RanksByDesc(begin, end)
}

// 获得指定排名区间的key-score(升序)
func (this *RedisRank) RanksByAscWithScores(begin, end int32) []string {
	list, _ := this.client.RanksByAscWithScores(this.RedisKey(), begin, end)
	return list
}

// 获得指定排行榜排名区间的key-score(降序)
func (this *RedisRank) RanksByDescWithScores(begin, end int32) []string {
	list, _ := this.client.RanksByDescWithScores(this.RedisKey(), begin, end)
	return list
}

// 获得指定排名区间key-score
func (this *RedisRank) RanksWithScores(begin, end int32) []string {

	if this.CheckMode(RankMode_Asc) {
		return this.RanksByAscWithScores(begin, end)
	}

	return this.RanksByDescWithScores(begin, end)
}

// 访问指定位置的目标
func (this *RedisRank) At(rank int32) (member string, score uint64, ok bool) {

	rets := this.RanksWithScores(rank, rank)

	if ok = (2 == len(rets)); !ok {
		return
	}

	member = rets[0]
	score = AtoUint64(rets[1])
	return
}

// page -> 获取第几页的数据, 0开始
// size -> 每一页的大小
func (this *RedisRank) Paging(page, size int32, handler func(string, uint64)) (total_page int32) {

	totalNum := this.Count()

	// 计算完整页
	total_page = totalNum / size

	// 不足一页的也算单独占一页
	if 0 != (totalNum % size) {
		total_page += 1
	}

	if page < total_page {
		begin := size * page
		end := MinInt32(begin+size, totalNum)
		if end > 0 {
			end -= 1
		}
		this.RanksWithScoresHelper(begin, end, handler)
	}

	return
}

// 获得排名区间的数据 key - value
func (this *RedisRank) RanksWithScoresHelper(begin, end int32, callback func(string, uint64)) {

	list := this.RanksWithScores(begin, end)

	page := len(list) / 2

	for i := 0; i < page; i++ {
		member := list[i*2]
		value := list[i*2+1]

		callback(member, AtoUint64(value))
	}
}

// 升序迭代 with scores
func (this *RedisRank) ForEachAscWithScores(begin, end, batch int32, callback func(page, totalPage int32, list []string)) {
	this.client.ForEachRankAscWithScores(this.RedisKey(), begin, end, batch, callback)
}

// 降序迭代 with scores
func (this *RedisRank) ForEachDescWithScores(begin, end, batch int32, callback func(page, totalPage int32, list []string)) {
	this.client.ForEachRankDescWithScores(this.RedisKey(), begin, end, batch, callback)
}

// 迭代排行榜 with scores
func (this *RedisRank) ForEachWithScores(begin, end, batch int32, callback func(page, totalPage int32, list []string)) {

	if this.CheckMode(RankMode_Asc) {
		this.ForEachAscWithScores(begin, end, batch, callback)
		return
	}

	this.ForEachDescWithScores(begin, end, batch, callback)
}

// 迭代排行榜 with socres
func (this *RedisRank) ForEachWithScoresHelper(begin, end, batch int32, callback func(string, uint64)) {

	this.ForEachWithScores(begin, end, batch, func(p, tP int32, list []string) {

		page := len(list) / 2

		for i := 0; i < page; i++ {
			member := list[i*2]
			value := list[i*2+1]

			callback(member, AtoUint64(value))
		}
	})
}

// 升序迭代排行榜
func (this *RedisRank) ForEachAsc(min, max, batch int32, callback func(page, totalPage int32, list []string)) {
	this.client.ForEachRankAsc(this.RedisKey(), min, max, batch, callback)
}

// 降序迭代排行榜
func (this *RedisRank) ForEachDesc(min, max, batch int32, callback func(page, totalPage int32, list []string)) {
	this.client.ForEachRankDesc(this.RedisKey(), min, max, batch, callback)
}

// 迭代排行榜
func (this *RedisRank) ForEach(min, max, batch int32, callback func(page, totalPage int32, list []string)) {

	if this.CheckMode(RankMode_Asc) {
		this.ForEachAsc(min, max, batch, callback)
		return
	}

	this.ForEachDesc(min, max, batch, callback)
}

// 统计一下有多少
func (this *RedisRank) Count() int32 {
	return this.client.CountSortedSet(this.RedisKey())
}

// 清空排行榜
func (this *RedisRank) Clear() {
	this.RemoveRange(0, -1)
}

// 销毁排行榜
func (this *RedisRank) Destroy() {
	this.client.Del(this.RedisKey())
}
