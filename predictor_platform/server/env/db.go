package env

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"server/conf"
	"server/libs/logger"
)

// init mysql
func InitMysql(conf *conf.Conf) *gorm.DB {
	mysqlDB, err := mysqlConn(conf)
	if err != nil {
		logger.Errorf("mysql conn error: %v", err)
	}
	logger.Infof("mysql connect succ!, mysqlDB:", mysqlDB)
	return mysqlDB
}

// mysql connect
func mysqlConn(conf *conf.Conf) (*gorm.DB, error) {
	mysql_conf := conf.MysqlDb
	var mysqldb *gorm.DB
	if mysql_conf.Driver == "sqlite3" {
		db, err := gorm.Open(mysql_conf.Driver, mysql_conf.Database)
		if err != nil {
			return nil, err
		}
		mysqldb = db
	} else {
		connArgs := fmt.Sprintf("%s:%s@(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			mysql_conf.User, mysql_conf.Passwd, mysql_conf.Host, mysql_conf.Port, mysql_conf.Database)
		db, err := gorm.Open(mysql_conf.Driver, connArgs)
		if err != nil {
			logger.Fatalf("db open error: %v", err)
			return nil, err
		}
		db.Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8")
		db.DB().SetMaxIdleConns(0)
		mysqldb = db
	}
	return mysqldb, nil
}
