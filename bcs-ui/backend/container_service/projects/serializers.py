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

from backend.container_service.projects.base.constants import ProjectKindID
from backend.iam.legacy_perms import ProjectActions


class UpdateProjectNewSLZ(serializers.Serializer):
    """更新项目的参数"""

    kind = serializers.ChoiceField(choices=[ProjectKindID], required=False)
    cc_app_id = serializers.IntegerField(required=False, min_value=1)


class UpdateNavProjectSLZ(serializers.Serializer):
    project_name = serializers.CharField()
    description = serializers.CharField()
    updator = serializers.CharField()


class CreateNavProjectSLZ(serializers.Serializer):
    project_name = serializers.CharField()
    english_name = serializers.CharField()
    creator = serializers.CharField()
    description = serializers.CharField(required=False)


class ProjectPermsSLZ(serializers.Serializer):
    with_apply_url = serializers.BooleanField(default=False)
    action_ids = serializers.ListField(
        child=serializers.ChoiceField(choices=[ProjectActions.CREATE.value]), default=[ProjectActions.CREATE.value]
    )


class ProjectInstPermsSLZ(ProjectPermsSLZ):
    action_ids = serializers.ListField(
        child=serializers.ChoiceField(choices=[ProjectActions.EDIT.value, ProjectActions.VIEW.value]),
        default=[ProjectActions.EDIT.value, ProjectActions.VIEW.value],
    )
    project_id = serializers.CharField()


class QueryAuthorizedUsersSLZ(serializers.Serializer):
    action_id = serializers.ChoiceField(
        choices=[ProjectActions.CREATE.value, ProjectActions.EDIT.value, ProjectActions.VIEW.value],
        default=ProjectActions.VIEW.value,
    )
