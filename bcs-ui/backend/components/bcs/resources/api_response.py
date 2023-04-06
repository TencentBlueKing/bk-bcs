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
import json
import logging
from functools import wraps
from typing import Any, Callable, Dict

from kubernetes.client.rest import ApiException

logger = logging.getLogger(__name__)


def response(format_data=True):
    # TODO: format_data 参数现在已经没有作用了，因为 KubeResponseTransformer 的实现会自动判断，结果是否可以调用 to_dict 方法
    # 如果可以的话，会自动调用，之后需要移除所有对该参数的使用。
    return KubeResponseTransformer()


# TODO: 这个响应转换属于更外层的逻辑，应该移动到更外层
class KubeResponseTransformer:
    """用于把可能返回 Kubernetes 客户端对象的函数响应，转换为可直接返回给前端的合法 JSON 响应"""

    _default_error_code = 4001
    _default_success_code = 0

    def __call__(self, func: Callable):
        @wraps(func)
        def decorated(*args, **kwargs):
            try:
                data = func(*args, **kwargs)
                # 普通 Kubernetes 对象提供了 to_dict() 方法
                if hasattr(data, 'to_dict'):
                    data = data.to_dict()

                return self.make_success_resp(data)
            except ApiException as err:
                if err.status == 404:
                    logger.info('resource not found, %s', err)
                else:
                    logger.error('request bcs api error, %s', err)

                error_msg = self.extract_message(err)
                return self.make_error_resp(error_msg, err.status)
            except Exception as err:
                logger.exception('request bcs api error, %s', err)
                return self.make_error_resp(str(err))

        return decorated

    @staticmethod
    def extract_message(error: ApiException) -> str:
        """从 ApiException 异常对象中尝试解析错误信息"""
        try:
            return json.loads(error.body)['message']
        except Exception:
            return f'request bcs api error, {error}'

    def make_success_resp(self, data: Any) -> Dict:
        return {'code': self._default_success_code, 'result': True, 'data': data, 'message': 'success'}

    def make_error_resp(self, message: str, code: int = _default_error_code) -> Dict:
        return {'code': code, 'result': False, 'message': message}
