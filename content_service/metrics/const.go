package metrics

const (
	PREFIX       = "ad_core_content_service"
	PROJECT_NAME = "content_service"

	// meter name
	METER_CONSUMING    = "consuming"
	METER_SERVER_ERROR = "server_error"

	// gauge name
	GAUGE_MODEL_PULL_ERR_TIMES = "model_pull_err_times"

	// timer name
	TIMER_CLEANING_SERVICE_CHECK_TIMER      = "cleaning_service_check_timer"
	TIMER_PULL_SINGLE_MODEL_TIMER           = "pull_single_model_timer"
	TIMER_P2P_MODEL_SERVICE_CHECK_TIMER     = "p2p_model_service_check_timer"
	TIMER_P2P_MODEL_SERVICE_PULLMODEL_TIMER = "p2p_model_service_pullmodel_timer"
	TIMER_HDFS_SERVICE_CHECK_TIMER          = "hdfs_service_check_timer"
	TIMER_PULL_SINGLE_HDFS_MODEL_TIMER      = "pull_single_hdfs_model_timer"

	// tag name
	TAG_P2P_MODEL_SERVICE                   = "p2p_model_service"
	TAG_CLEANING_SERVICE                    = "cleaning_service"
	TAG_HDFS_SERVICE                        = "hdfs_service"
	TAG_FILE_SYNC_SERVICE                   = "file_sync_service"
	TAG_CHECK_ERROR                         = "check_error"
	TAG_FILE_SYNC_ERROR                     = "file_sync_error"
	TAG_FETCHDB_ERROR                       = "fetchdb_error"
	TAG_FETCHDISK_ERROR                     = "fetchdisk_error"
	TAG_PULLMODEL_ERROR                     = "pullmodel_error"
	TAG_NOTIFY_ERROR                        = "notify_error"
	TAG_GETWEIGHT_ERROR                     = "getweight_error"
	TAG_FETCH_AND_PULL_SERVICE_MODELS_ERROR = "fetch_and_pull_service_models_error"
	TAG_FETCH_DB_SERVICES_ERROR             = "fetch_db_services_error"
	TAG_FETCH_HDFS_AND_UPDATE_STATUS_ERROR  = "fetch_hdfs_and_update_status_error"

	TAG_STRESS_SERVICE    			= "stress_service"
	TAG_NOTIFY_STRESS_ERROR 		= "notify_stress_error"
	TAG_SET_PREDICTOR_WORK_MODE_ERROR       = "set_predictor_work_mode_error"
	TAG_REGISTER_PREDICTOR_SERVICE_ERROR    = "register_predictor_service_error"
	TAG_GET_PARENT_IP_ERROR                 = "get_parent_ip_error"
	TAG_PULL_PREDICTOR_STATIC_LIST_ERROR    = "pull_predictor_static_list_error"
	TAG_FETCH_DB_GLOBAL_MODEL_SERVICE_MAP_ERROR    = "fetch_db_global_model_service_map_error"
	TAG_UPDATE_GLOBAL_MODEL_SERVICE_MAP_ERROR    = "update_global_model_service_map_error"
	TAG_GETCONFIG_ERROR                     = "getconfig_error"
)
