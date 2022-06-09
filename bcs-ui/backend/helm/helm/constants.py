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
import logging

from backend.utils.basic import ChoicesEnum

logger = logging.getLogger(__name__)

# release 名称，随机值长度，用于前端区分不同的release
CHART_RELEASE_SHOT_NAME_LENGTH = 6

# 用于更新app时选择不修改模板，
# 通过改选项可以保证模板是上次发布时归档的模板内容，
# 而不是当前模板id对应的最新内容
KEEP_TEMPLATE_UNCHANGED = -1
# release 版本的前缀，后面跟着真正的版本；用以区分release版本和仓库中版本，因为release版本和仓库中版本的values内容可能不一样用
RELEASE_VERSION_PREFIX = "(release-version)"

# app 与 release 有个创建顺序问题，release 先创建，
# 但是会指向 app，因此会先填个默认值 -1 ，待 app 创建好了再改成真实的 app id
TEMPORARY_APP_ID = -1


class ChartReleaseTypes(ChoicesEnum):
    RELEASE = "release"
    ROLLBACK = "rollback"
    _choices_labels = (
        (RELEASE, "release"),
        (ROLLBACK, "rollback"),
    )


RESOURCE_NAME_REGEX = r'^[a-z0-9]([-a-z0-9]*[a-z0-9])?$'


# default helm value file name
DEFAULT_VALUES_FILE_NAME = 'values.yaml'


# Harbor chart仓库项目名称
DEFAULT_CHART_REPO_PROJECT_NAME = ""
