package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	//"time"
	"bytes"
	"encoding/json"
	"regexp"

	election "bk-bcs/bcs-k8s/tkex-statefulsetplus-operator/pkg/util/kube-leaderelection"
)

const configFileSizeLimit = 10 << 20

var (
	electionConfigPath string
)

type ElectionConfig = election.Config

func init() {
	flag.StringVar(&electionConfigPath, "electionConfigPath", "", "Path to a electionConfig.  required if use custom election config for electing leader.")
}

func main() {
	flag.Parse()

	if electionConfigPath == "" {
		fmt.Println("please set electionConfigPath ")
		return
	}

	electionConfig, _err := LoadConfig(electionConfigPath)
	if _err != "" {
		fmt.Println(_err)
		return
	}

	elector, err := election.NewLeaderElector(*electionConfig)
	if err != nil {
		panic(err)
	}
	elector.Register(&listener{})
	elector.Run(context.Background())
}

type listener struct {
}

func (l *listener) StartedLeading() {
	log.Printf("[INFO] %s: started leading", hostname())
}

// invoked when this node stops being the leader
func (l *listener) StoppedLeading() {
	log.Printf("[INFO] %s: stopped leading", hostname())
}

// invoked when a new leader is elected
func (l *listener) NewLeader(id string) {
	log.Printf("[INFO] %s: new leader: %s", hostname(), id)
}

func hostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	return hostname
}

func LoadConfig(path string) (Config *ElectionConfig, Err string) {
	var config ElectionConfig
	config_file, err := os.Open(path)
	if err != nil {
		emit("Failed to open config file '%s': %s\n", path, err)
		return &config, err.Error()
	}

	fi, _ := config_file.Stat()
	if size := fi.Size(); size > (configFileSizeLimit) {
		emit("config file (%q) size exceeds reasonable limit (%d) - aborting", path, size)
		return &config, fmt.Sprintf("config file (%q) size exceeds reasonable limit (%d) - aborting", path, size) // REVU: shouldn't this return an error, then?
	}

	if fi.Size() == 0 {
		emit("config file (%q) is empty, skipping", path)
		return &config, fmt.Sprintf("config file (%q) is empty, skipping", path)
	}

	buffer := make([]byte, fi.Size())
	_, err = config_file.Read(buffer)
	//emit("\n %s\n", buffer)

	buffer, err = StripComments(buffer) //去掉注释
	if err != nil {
		emit("Failed to strip comments from json %q: %s\n", path, err)
		return &config, fmt.Sprintf("Failed to strip comments from json %q: %s\n", path, err)
	}

	buffer = []byte(os.ExpandEnv(string(buffer))) //特殊

	err = json.Unmarshal(buffer, &config) //解析json格式数据
	if err != nil {
		emit("Failed unmarshalling json %q: %s\n", path, err)
		return &config, fmt.Sprintf("Failed unmarshalling json %q: %s\n", path, err)
	}
	return &config, ""
}

func StripComments(data []byte) ([]byte, error) {
	data = bytes.Replace(data, []byte("\r"), []byte(""), 0) // Windows
	lines := bytes.Split(data, []byte("\n"))                //split to muli lines
	filtered := make([][]byte, 0)

	for _, line := range lines {
		match, err := regexp.Match(`^\s*#`, line)
		if err != nil {
			return nil, err
		}
		if !match {
			filtered = append(filtered, line)
		}
	}

	return bytes.Join(filtered, []byte("\n")), nil
}

func emit(msgfmt string, args ...interface{}) {
	log.Printf(msgfmt, args...)
}
