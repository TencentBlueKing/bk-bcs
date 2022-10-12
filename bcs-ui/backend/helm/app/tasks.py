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

from celery import shared_task, current_task

from backend.utils.local import local
from django.conf import settings

from .models import App

logger = logging.getLogger(__name__)


@shared_task
def destroy_app(app_id, access_token, username):
    App.objects.get(id=app_id).destroy_app_task(username, access_token)


@shared_task
def rollback_app(app_id, access_token, username, release_id):
    App.objects.get(id=app_id).rollback_app_task(username=username, access_token=access_token, release_id=release_id)


@shared_task
def upgrade_app(app_id, **kwargs):
    access_token = kwargs['access_token']
    local.new_dummy_request(access_token, "")

    app = App.objects.get(id=app_id)
    logger.info(
        "upgrading app task id %s, release detail: cluster_id: %s, namespace: %s name: %s",
        current_task.request.id,
        app.cluster_id,
        app.namespace,
        app.name,
    )
    app.upgrade_app_task(**kwargs)

    # 需要手动释放
    local.release()


@shared_task
def first_deploy(app_id, access_token, activity_log_id, deploy_options):
    local.new_dummy_request(access_token, "")

    App.objects.get(id=app_id).first_deploy_task(
        access_token=access_token,
        deploy_options=deploy_options,
        activity_log_id=activity_log_id,
    )

    # 需要手动释放
    local.release()


def sync_or_async(task_method):
    if settings.HELM_SYNC_DO_DEPLOY:
        return getattr(task_method, "apply")
    else:
        return getattr(task_method, "apply_async")
