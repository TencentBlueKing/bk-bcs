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

import redis
from logstash import formatter

from backend.utils.local import local


class LogstashRedisHandler(logging.Handler):
    def __init__(self, redis_url, queue_name='', message_type='logstash', tags=None):
        """
        初始化，延迟时间默认1秒钟
        """
        logging.Handler.__init__(self)
        self.queue_name = queue_name
        pool = redis.BlockingConnectionPool.from_url(redis_url, max_connections=600, timeout=1)
        self.client = redis.Redis(connection_pool=pool, health_check_interval=30)

        self.formatter = formatter.LogstashFormatterVersion1(message_type, tags, fqdn=False)

    def emit(self, record):
        """
        提交数据
        """
        try:
            self.client.rpush(self.queue_name, self.formatter.format(record))
        except Exception:
            logger.exception('LogstashRedisHandler push to redis error')


# 必须是非redis logger,否则循环错误
logger = logging.getLogger('console')


class RequestIdFilter(logging.Filter):
    def filter(self, record):
        record.request_id = local.request_id
        return True
