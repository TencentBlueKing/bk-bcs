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
from django.conf import settings

# 格式化器 单元测试目录
FORMATTER_UNITTEST_DIR = f'{settings.BASE_DIR}/backend/tests/resources/formatter/'

# 网络 类配置存放路径
NETWORK_CONFIG_DIR = f'{FORMATTER_UNITTEST_DIR}/networks/contents'

# 存储 类配置存放路径
STORAGE_CONFIG_DIR = f'{FORMATTER_UNITTEST_DIR}/storages/contents'

# 工作负载 类配置存放路径
WORKLOAD_CONFIG_DIR = f'{FORMATTER_UNITTEST_DIR}/workloads/contents'
