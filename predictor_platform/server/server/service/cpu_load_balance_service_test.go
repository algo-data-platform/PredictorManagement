package service

import (
	"reflect"
	"server/common"
	"server/conf"
	"server/env"
	"server/metrics"
	"server/mock"
	"server/schema"
	"server/server/dao"
	"server/util"
	"testing"
)

func TestGetAllServiceList(t *testing.T) {
	metrics.InitMetrics()
	conf := &conf.Conf{}
	conf.MysqlDb.Driver = "sqlite3"
	conf.MysqlDb.Database = "./cpu_load_balance_service_test.db"
	db := env.InitMysql(conf)
	dao.SetMysqlDB(db)
	mock.AutoMigrateAll(db, conf.MysqlDb.Driver)
	mock.CleanUp(db, conf.MysqlDb.Driver)

	db.Create(&schema.Host{ID: 2, Ip: "127.1.0.2"})
	db.Create(&schema.Host{ID: 3, Ip: "127.2.0.3"})
	db.Create(&schema.Host{ID: 4, Ip: "127.3.0.4"})
	db.Create(&schema.Host{ID: 5, Ip: "127.1.0.5"})
	db.Create(&schema.Host{ID: 6, Ip: "127.1.0.6"})
	db.Create(&schema.Host{ID: 7, Ip: "127.3.0.7"})
	db.Create(&schema.Host{ID: 8, Ip: "127.3.0.8"})
	db.Create(&schema.Host{ID: 9, Ip: "127.3.0.9"})
	db.Create(&schema.Host{ID: 10, Ip: "127.3.0.10"})
	db.Create(&schema.Host{ID: 11, Ip: "127.3.0.11"})
	db.Create(&schema.Host{ID: 12, Ip: "127.3.0.12"})
	db.Create(&schema.Host{ID: 13, Ip: "127.3.0.13"})
	db.Create(&schema.Host{ID: 13, Ip: "127.3.0.14"})

	db.Create(&schema.Service{Name: "service_1"}) // Sid: 1
	db.Create(&schema.Service{Name: "service_2"}) // Sid: 2
	db.Create(&schema.Service{Name: "service_3"}) // Sid: 3
	db.Create(&schema.Service{Name: "service_4"}) // Sid: 4

	db.Create(&schema.HostService{Hid: 2, Sid: 1, Desc: "127.1.0.2 -> service_1", LoadWeight: 100})
	db.Create(&schema.HostService{Hid: 5, Sid: 1, Desc: "127.1.0.5 -> service_1", LoadWeight: 100})
	db.Create(&schema.HostService{Hid: 6, Sid: 1, Desc: "127.1.0.6 -> service_1", LoadWeight: 100})
	db.Create(&schema.HostService{Hid: 3, Sid: 2, Desc: "127.2.0.3 -> service_2", LoadWeight: 100})
	db.Create(&schema.HostService{Hid: 3, Sid: 3, Desc: "127.2.0.3 -> service_3", LoadWeight: 100})
	db.Create(&schema.HostService{Hid: 4, Sid: 2, Desc: "127.3.0.4 -> service_2", LoadWeight: 100})
	db.Create(&schema.HostService{Hid: 4, Sid: 3, Desc: "127.3.0.4 -> service_3", LoadWeight: 100})
	db.Create(&schema.HostService{Hid: 7, Sid: 1, Desc: "127.3.0.7 -> service_1", LoadWeight: 100})
	db.Create(&schema.HostService{Hid: 8, Sid: 4, Desc: "127.1.0.8 -> service_4", LoadWeight: 100})
	db.Create(&schema.HostService{Hid: 9, Sid: 4, Desc: "127.1.0.9 -> service_4", LoadWeight: 100})
	db.Create(&schema.HostService{Hid: 10, Sid: 4, Desc: "127.1.0.10 -> service_4", LoadWeight: 100})
	db.Create(&schema.HostService{Hid: 11, Sid: 4, Desc: "127.1.0.11 -> service_4", LoadWeight: 100})
	db.Create(&schema.HostService{Hid: 12, Sid: 4, Desc: "127.1.0.12 -> service_4", LoadWeight: 100})
	db.Create(&schema.HostService{Hid: 13, Sid: 4, Desc: "127.1.0.13 -> service_4", LoadWeight: 100})
	db.Create(&schema.HostService{Hid: 13, Sid: 4, Desc: "127.1.0.14 -> service_4", LoadWeight: 100})

	conf.LoadThreshold.CpuLimit = 0.04
	conf.LoadThreshold.Method = "once"
	conf.LoadThreshold.Up_Gap = 300
	conf.LoadThreshold.Down_Gap = 100

	common.GNodeInfos = []util.NodeInfo{
		util.NodeInfo{
			Host: "127.1.0.2",
			ResourceInfo: util.NodeResourceInfo{
				CoreNum: 8,
				Cpu:     float64(0.30),
			},
			StatusInfo: []util.ServiceInfo{
				util.ServiceInfo{
					ServiceName:   "service_1",
					ServiceWeight: float64(100),
				},
			},
		},
		util.NodeInfo{
			Host: "127.1.0.5",
			ResourceInfo: util.NodeResourceInfo{
				CoreNum: 8,
				Cpu:     float64(0.90),
			},
			StatusInfo: []util.ServiceInfo{
				util.ServiceInfo{
					ServiceName:   "service_1",
					ServiceWeight: float64(200),
				},
			},
		},
		util.NodeInfo{
			Host: "127.1.0.6",
			ResourceInfo: util.NodeResourceInfo{
				CoreNum: 8,
				Cpu:     float64(0.50),
			},
			StatusInfo: []util.ServiceInfo{
				util.ServiceInfo{
					ServiceName:   "service_1",
					ServiceWeight: float64(300),
				},
			},
		},
		util.NodeInfo{
			Host: "127.2.0.3",
			ResourceInfo: util.NodeResourceInfo{
				CoreNum: 8,
				Cpu:     float64(0.60),
			},
			StatusInfo: []util.ServiceInfo{
				util.ServiceInfo{
					ServiceName:   "service_2",
					ServiceWeight: float64(300),
				},
				util.ServiceInfo{
					ServiceName:   "service_3",
					ServiceWeight: float64(150),
				},
			},
		},
		util.NodeInfo{
			Host: "127.3.0.4",
			ResourceInfo: util.NodeResourceInfo{
				CoreNum: 8,
				Cpu:     float64(0.20),
			},
			StatusInfo: []util.ServiceInfo{
				util.ServiceInfo{
					ServiceName:   "service_2",
					ServiceWeight: float64(100),
				},
				util.ServiceInfo{
					ServiceName:   "service_3",
					ServiceWeight: float64(200),
				},
			},
		},
		util.NodeInfo{
			Host: "127.3.0.7",
			ResourceInfo: util.NodeResourceInfo{
				CoreNum: 8,
				Cpu:     float64(0),
			},
			StatusInfo: []util.ServiceInfo{
				util.ServiceInfo{
					ServiceName:   "service_1",
					ServiceWeight: float64(100),
				},
			},
		},
		util.NodeInfo{
			Host: "127.1.0.8",
			ResourceInfo: util.NodeResourceInfo{
				CoreNum: 8,
				Cpu:     float64(0.40),
			},
			StatusInfo: []util.ServiceInfo{
				util.ServiceInfo{
					ServiceName:   "service_4",
					ServiceWeight: float64(300),
				},
			},
		},
		util.NodeInfo{
			Host: "127.1.0.9",
			ResourceInfo: util.NodeResourceInfo{
				CoreNum: 8,
				Cpu:     float64(0.70),
			},
			StatusInfo: []util.ServiceInfo{
				util.ServiceInfo{
					ServiceName:   "service_4",
					ServiceWeight: float64(100),
				},
			},
		},
		util.NodeInfo{
			Host: "127.1.0.10",
			ResourceInfo: util.NodeResourceInfo{
				CoreNum: 8,
				Cpu:     float64(0.70),
			},
			StatusInfo: []util.ServiceInfo{
				util.ServiceInfo{
					ServiceName:   "service_4",
					ServiceWeight: float64(100),
				},
			},
		},
		util.NodeInfo{
			Host: "127.1.0.11",
			ResourceInfo: util.NodeResourceInfo{
				CoreNum: 8,
				Cpu:     float64(0.70),
			},
			StatusInfo: []util.ServiceInfo{
				util.ServiceInfo{
					ServiceName:   "service_4",
					ServiceWeight: float64(100),
				},
			},
		},
		util.NodeInfo{
			Host: "127.1.0.12",
			ResourceInfo: util.NodeResourceInfo{
				CoreNum: 8,
				Cpu:     float64(0.70),
			},
			StatusInfo: []util.ServiceInfo{
				util.ServiceInfo{
					ServiceName:   "service_4",
					ServiceWeight: float64(100),
				},
			},
		},
		util.NodeInfo{
			Host: "127.1.0.13",
			ResourceInfo: util.NodeResourceInfo{
				CoreNum: 8,
				Cpu:     float64(0.70),
			},
			StatusInfo: []util.ServiceInfo{
				util.ServiceInfo{
					ServiceName:   "service_4",
					ServiceWeight: float64(100),
				},
			},
		},
		util.NodeInfo{
			Host: "127.1.0.14",
			ResourceInfo: util.NodeResourceInfo{
				CoreNum: 8,
				Cpu:     float64(0.30),
			},
			StatusInfo: []util.ServiceInfo{
				util.ServiceInfo{
					ServiceName:   "service_4",
					ServiceWeight: float64(200),
				},
			},
		},
	}
	// case 1
	{
		common.GLoadThresholdServices = []string{}
		expectIpServices := []util.IP_Service{}
		ipServices, err := GetAllServiceList(conf)
		if err != nil || !reflect.DeepEqual(expectIpServices, ipServices) {
			t.Errorf("TestGetAllServiceList() failed, \n[Got]:\n%v,\n[Expect]:\n%v",
				ipServices, expectIpServices)
		}
	}
	// case 2
	{
		common.GLoadThresholdServices = []string{"service_1"}
		expectIpServices := []util.IP_Service{
			util.IP_Service{
				IP:             "127.1.0.2",
				Service:        "service_1",
				Service_weight: uint(189),
			},
			util.IP_Service{
				IP:             "127.1.0.5",
				Service:        "service_1",
				Service_weight: uint(126),
			},
			util.IP_Service{
				IP:             "127.1.0.6",
				Service:        "service_1",
				Service_weight: uint(300),
			},
		}
		ipServices, err := GetAllServiceList(conf)
		mock.SortIPServiceByIP(expectIpServices)
		mock.SortIPServiceByIP(ipServices)
		if err != nil || !reflect.DeepEqual(expectIpServices, ipServices) {
			t.Errorf("TestGetAllServiceList() failed, \n[Got]:\n%v,\n[Expect]:\n%v",
				ipServices, expectIpServices)
		}
	}
	// case 2
	{
		common.GLoadThresholdServices = []string{"service_1", "service_2"}
		expectIpServices := []util.IP_Service{
			util.IP_Service{
				IP:             "127.3.0.4",
				Service:        "service_2",
				Service_weight: uint(200),
			},
			util.IP_Service{
				IP:             "127.1.0.2",
				Service:        "service_1",
				Service_weight: uint(189),
			},
			util.IP_Service{
				IP:             "127.1.0.5",
				Service:        "service_1",
				Service_weight: uint(126),
			},
			util.IP_Service{
				IP:             "127.1.0.6",
				Service:        "service_1",
				Service_weight: uint(300),
			},
			util.IP_Service{
				IP:             "127.2.0.3",
				Service:        "service_2",
				Service_weight: uint(200),
			},
		}
		ipServices, err := GetAllServiceList(conf)
		mock.SortIPServiceByIP(expectIpServices)
		mock.SortIPServiceByIP(ipServices)
		if err != nil || !reflect.DeepEqual(expectIpServices, ipServices) {
			t.Errorf("TestGetAllServiceList() failed, \n[Got]:\n%v,\n[Expect]:\n%v",
				ipServices, expectIpServices)
		}
	}
	// case 3
	{
		conf.LoadThreshold.Method = ""
		common.GLoadThresholdServices = []string{"service_1", "service_2"}
		expectIpServices := []util.IP_Service{
			util.IP_Service{
				IP:             "127.3.0.4",
				Service:        "service_2",
				Service_weight: uint(101),
			},
			util.IP_Service{
				IP:             "127.1.0.2",
				Service:        "service_1",
				Service_weight: uint(101),
			},
			util.IP_Service{
				IP:             "127.1.0.5",
				Service:        "service_1",
				Service_weight: uint(199),
			},
			util.IP_Service{
				IP:             "127.1.0.6",
				Service:        "service_1",
				Service_weight: uint(300),
			},
			util.IP_Service{
				IP:             "127.2.0.3",
				Service:        "service_2",
				Service_weight: uint(299),
			},
		}
		ipServices, err := GetAllServiceList(conf)
		mock.SortIPServiceByIP(expectIpServices)
		mock.SortIPServiceByIP(ipServices)
		if err != nil || !reflect.DeepEqual(expectIpServices, ipServices) {
			t.Errorf("TestGetAllServiceList() failed, \n[Got]:\n%v,\n[Expect]:\n%v",
				ipServices, expectIpServices)
		}
	}

	{
		// case 4
		conf.LoadThreshold.Up_Gap = 1000
		conf.LoadThreshold.Method = "once"
		common.GLoadThresholdServices = []string{"service_4"}
		expectIpServices := []util.IP_Service{
			util.IP_Service{
				IP:             "127.1.0.8",
				Service:        "service_4",
				Service_weight: uint(485),
			},
			util.IP_Service{
				IP:             "127.1.0.14",
				Service:        "service_4",
				Service_weight: uint(435),
			},
			util.IP_Service{
				IP:             "127.1.0.9",
				Service:        "service_4",
				Service_weight: uint(100),
			},
			util.IP_Service{
				IP:             "127.1.0.10",
				Service:        "service_4",
				Service_weight: uint(100),
			},
			util.IP_Service{
				IP:             "127.1.0.11",
				Service:        "service_4",
				Service_weight: uint(100),
			},
			util.IP_Service{
				IP:             "127.1.0.12",
				Service:        "service_4",
				Service_weight: uint(100),
			},
			util.IP_Service{
				IP:             "127.1.0.13",
				Service:        "service_4",
				Service_weight: uint(100),
			},
		}
		ipServices, err := GetAllServiceList(conf)
		mock.SortIPServiceByIP(expectIpServices)
		mock.SortIPServiceByIP(ipServices)
		if err != nil || !reflect.DeepEqual(expectIpServices, ipServices) {
			t.Errorf("TestGetAllServiceList() failed, \n[Got]:\n%v,\n[Expect]:\n%v",
				ipServices, expectIpServices)
		}
	}
	// clean up db
	mock.CleanUp(db, conf.MysqlDb.Driver)
}
