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
import codecs
import json

import six
from django.conf import settings
from rest_framework import renderers
from rest_framework.exceptions import ParseError
from rest_framework.parsers import BaseParser, JSONParser


class TextJsonParser(JSONParser):
    """JSON解释器，只有media_type不一样
    MyOA回调请求是这种类型
    """

    media_type = 'text/json'


class ESParser(BaseParser):
    media_type = 'application/json'

    def parse(self, stream, media_type=None, parser_context=None):
        """
        Parses the incoming bytestream as JSON and returns the resulting data.
        """
        parser_context = parser_context or {}
        encoding = parser_context.get('encoding', settings.DEFAULT_CHARSET)

        try:
            decoded_stream = codecs.getreader(encoding)(stream)
            stream = decoded_stream.read()
            return list(map(json.loads, stream.splitlines()))
        except ValueError as exc:
            raise ParseError('JSON parse error - %s' % six.text_type(exc))
