package dao

import (
	"fmt"
	"server/common"
	"server/libs/logger"
	"server/schema"
	"server/util"
	"strconv"
)

// 更新host_service表
func UpdateHostServiceData(hsid uint, hostServiceMap map[string]interface{}) error {
	errs := Mysql_db.Model(&schema.HostService{}).Where("id = ?", hsid).Update(hostServiceMap).GetErrors()
	if len(errs) != 0 {
		logger.Errorf("update table host_service error, err: %v", errs)
		return fmt.Errorf("update table host_service error, err: %v", errs)
	}
	return nil
}

// 获取host 和 service 的映射关系
// 返回 map[host]service_list
func GetHostServiceMap() (map[string][]string, error) {
	hostServiceMap := make(map[string][]string, 0)
	sql := `select h.ip as host,s.name as service_name from host_services hs 
		inner join hosts h on hs.hid=h.id  
		inner join services s on hs.sid = s.id;`
	rows, err := Mysql_db.Raw(sql).Rows()
	if err != nil {
		logger.Errorf("read db fail, sql=%s, err=%v", sql, err)
		return hostServiceMap, err
	}
	// scan rows
	for rows.Next() {
		var (
			host        string
			serviceName string
		)
		if err := rows.Scan(&host, &serviceName); err != nil {
			logger.Errorf("rows.Scan fail, err=%v", err)
			return hostServiceMap, fmt.Errorf("rows.Scan fail, err=%v", err)
		}

		if _, ok := hostServiceMap[host]; !ok {
			hostServiceMap[host] = []string{serviceName}
		} else {
			hostServiceMap[host] = append(hostServiceMap[host], serviceName)
		}
	}
	return hostServiceMap, nil
}

// 事务删除host_service和hosts表中的ip
func TranscatDeleteHost(hid uint, host_ip string) error {
	ts := Mysql_db.Begin()
	// 根据hid删除host_services中所有记录
	errs := ts.Where("hid = ?", hid).Delete(&schema.HostService{}).GetErrors()
	if len(errs) != 0 {
		ts.Rollback()
		return fmt.Errorf("gorm db err: err=%v", errs)
	}
	errs = ts.Where("ip = ?", host_ip).Delete(&schema.Host{}).GetErrors()
	if len(errs) != 0 {
		ts.Rollback()
		return fmt.Errorf("gorm db err: err=%v", errs)
	}
	ts.Commit()
	return nil
}

// 获取待分配机器列表
func GetHostsToAllocate() ([]schema.Host, error) {
	var hosts []schema.Host
	sql := `SELECT h.id,h.ip FROM hosts h 
		LEFT JOIN host_services hs ON hs.hid = h.id
		WHERE hs.hid IS NULL and h.desc=? ;`
	db := Mysql_db.Raw(sql, common.STATUS_ELASTIC_EXPANSION).Scan(&hosts)
	errs := db.GetErrors()
	if len(errs) != 0 {
		return []schema.Host{}, fmt.Errorf("gorm db err: err=%v, sql=%s, ?=%s", errs, sql, common.STATUS_ELASTIC_EXPANSION)
	}
	return hosts, nil
}

// 获取已分配机器跟service比例
func GetAllocatedHostService() ([]util.SidHostNum, error) {
	var sidHostNums []util.SidHostNum
	sql := `SELECT sid,count(hs.hid) as host_num FROM hosts h 
		LEFT JOIN host_services hs ON h.id=hs.hid 
		WHERE hs.hid IS NOT NULL AND h.desc=? 
		GROUP BY hs.sid;`
	db := Mysql_db.Raw(sql, common.STATUS_ELASTIC_EXPANSION).Scan(&sidHostNums)

	errs := db.GetErrors()
	if len(errs) != 0 {
		return []util.SidHostNum{}, fmt.Errorf("gorm db err: err=%v, sql=%s, ?=%s", errs, sql, common.STATUS_ELASTIC_EXPANSION)
	}
	return sidHostNums, nil
}

