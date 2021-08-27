package config

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
	"path"
)

const (
	IpvsConfigFileName = "ipvsConfig.yaml"
)

type IpvsConfig struct {
	Scheduler     string   `json:"scheduler"`
	VirtualServer string   `json:"vs"`
	RealServer    []string `json:"rs"`
}

func WriteIpvsConfig(dir string, config IpvsConfig) error {
	_, exist := os.Stat(dir)
	if os.IsNotExist(exist) {
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			fmt.Println("create ipvs persist dir failed")
			return err
		}
	}
	viper.SetConfigFile(path.Join(dir, IpvsConfigFileName))
	viper.Set("vs", config.VirtualServer)
	viper.Set("rs", config.RealServer)
	viper.Set("scheduler", config.Scheduler)
	err := viper.WriteConfigAs(path.Join(dir,IpvsConfigFileName))
	if err != nil {
		fmt.Println("persist ipvs config to file failed")
		return err
	}
	return nil
}

func ReadIpvsConfig(dir string) (*IpvsConfig, error) {
	_, exist := os.Stat(dir)
	if os.IsNotExist(exist) {
		err := fmt.Errorf("ipvs config persist dir [%s] not exists", dir)
		return nil, err
	}
	viper.SetConfigFile(path.Join(dir, IpvsConfigFileName))
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("read persist config failed, %v", err)
		return nil, err
	}
	config := &IpvsConfig{
		VirtualServer: viper.GetString("vs"),
		RealServer:    viper.GetStringSlice("rs"),
		Scheduler:     viper.GetString("scheduler"),
	}
	return config, nil
}
