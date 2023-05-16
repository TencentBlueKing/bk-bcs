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


class DashboardBaseError(Exception):
    """Dashboard 模块基础异常类"""

    message = 'Dashboard module exception'
    # 子异常在基类基础上，自带两位错误码后缀
    code = 40050

    def __init__(self, message=None, code=None):
        """初始化异常类，若无参数则使用默认值"""
        if message:
            self.message = message
        if code:
            self.code = code

    def __str__(self):
        return f'{self.code}: {self.message}'


class ResourceNotExist(DashboardBaseError):
    """指定资源不存在"""

    message = 'Resource not exist'
    code = 4005001


class CreateResourceError(DashboardBaseError):
    """创建资源失败"""

    message = 'Create Resource Error'
    # NOTE 前端对此错误码有特殊逻辑
    code = 4005002


class UpdateResourceError(DashboardBaseError):
    """更新资源失败"""

    message = 'Update Resource Error'
    # NOTE 前端对此错误码有特殊逻辑
    code = 4005003


class DeleteResourceError(DashboardBaseError):
    """删除资源失败"""

    message = 'Delete Resource Error'
    code = 4005004


class ResourceVersionExpired(DashboardBaseError):
    """ResourceVersion 过期"""

    message = 'ResourceVersion Expired'
    # NOTE 前端对此错误码有特殊逻辑
    code = 4005005


class OwnerReferencesNotExist(DashboardBaseError):
    """不存在父级资源"""

    message = "OwnerReferences Not Exist"
    code = 4005006


class ActionUnsupported(DashboardBaseError):
    """不支持指定的 Action"""

    message = "action unsupported"
    code = 4005007


class ResourceTypeUnsupported(DashboardBaseError):
    """不支持的用于资源视图鉴权的资源类型"""

    message = "resource type for dashboard perm validate unsupported"
    code = 4005008