// 插入hostService记录
func InsertHostService(hostService schema.HostService) error {
	if err := Mysql_db.Create(&hostService).Error; err != nil {
		return fmt.Errorf("gorm db err: err=%v", err)
	}
	return nil
}

// 获取服务下的host_service数据
func GetHostServiceBySid(sid uint) ([]schema.HostService, error) {
	var hostServices = []schema.HostService{}
	db := Mysql_db.Where(schema.HostService{Sid: sid}).Find(&hostServices)
	if db.RecordNotFound() {
		return hostServices, nil
	}
	errs := db.GetErrors()
	if len(errs) != 0 {
		return nil, fmt.Errorf("gorm db err: err=%v", errs)
	}
	return hostServices, nil
}

// 获取服务下的hsid及ip数据
func GetHostServiceInfoBySid(sid uint) ([]util.HostServiceInfo, error) {
	hostServiceInfos := []util.HostServiceInfo{}
	sidStr := strconv.Itoa(int(sid))
	sql := `SELECT hs.id as hsid, hs.hid, h.ip, hs.sid FROM host_services hs 
		inner JOIN hosts h ON h.id=hs.hid 
		WHERE hs.sid = ` + sidStr + `;`
	db := Mysql_db.Raw(sql).Scan(&hostServiceInfos)
	if errs := db.GetErrors(); len(errs) != 0 {
		return hostServiceInfos, fmt.Errorf("gorm db err: err=%v, sql=%s", errs, sql)
	}
	return hostServiceInfos, nil
}

// 根据主键获取host_service数据
func GetHostServiceById(id uint) (schema.HostService, error) {
	var hostService = schema.HostService{}
	db := Mysql_db.Where(schema.HostService{ID: id}).First(&hostService)
	if db.RecordNotFound() {
		return hostService, nil
	}
	errs := db.GetErrors()
	if len(errs) != 0 {
		return hostService, fmt.Errorf("gorm db err: err=%v", errs)
	}
	return hostService, nil
}

// 获取原始扩容前机器跟service比例
func GetOriginHostNumMap(sids []uint) (map[uint]int, error) {
	var sidHostNumMap = make(map[uint]int)
	var sidHostNums []util.SidHostNum
	if len(sids) == 0 {
		return sidHostNumMap, fmt.Errorf("sids is empty")
	}
	var sidStr string
	for idx, sid := range sids {
		sidStr = sidStr + fmt.Sprintf("%d", sid)
		if idx != (len(sids) - 1) {
			sidStr = sidStr + ","
		}
	}
	sql := `SELECT hs.sid,count(hs.hid) AS host_num 
		FROM host_services hs 
		LEFT JOIN hosts h ON h.id=hs.hid 
		WHERE hs.hid IS NOT NULL AND h.desc!=? AND hs.sid IN (` + sidStr + `) ` +
		`GROUP BY hs.sid;`
	db := Mysql_db.Raw(sql, common.STATUS_ELASTIC_EXPANSION).Scan(&sidHostNums)

	errs := db.GetErrors()
	if len(errs) != 0 {
		return sidHostNumMap, fmt.Errorf("gorm db err: err=%v, sql=%s, ?=%s %v", errs, sql, common.STATUS_ELASTIC_EXPANSION, sids)
	}
	for _, sidHostNum := range sidHostNums {
		sidHostNumMap[sidHostNum.Sid] = sidHostNum.HostNum
	}
	return sidHostNumMap, nil
}

