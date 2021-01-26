package metrics

// 注册 metrics
// 注：server中在加入metrics之前需要先注册一个metrics(meter or timer or guage)，然后再到server对应位置mark
func registerMetrics() {
	metrics := GetMetrics()
	// -------- register meter metrics ---------------

	meters[TAG_ELASTIC_EXPANSION_INSERT] = metrics.Tagged(map[string]string{"action_name": TAG_INSERT}).
		Meter(ELASTIC_EXPANSION_SERVER)
	meters[TAG_ELASTIC_EXPANSION_DELETE] = metrics.Tagged(map[string]string{"action_name": TAG_DELETE}).
		Meter(ELASTIC_EXPANSION_SERVER)
	meters[TAG_ELASTIC_EXPANSION_INSERT_ERROR] = metrics.Tagged(map[string]string{"error_name": TAG_INSERT_ERROR}).
		Meter(ELASTIC_EXPANSION_SERVER_ERROR)
	meters[TAG_ELASTIC_EXPANSION_DELETE_ERROR] = metrics.Tagged(map[string]string{"error_name": TAG_DELETE_ERROR}).
		Meter(ELASTIC_EXPANSION_SERVER_ERROR)
	meters[ELASTIC_EXPANSION_ALLOCATE] = metrics.Meter(ELASTIC_EXPANSION_ALLOCATE)
	meters[TAG_GET_HOSTS_TO_ALLOCATE_ERROR] = metrics.Tagged(map[string]string{"error_name": TAG_GET_HOSTS_TO_ALLOCATE_ERROR}).
		Meter(ELASTIC_EXPANSION_ERROR)
	meters[TAG_GET_ALLOCATED_HOST_SERVICE_ERROR] = metrics.Tagged(map[string]string{"error_name": TAG_GET_ALLOCATED_HOST_SERVICE_ERROR}).
		Meter(ELASTIC_EXPANSION_ERROR)
	meters[TAG_GET_All_SERVICE_MAP_ERROR] = metrics.Tagged(map[string]string{"error_name": TAG_GET_All_SERVICE_MAP_ERROR}).
		Meter(ELASTIC_EXPANSION_ERROR)
	meters[TAG_GET_ALLOCATE_CONFIG_ERROR] = metrics.Tagged(map[string]string{"error_name": TAG_GET_ALLOCATE_CONFIG_ERROR}).
		Meter(ELASTIC_EXPANSION_ERROR)
	meters[TAG_ALLOCATE_HOST_TO_SERVICE_ERROR] = metrics.Tagged(map[string]string{"error_name": TAG_ALLOCATE_HOST_TO_SERVICE_ERROR}).
		Meter(ELASTIC_EXPANSION_ERROR)

	meters[TAG_GET_DB_HOST_SERVICE_ERROR] = metrics.Tagged(map[string]string{"error_name": TAG_GET_DB_HOST_SERVICE_ERROR}).
		Meter(TAG_MODEL_LOAD_DIFF_ERROR)
	meters[TAG_GET_DB_MODEL_ERROR] = metrics.Tagged(map[string]string{"error_name": TAG_GET_DB_MODEL_ERROR}).
		Meter(TAG_MODEL_LOAD_DIFF_ERROR)
	meters[TAG_PARSE_TIMESTAMP_ERROR] = metrics.Tagged(map[string]string{"error_name": TAG_PARSE_TIMESTAMP_ERROR}).
		Meter(METER_MONITOR_ERROR)

	// -------- register timer metrics ---------------

	// -------- register gauge metrics ---------------
	servingTag := map[string]string{"status": "on"}
	gauges[GAUGE_SERVING] = metrics.Tagged(servingTag).Gauge(GAUGE_SERVING)

}
