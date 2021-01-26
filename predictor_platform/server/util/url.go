package util

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"server/libs/logger"
	"time"
)

// 获取指定网址的内容
func GetContentFromUrl(url_address string, timeout time.Duration, retry_cnt int) ([]byte, error) {
	var content []byte
	if retry_cnt <= 0 {
		retry_cnt = 1
	}
	http_client := http.Client{
		Timeout: timeout,
	}
	// retry http request
	var resp_content *http.Response
	var err error
	for round := 0; round < retry_cnt; round++ {
		resp_content, err = http_client.Get(url_address)
		if err != nil {
			logger.Errorf("url: %s, round: %d, resp: %v, error: %v", url_address, round, resp_content, err)
			time.Sleep(100 * time.Millisecond)
		} else {
			break
		}
	}
	if err != nil {
		return []byte{}, err
	}
	var err_ error
	if resp_content != nil {
		defer resp_content.Body.Close()
		content, err_ = ioutil.ReadAll(resp_content.Body)
		if err_ != nil {
			logger.Errorf("io read error: %v, content is: %v", err_, string(content))
			return content, err_
		}
		return content, err_
	} else {
		logger.Errorf("url: %s, retry_cnt: %d, still empty", url_address, retry_cnt)
		return []byte{}, fmt.Errorf("resp_content is nil")
	}
}

// post http url
// 请求类型为json
func HTTPPost(reqUrl string, contentType string, requestBody []byte, timeout time.Duration, retryCnt int) ([]byte, error) {
	var respData []byte
	if retryCnt <= 0 {
		retryCnt = 1
	}
	httpClient := http.Client{
		Timeout: timeout,
	}
	var respResp *http.Response
	var err error
	for round := 0; round < retryCnt; round++ {
		respResp, err = httpClient.Post(reqUrl, contentType, bytes.NewBuffer(requestBody))
		if err != nil {
			logger.Errorf("url: %s, round: %d, resp: %v, error: %v", reqUrl, round, respResp, err)
			time.Sleep(100 * time.Millisecond)
		} else {
			break
		}
	}
	if err != nil {
		return []byte{}, err
	}
	var err_ error
	if respResp != nil {
		defer respResp.Body.Close()
		respData, err_ = ioutil.ReadAll(respResp.Body)
		if err_ != nil {
			return []byte{}, err_
		}
		return respData, err_
	} else {
		return []byte{}, nil
	}
}
