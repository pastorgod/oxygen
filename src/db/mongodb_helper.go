package db

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	. "logger"
	"reflect"
	"strings"
)

type C bson.M
type KVPair bson.D
type RKVPair bson.RawD
type ObjectId bson.ObjectId
type SaveFields map[string]struct{}

type IRecord interface {

	// collection key.
	Key() C

	// collection name.
	Table() string

	// collection indexs
	IndexKey() []string
}

type IMutilIndexs interface {

	// mutil indexs
	MutilIndexKey() []string
}

func ObjectIdStr() string {
	return bson.NewObjectId().String()
}

func NewObjectId() ObjectId {
	return ObjectId(bson.NewObjectId())
}

func (this ObjectId) String() string {
	return bson.ObjectId(this).Hex()
}

func IsSlice(r interface{}) bool {

	val := reflect.ValueOf(r)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	return val.Kind() == reflect.Slice
}

func CheckFetchAll(r interface{}) (bool, string) {

	val := reflect.ValueOf(r)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	typ := reflect.TypeOf(r)

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	if val.Kind() == reflect.Slice {

		typ = typ.Elem()

		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}

		if re, ok := reflect.New(typ).Interface().(IRecord); ok {
			return true, re.Table()
		}

		LOG_FATAL("result must to implementation IRecord interface! %v", typ)
		return true, ""
	}

	if re, ok := reflect.New(typ).Interface().(IRecord); ok {
		return false, re.Table()
	}

	LOG_FATAL("result must to implementation IRecord interface! %v", typ)
	return false, ""
}

func FetchFieldNameValue(data IRecord, field string) (string, interface{}) {

	value := reflect.ValueOf(data).Elem()
	fieldVal := value.FieldByName(field)

	structField, found := reflect.TypeOf(data).Elem().FieldByName(field)

	if !found {
		LOG_FATAL("notFound Field! %s", field)
	}

	fieldName := structField.Tag.Get("bson")

	if "" == fieldName {
		fieldName = strings.ToLower(field)
	}

	return fieldName, fieldVal.Interface()
}

func EnsureIndex(DB *mgo.Database, data IRecord) bool {

	coll := DB.C(data.Table())

	if indexs := data.IndexKey(); len(indexs) > 0 {

		DEBUG("UniqueEnsureIndex: %s %s => %v", DB.Name, data.Table(), indexs)

		for _, key := range indexs {
			index := mgo.Index{
				Key:    []string{key},
				Unique: true,
			}

			if err := coll.EnsureIndex(index); err != nil {
				LOG_ERROR("UniqueEnsureIndex Fail. %s", err.Error())
				ERROR("UniqueEnsureIndex Fail. %s", err.Error())
				return false
			}
		}
	}

	if mt, ok := data.(IMutilIndexs); ok {

		indexs := mt.MutilIndexKey()

		if indexs != nil && len(indexs) > 0 {

			DEBUG("MutilEnsureIndex: %s %s => %v", DB.Name, data.Table(), indexs)

			for _, key := range indexs {
				index := mgo.Index{
					Key:      []string{key},
					Unique:   false,
					DropDups: false,
				}

				if err := coll.EnsureIndex(index); err != nil {
					LOG_ERROR("MutilEnsureIndex Fail. %s", err.Error())
					ERROR("MutilEnsureIndex Fail. %s", err.Error())
					return false
				}
			}
		}
	}

	return true
}

func FetchResult(isarray bool, table string, query *mgo.Query, result interface{}) bool {

	if isarray {
		if err := query.All(result); err != nil {
			LOG_ERROR("mongodb find [%s] %s", table, err.Error())
			return false
		}
	} else {
		if err := query.One(result); err != nil {

			LOG_ERROR("mongodb find [%s] %s", table, err.Error())
			return false
		}
	}

	return true
}

func Insert(DB *mgo.Database, data IRecord) bool {

	//	DEBUG( "Insert DB Collection: %s, Key: %s:%s", data.Table(), fmt.Sprintf( "%v = > %+v", data.Key(), data ) )

	coll := DB.C(data.Table())

	if err := coll.Insert(data); err != nil {
		LOG_ERROR("Insert [%s] %s", data.Table(), err.Error())
		return false
	}

	return true
}

func Update(DB *mgo.Database, data IRecord) bool {

	coll := DB.C(data.Table())

	if err := coll.Update(data.Key(), data); err != nil {
		LOG_ERROR("Update [%s] %v %s", data.Table(), data.Key(), err.Error())

		// 跳过更新不存在的目标
		if err != mgo.ErrNotFound {
			return false
		}
	}

	return true
}

func UpdateBy(DB *mgo.Database, table string, cond, updater C) bool {

	coll := DB.C(table)

	if err := coll.Update(cond, C{"$set": updater}); err != nil {

		LOG_ERROR("UpdateField [%v] [%s.%v] %s", cond, table, updater, err.Error())

		// 更新一个不存在的目标则跳过这个更新
		if err != mgo.ErrNotFound {
			return false
		}
	}

	return true
}

