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
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
)

// market images
// ImageOsList image list
var ImageOsList = []*proto.OsImage{
	{
		Alias:           "TencentOS Server 3.1 (TK4)",
		Arch:            "x86_64",
		ImageID:         "img-eb30mz89",
		OsCustomizeType: "GENERAL",
		OsName:          "tlinux3.1x86_64",
		SeriesName:      "TencentOS Server 3.1 (TK4)",
		Status:          "NORMAL",
		Provider:        common.PublicImageProvider,
	},
	{
		Alias:           "TencentOS Server 2.4",
		Arch:            "x86_64",
		ImageID:         "img-hdt9xxkt",
		OsCustomizeType: "GENERAL",
		OsName:          "tlinux2.4x86_64",
		SeriesName:      "TencentOS Server 2.4",
		Status:          "NORMAL",
		Provider:        common.PublicImageProvider,
	},
	{
		Alias:           "Ubuntu Server 20.04.1 LTS 64bit",
		Arch:            "x86_64",
		ImageID:         "img-22trbn9x",
		OsCustomizeType: "GENERAL",
		OsName:          "ubuntu20.04x86_64",
		SeriesName:      "ubuntu20.04x86_64",
		Status:          "NORMAL",
		Provider:        common.PublicImageProvider,
	},
	{
		Alias:           "Ubuntu Server 18.04.1 LTS 64bit",
		Arch:            "x86_64",
		ImageID:         "img-pi0ii46r",
		OsCustomizeType: "GENERAL",
		OsName:          "ubuntu18.04.1x86_64",
		SeriesName:      "ubuntu18.04.1x86_64",
		Status:          "NORMAL",
		Provider:        common.PublicImageProvider,
	},
	{
		Alias:           "Ubuntu Server 16.04.1 LTS 64bit",
		Arch:            "x86_64",
		ImageID:         "img-4wpaazux",
		OsCustomizeType: "GENERAL",
		OsName:          "ubuntu16.04.1 LTSx86_64",
		SeriesName:      "ubuntu16.04.1 LTSx86_64",
		Status:          "NORMAL",
		Provider:        common.PublicImageProvider,
	},
	{
		Alias:           "CentOS 8.0 64bit",
		Arch:            "x86_64",
		ImageID:         "img-25szkc8t",
		OsCustomizeType: "GENERAL",
		OsName:          "centos8.0x86_64",
		SeriesName:      "centos8.0x86_64",
		Status:          "NORMAL",
		Provider:        common.PublicImageProvider,
	},
	{
		Alias:           "CentOS 7.8 64bit",
		Arch:            "x86_64",
		ImageID:         "img-3la7wgnt",
		OsCustomizeType: "GENERAL",
		OsName:          "centos7.8.0_x64",
		SeriesName:      "centos7.8.0_x64",
		Status:          "NORMAL",
		Provider:        common.PublicImageProvider,
	},
	{
		Alias:           "CentOS 7.6 64bit",
		Arch:            "x86_64",
		ImageID:         "img-9qabwvbn",
		OsCustomizeType: "GENERAL",
		OsName:          "centos7.6.0_x64",
		SeriesName:      "centos7.6.0_x64",
		Status:          "NORMAL",
		Provider:        common.PublicImageProvider,
	},
	{
		Alias:           "CentOS 7.2 64bit",
		Arch:            "x86_64",
		ImageID:         "img-rkiynh11",
		OsCustomizeType: "GENERAL",
		OsName:          "centos7.2x86_64",
		SeriesName:      "centos7.2x86_64",
		Status:          "NORMAL",
		Provider:        common.PublicImageProvider,
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
