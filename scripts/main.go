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

// Package main  github dependabot alerts restful api
// 地址：https://docs.github.com/en/rest/dependabot/alerts?apiVersion=2022-11-28
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	excelize "github.com/xuri/excelize/v2"
)

func main() {
	DependabotAlertsExportToExcel()
}

// Alert dependabot alert model
type Alert struct {
	Number                int                   `json:"number"`
	State                 string                `json:"state"`
	Dependency            Dependency            `json:"dependency"`
	SecurityAdvisory      SecurityAdvisory      `json:"security_advisory"`
	SecurityVulnerability SecurityVulnerability `json:"security_vulnerability"`
}

// Dependency Dependency model
type Dependency struct {
	ManifestPath string  `json:"manifest_path"`
	Package      Package `json:"package"`
}

// Package Package model
type Package struct {
	Ecosystem string `json:"ecosystem"`
	Name      string `json:"name"`
}

// SecurityAdvisory SecurityAdvisory model
type SecurityAdvisory struct {
	Summary     string `json:"summary"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
}

// SecurityVulnerability SecurityVulnerability model
type SecurityVulnerability struct {
	VulnerableVersionRange string              `json:"vulnerable_version_range"`
	FirstPatchedVersion    FirstPatchedVersion `json:"first_patched_version"`
}

// FirstPatchedVersion FirstPatchedVersion model
type FirstPatchedVersion struct {
	Identifier string `json:"identifier"`
}

// excel标题及表格序号
var title = map[string]string{
	"alerts序号": "A",
	"路径":       "B",
	"级别":       "C",
	"组件名称":     "D",
	"开发语言":     "E",
	"受影响的版本":   "F",
	"补丁版本":     "G",
	"总结":       "H",
	"描述":       "I",
}

// 记录表格行数
var lineNum = 1

// alert地址，fork项目，%s项目组织名称从环境变量取：OrgName
const (
	alertUrl = "https://api.github.com/repos/%s/bk-bcs/dependabot/alerts"
	// 因为外网网络连接问题有时会失败，程序会重试继续，整个程序最多能重试5次，此参数可修改
	retryTimes = 5
)

var retryCount = 0

// DependabotAlertsExportToExcel DependabotAlerts导出到表格
func DependabotAlertsExportToExcel() {
	if os.Getenv("GHToken") == "" || os.Getenv("OrgName") == "" {
		fmt.Println("GHToken or OrgName is null")
		return
	}

	excelPath := "./dependabot-alerts.xlsx"

	// 创建excel表格
	err := createExcel(excelPath)
	if err != nil {
		fmt.Println("createExcel err: ", err)
		return
	}

	// dependabot alerts最大序号
	var maxNumber int
	// 失败重试
	var isFail bool
	for !isFail && retryCount != retryTimes {
		maxNumberUrl := fmt.Sprintf(alertUrl+"?state=open&per_page=1&direction=desc", os.Getenv("OrgName"))
		maxNumberValue, err := getMaxNumber(maxNumberUrl) // nolint
		if err != nil {
			// 网络原因重试
			if strings.Contains(err.Error(), io.EOF.Error()) {
				fmt.Println("getMaxNumber err: ", err)
				retryCount++
				continue
			}
			fmt.Println("getMaxNumber err: ", err)
			return
		}
		maxNumber = maxNumberValue
		isFail = true
	}
	err = writeToExcel(excelPath, maxNumber)
	if err != nil {
		fmt.Println("writeToExcel err: ", err)
		return
	}
}

// 写入excel表格
func writeToExcel(excelPath string, maxNumber int) error {
	// 开始读写excel表格
	xlsx, err := excelize.OpenFile(excelPath)
	if err != nil {
		return err
	}

	// 每页50条数据，github上限100条，100条请求时间有点慢
	var perPage = 50
	// nolint
	for i := maxNumber; i > 0; i -= perPage {
		url := fmt.Sprintf(alertUrl+"?per_page=%d&page=%d", os.Getenv("OrgName"), perPage, i/perPage+1)
		b, err := SimpleHttpGetRequest(url, os.Getenv("GHToken"))
		if err != nil {
			// 网络原因重试
			if strings.Contains(err.Error(), io.EOF.Error()) && retryCount != retryTimes {
				fmt.Println("SimpleHttpGetRequest err: ", err)
				i += 30
				retryCount++
				continue
			}
			fmt.Println("SimpleHttpGetRequest err: ", err)
			return err
		}

		alerts, err := getAlerts(b)
		if err != nil {
			return err
		}

		// 数据写入文件
		for j := len(alerts) - 1; j >= 0; j-- {
			// nolint
			// 通过配置筛选过滤的不写入
			if alerts[j].State != "auto_dismissed" {
				xlsx.SetCellValue("sheet1", title["alerts序号"]+fmt.Sprintf("%d", lineNum), alerts[j].Number)
				xlsx.SetCellValue("sheet1", title["路径"]+fmt.Sprintf("%d", lineNum), alerts[j].Dependency.ManifestPath)
				xlsx.SetCellValue("sheet1", title["组件名称"]+fmt.Sprintf("%d", lineNum),
					alerts[j].Dependency.Package.Name)
				xlsx.SetCellValue("sheet1", title["开发语言"]+fmt.Sprintf("%d", lineNum),
					alerts[j].Dependency.Package.Ecosystem)
				xlsx.SetCellValue("sheet1", title["级别"]+fmt.Sprintf("%d", lineNum),
					alerts[j].SecurityAdvisory.Severity)
				xlsx.SetCellValue("sheet1", title["描述"]+fmt.Sprintf("%d", lineNum),
					alerts[j].SecurityAdvisory.Description)
				xlsx.SetCellValue("sheet1", title["总结"]+fmt.Sprintf("%d", lineNum),
					alerts[j].SecurityAdvisory.Summary)
				xlsx.SetCellValue("sheet1", title["受影响的版本"]+fmt.Sprintf("%d", lineNum),
					alerts[j].SecurityVulnerability.VulnerableVersionRange)
				xlsx.SetCellValue("sheet1", title["补丁版本"]+fmt.Sprintf("%d", lineNum),
					alerts[j].SecurityVulnerability.FirstPatchedVersion.Identifier)
				lineNum++
			}
		}
		fmt.Println("writeToExcel success: ", i/perPage)
	}

	if err = xlsx.Save(); err != nil {
		return err
	}
	return nil
}

// 获取dependabot alert最大序号
func getMaxNumber(url string) (int, error) {
	b, err := SimpleHttpGetRequest(url, os.Getenv("GHToken"))
	if err != nil {
		return 0, err
	}
	alerts, err := getAlerts(b)
	if err != nil {
		return 0, err
	}
	if len(alerts) == 0 {
		return 0, nil
	}
	return alerts[0].Number, nil
}

// 创建表格，存在则删除再创建
func createExcel(excelPath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(excelPath); err == nil {
		if err != nil {
			return err
		}
		// 文件存在，删除文件
		if err := os.Remove(excelPath); err != nil {
			fmt.Println("删除文件失败:", err)
			return err
		}
	}
	// 文件不存在，创建
	xlsx := excelize.NewFile()
	// 写入标题
	for key, value := range title {
		xlsx.SetCellValue("sheet1", value+fmt.Sprintf("%d", lineNum), key) // nolint
	}
	if err := xlsx.SaveAs(excelPath); err != nil {
		return err
	}
	lineNum++
	return nil
}

// 解析alert
func getAlerts(b []byte) ([]Alert, error) {
	alerts := []Alert{}
	err := json.Unmarshal(b, &alerts)
	if err != nil {
		return nil, err
	}
	return alerts, nil
}

// SimpleHttpGetRequest method GET，发起github restful请求
// url: 请求地址
// patToken: github personal access tokens
func SimpleHttpGetRequest(url string, patToken string) ([]byte, error) {
	client := http.Client{Timeout: time.Second * 10}
	var respRaw []byte

	var request *http.Request

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return respRaw, err
	}
	request.Header.Add("Accept", "application/vnd.github+json")
	request.Header.Add("Authorization", "Bearer "+patToken)
	request.Header.Add("X-GitHub-Api-Version", "2022-11-28")

	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("code:%d,resp:%+v", resp.StatusCode, resp)
	}
	respRaw, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return respRaw, nil
}
