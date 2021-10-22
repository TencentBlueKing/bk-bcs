/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package utils

import (
	"bufio"
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-apiserver-proxy/pkg/ipvs"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const (
	IpvsPersistFileName = "ipvsConfig.yaml"
	RcLocalIpvsFlag     = "IPVS_START_UP"
)

// EndPoint wrap IP&Port
type EndPoint struct {
	IP   string
	Port uint32
}

// String trans endpoint to ip:port
func (ep EndPoint) String() string {
	port := strconv.Itoa(int(ep.Port))
	return ep.IP + ":" + port
}

// SplitServer split server to ip, port
func SplitServer(server string) (string, uint32) {
	s := strings.Split(server, ":")
	if len(s) != 2 {
		blog.Warn("SplitServer error: len(s) is not two.")
		return "", 0
	}
	blog.V(5).Infof("SplitServer debug: IP: %s, Port: %s", s[0], s[1])

	p, err := strconv.Atoi(s[1])
	if err != nil {
		blog.Warn("SplitServer error: ", err)
		return "", 0
	}
	return s[0], uint32(p)
}

// BuildVirtualServer build vip to ipvs.VirtualServer
func BuildVirtualServer(vip string, scheduler string) *ipvs.VirtualServer {
	ip, port := SplitServer(vip)
	virServer := &ipvs.VirtualServer{
		Address:   net.ParseIP(ip),
		Protocol:  "TCP",
		Port:      port,
		Scheduler: scheduler,
		Flags:     0,
		Timeout:   0,
	}
	return virServer
}

// BuildRealServer build real to ipvs.RealServer
func BuildRealServer(real string) *ipvs.RealServer {
	ip, port := SplitServer(real)
	realServer := &ipvs.RealServer{
		Address: net.ParseIP(ip),
		Port:    port,
		Weight:  1,
	}
	return realServer
}

func WriteToFile(filePath string, content string) error {
	var file *os.File
	var err error
	file, err = os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		log.Printf("open file %s failed; %v", filePath, err)
		return err
	}
	defer file.Close()
	_, err = file.Write([]byte(content))
	if err != nil {
		log.Printf("write to file %s failed", filePath)
		return err
	}
	log.Printf("write to file %s succeed!", filePath)
	return nil
}

func SetIpvsStartup(ipvsPersistDir string, toolPath string) error {
	command := "chmod +x /etc/rc.d/rc.local"
	cmd := exec.Command("/bin/sh", "-c", command)
	output, err := cmd.Output()
	if err != nil {
		log.Printf("command [%s] exec failed", command)
	}
	resp := string(output)
	log.Println(resp)

	exist, err := checkFlagExist("/etc/rc.local", RcLocalIpvsFlag)
	if err != nil {
		return err
	}
	if exist {
		return nil
	} else {
		command := fmt.Sprintf("%v -cmd reload -persistDir %v",
			toolPath, ipvsPersistDir)
		command = "# " + RcLocalIpvsFlag + "\n" + command + "\n"

		err = WriteToFile("/etc/rc.local", command)
		if err != nil {
			log.Printf("write command [%s] to rc.local failed", command)
			return err
		}
	}
	return nil
}

func checkFlagExist(path string, flag string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		log.Printf("open file [%v] failed", path)
		return false, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), flag) {
			log.Printf("ipvs startup flag already exists")
			return true, nil
		}
	}
	if err := scanner.Err(); err != nil {
		log.Printf("scan file [%s] failed, %v", path, err)
		return false, err
	}
	return false, nil
}
