package schema

import (
	"encoding/json"
	"fmt"
	"time"
)

type BaseModel struct {
	CreatedAt time.Time `gorm:"column:created_at" `
	UpdatedAt time.Time `gorm:"column:updated_at" `
}

type Host struct {
	BaseModel
	ID         uint   `gorm:"primary_key;"`
	Ip         string `gorm:"type:varchar(255);not null;unique_index"`
	DataCenter string `gorm:"type:varchar(255);"`
	Desc       string `gorm:"type:varchar(255);"`
}

func (h Host) String() string {
	return fmt.Sprintf("%v\t%v\t%v\t%v", h.ID, h.Ip, h.DataCenter, h.Desc)
}

func (h Host) MarshalJSON() ([]byte, error) {
	type Alias Host
	return json.Marshal(&struct {
		Alias
		CreatedAt string
		UpdatedAt string
	}{
		Alias:     (Alias)(h),
		CreatedAt: h.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: h.UpdatedAt.Format("2006-01-02 15:04:05"),
	})
}

func (h *Host) UnmarshalJSON(data []byte) error {
	type Alias Host
	host := &struct {
		Alias
		CreatedAt string
		UpdatedAt string
	}{
		Alias: (Alias)(*h),
	}
	var err error
	if err = json.Unmarshal(data, host); err != nil {
		return err
	}
	host.Alias.CreatedAt, err = time.Parse(`2006-01-02 15:04:05`, host.CreatedAt)
	if err != nil {
		return err
	}
	host.Alias.UpdatedAt, err = time.Parse(`2006-01-02 15:04:05`, host.UpdatedAt)
	if err != nil {
		return err
	}
	*h = Host(host.Alias)
	return nil
}

type Service struct {
	BaseModel
	ID   uint   `gorm:"primary_key;"`
	Name string `gorm:"type:varchar(255);not null;unique_index"`
	Desc string `gorm:"type:varchar(255);"`
}

func (s Service) String() string {
	return fmt.Sprintf("%v\t%v\t%v", s.ID, s.Name, s.Desc)
}

func (s Service) MarshalJSON() ([]byte, error) {
	type Alias Service
	return json.Marshal(&struct {
		Alias
		CreatedAt string
		UpdatedAt string
	}{
		Alias:     (Alias)(s),
		CreatedAt: s.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: s.UpdatedAt.Format("2006-01-02 15:04:05"),
	})
}

func (s *Service) UnmarshalJSON(data []byte) error {
	type Alias Service
	service := &struct {
		Alias
		CreatedAt string
		UpdatedAt string
	}{
		Alias: (Alias)(*s),
	}
	var err error
	if err = json.Unmarshal(data, service); err != nil {
		return err
	}
	service.Alias.CreatedAt, err = time.Parse(`2006-01-02 15:04:05`, service.CreatedAt)
	if err != nil {
		return err
	}
	service.Alias.UpdatedAt, err = time.Parse(`2006-01-02 15:04:05`, service.UpdatedAt)
	if err != nil {
		return err
	}
	*s = Service(service.Alias)
	return nil
}

type Model struct {
	BaseModel
	ID        uint   `gorm:"primary_key;"`
	Name      string `gorm:"type:varchar(255);not null;unique_index"`
	Path      string `gorm:"type:varchar(255);not null"`
	Desc      string `gorm:"type:varchar(255);"`
	Extension string `gorm:"type:text;"`
}

func (m Model) String() string {
	return fmt.Sprintf("%v\t%v\t%v\t%v", m.ID, m.Name, m.Path, m.Desc)
}

func (m Model) MarshalJSON() ([]byte, error) {
	type Alias Model
	return json.Marshal(&struct {
		Alias
		CreatedAt string
		UpdatedAt string
	}{
		Alias:     (Alias)(m),
		CreatedAt: m.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: m.UpdatedAt.Format("2006-01-02 15:04:05"),
	})
}

func (m *Model) UnmarshalJSON(data []byte) error {
	type Alias Model
	model := &struct {
		Alias
		CreatedAt string
		UpdatedAt string
	}{
		Alias: (Alias)(*m),
	}
	var err error
	if err = json.Unmarshal(data, model); err != nil {
		return err
	}
	model.Alias.CreatedAt, err = time.Parse(`2006-01-02 15:04:05`, model.CreatedAt)
	if err != nil {
		return err
	}
	model.Alias.UpdatedAt, err = time.Parse(`2006-01-02 15:04:05`, model.UpdatedAt)
	if err != nil {
		return err
	}
	*m = Model(model.Alias)
	return nil
}

type HostService struct {
	BaseModel
	ID         uint   `gorm:"primary_key;"`
	Hid        uint   `gorm:"unique_index:idx_hid_sid" sql:"type:integer not null REFERENCES hosts(id) on update cascade on delete cascade"`
	Sid        uint   `gorm:"unique_index:idx_hid_sid" sql:"type:integer not null REFERENCES services(id) on update cascade on delete cascade"`
	LoadWeight uint   `gorm:"type:integer;"`
	Desc       string `gorm:"type:varchar(255);"`
}

