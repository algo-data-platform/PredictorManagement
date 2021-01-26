package server

import (
	"sync"
	"content_service/libs/logger"
  "content_service/env"
  "fmt"
  "os/exec"
  "content_service/schema"
  "time"
  "path"
  "bufio"
  "io"
  "os"
  "strings"
)

type RegressionService struct {
}

var regressionInstance *RegressionService
var regressionOnce sync.Once

func GetRegressionInstance() *RegressionService {
  regressionOnce.Do(func() {
    regressionInstance = &RegressionService{}
  })
  return regressionInstance
}

func (service *RegressionService) Run(env *env.Env, packet_name string, mode string) (bool, error) {
  logger.Infof("LocalIp=%v: RegressionService start...", env.LocalIp)
  // 拉取待测试的algo_service部署包
  script := path.Join(env.Conf.RegressionService.ScriptPath, "regression.sh")
  cmd := exec.Command("sh", script, packet_name, mode)
  stdout, err := cmd.CombinedOutput()
  if err != nil {
    return false, fmt.Errorf("sh regression.sh failed, err : %v, stdout: %v, script: %v", err, string(stdout), script)
  }
  
  // 关联数据库,包括增加host表，增加service表，增加host service表
  db := env.Db
  var host = &schema.Host{}
	dbPtr := db.Where(schema.Host{Ip: env.LocalIp}).Find(host)
	if dbPtr.RecordNotFound() {
		host = &schema.Host{
      Ip:         env.LocalIp,
      DataCenter: "",
    }
    if err := db.Create(host).Error; err != nil {
      return false, fmt.Errorf("gorm db err: err=%v", err)
    }
  }
  
  var services = &schema.Service{}
	dbPtr = db.Where(schema.Service{Name: env.Conf.ValidateService.ServiceName}).Find(services)
	if dbPtr.RecordNotFound() {
    return false, nil
  }
  
  errs := dbPtr.GetErrors()
	if len(errs) != 0 {
		logger.Errorf("get algo services consistence failed, err: %v", errs)
		return false, fmt.Errorf("gorm db err: err=%v", errs)
	}

  var host_service = &schema.HostService{}
	dbPtr = db.Where(schema.HostService{Hid:host.ID, Sid:services.ID}).Find(host_service)
	if dbPtr.RecordNotFound() {
		host_service = &schema.HostService{
      Hid:         host.ID,
      Sid: services.ID,
    }
    if err := db.Create(host_service).Error; err != nil {
      return false, fmt.Errorf("gorm db err: err=%v", err)
    }
  }

  // algo service刚启动，http服务还没启动，所以需要睡眠一段时间，保证后面的http操作都能正常请求。
  time.Sleep(time.Duration(env.Conf.RegressionService.SleepTime)*time.Second)
  // 拉取模型操作，选取已经valid的模型
  t1 := time.Now()
  p2p_model_service := NewP2PModelService()
  err = p2p_model_service.check(env)
  if err != nil {
    return false, fmt.Errorf("p2p_model_service check error, %v", err);
  }
  elapsed := time.Since(t1)
  logger.Infof("pull model time is: %v" , elapsed)
  
  date := time.Now().AddDate(0, 0, 0).Format("2006-01-02")
  validate_service := GetValidateInstance()
	validate_service.GenerateAndSendDailyReport(path.Join(env.Conf.ValidateService.SummaryResultDir, "summary_result_" + date), date, env)
  // 解析邮件内容，判断所有的测试用例都通过
  summary_file := path.Join(env.Conf.ValidateService.SummaryResultDir, "summary_result_" + date + "_" + env.Conf.RegressionService.PacketName)
  logger.Infof("regression open summary result file: " + summary_file)
  fp, err_file := os.Open(summary_file)
  if err_file != nil {
    return false, fmt.Errorf("regression open summary result file failed!");
  }
  defer fp.Close()
  r := bufio.NewReader(fp)
  var values []string
  for {
    buf, err := r.ReadString('\n')
    if err == io.EOF || err != nil { //-1
        break
    }
    buf = strings.TrimSpace(buf)
    if(buf == "") {
      continue
    }
    values = strings.Split(buf, "\t")
    if len(values) != 5 {
      continue
    }
    if values[3] != "通过" {
      return false, fmt.Errorf("test case is not valid");
    }
  }
  
  return true, nil;
}