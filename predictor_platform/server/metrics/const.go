package metrics

const (
	PREFIX       = "ad_core_algo_platform"
	PROJECT_NAME = "algo_platform"

	// meter name
	METER_MONITOR_ERROR            = "monitor_error"
	ELASTIC_EXPANSION_ERROR        = "elastic_expansion_error"
	ELASTIC_EXPANSION_SERVER       = "elastic_expansion_server"
	ELASTIC_EXPANSION_SERVER_ERROR = "elastic_expansion_server_error"
	ELASTIC_EXPANSION_ALLOCATE     = "elastic_expansion_allocate"
	METER_LOAD_CHANGE              = "load_change"

	// timer name

	// gauge name
	GAUGE_SERVING                = "serving"
	GAUGE_MODEL_VERSION_INTERVAL = "model_version_interval"
	GAUGE_MODEL_DIFF             = "model_diff"
	GAUGE_MODEL_SIZE             = "model_size"

	// tag name
	TAG_PARSE_TIMESTAMP_ERROR            = "parse_timestamp_error"
	TAG_GET_DB_MODEL_ERROR               = "get_db_model_error"
	TAG_GET_DB_HOST_SERVICE_ERROR        = "get_db_host_service_error"
	TAG_MODEL_LOAD_DIFF_ERROR            = "model_load_diff_error"
	TAG_SERVICE_LOAD_DIFF                = "service_load_diff"
	TAG_MODEL_LOAD_DIFF                  = "model_load_diff"
	TAG_GET_HOSTS_TO_ALLOCATE_ERROR      = "get_hosts_to_allocate_error"
	TAG_GET_ALLOCATED_HOST_SERVICE_ERROR = "get_allocated_host_service_error"
	TAG_GET_All_SERVICE_MAP_ERROR        = "get_all_service_map_error"
	TAG_GET_ALLOCATE_CONFIG_ERROR        = "get_allocate_config_error"
	TAG_ALLOCATE_HOST_TO_SERVICE_ERROR   = "allocate_host_to_service_error"
	TAG_ELASTIC_EXPANSION_INSERT         = "elastic_expansion_insert"
	TAG_ELASTIC_EXPANSION_INSERT_ERROR   = "elastic_expansion_insert_error"
	TAG_ELASTIC_EXPANSION_DELETE         = "elastic_expansion_delete"
	TAG_ELASTIC_EXPANSION_DELETE_ERROR   = "elastic_expansion_delete_error"
	TAG_INSERT                           = "insert"
	TAG_INSERT_ERROR                     = "insert_error"
	TAG_DELETE                           = "delete"
	TAG_DELETE_ERROR                     = "delete_error"
	TAG_ELASTIC_EXPANSION_ALLOCATE       = "elastic_expansion_allocate"
)
