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

package manager

import (
	"compress/gzip"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpclient"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-process-executor/process-executor/types"

	"github.com/Microsoft/go-winio/archive/tar"
)

func (m *manager) initCli() {
	m.cli = httpclient.NewHttpClient()
	m.cli.SetHeader("Content-Type", "application/json")
	m.cli.SetHeader("Accept", "application/json")
}

func (m *manager) requestManager(method, uri string, data []byte, header map[string]string) ([]byte, error) {
	blog.V(3).Infof("request uri %s data %s", uri, string(data))

	if header != nil {
		for k, v := range header {
			m.cli.SetHeader(k, v)
		}
	}

	resp, err := m.cli.RequestEx(uri, method, nil, data)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(string(resp.Reply))
	}

	return resp.Reply, err
}

// RAR file：           /data/bcs/workspace/packages_dir/lgame_v1.tar.gz
// unzipped directory： /data/bcs/workspace/extract_dir/lgame_v1/***
// symbolic link：      /data/bcs/workspace/work_dir/lgame_v1  -> /data/bcs/extract_dir/lgame_v1

func (m *manager) downloadAndTarProcessPackages(processInfo *types.ProcessInfo) error {
	if processInfo.Uris == nil || len(processInfo.Uris) == 0 {
		return nil
	}
	uriPack := processInfo.Uris[0]
	m.lockObjectKey(uriPack.PackagesFile)
	defer m.unLockObjectKey(uriPack.PackagesFile)

	exist, err := m.downloadProcessPackages(processInfo)
	if err != nil {
		return err
	}

	if exist {
		blog.Infof("process %s PackagesFile %s is valid, and not need extract", processInfo.Id, uriPack.PackagesFile)
		return nil
	}

	//whether uriPack.ExtractDir exist
	_, err = os.Stat(uriPack.ExtractDir)
	if err != nil {
		blog.Errorf("process %s ExtractDir %s not exist, and need mkdir", processInfo.Id, uriPack.ExtractDir)
		err = os.MkdirAll(filepath.Dir(uriPack.ExtractDir), 0755)
		if err != nil {
			blog.Errorf("process %s mkdir %s error %s", processInfo.Id, filepath.Dir(uriPack.OutputDir), err.Error())
			return err
		}
	} else {
		blog.Infof("process %s ExtractDir %s exist, and need remove", processInfo.Id, uriPack.ExtractDir)
		err = os.RemoveAll(uriPack.ExtractDir)
		if err != nil {
			blog.Errorf("process %s remove file %s error %s", processInfo.Id, uriPack.OutputDir, err.Error())
			return err
		}
	}

	f, err := os.Open(uriPack.PackagesFile)
	if err != nil {
		blog.Errorf("process %s open file %s error %s", processInfo.Id, uriPack.PackagesFile, err.Error())
		return err
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		blog.Errorf("process %s package %s gzip.NewReader error %s", processInfo.Id, uriPack.PackagesFile, err.Error())
		return err
	}
	defer gr.Close()

	r := tar.NewReader(gr)
	for hdr, err := r.Next(); err != io.EOF; hdr, err = r.Next() {
		if err != nil {
			blog.Errorf("process %s tar file %s error %s", processInfo.Id, uriPack.PackagesFile, err.Error())
			return err
		}

		var fi *os.File
		if hdr.FileInfo().IsDir() {
			err = os.MkdirAll(filepath.Join(uriPack.ExtractDir, hdr.Name), hdr.FileInfo().Mode())
			if err != nil {
				blog.Errorf("process %s create dir %s error %s", processInfo.Id,
					filepath.Join(uriPack.ExtractDir, hdr.Name), err.Error())
				return err
			}
			goto ChownResp
		}

		if hdr.Linkname != "" {
			err = os.Symlink(hdr.Linkname, filepath.Join(uriPack.ExtractDir, hdr.Name))
			if err != nil {
				blog.Errorf("process %s create dir %s error %s", processInfo.Id,
					filepath.Join(uriPack.ExtractDir, hdr.Name), err.Error())
				return err
			}
			continue
		}

		fi, err = os.OpenFile(filepath.Join(uriPack.ExtractDir, hdr.Name), os.O_RDWR|os.O_CREATE, hdr.FileInfo().Mode())
		if err != nil {
			blog.Errorf("process %s create file %s error %s", processInfo.Id,
				filepath.Join(uriPack.ExtractDir, hdr.Name), err.Error())
			return err
		}

		_, err = io.Copy(fi, r)
		fi.Close()
		if err != nil {
			blog.Errorf("process %s copy file %s error %s", processInfo.Id,
				filepath.Join(uriPack.ExtractDir, hdr.Name), err.Error())
			return err
		}

		//chown file user:group
	ChownResp:
		u, err := user.Lookup(hdr.Uname)
		if err != nil {
			blog.Errorf("process %s lookup user %s error %s", processInfo.Id, hdr.Uname, err.Error())
			continue
		}
		uid, _ := strconv.Atoi(u.Uid)
		gid, _ := strconv.Atoi(u.Gid)
		err = os.Chown(filepath.Join(uriPack.ExtractDir, hdr.Name), uid, gid)
		if err != nil {
			blog.Errorf("process %s chown file %s user %s error %s", processInfo.Id,
				filepath.Join(uriPack.ExtractDir, hdr.Name), hdr.Uname, err.Error())
		}
	}

	return nil
}

