package logics

import (
	"server/schema"
	"server/server/dao"
)

// 批量插入机器
func BatchInsertHostIps(hostIps []string) error {
	// 判断机器是否存在
	for _, hostIp := range hostIps {
		isExists, err := dao.ExistsHost(hostIp)
		if err != nil {
			return err
		}
		if !isExists {
			host := &schema.Host{
				Ip: hostIp,
			}
			err = dao.CreateHost(host)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
