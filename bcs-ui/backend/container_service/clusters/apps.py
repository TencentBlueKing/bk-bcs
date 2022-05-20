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
from django.apps import AppConfig
from django.conf import settings

from backend.packages.blue_krill.data_types import enum

from .featureflag.featflags import cluster_mgr


def register_to_app():
    """注册额外的功能"""
    ext_feature_flags = [enum.FeatureFlagField(name='CLOUDTOKEN', label='云凭证', default=True)]
    for ext_feat_flag in ext_feature_flags:
        cluster_mgr.GlobalClusterFeatureFlag.register_feature_flag(ext_feat_flag)
        cluster_mgr.SingleClusterFeatureFlag.register_feature_flag(ext_feat_flag)


class ClusterConfig(AppConfig):
    name = 'backend.container_service.clusters'
    # 与重构前应用 label "cluster" 保持兼容
    label = 'cluster'

    def ready(self):
        # Multi-editions specific start
        if settings.EDITION == settings.COMMUNITY_EDITION:
            register_to_app()
            return

        try:
            from .apps_ext import contribute_to_app

            contribute_to_app(self.name)
        except ImportError:
            pass

        # Multi-editions specific end
