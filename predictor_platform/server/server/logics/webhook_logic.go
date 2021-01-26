package logics

import (
	"bytes"
	"encoding/json"
	"fmt"
	"server/conf"
	"server/env"
	"server/libs/logger"
	"server/server/dao"
	"server/util"
	"strings"
	"text/template"
)

const (
	TimeIntervalTag = "timeInterval"
)

type WebHook struct {
	ReqData *util.WebHookRequest
	Tag     string
}

func NewWebHook(reqData *util.WebHookRequest, tag string) *WebHook {
	return &WebHook{
		ReqData: reqData,
		Tag:     tag,
	}
}

// 模型分发报警
func (w *WebHook) DistributeAlert() error {
	if len(w.ReqData.EvalMatches) == 0 {
		return fmt.Errorf("metrics is zero")
	}
	// 构建model->match 的映射,及获取模型列表
	modelMatchMap, alertModelList := w.parseModelMatches()
	// 获取模型对应的负责人
	modelMailsMap, err := w.getModelMailsMap(alertModelList)
	if err != nil {
		return err
	}
	// 获取模型负责人对应的模型列表
	mailModelsMap := w.getMailModelsMap(modelMailsMap)
	// 判断模型负责人
	w.constructAndSendMail(modelMatchMap, mailModelsMap)
	return nil
}

func (w *WebHook) parseModelMatches() (map[string]util.EvalMatch, []string) {
	var modelMatchMap = make(map[string]util.EvalMatch)
	var modelList = []string{}
	var possibleModelKey = []string{"model_name", "category"}

	var isFoundModelKey bool
	var modelKey string
	for _, evalMatch := range w.ReqData.EvalMatches {
		w.formateEvalMatch(&evalMatch)
		tagMap := evalMatch.Tags.(map[string]interface{})
		if !isFoundModelKey {
			for _, modelKey = range possibleModelKey {
				if _, exists := tagMap[modelKey]; exists {
					isFoundModelKey = true
					break
				}
			}
		}
		// 如果没有找到model tag，退出
		if !isFoundModelKey {
			break
		}
		modelName := tagMap[modelKey].(string)
		modelList = append(modelList, modelName)
		modelMatchMap[modelName] = evalMatch
	}
	return modelMatchMap, modelList
}

func (w *WebHook) formateEvalMatch(evalMatch *util.EvalMatch) {
	if w.Tag == TimeIntervalTag {
		valueFloat64 := evalMatch.Value.(float64)
		valueInt64 := int64(valueFloat64)
		prettyTime := util.PrettyTime(valueInt64)
		evalMatch.Value = prettyTime
	}
}

// 构造邮件，然后发送邮件
func (w *WebHook) constructAndSendMail(modelMatchMap map[string]util.EvalMatch,
	mailModelMap map[string][]string) error {
	var allMails []string
	var err error
	if _, exists := mailModelMap["all"]; exists {
		allMails, err = w.getAllModelsMailRecipients()
		if err != nil {
			return err
		}
	}

	for mail, models := range mailModelMap {
		// 针对个人发送的模型数据
		realModelMatchMap := make(map[string]util.EvalMatch)
		for _, model := range models {
			if evalMatch, exists := modelMatchMap[model]; exists {
				realModelMatchMap[model] = evalMatch
			}
		}
		var recipients []string
		var needToClaim bool
		if mail == "all" {
			w.ReqData.Title = w.ReqData.Title + "(待认领)"
			recipients = allMails
			needToClaim = true
		} else {
			recipients = []string{mail}
		}
		// todo 构造邮件body
		mailHtml := w.constructMailHtml(realModelMatchMap, needToClaim)
		// todo 发送指定负责人
		err = env.Env.Mailer.SendMail("Predictor",
			conf.GetConf().FromEmail, recipients, w.ReqData.Title, []byte(mailHtml))
		if err != nil {
			logger.Errorf("send eamil fail:%v, to: %v", err, mail)
			return err
		} else {
			logger.Infof("send mail OK!, to: %v", mail)
		}
	}
	return nil
}

// 构造邮件html
func (w *WebHook) constructMailHtml(realModelMatchMap map[string]util.EvalMatch, needToClaim bool) string {
	tplFile := "html/webhook_model_alert.tpl"
	tpl := template.Must(template.ParseFiles(tplFile))
	buf := new(bytes.Buffer)
	tpl.Execute(buf, map[string]interface{}{
		"ModelMatchMap": realModelMatchMap,
		"Title":         w.ReqData.Title,
		"ImageUrl":      w.ReqData.ImageUrl,
		"RuleUrl":       w.ReqData.RuleUrl,
		"NeedToClaim":   needToClaim,
	})
	return buf.String()
}

// 获取邮箱对应的模型列表
// @return map[email][]model
func (w *WebHook) getMailModelsMap(modelMailsMap map[string][]string) map[string][]string {
	var mailModelsMap = make(map[string][]string)
	if len(modelMailsMap) == 0 {
		return mailModelsMap
	}
	for modelName, mails := range modelMailsMap {
		if len(mails) == 0 {
			if _, exists := mailModelsMap["all"]; exists {
				mailModelsMap["all"] = append(mailModelsMap["all"], modelName)
			} else {
				mailModelsMap["all"] = []string{modelName}
			}
		} else {
			for _, mail := range mails {
				if mail != "" {
					if _, exists := mailModelsMap[mail]; exists {
						mailModelsMap[mail] = append(mailModelsMap[mail], modelName)
					} else {
						mailModelsMap[mail] = []string{modelName}
					}
				}
			}
		}
	}
	return mailModelsMap
}

// 获取模型对应的邮箱列表
// @return map[model_name][]email
func (w *WebHook) getModelMailsMap(alertModelList []string) (map[string][]string, error) {
	var modelMailsMap = make(map[string][]string)
	modelExtensionMap, err := dao.GetModelExtensionMap(alertModelList)
	if err != nil {
		return modelMailsMap, err
	}
	for modelName, extension := range modelExtensionMap {
		var mails = []string{}
		extension = strings.Trim(extension, " ")
		if extension == "" {
			modelMailsMap[modelName] = mails
			continue
		}
		var extensionInfo util.ModelExtensionInfo
		err := json.Unmarshal([]byte(extension), &extensionInfo)
		if err != nil {
			logger.Errorf("json.Unmashal err: %v", err)
			return modelMailsMap, err
		}
		if len(extensionInfo.MailRecipients) == 0 {
			modelMailsMap[modelName] = mails
			continue
		}
		for _, mail := range extensionInfo.MailRecipients {
			if mail != "" {
				mails = append(mails, mail)
			}
		}
		modelMailsMap[modelName] = mails
	}
	return modelMailsMap, nil
}

// 获取所有模型负责人
func (w *WebHook) getAllModelsMailRecipients() ([]string, error) {
	var allMails = []string{}
	modelExtensions, err := dao.GetAllModelExtensions()
	if err != nil {
		return allMails, err
	}
	for _, modelExtension := range modelExtensions {
		extension := strings.Trim(modelExtension.Extension, " ")
		if extension == "" {
			continue
		}
		var extensionInfo util.ModelExtensionInfo
		err := json.Unmarshal([]byte(extension), &extensionInfo)
		if err != nil {
			logger.Errorf("json.Unmashal err: %v", err)
			return allMails, err
		}
		if len(extensionInfo.MailRecipients) == 0 {
			continue
		}
		for _, mail := range extensionInfo.MailRecipients {
			if mail != "" && !util.IsInSliceString(mail, allMails) {
				allMails = append(allMails, mail)
			}
		}
	}
	return allMails, nil
}
