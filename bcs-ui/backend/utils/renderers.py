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
import dataclasses

from rest_framework.renderers import JSONRenderer
from rest_framework.utils import encoders

from backend.utils.local import local


class JSONEncoder(encoders.JSONEncoder):
    """支持 dataclasses 序列化"""

    def default(self, obj):
        if dataclasses.is_dataclass(obj):
            # 如果 asdict 出现异常，则使用默认的 encoder 逻辑.
            # note: kubernetes.dynamic.ResourceField 由于 __getattr__ 的逻辑，会通过 is_dataclass 检查，但实际并非dataclass
            try:
                return dataclasses.asdict(obj)
            except Exception:
                return super().default(obj)

        return super().default(obj)


class BKAPIRenderer(JSONRenderer):
    """
    采用统一的结构封装返回内容
    """

    encoder_class = JSONEncoder
    SUCCESS_CODE = 0
    SUCCESS_MESSAGE = 'OK'

    def render(self, data, accepted_media_type=None, renderer_context=None):
        if isinstance(data, dict) and 'code' in data:
            data['request_id'] = local.request_id
        else:
            data = {
                'data': data,
                'code': self.SUCCESS_CODE,
                'message': self.SUCCESS_MESSAGE,
                'request_id': local.request_id,
            }

        if renderer_context:
            for key in ['permissions', 'message', 'web_annotations']:
                if renderer_context.get(key):
                    data[key] = renderer_context[key]

        response = super(BKAPIRenderer, self).render(data, accepted_media_type, renderer_context)
        return response
