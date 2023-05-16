# -*- coding: utf-8 -*-
"""
Tencent is pleased to support the open source community by making 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community
Edition) available.
Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://opensource.org/licenses/MIT

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
specific language governing permissions and limitations under the License.
"""
import re
from collections import OrderedDict

CLUSTER_IMPORT_TPL = ""

# NOTE: 因为现阶段集群版本已经较高(包含导入集群)，如果匹配不到 K8S 版本，则默认使用 v2
DEFAULT_DASHBOARD_CTL_VERSION = "v2"

DASHBOARD_CTL_VERSION = OrderedDict(
    {
        "v1": [
            re.compile(r"^[vV]?1\.8\.\S+$"),
            re.compile(r"^[vV]?1\.12\.\S+$"),
            re.compile(r"^[vV]?1\.14\.\S+$"),
            re.compile(r"^[vV]?1\.16\.\S+$"),
        ],
        "v2": [
            re.compile(r"^[vV]?1\.18\.\S+$"),
            re.compile(r"^[vV]?1\.20\.\S+$"),
            re.compile(r"^[vV]?1\.21\.\S+$"),
            re.compile(r"^[vV]?1\.22\.\S+$"),
        ],
    }
)
