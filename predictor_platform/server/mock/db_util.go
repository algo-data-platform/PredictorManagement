package mock

import (
	"github.com/jinzhu/gorm"
	"server/schema"
)

func AutoMigrateAll(db *gorm.DB, driver string) {
	if driver == "sqlite3" {
		db.AutoMigrate(&schema.Host{})
		db.AutoMigrate(&schema.Service{})
		db.AutoMigrate(&schema.Model{})
		db.AutoMigrate(&schema.HostService{})
		db.AutoMigrate(&schema.ServiceModel{})
		db.AutoMigrate(&schema.ModelHistory{})
		db.AutoMigrate(&schema.StressInfo{})
		db.AutoMigrate(&schema.Config{})
		db.AutoMigrate(&schema.ServiceConfig{})
		db.Exec("PRAGMA foreign_keys = ON") // enable cascade
	} else if driver == "mysql" {
		db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8").AutoMigrate(&schema.Host{})
		db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8").AutoMigrate(&schema.Service{})
		db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8").AutoMigrate(&schema.Model{})
		db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8").AutoMigrate(&schema.HostService{})
		db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8").AutoMigrate(&schema.ServiceModel{})
		db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8").AutoMigrate(&schema.ModelHistory{})
		db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8").AutoMigrate(&schema.StressInfo{})
	}
}

func CleanUp(db *gorm.DB, driver string) {
	// delete records
	db.Exec("delete from host_services where 1;")
	db.Exec("delete from service_models where 1;")
	db.Exec("delete from model_histories where 1;")
	db.Exec("delete from hosts where 1;")
	db.Exec("delete from services where 1;")
	db.Exec("delete from models where 1;")
	db.Exec("delete from stress_infos where 1;")
	db.Exec("delete from configs where 1;")
	db.Exec("delete from service_configs where 1;")

	if driver == "sqlite3" {
		// reset auto increment
		db.Exec("delete from sqlite_sequence where name='hosts'")
		db.Exec("delete from sqlite_sequence where name='services'")
		db.Exec("delete from sqlite_sequence where name='models'")
		db.Exec("delete from sqlite_sequence where name='host_services'")
		db.Exec("delete from sqlite_sequence where name='service_models'")
		db.Exec("delete from sqlite_sequence where name='model_histories'")
		db.Exec("delete from sqlite_sequence where name='stress_infos'")
		db.Exec("delete from sqlite_sequence where name='configs'")
		db.Exec("delete from sqlite_sequence where name='service_configs'")
	} else if driver == "mysql" {
		db.Exec("ALTER TABLE hosts AUTO_INCREMENT = 1")
		db.Exec("ALTER TABLE services AUTO_INCREMENT = 1")
		db.Exec("ALTER TABLE models AUTO_INCREMENT = 1")
		db.Exec("ALTER TABLE host_services AUTO_INCREMENT = 1")
		db.Exec("ALTER TABLE service_models AUTO_INCREMENT = 1")
		db.Exec("ALTER TABLE model_histories AUTO_INCREMENT = 1")
		db.Exec("ALTER TABLE stress_infos AUTO_INCREMENT = 1")
	}
}

