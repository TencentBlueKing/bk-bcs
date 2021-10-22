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
from rest_framework import serializers
from rest_framework.exceptions import ValidationError

from .constants import ViewMode
from .featflag import UNSELECTED_CLUSTER, ClusterFeatureType


class ClusterFeatureTypeSLZ(serializers.Serializer):
    cluster_id = serializers.CharField()
    cluster_feature_type = serializers.ChoiceField(choices=ClusterFeatureType.get_choices(), required=False)
    view_mode = serializers.ChoiceField(
        choices=ViewMode.get_choices(), default=ViewMode.ClusterManagement, required=False
    )

    def validate(self, data):
        # cluster_id 为 -, 表示未指定具体集群
        if data['cluster_id'] != UNSELECTED_CLUSTER and 'cluster_feature_type' not in data:
            raise ValidationError("missing valid parameter cluster_feature_type")
        return data
