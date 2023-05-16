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
import copy
import logging
from contextlib import contextmanager

from backend.metrics import Result, counter_inc

from .constants import ActivityType
from .models import UserActivityLog

logger = logging.getLogger(__name__)
undefined = object()


class UserActivityLogClient(object):
    def __init__(
        self,
        user=undefined,
        project_id=undefined,
        activity_type="note",
        resource=undefined,
        resource_type=undefined,
        resource_id=undefined,
        activity_status="completed",
        activity_time=undefined,
        description=undefined,
        extra=undefined,
    ):
        params = {
            "user": user,
            "project_id": project_id,
            "activity_type": activity_type,
            "resource": resource,
            "resource_type": resource_type,
            "resource_id": resource_id,
            "activity_status": activity_status,
            "activity_time": activity_time,
            "description": description,
            "extra": extra,
        }
        self.params = self.remove_undefined(params)
        self.activity_log = None

    def remove_undefined(self, params):
        return {k: v for k, v in params.items() if v is not undefined}

    def update_log(self, **params):
        activity_log = self.activity_log
        if not activity_log:
            return False
        for k, v in params.items():
            setattr(activity_log, k, v)
        self.activity_log.save()
        return True

    @contextmanager
    def context_log(self):
        try:
            yield self
        except Exception as err:
            counter_inc(self.activity_log.resource_type, self.activity_log.activity_type, Result.Failure.value)
            activity_status = "error"
            description = str(err)
            if not self.update_log(activity_status=activity_status, description=description):
                self.log_note(user="system", activity_status=activity_status, description=description)
            raise err
        else:
            counter_inc(self.activity_log.resource_type, self.activity_log.activity_type, Result.Success.value)

    def get_params(self, params):
        params = self.remove_undefined(params)
        _params = copy.deepcopy(self.params)
        _params.update(params)
        return _params

    def query_logs(self, start_time=None, end_time=None, **kwargs):
        params = self.get_params(kwargs)
        if start_time:
            params["activity_time__gte"] = start_time
        if end_time:
            params["activity_time__lt"] = end_time
        return UserActivityLog.objects.filter(**params)

    def log(self, **kwargs):
        params = self.get_params(kwargs)
        try:
            self.activity_log = UserActivityLog.objects.create(**params)
        except Exception as err:
            logger.exception("activity log failed: %s", err)
            raise

    def log_note(
        self,
        user=undefined,
        project_id=undefined,
        resource=undefined,
        resource_type=undefined,
        resource_id=undefined,
        activity_status="completed",
        description=undefined,
        **kw
    ):
        return self.log(
            activity_type="note",
            user=user,
            project_id=project_id,
            resource=resource,
            resource_type=resource_type,
            resource_id=resource_id,
            activity_status=activity_status,
            description=description,
            **kw
        )

    def log_add(
        self,
        user=undefined,
        project_id=undefined,
        resource=undefined,
        resource_type=undefined,
        resource_id=undefined,
        activity_status="completed",
        description=undefined,
        **kw
    ):
        return self.log(
            activity_type="add",
            user=user,
            project_id=project_id,
            resource=resource,
            resource_type=resource_type,
            resource_id=resource_id,
            activity_status=activity_status,
            description=description,
            **kw
        )

    def log_modify(
        self,
        user=undefined,
        project_id=undefined,
        resource=undefined,
        resource_type=undefined,
        resource_id=undefined,
        activity_status="completed",
        description=undefined,
        **kw
    ):
        return self.log(
            activity_type="modify",
            user=user,
            project_id=project_id,
            resource=resource,
            resource_type=resource_type,
            resource_id=resource_id,
            activity_status=activity_status,
            description=description,
            **kw
        )

    def log_scale(
        self,
        user=undefined,
        project_id=undefined,
        resource=undefined,
        resource_type=undefined,
        resource_id=undefined,
        activity_status="completed",
        description=undefined,
        **kw
    ):
        return self.log(
            activity_type=ActivityType.Scale.value,
            user=user,
            project_id=project_id,
            resource=resource,
            resource_type=resource_type,
            resource_id=resource_id,
            activity_status=activity_status,
            description=description,
            **kw
        )

    def log_recreate(
        self,
        user=undefined,
        project_id=undefined,
        resource=undefined,
        resource_type=undefined,
        resource_id=undefined,
        activity_status="completed",
        description=undefined,
        **kw
    ):
        return self.log(
            activity_type=ActivityType.Recreate.value,
            user=user,
            project_id=project_id,
            resource=resource,
            resource_type=resource_type,
            resource_id=resource_id,
            activity_status=activity_status,
            description=description,
            **kw
        )

    def log_delete(
        self,
        user=undefined,
        project_id=undefined,
        resource=undefined,
        resource_type=undefined,
        resource_id=undefined,
        activity_status="completed",
        description=undefined,
        **kw
    ):
        return self.log(
            activity_type="delete",
            user=user,
            project_id=project_id,
            resource=resource,
            resource_type=resource_type,
            resource_id=resource_id,
            activity_status=activity_status,
            description=description,
            **kw
        )

    def log_begin(
        self,
        user=undefined,
        project_id=undefined,
        resource=undefined,
        resource_type=undefined,
        resource_id=undefined,
        activity_status="completed",
        description=undefined,
        **kw
    ):
        return self.log(
            activity_type="begin",
            user=user,
            project_id=project_id,
            resource=resource,
            resource_type=resource_type,
            resource_id=resource_id,
            activity_status=activity_status,
            description=description,
            **kw
        )

    def log_end(
        self,
        user=undefined,
        project_id=undefined,
        resource=undefined,
        resource_type=undefined,
        resource_id=undefined,
        activity_status="completed",
        description=undefined,
        **kw
    ):
        return self.log(
            activity_type="end",
            user=user,
            project_id=project_id,
            resource=resource,
            resource_type=resource_type,
            resource_id=resource_id,
            activity_status=activity_status,
            description=description,
            **kw
        )

    def log_start(
        self,
        user=undefined,
        project_id=undefined,
        resource=undefined,
        resource_type=undefined,
        resource_id=undefined,
        activity_status="completed",
        description=undefined,
        **kw
    ):
        return self.log(
            activity_type="start",
            user=user,
            project_id=project_id,
            resource=resource,
            resource_type=resource_type,
            resource_id=resource_id,
            activity_status=activity_status,
            description=description,
            **kw
        )

    def log_pause(
        self,
        user=undefined,
        project_id=undefined,
        resource=undefined,
        resource_type=undefined,
        resource_id=undefined,
        activity_status="completed",
        description=undefined,
        **kw
    ):
        return self.log(
            activity_type="pause",
            user=user,
            project_id=project_id,
            resource=resource,
            resource_type=resource_type,
            resource_id=resource_id,
            activity_status=activity_status,
            description=description,
            **kw
        )

    def log_carryon(
        self,
        user=undefined,
        project_id=undefined,
        resource=undefined,
        resource_type=undefined,
        resource_id=undefined,
        activity_status="completed",
        description=undefined,
        **kw
    ):
        return self.log(
            activity_type="carryon",
            user=user,
            project_id=project_id,
            resource=resource,
            resource_type=resource_type,
            resource_id=resource_id,
            activity_status=activity_status,
            description=description,
            **kw
        )

    def log_stop(
        self,
        user=undefined,
        project_id=undefined,
        resource=undefined,
        resource_type=undefined,
        resource_id=undefined,
        activity_status="completed",
        description=undefined,
        **kw
    ):
        return self.log(
            activity_type="stop",
            user=user,
            project_id=project_id,
            resource=resource,
            resource_type=resource_type,
            resource_id=resource_id,
            activity_status=activity_status,
            description=description,
            **kw
        )

    def log_restart(
        self,
        user=undefined,
        project_id=undefined,
        resource=undefined,
        resource_type=undefined,
        resource_id=undefined,
        activity_status="completed",
        description=undefined,
        **kw
    ):
        return self.log(
            activity_type="restart",
            user=user,
            project_id=project_id,
            resource=resource,
            resource_type=resource_type,
            resource_id=resource_id,
            activity_status=activity_status,
            description=description,
            **kw
        )

    def log_query(
        self,
        user=undefined,
        project_id=undefined,
        resource=undefined,
        resource_type=undefined,
        resource_id=undefined,
        activity_status="completed",
        description=undefined,
        **kw
    ):
        return self.log(
            activity_type="query",
            user=user,
            project_id=project_id,
            resource=resource,
            resource_type=resource_type,
            resource_id=resource_id,
            activity_status=activity_status,
            description=description,
            **kw
        )


class ContextActivityLogClient(UserActivityLogClient):
    def log(self, **kwargs):
        super().log(**kwargs)
        return self.context_log()


def get_log_client_by_activity_log_id(activity_log_id):
    log_client = UserActivityLogClient()
    log_client.activity_log = UserActivityLog.objects.get(id=activity_log_id)
    return log_client
