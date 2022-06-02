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

package api

import (
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
)

// market images
// 没有特定接口，按照 tke 节点池页面数据硬编码
var imageOsList = []*proto.OsImage{
	{
		Alias:           "CentOS 7.2 64bit",
		Arch:            "x86_64",
		ImageID:         "img-rkiynh11",
		OsCustomizeType: "GENERAL",
		OsName:          "centos7.2x86_64",
		SeriesName:      "centos7.2x86_64",
		Status:          "NORMAL",
		Provider:        common.MarketImageProvider,
	},
	{
		Alias:           "Ubuntu Server 16.04.1 LTS 64bit",
		Arch:            "x86_64",
		ImageID:         "img-4wpaazux",
		OsCustomizeType: "GENERAL",
		OsName:          "ubuntu16.04.1 LTSx86_64",
		SeriesName:      "ubuntu16.04.1 LTSx86_64",
		Status:          "NORMAL",
		Provider:        common.MarketImageProvider,
	},
	// https://console.cloud.tencent.com/cvm/image/detail?rid=1&id=img-fv2263iz
	{
		Alias:           "TencentOS Server 3.1 (TK4)",
		Arch:            "x86_64",
		ImageID:         "img-fv2263iz",
		OsCustomizeType: "GENERAL",
		OsName:          "tlinux3.1x86_64",
		SeriesName:      "TencentOS Server 3.1 (TK4)",
		Status:          "NORMAL",
		Provider:        common.MarketImageProvider,
	},
	// https://console.cloud.tencent.com/cvm/image/detail?rid=1&id=img-jebhne9p
	{
		Alias:           "TencentOS Server 3.1 (TK4)(支持混部)",
		Arch:            "x86_64",
		ImageID:         "img-jebhne9p",
		OsCustomizeType: "GENERAL",
		OsName:          "tlinux3.1x86_64",
		SeriesName:      "TencentOS Server 3.1 (TK4)",
		Status:          "NORMAL",
		Provider:        common.MarketImageProvider,
	},
	// https://console.cloud.tencent.com/cvm/image/detail?rid=1&id=img-1isywgop
	{
		Alias:           "Tencent Linux release 3.2 (支持混部)",
		Arch:            "x86_64",
		ImageID:         "img-1isywgop",
		OsCustomizeType: "GENERAL",
		OsName:          "tlinux3.2x86_64",
		SeriesName:      "TencentOS Server 3.2 (Final)",
		Status:          "NORMAL",
		Provider:        common.MarketImageProvider,
	},
}
