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
from rest_framework.decorators import action
from rest_framework.response import Response

from backend.bcs_web.viewsets import SystemViewSet
from backend.dashboard.examples.serializers import FetchResourceDemoManifestSLZ
from backend.dashboard.examples.utils import load_demo_manifest, load_resource_references, load_resource_template
from backend.utils.i18n import get_lang_from_cookies


class TemplateViewSet(SystemViewSet):
    """模板相关接口"""

    @action(methods=['GET'], url_path='manifests', detail=False)
    def manifests(self, request, project_id, cluster_id):
        """指定资源类型的 Demo 配置信息"""
        params = self.params_validate(FetchResourceDemoManifestSLZ)
        lang = get_lang_from_cookies(request.COOKIES)
        config = load_resource_template(params['kind'], lang)
        config['references'] = load_resource_references(params['kind'], lang)
        for t in config['items']:
            t['manifest'] = load_demo_manifest(f"{config['class']}/{t['name']}")
        return Response(config)
