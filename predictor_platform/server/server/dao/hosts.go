package dao

import (
	"fmt"
	"server/common"
	"server/schema"
)

func GetAllHosts() ([]schema.Host, error) {
	var hosts []schema.Host
	db := Mysql_db.Find(&hosts)
	if db.RecordNotFound() {
		return nil, fmt.Errorf("Host is empty")
	}
	errs := db.GetErrors()
	if len(errs) != 0 {
		return nil, fmt.Errorf("gorm db err: err=%v", errs)
	}
	return hosts, nil
}

// 判断机器是否存在
func ExistsHost(host_ip string) (bool, error) {
	notFound := Mysql_db.Where(schema.Host{Ip: host_ip}).Find(&schema.Host{}).RecordNotFound()
	errs := Mysql_db.GetErrors()
	if len(errs) > 0 {
		err := fmt.Errorf("gorm db err: err=%v", errs)
		return false, err
	}
	return !notFound, nil
}

// 判断机器id是否存在
func ExistsHostById(hostId uint) (bool, error) {
	notFound := Mysql_db.Where(schema.Host{ID: hostId}).Find(&schema.Host{}).RecordNotFound()
	errs := Mysql_db.GetErrors()
	if len(errs) > 0 {
		err := fmt.Errorf("gorm db err: err=%v", errs)
		return false, err
	}
	return !notFound, nil
}

// 获取机器记录
func GetElasticHost(host_ip string) (*schema.Host, error) {
	host := &schema.Host{}
	Mysql_db.Where(schema.Host{Ip: host_ip, Desc: common.STATUS_ELASTIC_EXPANSION}).Find(host)
	errs := Mysql_db.GetErrors()
	if len(errs) > 0 {
		err := fmt.Errorf("gorm db err: err=%v", errs)
		return host, err
	}
	return host, nil
}

// 获取机器记录
func GetHostByID(hid uint) (*schema.Host, error) {
	host := &schema.Host{}
	Mysql_db.Where(schema.Host{ID: hid}).Find(host)
	errs := Mysql_db.GetErrors()
	if len(errs) > 0 {
		err := fmt.Errorf("gorm db err: err=%v", errs)
		return host, err
	}
	return host, nil
}

// 批量获取机器记录
func GetHostMapByIds(hids []uint) (map[uint]*schema.Host, error) {
	hostMap := make(map[uint]*schema.Host, 0)
	hosts := []*schema.Host{}
	if len(hids) == 0 {
		return hostMap, nil
	}
	var hidsStr string
	for idx, hid := range hids {
		hidsStr = hidsStr + fmt.Sprintf("%d", hid)
		if idx != (len(hids) - 1) {
			hidsStr = hidsStr + ","
		}
	}
	sql := `SELECT * FROM hosts
		WHERE id IN (` + hidsStr + `)`
	db := Mysql_db.Raw(sql).Scan(&hosts)
	if errs := db.GetErrors(); len(errs) != 0 {
		return hostMap, fmt.Errorf("gorm db err: err=%v, sql=%s, hids=%v", errs, sql, hids)
	}
	for _, host := range hosts {
		hostMap[host.ID] = host
	}
	return hostMap, nil
}

// 插入弹性扩容机器
func InsertElasticHost(host_ip string) error {
	host := schema.Host{
		Ip:         host_ip,
		DataCenter: "",
		Desc:       common.STATUS_ELASTIC_EXPANSION,
	}
	if err := Mysql_db.Create(&host).Error; err != nil {
		return fmt.Errorf("gorm db err: err=%v", err)
	}
	return nil
}

// 获取模型所在机器
func GetHostByModel(model_name string, excludeSid uint) ([]schema.Host, error) {
	var hosts []schema.Host
	sql := `SELECT h.* FROM hosts h 
	    LEFT JOIN host_services hs ON hs.hid=h.id  
		LEFT JOIN service_models sm ON sm.sid=hs.sid 
		LEFT JOIN models m ON m.id=sm.mid 
		WHERE m.name=? AND hs.sid!=?;`
	db := Mysql_db.Raw(sql, model_name, excludeSid).Scan(&hosts)

	errs := db.GetErrors()
	if len(errs) != 0 {
		return []schema.Host{}, fmt.Errorf("gorm db err: err=%v, sql=%s, model_name=%s, excludeSid=%d", errs, sql, model_name, excludeSid)
	}
	return hosts, nil
}

func CreateHost(host *schema.Host) error {
	if errs := Mysql_db.Create(host).GetErrors(); len(errs) != 0 {
		return fmt.Errorf("insert record into table fail, err: %v", errs)
	}
	return nil
}
