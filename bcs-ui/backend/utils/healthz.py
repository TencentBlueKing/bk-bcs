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

from django.http import JsonResponse

logger = logging.getLogger(__name__)

PREFIX = 'handler_'


class HealthzError(Exception):
    pass


def healthz_view(request):
    data = healthz_manager()
    result = {'data': data, 'result': True, 'message': ''}
    return JsonResponse(result)


def healthz_manager():
    data = {}
    handlers = filter(lambda x: x[0].startswith(PREFIX), globals().items())
    for handler in handlers:
        name = handler[0][len(PREFIX) :]
        func = handler[1]
        # 默认返回0
        data[name] = 0
        try:
            func()
        except HealthzError as error:
            data[name] = '%s' % error
        except Exception as error:
            data[name] = '%s' % error
            logger.exception('healthz handler %s error', name)
    return data


def handler_redis():
    from backend.utils.cache import rd_client

    if rd_client.set("__healthz__", 1) is False:
        raise HealthzError("redis set command failed")
    if rd_client.get("__healthz__") != b'1':
        raise HealthzError("redis get command failed")


def handler_gcs_mysql():
    from backend.container_service.projects.models import ProjectUser

    ProjectUser.objects.first()


def main():
    print(healthz_manager())


def test_sentry(request):
    a = int('')
    return JsonResponse({'test': a})


if __name__ == '__main__':
    main()
