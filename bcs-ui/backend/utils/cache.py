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
import redis
from django.conf import settings
from dogpile.cache import make_region

rd_pool = redis.ConnectionPool(
    host=settings.REDIS_HOST, port=settings.REDIS_PORT, db=settings.REDIS_DB, password=settings.REDIS_PASSWORD
)

rd_client = redis.Redis(connection_pool=rd_pool, health_check_interval=30)


region = make_region().configure(
    'dogpile.cache.redis',
    expiration_time=7200,
    arguments={'connection_pool': rd_pool, 'redis_expiration_time': 7 * 24 * 3600},  # 最长cache时间
)
