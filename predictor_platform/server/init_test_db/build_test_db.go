package main

import (
	"log"
	"server/conf"
	"server/env"
	"server/mock"
	"server/server/dao"
)

func main() {
	log.Println("build test db....")
	conf := &conf.Conf{}
	conf.MysqlDb.Driver = "sqlite3"
	conf.MysqlDb.Database = "./ad_test_db.db"
	db := env.InitMysql(conf)
	dao.SetMysqlDB(db)
	mock.AutoMigrateAll(db, conf.MysqlDb.Driver)
	mock.CleanUp(db, conf.MysqlDb.Driver)
	conf.HttpHost="LOCAL_IP"
	localIp, _:= env.GetLocalIp(conf)
	mock.BuildTestDB(db, localIp)
	log.Println("test db build success")
}
