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
import time

from redis import WatchError


class BaseRateLimiter(object):
    def __init__(self, redisdb, identifier, tokens=None, period=None):
        """Init a RateLimiter class"""
        self.redisdb = redisdb
        self.identifier = identifier
        self.rules = []

        # Add rule
        if tokens is not None and period:
            self.add_rule(tokens, period)

        self.prepare()

    def prepare(self):
        """Prepare to work"""
        pass

    def add_rule(self, tokens, period):
        """Add multiple rules for this limiter, see `__init__` for parameter details"""
        rule = Rule(tokens, Rule.period_to_seonds(period))
        self.rules.append(rule)

    def acquire(self, tokens=1):
        """Acquire for a single request

        :param int tokens: tokens to consume for this request, default to 1
        """
        if not self.rules:
            return {'allowed': True, 'remaining_tokens': 0}

        rets = []
        for rule in self.rules:
            ret = self.acquire_by_single_rule(rule, tokens)
            if not ret['allowed']:
                return ret

            rets.append(ret)
        return {'allowed': True, 'remaining_tokens': min(x['remaining_tokens'] for x in rets)}


class RateLimiter(BaseRateLimiter):
    def prepare(self):
        self.simple_incr = self.redisdb.register_script(
            '''\
local current
current = redis.call("incr", KEYS[1])
if tonumber(current) == 1 then
    redis.call("expire", KEYS[1], ARGV[1])
end
return current'''
        )

    def acquire_by_single_rule(self, rule, tokens=1):
        """Acquire an request quota from limiter"""
        rk_counter = 'rlim::identifier::%s::rule::%s' % (self.identifier, rule.to_string())
        old_cnt = self.redisdb.get(rk_counter)
        if old_cnt is not None and int(old_cnt) >= rule.tokens:
            return {'allowed': False, 'remaining_tokens': 0.0}

        new_cnt = self.simple_incr(keys=[rk_counter], args=[rule.period_seconds])
        return {'allowed': True, 'remaining_tokens': max(0, rule.tokens - new_cnt)}


class Rule(object):
    """Rule class for RateLimiter"""

    time_unit_to_seconds = {
        'second': 1,
        'minute': 60,
        'hour': 3600,
        'day': 3600 * 24,
    }

    @classmethod
    def period_to_seonds(cls, period):
        for unit, seconds in cls.time_unit_to_seconds.items():
            if unit in period:
                period_seconds = period[unit] * seconds
                break
        else:
            raise ValueError(('Invalid period %s given, should be ' '{"second/minute/hour/day": NUMBER}') % period)
        return period_seconds

    def __init__(self, tokens, period_seconds):
        self.tokens = tokens
        # Precision of seconds only to second
        self.period_seconds = int(period_seconds)

    def to_string(self):
        return "%s_%s" % (self.tokens, self.period_seconds)

    def fresh_tokens_by_seconds(self, seconds):
        return self.rate_per_seconds * seconds

    @property
    def rate_per_seconds(self):
        return self.tokens / float(self.period_seconds)

    def __repr__(self):
        return '<Rule %s>' % self.to_string()
