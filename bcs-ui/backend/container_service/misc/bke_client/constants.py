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

BCS_USER_NAME = "bke-dynamic-user"

# cluster not found code
CLUSTER_NOT_FOUND_CODE_NAME = 'CLUSTER_NOT_FOUND'

# perm fail code
CLUSTER_PERM_FAIL_CODE_NAMES = ['CHECK_USER_CLUSTER_PERM_FAIL', 'UNAUTHORIZED']

# token not found code
TOKEN_NOT_FOUND_CODE_NAME = 'RTOKEN_NOT_FOUND'

# cluster exist code
CLUSTER_EXIST_CODE_NAME = 'CLUSTER_ALREADY_EXISTS'

# credentials not found code
CREDENTIALS_NOT_FOUND_CODE_NAME = 'CREDENTIALS_NOT_FOUND'

# default kubectl version
DEFAULT_KUBECTL_VERSION = '1.20.13'

# KUBECTL VERSION
KUBECTL_VERSION = OrderedDict(
    {
        "1.20.13": [re.compile(r"^[vV]?1\.20\.\S+$")],
        "1.18.12": [re.compile(r"^[vV]?1\.18\.\S+$")],
        "1.16.3": [re.compile(r"^[vV]?1\.16\.\S+$")],
        "1.14.9": [re.compile(r"^[vV]?1\.14\.\S+$")],
        "1.12.3": [re.compile(r"^[vV]?1\.12\.\S+$")],
        "1.8.3": [re.compile(r"^[vV]?1\.8\.\S+$")],
    }
)
