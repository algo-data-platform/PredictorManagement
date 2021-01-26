package server

import (
	"content_service/env"
	"content_service/libs/logger"
	"content_service/metrics"
)

func Start(env *env.Env) {
	// init metrics before http service
	metrics.InitMetrics()

	// always start http service
	http_service := NewHttpService()
	go http_service.Run(env)

	// start services specified in conf
	for _, name := range env.Conf.Services {
		switch name {
		case "p2pModelService":
			p2p_model_service := NewP2PModelService()
			go p2p_model_service.Run(env)
		case "cleaningService":
			cleaning_service := NewCleaningService()
			go cleaning_service.Run(env)
		case "validateService":
			if env.LocalIp == env.Conf.ValidateService.Host {
				validate_service := GetValidateInstance()
				go validate_service.Start(env)
			}
		case "hdfsService":
			if env.LocalIp == env.Conf.HdfsService.Host {
				hdfs_service := NewHdfsService()
				go hdfs_service.Run(env)
			}
		case "fileSyncService":
			file_sync_service := NewFileSyncService()
			go file_sync_service.Run(env)
		default:
			logger.Infof("unknown service name: %s, not starting it.", name)
		}
	}

	// prevent main from exiting
	select {}
}
