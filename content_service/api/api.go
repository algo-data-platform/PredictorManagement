package api

type PredictorModelRecord struct {
  Name       string   `json:"name"`
  Timestamp  string   `json:"timestamp"`
  FullName   string   `json:"fullName"`
  ConfigName string   `json:"configName"`
  IsLocked   uint     `json:"is_locked"`
  Md5        string   `json:"md5"`
}

type PredictorService struct {
  ServiceName string `json:"service_name"`
  ServiceWeight int `json:"service_weight"`
  ServiceConfig map[string]interface{} `json:"service_config"`
  ModelRecords []PredictorModelRecord `json:"model_records"`
}

type PredictorPayload struct {
  Services []PredictorService `json:"services"`
}

//data schema for check model load status reply from predictor
type ModelRecord struct {
  SuccessTime string `json:"success_time"`
  FullName string `json:"fullName"`
  ConfigName string `json:"configName"`
  Name string `json:"name"`
  Timestamp string `json:"timestamp"`
  Md5 string `json:"md5"`
  State string `json:state`
}
type Service struct {
  ModelRecords []ModelRecord `json:"model_records"`
  ServiceName string `json:"service_name"`
}
type MsgInfo struct {
  Services []Service `json:"services"`
}
type ModelServiceInfo struct {
  Msg MsgInfo `json:"msg"`
  Code int64 `json:"code"`
}

// data schema for model feature info from feature service
type DimInfo struct {
  Name string `json:"Name"`
}
type FeatureInfo struct {
  Name string `json:"Name"`
  Type int64 `json:"Type"`
  Dim DimInfo `json:"Dim"`
}
type FeatureItem struct {
  Feature FeatureInfo `json:"Feature"`
}
type Data struct {
  Name string `json:"Name"`
  Features []FeatureItem `json:"Features"`
}
type ModelFeatureInfo struct {
  Data Data `json:"Data"`
}

type ModelFeature struct {
  Type int64
  Dim string
}

type ModelExtensionInfo struct {
  MailRecipients []string `json:MailRecipients`
}

type PredictorStressInfoPayload struct {
  ModelNames      string   `json:"model_names"`
  Qps             string   `json:"qps"`
  Service         string   `json:"service"`
}
