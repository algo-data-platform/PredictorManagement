package mock

import (
  "content_service/env"
  "content_service/schema"
)

func AutoMigrateAll(env *env.Env) {
  db := env.Db
  if env.Conf.Db.Driver == "sqlite3" {
    db.AutoMigrate(&schema.Host{})
    db.AutoMigrate(&schema.Service{})
    db.AutoMigrate(&schema.Model{})
    db.AutoMigrate(&schema.HostService{})
    db.AutoMigrate(&schema.ServiceModel{})
    db.AutoMigrate(&schema.ModelHistory{})
    db.AutoMigrate(&schema.StressInfo{})
    db.AutoMigrate(&schema.Config{})
    db.AutoMigrate(&schema.ServiceConfig{})
    db.Exec("PRAGMA foreign_keys = ON")  // enable cascade
  } else if env.Conf.Db.Driver == "mysql" {
    db.Set("gorm:table_options","ENGINE=InnoDB DEFAULT CHARSET=utf8").AutoMigrate(&schema.Host{})
    db.Set("gorm:table_options","ENGINE=InnoDB DEFAULT CHARSET=utf8").AutoMigrate(&schema.Service{})
    db.Set("gorm:table_options","ENGINE=InnoDB DEFAULT CHARSET=utf8").AutoMigrate(&schema.Model{})
    db.Set("gorm:table_options","ENGINE=InnoDB DEFAULT CHARSET=utf8").AutoMigrate(&schema.HostService{})
    db.Set("gorm:table_options","ENGINE=InnoDB DEFAULT CHARSET=utf8").AutoMigrate(&schema.ServiceModel{})
    db.Set("gorm:table_options","ENGINE=InnoDB DEFAULT CHARSET=utf8").AutoMigrate(&schema.ModelHistory{})
  }
}

func getDataCenter(ip string) string {
  return "默认数据中心"
}

func CreateLocalIpRecords(env *env.Env) {
  db := env.Db
  local_ip := env.LocalIp

  // insert service record if not exist yet
  var service schema.Service
  if db.Where("name = ?", "algo_service").First(&service).RecordNotFound() {
    db.Create(&schema.Service{Name: "algo_service", Desc: "stage: " + env.Conf.Stage})
  }
  // insert current host record if not exist
  var host schema.Host
  if db.Where("ip = ?", local_ip).First(&host).RecordNotFound() {
    db.Create(&schema.Host{Ip: local_ip, DataCenter: getDataCenter(local_ip), Desc: "NA"})
  }

  // insert host_service record if not exist
  var hs schema.HostService
  if db.Where("hid = ? AND sid = ?", host.ID, service.ID).First(&hs).RecordNotFound() {
    db.Where("ip = ?", local_ip).First(&host)
    db.Where("name = ?", "algo_service").First(&service)
    db.Create(&schema.HostService{Hid: host.ID, Sid: service.ID, Desc: host.Ip + " -> " + service.Name})
  }

  if env.Conf.Log.Level == "debug" {
    PrintAllTables(env)
  }
}

func CreateDummyRecords(env *env.Env) {
  db := env.Db
  db.Create(&schema.Host{Ip: "127.1.101.172", DataCenter: "数据中心1"})
  db.Create(&schema.Host{Ip: "127.3.14.214", DataCenter: "数据中心2"})
  db.Create(&schema.Host{Ip: "127.2.11.10", DataCenter: "数据中心3"})
  db.Create(&schema.Host{Ip: "127.4.0.8", DataCenter: "数据中心4", Desc: "数据中心4"})

	db.Create(&schema.Service{Name: "service_1", Desc: "服务1"})
	db.Create(&schema.Service{Name: "service_2", Desc: "服务2"})

  db.Create(&schema.Model{Name: "model_1", Path: "/a/b/c", Desc: "1"})
  db.Create(&schema.Model{Name: "model_2", Path: "/d/e/f", Desc: "2"})
  db.Create(&schema.Model{Name: "model_3", Path: "/g/h/c", Desc: "3"})

	db.Create(&schema.ModelHistory{ModelName: "model_1", Timestamp: "20190829_175600", Md5: "a1b2c3d4e5f6"})
	db.Create(&schema.ModelHistory{ModelName: "model_1", Timestamp: "20190428_100045", Md5: "a1b2c3d4e5f7"})
	db.Create(&schema.ModelHistory{ModelName: "model_2", Timestamp: "20190712_103100", Md5: "a1b2c3d4e5f9"})
	db.Create(&schema.ModelHistory{ModelName: "model_2", Timestamp: "20190820_175505", Md5: "a1b2c3d4e5f8"})
	db.Create(&schema.ModelHistory{ModelName: "model_3", Timestamp: "20190821_175508", Md5: "a1b2c3d4e5h8"})
}

func CleanUp(env *env.Env) {
  db := env.Db
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

  if env.Conf.Db.Driver == "sqlite3" {
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
  } else if env.Conf.Db.Driver == "mysql" {
    db.Exec("ALTER TABLE hosts AUTO_INCREMENT = 1")
    db.Exec("ALTER TABLE services AUTO_INCREMENT = 1")
    db.Exec("ALTER TABLE models AUTO_INCREMENT = 1")
    db.Exec("ALTER TABLE host_services AUTO_INCREMENT = 1")
    db.Exec("ALTER TABLE service_models AUTO_INCREMENT = 1")
    db.Exec("ALTER TABLE model_histories AUTO_INCREMENT = 1")
  }
}
