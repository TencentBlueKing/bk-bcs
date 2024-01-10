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

package service

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/jsoni"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/thirdparty/repo"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
)

const (
	// nodeDetailFileName node detail save file name.
	nodeDetailFileName = "metadata"
	// errFileNotFound is open not exist file error.
	errFileNotFound = "no such file or directory"
)

func (s *Service) queryMetadataInfo(w http.ResponseWriter, r *http.Request) {
	resp := new(BaseResp)

	element := strings.Split(r.URL.Path, "/")
	if len(element) != 8 {
		resp.Err(w, errf.New(errf.InvalidParameter, fmt.Sprintf("request %s path not right", r.URL.Path)))
		return
	}
	project := element[4]
	repoName := element[5]
	sign := element[7]
	if len(project) == 0 || len(repoName) == 0 || len(sign) == 0 {
		resp.Err(w, errf.New(errf.InvalidParameter, fmt.Sprintf("project or repo or sign is nil, project: %s, repo: "+
			"%s, sign: %s", project, repoName, sign)))
		return
	}

	metaPath := nodeMetadataPath(s.Workspace, project, repoName)
	file, err := os.Open(metaPath)
	if err != nil {
		if strings.Contains(err.Error(), errFileNotFound) {
			resp.WriteResp(w, make(map[string]interface{}, 0))
			return
		}
		return
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if len(bytes) == 0 || err != nil {
		resp.WriteResp(w, make(map[string]interface{}, 0))
		return
	}

	meta := make(map[string]*NodeMetadata, 0)
	if err = jsoni.Unmarshal(bytes, &meta); err != nil {
		resp.Err(w, fmt.Errorf("unmarshal metadata failed, err: %v", err))
		return
	}

	metadata, exist := meta[sign]
	if exist {
		appIDs, err := jsoni.Marshal(metadata.AppID)
		if err != nil {
			resp.Err(w, fmt.Errorf("marshal node metadata failed, err: %v", err))
			return
		}

		resp.WriteResp(w, map[string]string{
			"biz_id": strconv.Itoa(int(metadata.BizID)),
			"app_id": string(appIDs),
		})
		return
	}

	resp.WriteResp(w, make(map[string]interface{}, 0))
	return
}

func (s *Service) getNodeInfo(w http.ResponseWriter, r *http.Request) {
	resp := new(BaseResp)

	element := strings.Split(r.URL.Path, "/")
	if len(element) != 6 {
		resp.Err(w, errf.New(errf.InvalidParameter, fmt.Sprintf("request %s path not right", r.URL.Path)))
		return
	}
	project := element[2]
	repoName := element[3]
	sign := element[5]
	if len(project) == 0 || len(repoName) == 0 || len(sign) == 0 {
		resp.Err(w, errf.New(errf.InvalidParameter, fmt.Sprintf("project or repo or sign is nil, project: %s, repo: "+
			"%s, sign: %s", project, repoName, sign)))
		return
	}

	path := nodePath(s.Workspace, project, repoName, sign)
	file, err := os.Open(path)
	if err != nil {
		if strings.Contains(err.Error(), errFileNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		resp.Err(w, fmt.Errorf("open file failed, err: %v", err))
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		resp.Err(w, fmt.Errorf("get file stat failed, err: %v", err))
		return
	}

	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Length", strconv.Itoa(int(stat.Size())))
	return
}

func (s *Service) uploadNode(w http.ResponseWriter, r *http.Request) {
	resp := new(BaseResp)
	sha256 := r.Header.Get(repo.HeaderKeySHA256)

	element := strings.Split(r.URL.Path, "/")
	if len(element) != 6 {
		resp.Err(w, errf.New(errf.InvalidParameter, fmt.Sprintf("request %s path not right", r.URL.Path)))
		return
	}
	project := element[2]
	repoName := element[3]
	sign := element[5]
	if len(project) == 0 || len(repoName) == 0 || len(sign) == 0 {
		resp.Err(w, errf.New(errf.InvalidParameter, fmt.Sprintf("project or repo or sign is nil, project: %s, repo: "+
			"%s, sign: %s", project, repoName, sign)))
		return
	}

	b, err := io.ReadAll(r.Body)
	if err != nil {
		resp.Err(w, fmt.Errorf("read body failed, err: %v", err))
		return
	}

	fileSign := tools.SHA256(string(b))
	if fileSign != sha256 {
		resp.Err(w, fmt.Errorf("upload file sign validate failed, file sign: %s, sha256: %s", fileSign, sha256))
		return
	}

	nodePath := nodePath(s.Workspace, project, repoName, sign)
	file, err := os.OpenFile(nodePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		resp.Err(w, fmt.Errorf("open file fialed, err: %v", err))
		return
	}
	defer file.Close()

	_, err = file.Write(b)
	if err != nil {
		resp.Err(w, fmt.Errorf("write file fialed, err: %v", err))
		return
	}

	metaPath := nodeMetadataPath(s.Workspace, project, repoName)
	metaData := r.Header.Get(repo.HeaderKeyMETA)
	if err := s.recordNodeDetail(metaPath, sign, metaData); err != nil {
		resp.Err(w, err)
		return
	}

	resp.WriteResp(w, &UploadNodeRespData{})
	return
}

func (s *Service) downloadNode(w http.ResponseWriter, r *http.Request) {
	resp := new(BaseResp)

	element := strings.Split(r.URL.Path, "/")
	if len(element) != 6 {
		resp.Err(w, errf.New(errf.InvalidParameter, fmt.Sprintf("request %s path not right", r.URL.Path)))
		return
	}
	project := element[2]
	repoName := element[3]
	sign := element[5]
	if len(project) == 0 || len(repoName) == 0 || len(sign) == 0 {
		resp.Err(w, errf.New(errf.InvalidParameter, fmt.Sprintf("project or repo or sign is nil, project: %s, repo: "+
			"%s, sign: %s", project, repoName, sign)))
		return
	}

	path := nodePath(s.Workspace, project, repoName, sign)
	dlRange := r.Header.Get("Range")
	if len(dlRange) == 0 {
		if err := s.downloadAll(w, path); err != nil {
			resp.Err(w, err)
			return
		}

		return
	}

	if err := s.downloadRange(w, path, dlRange); err != nil {
		resp.Err(w, err)
		return
	}

	return
}

func (s *Service) downloadRange(w http.ResponseWriter, path, dlRange string) error {
	start, offset, err := parseDlRange(dlRange)
	if err != nil {
		return fmt.Errorf("parse download range failed, err: %v", err)
	}

	file, err := os.Open(path)
	if err != nil {
		if strings.Contains(err.Error(), errFileNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return nil
		}

		return fmt.Errorf("open file failed, err: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("get file stat failed, err: %v", err)
	}

	if stat.Size() < start {
		w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
		return nil
	}

	_, err = file.Seek(start, io.SeekStart)
	if err != nil {
		return fmt.Errorf("set file seek failed, start: %d", start)
	}

	if offset == 0 {
		buf := make([]byte, stat.Size()-start)
		if _, err = file.Read(buf); err != nil {
			return fmt.Errorf("read node failed, err: %v", err)
		}

		w.Write(buf)
		return nil
	}

	buf := make([]byte, offset)
	if _, err = file.Read(buf); err != nil {
		return fmt.Errorf("read node failed, err: %v", err)
	}

	w.Write(buf)
	w.WriteHeader(http.StatusPartialContent)
	return nil
}

func parseDlRange(rg string) (int64, int64, error) {
	els := strings.Split(rg, "=")
	if len(els) != 2 {
		return 0, 0, errors.New("file download range is not right format")
	}

	if strings.TrimSpace(els[0]) != "bytes" {
		return 0, 0, fmt.Errorf("file download range %s type not bytes", els[0])
	}

	split := strings.Split(strings.TrimSpace(els[1]), "-")
	if len(split) == 0 {
		return 0, 0, fmt.Errorf("file download range not has limit")
	}

	startStr := strings.TrimSpace(split[0])
	start, err := strconv.ParseInt(startStr, 10, 64)
	if err != nil {
		return 0, 0, err
	}

	if len(split) == 1 || len(split[1]) == 0 {
		return start, 0, nil
	}

	offsetStr := strings.TrimSpace(split[1])
	offset, err := strconv.ParseInt(offsetStr, 10, 64)
	if err != nil {
		return 0, 0, err
	}

	return start, offset, nil
}

func (s *Service) downloadAll(w http.ResponseWriter, path string) error {
	file, err := os.Open(path)
	if err != nil {
		if strings.Contains(err.Error(), errFileNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return nil
		}

		return fmt.Errorf("open file failed, err: %v", err)
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("read node failed, err: %v", err)
	}

	w.Write(bytes)
	return nil
}

func (s *Service) recordNodeDetail(path, sign, metadata string) error {
	readFile, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return fmt.Errorf("open file failed, err: %v", err)
	}
	defer readFile.Close()

	bytes, err := io.ReadAll(readFile)
	if err != nil {
		return fmt.Errorf("read node detail failed, err: %v", err)
	}

	meta := make(map[string]*NodeMetadata, 0)
	if len(bytes) != 0 {
		if err := jsoni.Unmarshal(bytes, &meta); err != nil {
			return err
		}
	}

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer file.Close()

	bizID, appIDs, err := parseMetadata(metadata)
	if err != nil {
		return err
	}

	meta[sign] = &NodeMetadata{
		BizID: bizID,
		AppID: appIDs,
	}

	marshal, err := jsoni.Marshal(meta)
	if err != nil {
		return err
	}

	if _, err = file.Write(marshal); err != nil {
		return err
	}

	return nil
}

func nodeMetadataPath(workspace, project, repo string) string {
	return filepath.Clean(fmt.Sprintf("%s/%s/%s/%s", workspace, project, repo, nodeDetailFileName))
}

func nodePath(workspace, project, repo, sigh string) string {
	return filepath.Clean(fmt.Sprintf("%s/%s/%s/%s", workspace, project, repo, sigh))
}
