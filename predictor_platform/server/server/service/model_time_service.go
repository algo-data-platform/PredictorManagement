package service

import (
	"bytes"
	"github.com/robfig/cron"
	"server/common"
	"server/conf"
	"server/env"
	"server/libs/logger"
	"server/schema"
	"server/server/dao"
	"server/server/logics"
	"sort"
	"text/template"
	"time"
)

func ModelTimeMail(conf *conf.Conf) {
	cron_tab := cron.New()
	cron_tab.AddFunc("00 30 09,17 * *", func() {
		logger.Infof("sending email time: %v", time.Now().Format("2006-01-02 15:04:05"))
		EmailCrontabJob(conf)
	})
	cron_tab.Start()
}

// 获取需要邮件html展示的数据源
func getModelTimingInfoData(conf *conf.Conf) string {
	var model_info_list []schema.Model
	var err error
	model_info_list, err = dao.GetAllModels()
	if err != nil {
		logger.Errorf("GetAllModels fail, err: %v", err)
	}
	// sort by model_name
	sort.Sort(common.ModelInfoWrapper{model_info_list, func(p, q *schema.Model) bool {
		//return (p.Desc < q.Desc)
		if p.Desc < q.Desc {
			return true
		}
		if p.Desc > q.Desc {
			return false
		}
		return p.Name < q.Name
	}})
	var modelList []string
	for index := 0; index < len(model_info_list); index++ {
		modelList = append(modelList, model_info_list[index].Name)
	}
	_, html_data := logics.ModelListUpdateTimeWithinWeek(modelList, conf)
	// go html/template
	html_path := conf.HtmlTemplatePath
	tpl := template.Must(template.ParseFiles(html_path))
	buf := new(bytes.Buffer)
	tpl.Execute(buf, html_data)
	html_str := buf.String()
	return html_str
}

// 发送邮件的crontab job
func EmailCrontabJob(conf *conf.Conf) {
	html_data := getModelTimingInfoData(conf)
	err := env.Env.Mailer.SendMail("Predictor", conf.FromEmail, conf.Recipients, "模型时效性统计报告", []byte(html_data))
	if err != nil {
		logger.Errorf("send eamil fail:%v", err)
	} else {
		logger.Infof("send mail OK!, conf.Recipients: %v", conf.Recipients)
	}
}