func (hs HostService) String() string {
	return fmt.Sprintf("%v\t%v\t%v\t%v\t%v", hs.ID, hs.Hid, hs.Sid, hs.LoadWeight, hs.Desc)
}

func (hs HostService) MarshalJSON() ([]byte, error) {
	type Alias HostService
	return json.Marshal(&struct {
		Alias
		CreatedAt string
		UpdatedAt string
	}{
		Alias:     (Alias)(hs),
		CreatedAt: hs.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: hs.UpdatedAt.Format("2006-01-02 15:04:05"),
	})
}

func (hs *HostService) UnmarshalJSON(data []byte) error {
	type Alias HostService
	hostService := &struct {
		Alias
		CreatedAt string
		UpdatedAt string
	}{
		Alias: (Alias)(*hs),
	}
	var err error
	if err = json.Unmarshal(data, hostService); err != nil {
		return err
	}
	hostService.Alias.CreatedAt, err = time.Parse(`2006-01-02 15:04:05`, hostService.CreatedAt)
	if err != nil {
		return err
	}
	hostService.Alias.UpdatedAt, err = time.Parse(`2006-01-02 15:04:05`, hostService.UpdatedAt)
	if err != nil {
		return err
	}
	*hs = HostService(hostService.Alias)
	return nil
}

type ServiceModel struct {
	BaseModel
	ID   uint   `gorm:"primary_key;"`
	Sid  uint   `gorm:"unique_index:idx_sid_mid" sql:"type:integer not null REFERENCES services(id) on update cascade on delete cascade"`
	Mid  uint   `gorm:"unique_index:idx_sid_mid" sql:"type:integer not null REFERENCES models(id) on update cascade on delete cascade"`
	Desc string `gorm:"type:varchar(255);"`
}

func (sm ServiceModel) String() string {
	return fmt.Sprintf("%v\t%v\t%v\t%v", sm.ID, sm.Sid, sm.Mid, sm.Desc)
}

func (sm ServiceModel) MarshalJSON() ([]byte, error) {
	type Alias ServiceModel
	return json.Marshal(&struct {
		Alias
		CreatedAt string
		UpdatedAt string
	}{
		Alias:     (Alias)(sm),
		CreatedAt: sm.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: sm.UpdatedAt.Format("2006-01-02 15:04:05"),
	})
}

func (sm *ServiceModel) UnmarshalJSON(data []byte) error {
	type Alias ServiceModel
	serviceModel := &struct {
		Alias
		CreatedAt string
		UpdatedAt string
	}{
		Alias: (Alias)(*sm),
	}
	var err error
	if err = json.Unmarshal(data, serviceModel); err != nil {
		return err
	}
	serviceModel.Alias.CreatedAt, err = time.Parse(`2006-01-02 15:04:05`, serviceModel.CreatedAt)
	if err != nil {
		return err
	}
	serviceModel.Alias.UpdatedAt, err = time.Parse(`2006-01-02 15:04:05`, serviceModel.UpdatedAt)
	if err != nil {
		return err
	}
	*sm = ServiceModel(serviceModel.Alias)
	return nil
}

type ModelHistory struct {
	BaseModel
	ID        uint   `gorm:"primary_key;"`
	ModelName string `gorm:"type:varchar(255);not null;unique_index:idx_mn_ts"`
	Timestamp string `gorm:"type:varchar(255);not null;unique_index:idx_mn_ts"`
	Md5       string `gorm:"type:varchar(255)"`
	IsLocked  uint   `gorm:"type:integer"`
	Desc      string `gorm:"type:varchar(255);"`
}

func (mh ModelHistory) String() string {
	return fmt.Sprintf("%v\t%v\t%v\t%v\t%v\t%v", mh.ID, mh.ModelName, mh.Timestamp, mh.Md5, mh.IsLocked, mh.Desc)
}

func (mh ModelHistory) MarshalJSON() ([]byte, error) {
	type Alias ModelHistory
	return json.Marshal(&struct {
		Alias
		CreatedAt string
		UpdatedAt string
	}{
		Alias:     (Alias)(mh),
		CreatedAt: mh.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: mh.UpdatedAt.Format("2006-01-02 15:04:05"),
	})
}

func (mh *ModelHistory) UnmarshalJSON(data []byte) error {
	type Alias ModelHistory
	modelHistory := &struct {
		Alias
		CreatedAt string
		UpdatedAt string
	}{
		Alias: (Alias)(*mh),
	}
	var err error
	if err = json.Unmarshal(data, modelHistory); err != nil {
		return err
	}
	modelHistory.Alias.CreatedAt, err = time.Parse(`2006-01-02 15:04:05`, modelHistory.CreatedAt)
	if err != nil {
		return err
	}
	modelHistory.Alias.UpdatedAt, err = time.Parse(`2006-01-02 15:04:05`, modelHistory.UpdatedAt)
	if err != nil {
		return err
	}
	*mh = ModelHistory(modelHistory.Alias)
	return nil
}