func UpdateField(DB *mgo.Database, data IRecord, field string) bool {

	/* DEBUG ONLY */
	defer func() {
		if err := recover(); err != nil {
			LOG_ERROR("UpdateField Error: %s", err)
		}
	}()

	fieldName, fieldValue := FetchFieldNameValue(data, field)

	//DEBUG( "UpdateField => %s %s", data.Table(), fieldName )

	coll := DB.C(data.Table())

	if err := coll.Update(data.Key(), C{"$set": C{fieldName: fieldValue}}); err != nil {
		LOG_ERROR("UpdateField [%v] [%s.%s] %s", data.Key(), data.Table(), fieldName, err.Error())

		// 跳过更新一个不存在的目标
		if err != mgo.ErrNotFound {
			return false
		}
	}

	return true
}

func UpdateFields(DB *mgo.Database, data IRecord, fields []string) bool {
	/* DEBUG ONLY */
	defer func() {
		if err := recover(); err != nil {
			LOG_ERROR("UpdateFields Error: %s", err)
		}
	}()

	// { "$set" : { "name" : "xxx", "field1" : { ... }, "field2" : { ... }  } }

	updater := C{}

	for _, field := range fields {
		fieldName, fieldValue := FetchFieldNameValue(data, field)
		updater[fieldName] = fieldValue
	}

	coll := DB.C(data.Table())

	if err := coll.Update(data.Key(), C{"$set": updater}); err != nil {
		LOG_ERROR("UpdateFields [%v] [%v] %v %s", data.Key(), data.Table(), updater, err.Error())

		// 跳过更新一个不存在的目标
		if err != mgo.ErrNotFound {
			return false
		}
	}

	return true
}

func UpdateFieldsSet(DB *mgo.Database, data IRecord, fields SaveFields) bool {
	return UpdateFieldsSetBy(DB, data.Key(), data, fields)
}

func UpdateFieldsSetBy(DB *mgo.Database, cond C, data IRecord, fields SaveFields) bool {
	/* DEBUG ONLY */
	defer func() {
		if err := recover(); err != nil {
			LOG_ERROR("UpdateFieldsSet Error: %s", err)
		}
	}()

	updater := C{}

	for field, _ := range fields {
		fieldName, fieldValue := FetchFieldNameValue(data, field)
		updater[fieldName] = fieldValue
	}

	// 没有需要保存的
	if 0 == len(updater) {
		return true
	}

	coll := DB.C(data.Table())

	// 批量更新
	if err := coll.Update(cond, C{"$set": updater}); err != nil {
		LOG_ERROR("UpdateFieldsSet err: %s table: %s cond: [%v] updater: [%v]", err.Error(), data.Table(), cond, updater)

		// 跳过更新一个不存在的目标
		if err != mgo.ErrNotFound {
			return false
		}
	}

	return true
}

func Delete(DB *mgo.Database, data IRecord) bool {

	coll := DB.C(data.Table())

	if err := coll.Remove(data.Key()); err != nil {
		LOG_ERROR("Delete [%s] %s", data.Table(), err.Error())
		return false
	}

	return true
}

func DeleteBy(DB *mgo.Database, table string, cond C) bool {

	coll := DB.C(table)

	if err := coll.Remove(cond); err != nil {
		LOG_ERROR("Delete [%s] %v %s", table, cond, err.Error())
		return false
	}

	return true
}

func DeleteAll(DB *mgo.Database, table string, cond C) bool {

	coll := DB.C(table)

	if info, err := coll.RemoveAll(cond); err != nil {
		LOG_ERROR("Delete [%s] %s", table, err.Error())
		return false
	} else {
		LOG_DEBUG("DeleteAll.%s [%d, %d]", table, info.Updated, info.Removed)
	}

	return true
}

func Find(DB *mgo.Database, cond C, result interface{}) bool {

	isarray, table := CheckFetchAll(result)

	if table == "" {
		return false
	}

	query := DB.C(table).Find(cond)

	return FetchResult(isarray, table, query, result)
}

// result := &Account{ Id : "linbo" }
func FindOne(DB *mgo.Database, result interface{}) bool {

	if data, ok := result.(IRecord); ok {
		coll := DB.C(data.Table())

		if err := coll.Find(data.Key()).One(result); err != nil {

			if err != mgo.ErrNotFound {
				LOG_ERROR("mongodb find [%s] @ %v %s", data.Table(), data.Key(), err.Error())
			}

			return false
		}
	} else {
		LOG_FATAL("result must to implementation IRecord interface!")
	}

	return true
}

// result := []Account{}
func FindAll(DB *mgo.Database, table string, key C, result interface{}) bool {

	coll := DB.C(table)

	if err := coll.Find(key).All(result); err != nil {

		LOG_ERROR("mongodb find all [%s] %s", table, err.Error())
		return false
	}

	return true
}

