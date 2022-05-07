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

from django.conf import settings
from rest_framework.exceptions import PermissionDenied

from backend.components import paas_cc
from backend.helm.app.utils import yaml_dump, yaml_load
from backend.helm.permissions import check_cluster_perm

from .base import BaseProvider

logger = logging.getLogger(__name__)


class SingleHelmAppUpdateProvider(BaseProvider):
    """a provider which support update helm app with token"""

    NAME = "Single Helm App Update"
    CONFIG_SCHEMA = {
        "type": "object",
        "required": ["cluster_id", "app_id"],
        "properties": {
            "cluster_id": {"type": "string", "maxLength": 32, "minLength": len("BCS-K8s-"), "pattern": "[\w\-\d]+"},
            "app_id": {
                "type": "number",
                "minimum": 1,
                "maximum": 10**6,
            },
        },
    }
    REQUEST_SCHEMA = CONFIG_SCHEMA

    def validate_project_id(self, user, project_id):
        result = paas_cc.get_project(user.token.access_token, project_id)
        if result.get('code') != 0:
            raise PermissionDenied("project not found")

    @staticmethod
    def provide(user, config):
        from backend.helm.app.models import App
        from backend.utils.client import get_kubectl_config_context

        if not App.objects.filter(
            project_id=config.get("project_id"), cluster_id=config.get("cluster_id"), id=config.get("app_id")
        ).exists():
            raise PermissionDenied()

        check_cluster_perm(user, config.get("project_id"), config.get("cluster_id"))

        # bke_client = get_bke_client(
        #     project_id=config.get("project_id"),
        #     cluster_id=config.get("cluster_id"),
        #     access_token=user.token.access_token
        # )
        # bke_client.active_admin()

        kubeconfig = get_kubectl_config_context(
            access_token=user.token.access_token,
            project_id=config.get("project_id"),
            cluster_id=config.get("cluster_id"),
        )
        kubeconfig_obj = yaml_load(kubeconfig)
        kubeconfig_obj["users"][0]["user"]["token"] = settings.BKE_ADMIN_TOKEN
        config["kubeconfig"] = yaml_dump(kubeconfig_obj)
        return config

    @staticmethod
    def validate(token, request_data):
        config = token.config
        if not all(
            [
                config["project_id"] == request_data["project_id"],
                config["cluster_id"] == request_data["cluster_id"],
                config["app_id"] == request_data["app_id"],
            ]
        ):
            raise PermissionDenied()
