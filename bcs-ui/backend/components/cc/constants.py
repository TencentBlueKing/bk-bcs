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
# 默认查询主机字段
DEFAULT_HOST_FIELDS = [
    'bk_bak_operator',
    'classify_level_name',
    'svr_device_class',
    'bk_svr_type_id',
    'svr_type_name',
    'hard_memo',
    'bk_host_id',
    'bk_host_name',
    'idc_name',
    'bk_idc_area',
    'bk_idc_area_id',
    'idc_id',
    'idc_unit_name',
    'idc_unit_id',
    'bk_host_innerip',
    'bk_comment',
    'module_name',
    'operator',
    'bk_os_name',
    'bk_os_version',
    'bk_host_outerip',
    'rack',
    'bk_cloud_id',
]

# 默认从 0 开始查询
DEFAULT_START_AT = 0

# 用于查询 count 的 Limit，最小为 1
LIMIT_FOR_COUNT = 1

# CMDB 通用的最大 Limit 限制
CMDB_MAX_LIMIT = 200

# CMDB 请求主机列表最大 Limit 限制
CMDB_LIST_HOSTS_MAX_LIMIT = 500