func GetIPWeightsBySid(sid uint) ([]util.IPWeight, error) {
	ipWeights := []util.IPWeight{}
	sql := fmt.Sprintf("select h.ip,hs.load_weight from hosts h,host_services hs where h.desc!='%s' and hs.sid=%d and h.id=hs.hid;",
		common.STATUS_ELASTIC_EXPANSION, sid)
	rows, err := Mysql_db.Raw(sql).Rows()
	if err != nil {
		logger.Errorf("read db fail, sql=%s, err=%v", sql, err)
		return ipWeights, err
	}
	// scan rows
	for rows.Next() {
		var (
			hostIp     string
			loadWeight int
		)
		if err := rows.Scan(&hostIp, &loadWeight); err != nil {
			logger.Errorf("rows.Scan fail, err=%v", err)
			continue
		}
		ipWeights = append(ipWeights, util.IPWeight{HostIp: hostIp, LoadWeight: loadWeight})
	}
	return ipWeights, nil
}

func UpdateHostService(hid uint, sid uint) error {
	// 修改host service表
	hs := &schema.HostService{
		Hid:        hid,
		Sid:        sid,
		Desc:       "",
		LoadWeight: 0,
	}
	errs := Mysql_db.Model(&schema.HostService{}).Where("hid = ?", hid).Update(hs).GetErrors()
	if len(errs) != 0 {
		logger.Errorf("updateHostServiceById id: %d, err: %+v", hid, errs)
		return fmt.Errorf("gorm db err: err=%v", errs)
	}
	return nil
}

func CreateHostService(hostService *schema.HostService) error {
	if errs := Mysql_db.Create(hostService).GetErrors(); len(errs) != 0 {
		return fmt.Errorf("insert record into host_service fail, err: %v", errs)
	}
	return nil
}

func GetHostServicesByHid(hid uint) ([]schema.HostService, error) {
	var hostServices = []schema.HostService{}
	db := Mysql_db.Where(schema.HostService{Hid: hid}).Find(&hostServices)
	if db.RecordNotFound() {
		return hostServices, nil
	}
	errs := db.GetErrors()
	if len(errs) != 0 {
		return nil, fmt.Errorf("gorm db err: err=%v", errs)
	}
	return hostServices, nil
}

// 获取当前service下的所有ip
func GetIpListBySid(sid uint) ([]string, error) {
	hosts := []string{}
	sql := `SELECT h.ip FROM host_services hs 
		INNER JOIN hosts h ON h.id = hs.hid
		WHERE hs.sid = ?`
	dbPtr := Mysql_db.Raw(sql, sid).Pluck("h.ip", &hosts)
	if errs := dbPtr.GetErrors(); len(errs) > 0 {
		logger.Errorf("gorm db err: sql=%s err=%v", sql, errs)
		return hosts, fmt.Errorf("gorm db err: sql=%s err=%v", sql, errs)
	}
	return hosts, nil
}

// 获取所有sid及ip
func GetAllIpSid() ([]util.IpSid, error) {
	ipSids := []util.IpSid{}
	sql := `SELECT h.ip,hs.sid  FROM host_services hs 
		inner JOIN hosts h ON h.id=hs.hid;`
	db := Mysql_db.Raw(sql).Scan(&ipSids)
	if errs := db.GetErrors(); len(errs) != 0 {
		return ipSids, fmt.Errorf("gorm db err: err=%v, sql=%s", errs, sql)
	}
	return ipSids, nil
}

// 通过ip列表获取ipSid列表
func GetIpSidsByIps(ipList []string) ([]util.IpSid, error) {
	ipSids := []util.IpSid{}
	if len(ipList) == 0 {
		return ipSids, nil
	}
	var ipsStr string
	for idx, ip := range ipList {
		ipsStr = ipsStr + fmt.Sprintf("'%s'", ip)
		if idx != (len(ipList) - 1) {
			ipsStr = ipsStr + ","
		}
	}
	sql := `SELECT h.ip,hs.sid  FROM host_services hs 
		inner JOIN hosts h ON h.id=hs.hid 
		WHERE h.ip in (` + ipsStr + `);`
	db := Mysql_db.Raw(sql).Scan(&ipSids)
	if errs := db.GetErrors(); len(errs) != 0 {
		return ipSids, fmt.Errorf("gorm db err: err=%v, sql=%s", errs, sql)
	}
	return ipSids, nil
}

