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

存放一些 URL 相关小工具 & 常量
"""

# 默认的 URL 占位符，用于 url 参数无需指定的情况，如 /pod/-/containers/
URL_DEFAULT_PLACEHOLDER = '-'

# k8s 命名空间格式
NAMESPACE_REGEX = "[a-z0-9]([-a-z0-9]*[a-z0-9])?"

# k8s 资源名称格式
KUBE_NAME_REGEX = "[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"

# IPV4 正则表达式
IPV4_REGEX = r'(((\d{1,2})|(1\d{2})|(2[0-4]\d)|(25[0-5]))\.){3}((\d{1,2})|(1\d{2})|(2[0-4]\d)|(25[0-5]))'
