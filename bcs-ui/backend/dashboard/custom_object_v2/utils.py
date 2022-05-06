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
from typing import Dict, Union

from django.utils.translation import ugettext_lazy as _

from backend.dashboard.exceptions import ResourceNotExist
from backend.dashboard.utils.web import gen_base_web_annotations
from backend.resources.custom_object import CustomResourceDefinition


def gen_cobj_web_annotations(request, project_id, cluster_id, namespace: Union[str, None], crd_name: str) -> Dict:
    """构造 custom_object 相关 web_annotations"""
    client = CustomResourceDefinition(request.ctx_cluster)
    crd = client.get(name=crd_name, is_format=False)
    if not crd:
        raise ResourceNotExist(_('集群 {} 中未注册自定义资源 {}').format(cluster_id, crd_name))
    # 先获取基础的，仅包含权限信息的 web_annotations
    web_annotations = gen_base_web_annotations(request.user.username, project_id, cluster_id, namespace)
    web_annotations.update({'additional_columns': crd.additional_columns})
    return web_annotations