type StressInfo struct {
	BaseModel
	ID         uint   `gorm:"primary_key;"`
	Hid        uint   `sql:"type:integer not null REFERENCES hosts(id) on update cascade on delete cascade"`
	Mids       string `gorm:"type:varchar(255);"`
	Qps        string `gorm:"type:varchar(255);"`
	IsEnable   uint   `gorm:"type:integer"`
	OriginSids string `gorm:"type:varchar(64);"`
}

func (si StressInfo) String() string {
	return fmt.Sprintf("%v\t%v\t%v\t%v\t%v\t%v", si.ID, si.Hid, si.Mids, si.Qps, si.IsEnable, si.OriginSids)
}

func (si StressInfo) MarshalJSON() ([]byte, error) {
	type Alias StressInfo
	return json.Marshal(&struct {
		Alias
		CreatedAt string
		UpdatedAt string
	}{
		Alias:     (Alias)(si),
		CreatedAt: si.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: si.UpdatedAt.Format("2006-01-02 15:04:05"),
	})
}

func (si *StressInfo) UnmarshalJSON(data []byte) error {
	type Alias StressInfo
	stressInfo := &struct {
		Alias
		CreatedAt string
		UpdatedAt string
	}{
		Alias: (Alias)(*si),
	}
	var err error
	if err = json.Unmarshal(data, stressInfo); err != nil {
		return err
	}
	stressInfo.Alias.CreatedAt, err = time.Parse(`2006-01-02 15:04:05`, stressInfo.CreatedAt)
	if err != nil {
		return err
	}
	stressInfo.Alias.UpdatedAt, err = time.Parse(`2006-01-02 15:04:05`, stressInfo.UpdatedAt)
	if err != nil {
		return err
	}
	*si = StressInfo(stressInfo.Alias)
	return nil
}

type Config struct {
	BaseModel
	ID          uint   `gorm:"primary_key;"`
	Description string `gorm:"type:varchar(255);not null;unique_index"`
	Config      string `gorm:"type:text;"`
}

func (c Config) MarshalJSON() ([]byte, error) {
	type Alias Config
	return json.Marshal(&struct {
		Alias
		CreatedAt string
		UpdatedAt string
	}{
		Alias:     (Alias)(c),
		CreatedAt: c.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: c.UpdatedAt.Format("2006-01-02 15:04:05"),
	})
}

func (c *Config) UnmarshalJSON(data []byte) error {
	type Alias Config
	config := &struct {
		Alias
		CreatedAt string
		UpdatedAt string
	}{
		Alias: (Alias)(*c),
	}
	var err error
	if err = json.Unmarshal(data, config); err != nil {
		return err
	}
	config.Alias.CreatedAt, err = time.Parse(`2006-01-02 15:04:05`, config.CreatedAt)
	if err != nil {
		return err
	}
	config.Alias.UpdatedAt, err = time.Parse(`2006-01-02 15:04:05`, config.UpdatedAt)
	if err != nil {
		return err
	}
	*c = Config(config.Alias)
	return nil
}

type ServiceConfig struct {
	BaseModel
	ID          uint   `gorm:"primary_key;"`
	Description string `gorm:"type:varchar(255);not null;unique_index"`
	Sid         uint   `gorm:"unique_index:idx_cid_sid" sql:"type:integer not null REFERENCES services(id) on update cascade on delete cascade"`
	Cid         uint   `gorm: sql:"type:integer not null REFERENCES configs(id) on update cascade on delete cascade"`
}

func (sc ServiceConfig) MarshalJSON() ([]byte, error) {
	type Alias ServiceConfig
	return json.Marshal(&struct {
		Alias
		CreatedAt string
		UpdatedAt string
	}{
		Alias:     (Alias)(sc),
		CreatedAt: sc.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: sc.UpdatedAt.Format("2006-01-02 15:04:05"),
	})
}

func (sc *ServiceConfig) UnmarshalJSON(data []byte) error {
	type Alias ServiceConfig
	service_config := &struct {
		Alias
		CreatedAt string
		UpdatedAt string
	}{
		Alias: (Alias)(*sc),
	}
	var err error
	if err = json.Unmarshal(data, service_config); err != nil {
		return err
	}
	service_config.Alias.CreatedAt, err = time.Parse(`2006-01-02 15:04:05`, service_config.CreatedAt)
	if err != nil {
		return err
	}
	service_config.Alias.UpdatedAt, err = time.Parse(`2006-01-02 15:04:05`, service_config.UpdatedAt)
	if err != nil {
		return err
	}
	*sc = ServiceConfig(service_config.Alias)
	return nil
}
