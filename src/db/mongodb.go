package db

import (
	"net/url"
	"time"

	"gopkg.in/mgo.v2"
	. "logger"
)

func SubStringRange(str string, begin, length int) string {

	rs := []rune(str)
	lth := len(rs)

	if begin < 0 {
		begin = 0
	}

	if length <= 0 {
		length = lth
	}

	if begin >= lth {
		begin = lth
	}

	end := begin + length

	if end > lth {
		end = lth
	}

	return string(rs[begin:end])
}

// 切割字符串
func SubString(str string, offset int) string {
	return SubStringRange(str, offset, -1)
}

// example: mongodb://linbo:password@192.168.1.2:4323/account

func OpenDatabase(urlStr string) (*mgo.Session, *mgo.Database) {

	url, err := url.Parse(urlStr)

	if err != nil {
		LOG_ERROR("url error: %s", err.Error())
		return nil, nil
	}

	//	LOG_DEBUG( "connecting %s@%s ...", url.Scheme, url.Host )

	session := OpenDBSession(url.Host)

	if nil == session {
		return nil, nil
	}

	LOG_DEBUG("connected %s@%s", url.Scheme, url.Host)

	if url.Path == "" || url.Path == "/" {
		return session, nil
	}

	db_name := SubString(url.Path, 1)

	db_data := session.DB(db_name)

	if db_data == nil {
		LOG_ERROR("select database fail. %s", db_name)
		return session, nil
	}

	var (
		username string
		password string
	)

	if url.User != nil {
		username = url.User.Username()
		password, _ = url.User.Password()
	}

	db := LoginDatabase(session, username, password, db_name)

	if db != nil {
		LOG_DEBUG("database selected %s", db.Name)
	} else {
		LOG_ERROR("select to DB[%s] failed.", db_name)
	}

	return session, db
}

func OpenDBSession(addr string) *mgo.Session {

	session, err := mgo.Dial(addr)

	if err != nil {
		LOG_ERROR("fail to connect mongodb: %s %s\n", addr, err.Error())
		return nil
	}

	session.SetMode(mgo.Monotonic, true)

	return session
}

func LoginDatabase(session *mgo.Session, username, password, database string) *mgo.Database {

	if session == nil {
		FATAL("db_session is nil.")
	}

	db_database := session.DB(database)

	if db_database != nil && username != "" {

		LOG_DEBUG("login %s => %s %s", database, username, password)

		if err := db_database.Login(username, password); err != nil {
			LOG_ERROR("db [%s@%s] login fail. %s", username, database, err.Error())
			return nil
		}

		LOG_DEBUG("mongodb login user: %s", username)
	}

	return db_database
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////

type MongoSession struct {
	addr    string
	session *mgo.Session   // 数据库会话
	ping    chan bool      // ping
	db      *MongoDatabase // 默认数据库
	setter  func(*MongoDatabase)
}

// 创建一个数据库会话
func NewMongoSession(addr string, setter func(*MongoDatabase)) *MongoSession {

	session, db := OpenDatabase(addr)

	if nil == session {
		LOG_ERROR("open session fail. %s", addr)
		return nil
	}

	mongo := &MongoSession{
		addr:    addr,
		session: session,
		ping:    make(chan bool),
		db:      NewMongoDatabase(db),
		setter:  setter,
	}

	// keep-alive
	go mongo.keepAlive()

	// 回调给上面
	setter(mongo.db)

	// 服务器关闭的时候顺便关闭这个数据库会话
	onDestroy(mongo)

	return mongo
}

func (this *MongoSession) Key() string {
	return this.addr
}

func (this *MongoSession) reconnect() {

	session, db := OpenDatabase(this.addr)

	if nil == session {
		LOG_ERROR("open session fail. %s", this.addr)
		return
	}

	// 设置会话
	this.session = session

	// 创建数据库
	this.db = NewMongoDatabase(db)

	// 回调给设置
	this.setter(this.db)

	LOG_DEBUG("重连数据库成功! %s", this.addr)
}

// keep-alive
func (this *MongoSession) keepAlive() {

	for {
		select {
		case <-this.ping:
			return
		case <-time.After(time.Second * 1):
			if err := this.session.Ping(); err != nil {
				LOG_ERROR("ping error, %s @ %s", err.Error(), this.Key())
				this.reconnect()
			}
		}
	}

}

// 注销所有数据库会话 & 退出本次数据库会话
func (this *MongoSession) Close() {

	if this.ping != nil {
		close(this.ping)
		this.ping = nil
	}

	if this.db != nil {
		this.db.Logout()
		this.db = nil
	}

	if this.session != nil {
		this.session.LogoutAll()
		this.session.Close()
		this.session = nil
	}

	LOG_INFO("mongod logout: %s", this.Key())
}

// 选择指定数据库
func (this *MongoSession) Select(db_name string) *MongoDatabase {

	if nil != this.db && this.db.Name() == db_name {
		return this.db
	}

	return this.SelectBy(db_name, "", "")
}

// 使用指定账户密码登录数据库
func (this *MongoSession) SelectBy(db_name, username, password string) *MongoDatabase {

	db := LoginDatabase(this.session, username, password, db_name)

	if nil == db {
		LOG_ERROR("选择数据库失败: %s", db_name)
		return nil
	}

	return NewMongoDatabase(db)
}