func (m *manager) downloadProcessPackages(processInfo *types.ProcessInfo) (bool, error) {
	//"http://xxxx.artifactory.xxxx.com/generic-local/xxxx/pack-master.tar.gz" convert to
	//"http://xxxx.artifactory.xxxx.com/api/storage/generic-local/xxxx/pack-master.tar.gz"
	uriPack := processInfo.Uris[0]
	u, err := url.Parse(uriPack.Value)
	if err != nil {
		blog.Errorf("process %s url.parse %s error %s", processInfo.Id, uriPack.Value, err.Error())
		return false, err
	}

	uri := fmt.Sprintf("%s://%s/api/storage%s", u.Scheme, u.Host, u.Path)
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", uriPack.User, uriPack.Pwd)))
	header := map[string]string{"Authorization": fmt.Sprintf("Basic %s", auth)}
	by, err := m.requestManager("GET", uri, nil, header)
	if err != nil {
		blog.Errorf("check process %s packages uri %s error %s", processInfo.Id, uri, err.Error())
		return false, err
	}
	blog.Infof("process %s check jfroginfo %s", processInfo.Id, string(by))

	var packegeInfo *types.JfrogRegistry
	err = json.Unmarshal(by, &packegeInfo)
	if err != nil {
		blog.Errorf("process %s unmarshal data %s to types.JfrogRegistry error %s", processInfo.Id, string(by), err.Error())
		return false, err
	}

	/*if uriPack.Value!=packegeInfo.DownloadUri {
		blog.Errorf("process %s DownloadUri %s is invalid",processInfo.Id,uriPack.Value)
		return fmt.Errorf("DownloadUri %s is invalid",uriPack.Value)
	}*/

	var packMd5 string
	h := md5.New()
	f, err := os.Open(uriPack.PackagesFile)
	if err != nil {
		blog.Errorf("process %s open PackagesFile %s error %s", processInfo.Id, uriPack.PackagesFile, err.Error())
		goto DownloadRESP
	}

	_, err = io.Copy(h, f)
	f.Close()
	if err != nil {
		blog.Errorf("process %s copy PackagesFile %s error %s", processInfo.Id, uriPack.Value, err.Error())
		goto DownloadRESP
	}

	packMd5 = fmt.Sprintf("%x", h.Sum(nil))
	if packMd5 == packegeInfo.Checksums.Md5 {
		blog.Infof("process %s PackagesFile %s md5 %s is valid, and not need download", processInfo.Id, uriPack.PackagesFile, packegeInfo.Checksums.Md5)
		return true, nil
	}

DownloadRESP:
	blog.Errorf("process %s PackagesFile %s md5 %s is invalid, and need download", processInfo.Id, uriPack.PackagesFile, packMd5)
	err = os.MkdirAll(filepath.Dir(uriPack.PackagesFile), 0755)
	if err != nil {
		blog.Errorf("process %s mkdir %s error %s", processInfo.Id, filepath.Dir(uriPack.PackagesFile), err.Error())
		return false, err
	}

	file, err := os.OpenFile(uriPack.PackagesFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		blog.Errorf("process %s openfile %s error %s", processInfo.Id, uriPack.PackagesFile, err.Error())
		return false, err
	}
	defer file.Close()

	by, err = m.requestManager("GET", uriPack.Value, nil, header)
	if err != nil {
		blog.Errorf("download process %s packages uri %s error %s", processInfo.Id, uriPack.Value, err.Error())
		return false, err
	}

	_, err = file.Write(by)
	if err != nil {
		blog.Errorf("process %s write file %s error %s", processInfo.Id, uriPack.PackagesFile, err.Error())
		return false, err
	}
	blog.Infof("process %s download packages %s success", processInfo.Id, uriPack.PackagesFile)
	return false, nil
}
