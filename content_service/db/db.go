package db

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
  "fmt"
  "content_service/libs/logger"
  "content_service/conf"
)

func New(conf *conf.Conf) *gorm.DB {
  var db_args string
  switch conf.Db.Driver {
  case "sqlite3":
    db_args = conf.Db.Name
  case "mysql":
    db_args = fmt.Sprintf("%s:%s@(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
      conf.Db.User, conf.Db.Passwd, conf.Db.Host, conf.Db.Port, conf.Db.Name)
  default:
    logger.Panicf("unrecognized db driver: %v, conf=%+v", conf.Db.Driver, *conf)
  }

  db, err := gorm.Open(conf.Db.Driver, db_args)
  if err != nil {
    logger.Panicf("failed to connect database: err=%v, conf=%+v", err, *conf)
  } else {
		logger.Infof("db initiated and connected: dbDriver=%v dbName=%v", conf.Db.Driver, conf.Db.Name)
	}

  if conf.Db.Driver == "mysql" {
    db.DB().SetMaxIdleConns(0)
  }

	return db
}
