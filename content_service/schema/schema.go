package schema

import (
	"content_service/api"
	"content_service/common"
	"content_service/env"
	"fmt"
	"path"
	"time"
)

type Host struct {
	ID         uint   `gorm:"primary_key;"`
	Ip         string `gorm:"type:varchar(255);not null;unique_index"`
	DataCenter string `gorm:"type:varchar(255);"`
	Desc       string `gorm:"type:varchar(255);"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (h Host) String() string {
	return fmt.Sprintf("%v\t%v\t%v\t%v", h.ID, h.Ip, h.DataCenter, h.Desc)
}

type Service struct {
	ID        uint   `gorm:"primary_key;"`
	Name      string `gorm:"type:varchar(255);not null;unique_index"`
	Desc      string `gorm:"type:varchar(255);"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (s Service) String() string {
	return fmt.Sprintf("%v\t%v\t%v", s.ID, s.Name, s.Desc)
}

type Model struct {
	ID        uint   `gorm:"primary_key;"`
	Name      string `gorm:"type:varchar(255);not null;unique_index"`
	Path      string `gorm:"type:varchar(255);not null"`
	Desc      string `gorm:"type:varchar(255);"`
	Extension string `gorm:"type:text;"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (m Model) String() string {
	return fmt.Sprintf("%v\t%v\t%v\t%v", m.ID, m.Name, m.Path, m.Desc)
}

type HostService struct {
	ID         uint   `gorm:"primary_key;"`
	Hid        uint   `gorm:"unique_index:idx_hid_sid" sql:"type:integer not null REFERENCES hosts(id) on update cascade on delete cascade"`
	Sid        uint   `gorm:"unique_index:idx_hid_sid" sql:"type:integer not null REFERENCES services(id) on update cascade on delete cascade"`
	LoadWeight uint   `gorm:"type:integer;"`
	Desc       string `gorm:"type:varchar(255);"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (hs HostService) String() string {
	return fmt.Sprintf("%v\t%v\t%v\t%v\t%v", hs.ID, hs.Hid, hs.Sid, hs.LoadWeight, hs.Desc)
}

type ServiceModel struct {
	ID        uint   `gorm:"primary_key;"`
	Sid       uint   `gorm:"unique_index:idx_sid_mid" sql:"type:integer not null REFERENCES services(id) on update cascade on delete cascade"`
	Mid       uint   `gorm:"unique_index:idx_sid_mid" sql:"type:integer not null REFERENCES models(id) on update cascade on delete cascade"`
	Desc      string `gorm:"type:varchar(255);"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (sm ServiceModel) String() string {
	return fmt.Sprintf("%v\t%v\t%v\t%v", sm.ID, sm.Sid, sm.Mid, sm.Desc)
}

type ModelHistory struct {
	ID        uint   `gorm:"primary_key;"`
	ModelName string `gorm:"type:varchar(255);not null;unique_index:idx_mn_ts"`
	Timestamp string `gorm:"type:varchar(255);not null;unique_index:idx_mn_ts"`
	Md5       string `gorm:"type:varchar(255)"`
	IsLocked  uint   `gorm:"type:integer`
	Desc      string `gorm:"type:varchar(255);"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (mh ModelHistory) String() string {
	return fmt.Sprintf("%v\t%v\t%v\t%v\t%v\t%v", mh.ID, mh.ModelName, mh.Timestamp, mh.Md5, mh.IsLocked, mh.Desc)
}
func (mh ModelHistory) FullName() string {
	if mh.ModelName == "" || mh.Timestamp == "" {
		return ""
	}
	return fmt.Sprintf("%v%v%v", mh.ModelName, common.Delimiter, mh.Timestamp)
}
func (mh ModelHistory) ToPredictorModelRecord(env *env.Env) api.PredictorModelRecord {
	return api.PredictorModelRecord{
		Name:       mh.ModelName,
		Timestamp:  mh.Timestamp,
		FullName:   mh.FullName(),
		ConfigName: path.Join(env.Conf.P2PModelService.DestPath, mh.FullName(), mh.ModelName+common.ConfigSuffix),
		IsLocked:   mh.IsLocked,
		Md5:        mh.Md5}
}

type StressInfo struct {
  ID         uint   `gorm:"primary_key;"`
	Hid        uint   `sql:"type:integer not null REFERENCES hosts(id) on update cascade on delete cascade"`
	Mids       string `gorm:"type:varchar(255);"`
	Qps        string `gorm:"type:varchar(255);"`
	IsEnable   uint   `gorm:"type:integer"`
	OriginSids string `gorm:"type:varchar(64);"`
}

type Config struct {
	ID          uint   `gorm:"primary_key;"`
	Description string `gorm:"type:varchar(255);not null;unique_index"`
	Config      string `gorm:"type:text;"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ServiceConfig struct {
	ID          uint   `gorm:"primary_key;"`
	Description string `gorm:"type:varchar(255);not null;unique_index"`
	Sid         uint   `gorm:"unique_index:idx_cid_sid" sql:"type:integer not null REFERENCES services(id) on update cascade on delete cascade"`
	Cid         uint   `gorm: sql:"type:integer not null REFERENCES configs(id) on update cascade on delete cascade"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}