func Count(DB *mgo.Database, table string, cond C) int {

	coll := DB.C(table)

	// 没有提供条件则直接调用自带方法
	if cond == nil || 0 == len(cond) {
		if num, err := coll.Count(); err != nil {

			LOG_ERROR("mongodb count [%s] %s", table, err.Error())
			return 0
		} else {
			return num
		}
	}

	if num, err := coll.Find(cond).Count(); err != nil {

		LOG_ERROR("mongodb count [%s] %s", table, err.Error())
		return 0
	} else {
		return num
	}

	return 0
}

// ( []string{ "-age", "name" }, 10, &[]Character )
func Select(DB *mgo.Database, cond, filter C, sort []string, offset, limit int, result interface{}) bool {

	isarray, table := CheckFetchAll(result)

	if table == "" {
		return false
	}

	query := DB.C(table).Find(cond).Select(filter)

	if len(sort) > 0 {
		query = query.Sort(sort...)
	}

	if offset > 0 {
		query = query.Skip(offset)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	return FetchResult(isarray, table, query, result)
}

// 运行命令
func Run(DB *mgo.Database, cmd interface{}, result interface{}) bool {

	if err := DB.Run(cmd, result); err != nil {
		LOG_ERROR("mongodb run: ", err.Error())
		return false
	}

	return true
}

// 聚合管道
/*
var results []interface{}

ok := db.DB.Aggregate( "Character", []db.C {
		db.C{ "$match" : db.Condition{ "last_login" : db.Condition{ "$gte" : begin.Unix(), "$lte" : end.Unix() } } },
		db.C{ "$group" : db.Condition{ "_id" : "$accid", "gem" : db.Condition{ "$sum" : "$attrib.gem" } } },
}, &results )

*/
func Pipe(DB *mgo.Database, table string, pipeline []C, result interface{}) bool {

	query := DB.C(table).Pipe(pipeline)

	var err error

	if IsSlice(result) {
		err = query.All(result)
	} else {
		err = query.One(result)
	}

	if err != nil {

		if err != mgo.ErrNotFound {
			LOG_ERROR("mongodb pipe: %s %s", table, err.Error())
		}

		return false
	}

	return true
}

// 去重
func Distinct(DB *mgo.Database, table string, cond C, field string, result interface{}) bool {

	coll := DB.C(table)

	if err := coll.Find(cond).Distinct(field, result); err != nil {
		LOG_ERROR("mongodb distinct: %v", err)
		return false
	}

	return true
}

// 清空表数据
func Truncate(DB *mgo.Database, table string) bool {

	coll := DB.C(table)

	if err := coll.DropCollection(); err != nil {
		LOG_WARN("mongodb truncate: %s %s", table, err.Error())
		return false
	}

	return true
}

// 查找& 修改
func FindAndModify(DB *mgo.Database, table string, cond, updater C, result IRecord, upsert ...bool) bool {

	coll := DB.C(table)

	change := mgo.Change{
		Update:    updater, // C{ "$set" : { key : value } }
		Upsert:    true,
		ReturnNew: true,
	}

	// 如果有指定这按照指定的来执行
	if len(upsert) > 0 {
		// false -> 不存在时不插入, true -> 不存在时插入并且更新
		change.Upsert = upsert[0]
	}

	if _, err := coll.Find(cond).Apply(change, result); err != nil {
		LOG_ERROR("mongodb FindAndModify: %s", err.Error())
		return false
	}

	return true
}

// 迭代指定条件
// 更具指定条件过滤器分批查询

func ForEach(DB *mgo.Database,
	ins IRecord,
	cond, filter C,
	sort []string,
	batchCount int,
	callback func(page, totoalPage int, results interface{})) (taskNum int, taskFun func(taskNo int)) {

	if batchCount <= 0 {
		panic("use error, batchCount must > 0")
	}

	total := Count(DB, ins.Table(), cond)

	if total <= 0 {
		LOG_DEBUG("ForEach.%s by %v of 0", ins.Table(), cond)
		return 0, nil
	}

	// 计算一下有多少页
	totalPage := total / batchCount

	if 0 != (total % batchCount) {
		totalPage += 1
	}

	LOG_DEBUG("ForEach.%s, batchCount: %d, totalPage: %d, total: %d", ins.Table(), batchCount, totalPage, total)

	sliceT := reflect.SliceOf(reflect.TypeOf(ins).Elem())

	return totalPage, func(cur_page int) {

		resultV := reflect.New(sliceT)

		slicePtr := resultV.Interface()

		sliceV := reflect.ValueOf(slicePtr)

		if !Select(DB, cond, filter, sort, cur_page*batchCount, batchCount, sliceV.Elem().Addr().Interface()) {
			LOG_ERROR("Select.%s by %v to fail.", ins.Table(), cond)
			return
		}

		callback(cur_page, totalPage, resultV.Elem().Interface())
	}
}
