package prometheus

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type Prom struct {
	Address string
}

type PrometheusResponse struct {
	Status string     `json:"status"` //  "success" | "error",
	Data   ResultData `json:"data"`   //  <data>,
	// Only set if status is "error". The data field may still hold
	// additional data.
	ErrorType string `json:"errorType"`
	Error     string `json:"error"`
}

type ResultData struct {
	ResultType string        `json:"resultType"` // "matrix" | "vector" | "scalar" | "string"
	Result     []interface{} `json:"result"`
}

type VectorResult struct {
	Metric map[string]string `json:"metric"` // { "<label_name>": "<label_value>", ... },
	Value  []interface{}     `json:"value"`  //  [ <unix_time>, "<sample_value>" ]
}

// 初始化prometheus
func New(address string) *Prom {
	return &Prom{
		Address: address,
	}
}

// 瞬时请求
func (pm *Prom) Query(pmsql string, timeout time.Duration) (*PrometheusResponse, error) {
	var pmResp = &PrometheusResponse{}
	// 构建请求
	baseUrl, err := url.Parse(pm.Address)
	if err != nil {
		return nil, fmt.Errorf("address is not valid %v", err)
	}
	baseUrl.Path += "/api/v1/query"
	params := url.Values{}
	params.Add("query", pmsql)
	baseUrl.RawQuery = params.Encode()
	resp, err := pm.httpGet(baseUrl.String(), timeout)
	if err != nil {
		return nil, fmt.Errorf("httpGet url fail, err: %v, url: %s", err, baseUrl.String())
	}
	err = json.Unmarshal(resp, pmResp)
	if err != nil {
		return nil, fmt.Errorf("unmarshal fail, err: %v, resp: %s", err, string(resp))
	}
	return pmResp, nil
}

func (pm *Prom) httpGet(url_address string, timeout time.Duration) ([]byte, error) {
	var content []byte
	http_client := http.Client{
		Timeout: timeout,
	}
	var resp_content *http.Response
	var err error
	resp_content, err = http_client.Get(url_address)
	if err != nil {
		return []byte{}, err
	}

	var err_ error
	if resp_content != nil {
		defer resp_content.Body.Close()
		content, err_ = ioutil.ReadAll(resp_content.Body)
		if err_ != nil {
			return content, err_
		}
		return content, err_
	} else {
		return []byte{}, fmt.Errorf("resp_content is nil")
	}
}
