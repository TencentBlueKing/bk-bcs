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
import json
import logging

from backend.components import paas_cc

logger = logging.getLogger(__name__)


def get_cluster_version(access_token, project_id, cluster_id):
    snapshot = paas_cc.get_cluster_snapshot(access_token, project_id, cluster_id)
    try:
        configure = json.loads(snapshot["data"]["configure"])
        version = configure["version"]  # "1.12.3"
    except Exception as e:
        version = ''
        logger.exception("get_cluster_version failed, %s", e)
    return version
