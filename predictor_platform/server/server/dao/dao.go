package dao

import (
	"github.com/jinzhu/gorm"
	"server/libs/logger"
	"server/schema"
)

var Mysql_db *gorm.DB

const RecordNotFound = "record not found"

func SetMysqlDB(db *gorm.DB) {
	// mysql handler
	Mysql_db = db
}

func TableCheck() {
	if !TablesExist() {
		logger.Errorf("table not exists! AutoMigrateAll")
		AutoMigrateAll()
	}
}

func TablesExist() bool {
	return Mysql_db.HasTable(&schema.Host{}) &&
		Mysql_db.HasTable(&schema.Model{}) &&
		Mysql_db.HasTable(&schema.Service{}) &&
		Mysql_db.HasTable(&schema.ModelHistory{}) &&
		Mysql_db.HasTable(&schema.HostService{}) &&
		Mysql_db.HasTable(&schema.ServiceModel{}) &&
		Mysql_db.HasTable(&schema.Config{}) &&
		Mysql_db.HasTable(&schema.ServiceConfig{})
}

//to do, will be moved into script
func AutoMigrateAll() {
	Mysql_db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8").AutoMigrate(&schema.Host{})
	Mysql_db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8").AutoMigrate(&schema.Service{})
	Mysql_db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8").AutoMigrate(&schema.Model{})
	Mysql_db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8").AutoMigrate(&schema.HostService{})
	Mysql_db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8").AutoMigrate(&schema.ServiceModel{})
	Mysql_db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8").AutoMigrate(&schema.ModelHistory{})
	Mysql_db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8").AutoMigrate(&schema.Config{})
	Mysql_db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8").AutoMigrate(&schema.ServiceConfig{})
}
