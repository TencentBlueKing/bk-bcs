package config

import (
	"time"
)

const (
	DevEnv  = "dev"
	StagEnv = "stag"
	ProdEnv = "prod"
)

// BaseConf :
type BaseConf struct {
	AppCode       string         `yaml:"app_code"`
	AppSecret     string         `yaml:"app_secret"`
	TimeZone      string         `yaml:"time_zone"`
	LangeuageCode string         `yaml:"langeuage_code"`
	Managers      []string       `yaml:"managers"`
	Debug         bool           `yaml:"debug"`
	Env           string         `yaml:"env"`
	Location      *time.Location `yaml:"-"`
}

// Init : init default redis config
func (c *BaseConf) Init() error {
	// only for development
	var err error
	c.AppCode = ""
	c.AppSecret = ""
	c.TimeZone = "Asia/Shanghai"
	c.LangeuageCode = "zh-hans"
	c.Managers = []string{}
	c.Debug = false
	c.Env = DevEnv
	c.Location, err = time.LoadLocation(c.TimeZone)
	if err != nil {
		return err
	}
	return nil
}
