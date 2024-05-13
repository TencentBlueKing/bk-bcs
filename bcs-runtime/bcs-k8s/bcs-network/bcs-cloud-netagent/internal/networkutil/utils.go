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
 */

package networkutil

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/internal/constant"
)

func appnedFile(filepath, value string) error {
	f, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.WriteString(value); err != nil {
		return err
	}
	return nil
}

func addRouteTables(idMap map[string]string) bool {
	var err error
	for id, name := range idMap {
		err = appnedFile("/etc/iproute2/rt_tables", fmt.Sprintf("%s\t%s\n", id, name))
		if err != nil {
			blog.Errorf("error to add '%s\t%s' into /etc/iproute2/rt_tables, err %s", id, name, err.Error())
			return false
		}
		blog.Infof("add '%s\t%s' successfully", id, name)
	}
	return true
}

func getEniRtTableIDs(content string) map[string]string {
	if !strings.Contains(content, constant.EniPrefix) {
		return nil
	}
	strs := strings.Split(content, "\n")
	if len(strs) > 1 && strs[len(strs)-1] == "" {
		strs = strs[0 : len(strs)-1]
	}
	ret := make(map[string]string)
	for _, str := range strs {
		if strings.Contains(str, constant.EniPrefix) {
			words := strings.Split(str, "\t")
			if len(words) == 2 {
				key := strings.TrimSpace(words[0])
				value := strings.TrimSpace(words[1])
				ret[key] = value
			}
		}
	}
	return ret
}

func ensureRouteTables(tableIDMap map[string]string) bool {
	content, err := ioutil.ReadFile("/etc/iproute2/rt_tables")
	if err != nil {
		blog.Errorf("read file /etc/iproute2/rt_tables failed, err %s", err.Error())
		return false
	}
	curMap := getEniRtTableIDs(string(content))
	mapNeedAdd := make(map[string]string)
	for key, value := range tableIDMap {
		if _, ok := curMap[key]; !ok {
			mapNeedAdd[key] = value
		}
	}
	if len(mapNeedAdd) == 0 {
		blog.Infof("no new ids to ensure")
		return true
	}
	return addRouteTables(mapNeedAdd)
}

func cleanEniRtTable(content string) string {
	if !strings.Contains(content, constant.EniPrefix) {
		return content
	}
	strs := strings.Split(content, "\n")
	if len(strs) > 1 && strs[len(strs)-1] == "" {
		strs = strs[0 : len(strs)-1]
	}
	ret := ""
	for _, str := range strs {
		if strings.Contains(str, constant.EniPrefix) {
			continue
		}
		ret = ret + str + "\n"
	}
	return ret
}

func delRouteTables() bool {
	bytes, err := ioutil.ReadFile("/etc/iproute2/rt_tables")
	if err != nil {
		blog.Errorf("read file /etc/iproute2/rt_tables failed, err %s", err.Error())
		return false
	}
	str := string(bytes)
	retStr := cleanEniRtTable(str)
	if retStr == str {
		blog.Infof("no eni record in rt_table, no need to clean")
		return true
	}
	err = ioutil.WriteFile("/etc/iproute2/rt_tables", []byte(retStr), 0644)
	if err != nil {
		blog.Errorf("failed to clean /etc/iproute2/rt_tables")
		return false
	}
	return true
}

func getIPForward() (int, error) {
	var err error
	content, err := ioutil.ReadFile("/proc/sys/net/ipv4/ip_forward")
	if err != nil {
		blog.Errorf("failed to get /proc/sys/net/ipv4/ip_forward, err %s", err.Error())
		return -1, fmt.Errorf("failed to get /proc/sys/net/ipv4/ip_forward, err %s", err.Error())
	}
	value, err := strconv.Atoi(strings.TrimRight(string(content), "\n"))
	if err != nil {
		blog.Errorf("convert %s to int failed, err %s", content, err.Error())
		return -1, fmt.Errorf("convert %s to int failed, err %s", content, err.Error())
	}
	return value, nil
}

func setIPForward(value bool) bool {
	var err error
	v := []byte{'1', '\n'}
	if !value {
		v = []byte{'0', '\n'}
	}
	if err = ioutil.WriteFile("/proc/sys/net/ipv4/ip_forward", v, 0644); err != nil {
		blog.Errorf("failed to set ipv4.ip_forward to %s, err %s", string(v), err.Error())
		return false
	}
	blog.Infof("success to set ipv4.ip_forward to %s", string(v))
	return true
}

func getRpFilter(ifname string) (int, error) {
	var err error
	procPath := fmt.Sprintf("/proc/sys/net/ipv4/conf/%s/rp_filter", ifname)
	content, err := ioutil.ReadFile(procPath)
	if err != nil {
		blog.Errorf("failed to get ipv4.conf.%s.rp_filter, err %s", ifname, err.Error())
		return -1, fmt.Errorf("failed to get ipv4.conf.%s.rp_filter, err %s", ifname, err.Error())
	}
	value, err := strconv.Atoi(strings.TrimRight(string(content), "\n"))
	if err != nil {
		blog.Errorf("convert %s to int failed, err %s", content, err.Error())
		return -1, fmt.Errorf("convert %s to int failed, err %s", content, err.Error())
	}
	return value, nil
}

func setRpFilter(ifname string, value bool) bool {
	var err error
	v := []byte{'1', '\n'}
	if !value {
		v = []byte{'0', '\n'}
	}
	procPath := fmt.Sprintf("/proc/sys/net/ipv4/conf/%s/rp_filter", ifname)
	if err = ioutil.WriteFile(procPath, v, 0644); err != nil {
		blog.Errorf("failed to set ipv4.conf.%s.rp_filter to %s, err %s", ifname, string(v), err.Error())
		return false
	}
	blog.Infof("success to set ipv4.conf.%s.rp_filter to %s", ifname, string(v))
	return true
}
