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
from typing import Dict, Optional

from celery import Celery
from celery.schedules import crontab
from django.apps import AppConfig
from django.conf import settings  # noqa

logger = logging.getLogger(__name__)


def get_celery_major_version() -> Optional[int]:
    """获取 Celery 大版本号，比如 3.1.0 应该返回 3"""
    from celery import __version__

    try:
        return int(__version__.split('.')[0])
    except Exception as e:
        logger.debug('Unable to get major version for celery, error: %s', e)
        return None


app = Celery('backend')

major_ver = get_celery_major_version()

if major_ver and major_ver > 3:
    app.config_from_object('django.conf:settings', namespace='CELERY')
    app.autodiscover_tasks()
else:
    # 目前需要同时兼容 Celery 3 与更高版本，celery 3 不需要指定 namespace 属性
    # 详见：https://stackoverflow.com/questions/54834228/
    # how-to-solve-the-a-celery-worker-configuration-with-keyword-argument-namespace
    app.config_from_object('django.conf:settings')


class CeleryConfig(AppConfig):
    name = "backend.celery_app"
    verbose_name = "celery_app"

    def ready(self):
        from backend.container_service.infras.hosts.terraform import tasks as host_tasks  # noqa
        from backend.helm.app import tasks as helm_app_tasks  # noqa
        from backend.helm.helm import tasks as helm_chart_tasks  # noqa
        from backend.packages.blue_krill.async_utils import poll_task  # noqa
        from backend.templatesets.legacy_apps.configuration import tasks as backend_instance_status  # noqa
        from backend.utils import notify  # noqa
