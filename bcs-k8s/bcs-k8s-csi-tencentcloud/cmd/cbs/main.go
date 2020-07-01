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

package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-k8s-csi-tencentcloud/driver/cbs"
	"github.com/dbdd4us/qcloudapi-sdk-go/metadata"
	"github.com/golang/glog"
)

const (
	TENCENTCLOUD_CBS_API_SECRET_ID  = "TENCENTCLOUD_CBS_API_SECRET_ID"
	TENCENTCLOUD_CBS_API_SECRET_KEY = "TENCENTCLOUD_CBS_API_SECRET_KEY"
)

var (
	endpoint  = flag.String("endpoint", fmt.Sprintf("unix:///var/lib/kubelet/plugins/%s/csi.sock", cbs.DriverName), "CSI endpoint")
	secretId  = flag.String("secret_id", "", "tencent cloud api secret id")
	secretKey = flag.String("secret_key", "", "tencent cloud api secret key")
	region    = flag.String("region", "", "tencent cloud api region")
	zone      = flag.String("zone", "", "cvm instance region")
	cbsUrl    = flag.String("cbs_url", "cbs.internal.tencentcloudapi.com", "cbs api domain")
)

func main() {
	flag.Parse()
	defer glog.Flush()

	metadataClient := metadata.NewMetaData(http.DefaultClient)

	if *region == "" {
		r, err := metadataClient.Region()
		if err != nil {
			glog.Fatal(err)
		}
		region = &r
	}
	if *zone == "" {
		z, err := metadataClient.Zone()
		if err != nil {
			glog.Fatal(err)
		}
		zone = &z
	}

	if *secretId == "" {
		if secretIdFromEnv := os.Getenv(TENCENTCLOUD_CBS_API_SECRET_ID); secretIdFromEnv != "" {
			secretId = &secretIdFromEnv
		}
	}
	if *secretKey == "" {
		if secretKeyFromEnv := os.Getenv(TENCENTCLOUD_CBS_API_SECRET_KEY); secretKeyFromEnv != "" {
			secretKey = &secretKeyFromEnv
		}
	}

	if *secretId == "" || *secretKey == "" {
		glog.Fatal("tencent cloud credential must be specified")
	}

	u, err := url.Parse(*endpoint)
	if err != nil {
		glog.Fatalf("parse endpoint err: %s", err.Error())
	}

	if u.Scheme != "unix" {
		glog.Fatal("only unix socket is supported currently")
	}

	drv, err := cbs.NewDriver(*region, *zone, *secretId, *secretKey)
	if err != nil {
		glog.Fatal(err)
	}

	if err := drv.Run(u, *cbsUrl); err != nil {
		glog.Fatal(err)
	}

	return
}