// 获取host_services 数据
func GetAllHostServices() ([]schema.HostService, error) {
	hostServices := []schema.HostService{}
	db := Mysql_db.Find(&hostServices)
	if db.RecordNotFound() {
		return nil, fmt.Errorf("hostServices is empty")
	}
	if errs := db.GetErrors(); len(errs) != 0 {
		return nil, fmt.Errorf("gorm db err: err=%v", errs)
	}
	return hostServices, nil
}

// 通过ip列表获取host service信息列表
func GetHostServiceInfoByIps(ipList []string) ([]util.HostServiceInfo, error) {
	hostServiceInfos := []util.HostServiceInfo{}
	if len(ipList) == 0 {
		return hostServiceInfos, nil
	}
	var ipsStr string
	for idx, ip := range ipList {
		ipsStr = ipsStr + fmt.Sprintf("'%s'", ip)
		if idx != (len(ipList) - 1) {
			ipsStr = ipsStr + ","
		}
	}
	sql := `SELECT hs.id as hsid, hs.hid, h.ip, hs.sid FROM host_services hs 
		inner JOIN hosts h ON h.id=hs.hid 
		WHERE h.ip in (` + ipsStr + `);`
	db := Mysql_db.Raw(sql).Scan(&hostServiceInfos)
	if errs := db.GetErrors(); len(errs) != 0 {
		return hostServiceInfos, fmt.Errorf("gorm db err: err=%v, sql=%s", errs, sql)
	}
	return hostServiceInfos, nil
}

// 事务迁移host_service 中的机器
func TranscatMigrateHost(fromSids []uint, toSids []uint, toWeights []uint, hids []uint) error {
	ts := Mysql_db.Begin()
	var hidsStr string
	for idx, hid := range hids {
		hidsStr += strconv.Itoa(int(hid))
		if idx != (len(hids) - 1) {
			hidsStr += ","
		}
	}
	// 根据hid删除host_services中fromSid所有记录
	errs := ts.Where(fmt.Sprintf("hid in (%s)", hidsStr)).Delete(&schema.HostService{}).GetErrors()
	if len(errs) != 0 {
		ts.Rollback()
		return fmt.Errorf("gorm db err: err=%v", errs)
	}
	if len(toWeights) != len(toSids) {
		ts.Rollback()
		return fmt.Errorf("the length of toWeights and toSids is not equal")
	}
	// 获取toSid下的数据，避免insert duplicate
	var sidsStr string
	for idx, sid := range toSids {
		sidsStr += strconv.Itoa(int(sid))
		if idx != (len(toSids) - 1) {
			sidsStr += ","
		}
	}
	var toHostServices = []schema.HostService{}
	db := ts.Where("sid in (" + sidsStr + ")").Find(&toHostServices)
	if errs := db.GetErrors(); len(errs) != 0 {
		ts.Rollback()
		return fmt.Errorf("gorm db err: err=%v", errs)
	}
	hidSidMap := make(map[string]struct{})
	for _, toHostService := range toHostServices {
		uniqKey := fmt.Sprintf("%d_%d", toHostService.Hid, toHostService.Sid)
		hidSidMap[uniqKey] = struct{}{}
	}
	// 插入数据
	var hostService = &schema.HostService{}
	for _, hid := range hids {
		for idx, toSid := range toSids {
			if _, exists := hidSidMap[fmt.Sprintf("%d_%d", hid, toSid)]; exists {
				continue
			}
			hostService = &schema.HostService{
				Hid:        hid,
				Sid:        toSid,
				LoadWeight: toWeights[idx],
			}
			if err := ts.Create(hostService).Error; err != nil {
				ts.Rollback()
				return fmt.Errorf("gorm db err: err=%v", err)
			}
			if len(errs) != 0 {
				ts.Rollback()
				return fmt.Errorf("gorm db err: err=%v", errs)
			}
		}
	}
	ts.Commit()
	return nil
}