func BuildTestDB(db *gorm.DB, localIp string) {
	db.Create(&schema.Service{Name: "service_1"}) // Sid: 1
	db.Create(&schema.Service{Name: "service_2"}) // Sid: 2
	db.Create(&schema.Service{Name: "service_3"}) // Sid: 3
	db.Create(&schema.Service{Name: "predictor_service_dev"}) // Sid: 4

	db.Create(&schema.Model{Name: "model_a", Desc: "模型A", Path: "/data0/dummy/path"}) // Mid: 1
	db.Create(&schema.Model{Name: "model_b", Desc: "模型B", Path: "/data0/dummy/path"}) // Mid: 2
	db.Create(&schema.Model{Name: "model_c", Desc: "模型C", Path: "/data0/dummy/path"}) // Mid: 3
	db.Create(&schema.Model{Name: "model_d", Desc: "模型D", Path: "/data0/dummy/path"}) // Mid: 4
	db.Create(&schema.Model{Name: "catboost_direct_v0_demo_model", Desc: "demo模型", Path: "./data"}) // Mid: 5

	db.Create(&schema.ServiceModel{Sid: 1, Mid: 1, Desc: "service_1 -> model_a"})
	db.Create(&schema.ServiceModel{Sid: 1, Mid: 2, Desc: "service_1 -> model_b"})
	db.Create(&schema.ServiceModel{Sid: 1, Mid: 3, Desc: "service_1 -> model_c"})
	db.Create(&schema.ServiceModel{Sid: 2, Mid: 1, Desc: "service_2 -> model_a"})
	db.Create(&schema.ServiceModel{Sid: 2, Mid: 4, Desc: "service_2 -> model_d"})
	db.Create(&schema.ServiceModel{Sid: 4, Mid: 5, Desc: "predictor_service_dev -> catboost_direct_v0_demo_model"})

	db.Create(&schema.ModelHistory{ModelName: "model_a", Timestamp: "20201011_090000", Desc: "Validated"}) // ID: 1
	db.Create(&schema.ModelHistory{ModelName: "model_a", Timestamp: "20201020_090000", Desc: "Validated"}) // ID: 2
	db.Create(&schema.ModelHistory{ModelName: "model_b", Timestamp: "20201020_090000", Desc: "Validated"}) // ID: 3
	db.Create(&schema.ModelHistory{ModelName: "model_b", Timestamp: "20201020_010000", Desc: "Validated"}) // ID: 4
	db.Create(&schema.ModelHistory{ModelName: "model_c", Timestamp: "20201020_091000", Desc: "Validated"}) // ID: 5
	db.Create(&schema.ModelHistory{ModelName: "model_d", Timestamp: "20201001_010000", Desc: "Validated"}) // ID: 6
	db.Create(&schema.ModelHistory{ModelName: "catboost_direct_v0_demo_model", Timestamp: "20200826_103000", Desc: "Validated"}) // ID: 7

	db.Create(&schema.Host{ID: 2, Ip: "127.1.0.2"})
	db.Create(&schema.Host{ID: 3, Ip: "127.2.0.3"})
	db.Create(&schema.Host{ID: 4, Ip: "127.3.0.4"})
	db.Create(&schema.Host{ID: 5, Ip: "127.1.0.5"})
	db.Create(&schema.Host{ID: 6, Ip: "127.1.0.6", Desc: "elastic_expansion"})
	db.Create(&schema.Host{ID: 7, Ip: "127.3.0.7"})
	db.Create(&schema.Host{ID: 8, Ip: localIp})

	db.Create(&schema.HostService{Hid: 2, Sid: 1, Desc: "127.1.0.2 -> service_1", LoadWeight: 120})
	db.Create(&schema.HostService{Hid: 5, Sid: 1, Desc: "127.1.0.5 -> service_1", LoadWeight: 130})
	db.Create(&schema.HostService{Hid: 6, Sid: 1, Desc: "127.1.0.6 -> service_1", LoadWeight: 140})
	db.Create(&schema.HostService{Hid: 3, Sid: 2, Desc: "127.2.0.3 -> service_2", LoadWeight: 100})
	db.Create(&schema.HostService{Hid: 3, Sid: 3, Desc: "127.2.0.3 -> service_3", LoadWeight: 200})
	db.Create(&schema.HostService{Hid: 4, Sid: 2, Desc: "127.3.0.4 -> service_2", LoadWeight: 300})
	db.Create(&schema.HostService{Hid: 4, Sid: 3, Desc: "127.3.0.4 -> service_3", LoadWeight: 90})
	db.Create(&schema.HostService{Hid: 7, Sid: 1, Desc: "127.3.0.7 -> service_1", LoadWeight: 80})
	db.Create(&schema.HostService{Hid: 8, Sid: 4, Desc: "localip -> predictor_service_dev", LoadWeight: 80})

	db.Create(&schema.Config{Description: "default", Config: `{"tf_thread_ratio": 0.5, "heavy_tasks_thread_ratio": 0.5, "request_cpu_thread_ratio": 1.0, "feature_extract_tasks_ratio":0.5}`})

	db.Create(&schema.ServiceConfig{Sid: 1, Description: "service_1->default", Cid: 1})
	db.Create(&schema.ServiceConfig{Sid: 2, Description: "service_2->default", Cid: 1})
	db.Create(&schema.ServiceConfig{Sid: 3, Description: "service_3->default", Cid: 1})
	db.Create(&schema.ServiceConfig{Sid: 4, Description: "predictor_service_dev->default", Cid: 1})
}
