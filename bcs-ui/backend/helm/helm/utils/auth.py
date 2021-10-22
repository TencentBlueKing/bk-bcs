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
import random
import string


class BasicAuthGenerator:
    alphabet = string.ascii_letters + string.digits
    pw_length = 16
    username_random_length = 4
    max_retries = 5

    def _generate_random_string(self, length):
        """
        随机生成字符串 生成 大小写数字, 且包含至少一位数字
        """
        password_chars = [random.choice(self.alphabet) for _ in range(length - 1)]
        password_chars.append(random.choice(string.digits))
        random.shuffle(password_chars)
        return ''.join(password_chars)

    def generate_username(self):
        return self._generate_random_string(self.username_random_length)

    def generate_password(self):
        return self._generate_random_string(self.pw_length)

    def _generate_multiple_random_string(self, number, length):
        """
        产生一组不相同的随机串 预留
        """
        random_string_list = set()
        retied = 0
        while retied < self.max_retries:
            for i in range(number):
                random_string_list.add(self._generate_random_string(length))

            # if random string duplicated, length of set will lower than number
            if len(random_string_list) == number:
                return random_string_list

        raise RuntimeError("Can not generate unique string after tried for %s times!" % self.max_retries)

    def generate_basic_auth_by_role(self, role):
        return {'username': '%s-%s' % (role, self.generate_username()), 'password': self.generate_password()}
