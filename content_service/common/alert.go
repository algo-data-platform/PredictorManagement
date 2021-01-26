package common

import (
  "content_service/libs/logger"
  "time"
  "bytes"
  "encoding/json"
  "io/ioutil"
  "net/http"
)

var cache map[uint32]time.Time

func ShouldTrigger(str string, rate int) bool {
  cur_time := time.Now();
  str_hash := Hash(str)

  if cache == nil {
    cache = make(map[uint32]time.Time)  // init map
  }

  last_time, found := cache[str_hash];
  if !found {
    // not found, store and return true
    cache[str_hash] = cur_time
    return true
  }
  // found, check rate
  if cur_time.Sub(last_time).Seconds() <= float64(rate) {
    // triggered within "rate" seconds, not trigger again
    return false
  }

  // update timestamp
  cache[str_hash] = cur_time
  return true
}

func Alert(recipients []string, subject string, msg string, rate int) {
  if !ShouldTrigger(msg, rate) {
    return
  }

  auth := NewLoginAuth("your_username", "your_password")
  msg_byte := []byte(msg)
  err := SendMail("github.com:888", auth, "your_username@github.com", recipients, subject, msg_byte)
  if err != nil {
    logger.Errorf("err in sending alert email: %v", err)
  }
}

func DingDingAlert(webhook_url string, s string) {
    content, data := make(map[string]string), make(map[string]interface{})
    content["content"] = s 
    data["msgtype"] = "text"
    data["text"] = content
    b, _ := json.Marshal(data)

    resp, err := http.Post(webhook_url,
        "application/json",
        bytes.NewBuffer(b))
    if err != nil {
        logger.Errorf("err in sending DingDing alert: %v", err)
    }   
    defer resp.Body.Close()
    body, _ := ioutil.ReadAll(resp.Body)
    logger.Infof(string(body))
}
