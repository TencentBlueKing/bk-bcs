/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package bkrepo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"bk-bscp/pkg/common"
	dl "bk-bscp/pkg/downloader"
)

const (
	// defaultDownloadConcurrent is default download concurrent num.
	defaultDownloadConcurrent = 5

	// defaultDownloadLimitBytesInSecond is default download limit in second, 50MB.
	defaultDownloadLimitBytesInSecond = int64(1024 * 1024 * 50)
)

// CreateProject creates new project from bkrepo.
func CreateProject(host string, auth *Auth, req *CreateProjectReq, timeout time.Duration) error {
	// NOTE: change bscp biz_id to "bscp-{CMDB biz_id}" for bkrepo.
	req.Name = BSCPBIZIDPREFIX + req.Name
	req.DisplayName = BSCPBIZIDPREFIX + req.DisplayName

	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	request, err := http.NewRequest("POST",
		fmt.Sprintf("%s/repository/api/project", host), strings.NewReader(string(data)))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", fmt.Sprintf("Platform %s", auth.Token))
	request.Header.Set(BKRepoUIDHeaderKey, auth.UID)

	clientWithTimeout := &http.Client{Timeout: timeout}
	response, err := clientWithTimeout.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusBadRequest {
		return fmt.Errorf("response status[%+v]", response.StatusCode)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	resp := &CommonResp{}
	if err := json.Unmarshal(body, resp); err != nil {
		return err
	}

	if resp.Code != 0 && resp.Code != BKRepoErrCodeProjectAlreadyExist {
		return fmt.Errorf("errcode[%d] errmsg[%+v]", resp.Code, resp.Message)
	}
	return nil
}

// CreateRepo creates new project repo from bkrepo.
func CreateRepo(host string, auth *Auth, req *CreateRepoReq, timeout time.Duration) error {
	// NOTE: change bscp biz_id to "bscp-{CMDB biz_id}" for bkrepo.
	req.ProjectID = BSCPBIZIDPREFIX + req.ProjectID

	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	request, err := http.NewRequest("POST",
		fmt.Sprintf("%s/repository/api/repo", host), strings.NewReader(string(data)))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", fmt.Sprintf("Platform %s", auth.Token))
	request.Header.Set(BKRepoUIDHeaderKey, auth.UID)

	clientWithTimeout := &http.Client{Timeout: timeout}
	response, err := clientWithTimeout.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusBadRequest {
		return fmt.Errorf("response status[%+v]", response.StatusCode)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	resp := &CommonResp{}
	if err := json.Unmarshal(body, resp); err != nil {
		return err
	}

	if resp.Code != 0 && resp.Code != BKRepoErrCodeRepoAlreadyExist {
		return fmt.Errorf("errcode[%d] errmsg[%+v]", resp.Code, resp.Message)
	}
	return nil
}

// DownloadContent downloads target content base on cid.
func DownloadContent(option *DownloadContentOption, auth *Auth, timeout time.Duration) (time.Duration, error) {
	timenow := time.Now()

	headers := make(map[string]string)
	headers["Authorization"] = fmt.Sprintf("Platform %s", auth.Token)
	headers[BKRepoUIDHeaderKey] = auth.UID

	if option.Concurrent == 0 {
		option.Concurrent = defaultDownloadConcurrent
	}
	if option.LimitBytesInSecond == 0 {
		option.LimitBytesInSecond = defaultDownloadLimitBytesInSecond
	}

	downloader := dl.NewDownloader(option.URL, option.Concurrent, headers, option.NewFile)
	downloader.SetRateLimiterOption(dl.NewSimpleRateLimiter(option.LimitBytesInSecond))

	// download.
	if err := downloader.Download(timeout); err != nil {
		downloader.Clean()
		return time.Since(timenow), fmt.Errorf("download failed, %+v", err)
	}

	// check cid.
	contentID, err := common.FileSHA256(option.NewFile)
	if err != nil {
		downloader.Clean()
		return time.Since(timenow), fmt.Errorf("check file cid failed, %+v", err)
	}

	if contentID != option.ContentID {
		downloader.Clean()
		return time.Since(timenow), errors.New("download invalid cid")
	}

	return time.Since(timenow), nil
}

// DownloadContentInMemory downloads target content base on cid in memory.
func DownloadContentInMemory(option *DownloadContentOption, auth *Auth,
	timeout time.Duration) (string, time.Duration, error) {

	timenow := time.Now()

	request, err := http.NewRequest("GET", option.URL, nil)
	if err != nil {
		return "", time.Since(timenow), err
	}
	request.Header.Set("Authorization", fmt.Sprintf("Platform %s", auth.Token))
	request.Header.Set(BKRepoUIDHeaderKey, auth.UID)

	client := &http.Client{Timeout: timeout}
	response, err := client.Do(request)
	if err != nil {
		return "", time.Since(timenow), err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusBadRequest &&
		response.StatusCode != http.StatusNotFound {
		return "", time.Since(timenow), fmt.Errorf("response status[%+v]", response.StatusCode)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", time.Since(timenow), err
	}

	if response.StatusCode != http.StatusOK {
		resp := &CommonResp{}
		if err := json.Unmarshal(body, resp); err != nil {
			return "", time.Since(timenow), err
		}
		return "", time.Since(timenow), fmt.Errorf("download failed, errcode[%d] errmsg[%+v]", resp.Code, resp.Message)
	}

	contentID := common.SHA256(string(body))
	if contentID != option.ContentID {
		return "", time.Since(timenow), errors.New("download invalid cid")
	}
	return string(body), time.Since(timenow), nil
}

// UploadContentInMemory uploads content in memory.
func UploadContentInMemory(option *UploadContentOption, auth *Auth, content string,
	timeout time.Duration) (time.Duration, error) {

	timenow := time.Now()

	contentID := common.SHA256(content)
	if contentID != option.ContentID {
		return time.Since(timenow), errors.New("upload invalid cid")
	}

	request, err := http.NewRequest("PUT", option.URL, strings.NewReader(content))
	if err != nil {
		return time.Since(timenow), err
	}
	request.Header.Set("Authorization", fmt.Sprintf("Platform %s", auth.Token))
	request.Header.Set(BKRepoUIDHeaderKey, auth.UID)
	request.Header.Set(BKRepoSHA256HeaderKey, option.ContentID)

	client := &http.Client{Timeout: timeout}
	response, err := client.Do(request)
	if err != nil {
		return time.Since(timenow), err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusBadRequest {
		return time.Since(timenow), fmt.Errorf("response status[%+v]", response.StatusCode)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return time.Since(timenow), err
	}

	resp := &CommonResp{}
	if err := json.Unmarshal(body, resp); err != nil {
		return time.Since(timenow), err
	}

	if resp.Code != 0 && resp.Code != BKRepoErrCodeNodeAlreadyExist {
		return time.Since(timenow), fmt.Errorf("errcode[%d] errmsg[%+v]", resp.Code, resp.Message)
	}
	return time.Since(timenow), nil
}

// ValidateContentExistence validates existence of target content.
func ValidateContentExistence(url string, auth *Auth, timeout time.Duration) error {
	request, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", fmt.Sprintf("Platform %s", auth.Token))
	request.Header.Set(BKRepoUIDHeaderKey, auth.UID)

	client := &http.Client{Timeout: timeout}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusNotFound {
		return fmt.Errorf("get content metadata failed, %+v", response.StatusCode)
	}

	if response.StatusCode == http.StatusNotFound {
		return errors.New("content metadata not found")
	}
	return nil
}
