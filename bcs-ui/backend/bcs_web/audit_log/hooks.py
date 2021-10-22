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
from collections import OrderedDict

from django.db.models.signals import post_delete, post_save
from django.dispatch import receiver
from django.utils.encoding import force_str

from backend.utils.local import local

from .client import UserActivityLogClient

logger = logging.getLogger(__name__)


class DjangoSignalHooks(object):
    def __init__(self):
        self.ignore_models = set()
        self.model_checkers = OrderedDict()
        self.activity_log_client = UserActivityLogClient(activity_status="completed")

    def activity_check(self, params):
        return params

    def register(self, model):
        checker = getattr(model, "activity_check")
        if isinstance(checker, str):
            checker = getattr(model, checker)
        if checker:
            self.model_checkers[model] = checker

    def find_checker_by_instance(self, instance):
        model_cls = instance.__class__
        checker = self.model_checkers.get(model_cls)
        if checker:
            return checker
        if model_cls in self.ignore_models:
            return
        for model, checker in self.model_checkers.items():
            if isinstance(instance, model):
                return checker
        self.ignore_models.add(model_cls)

    def setup_hook(self):
        @receiver(post_save)
        def log_model_activity_create_and_modify(sender, instance, created, **kwargs):
            if created:
                activity_type = "add"
            else:
                activity_type = "modify"

            check_and_log(instance, dict(activity_type=activity_type))

        @receiver(post_delete)
        def log_model_activity_delete(sender, instance, **kwargs):
            check_and_log(instance, dict(activity_type="delete"))

        def check_and_log(instance, params):
            checker = self.find_checker_by_instance(instance)

            try:
                username = local.request.user.username or force_str(local.request.user)
            except Exception:
                username = "*SYSTEM*"  # celery backend process
            params.setdefault("user", username)

            if not checker:
                return
            try:
                params = checker(params)
                if params is not None:
                    self.activity_log_client.log(**params)
            except Exception as err:
                logger.exception("activity model log error: %s", err)
                return


SignalActivityLogHook = DjangoSignalHooks()
