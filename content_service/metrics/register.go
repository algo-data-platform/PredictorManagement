package metrics

// 注册 metrics
// 注：server中在加入metrics 之前需要先注册一个metrics(meter or timer or guage)，然后再到server对应位置mark
func registerMetrics() {
	metrics := GetMetrics()
	// -------- register meter metrics ---------------
	// error meter 可通过 GetErrorMeter 方法获得，动态注册。
	// 建议其他简单meter要放到下面注册，清晰且不易出错
	// 如果需要第一次错误展示并报警，需要初始化
	GetErrorMeter(TAG_P2P_MODEL_SERVICE, TAG_CHECK_ERROR)
	GetErrorMeter(TAG_P2P_MODEL_SERVICE, TAG_FETCHDISK_ERROR)
	GetErrorMeter(TAG_P2P_MODEL_SERVICE, TAG_NOTIFY_ERROR)
	GetErrorMeter(TAG_P2P_MODEL_SERVICE, TAG_GETWEIGHT_ERROR)
	GetErrorMeter(TAG_P2P_MODEL_SERVICE, TAG_FETCH_AND_PULL_SERVICE_MODELS_ERROR)
	GetErrorMeter(TAG_P2P_MODEL_SERVICE, TAG_FETCH_DB_SERVICES_ERROR)
	GetErrorMeter(TAG_CLEANING_SERVICE, TAG_CHECK_ERROR)
	GetErrorMeter(TAG_HDFS_SERVICE, TAG_CHECK_ERROR)
	GetErrorMeter(TAG_HDFS_SERVICE, TAG_FETCHDB_ERROR)
	GetErrorMeter(TAG_HDFS_SERVICE, TAG_FETCH_HDFS_AND_UPDATE_STATUS_ERROR)

	// -------- register timer metrics ---------------
	// model_service timer
	timers[TIMER_P2P_MODEL_SERVICE_CHECK_TIMER] = metrics.Timer(TIMER_P2P_MODEL_SERVICE_CHECK_TIMER)
	timers[TIMER_P2P_MODEL_SERVICE_PULLMODEL_TIMER] = metrics.Timer(TIMER_P2P_MODEL_SERVICE_PULLMODEL_TIMER)
	timers[TIMER_HDFS_SERVICE_CHECK_TIMER] = metrics.Timer(TIMER_HDFS_SERVICE_CHECK_TIMER)

	// cleaning_service timer
	timers[TIMER_CLEANING_SERVICE_CHECK_TIMER] = metrics.Timer(TIMER_CLEANING_SERVICE_CHECK_TIMER)
}
