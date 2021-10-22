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

"""A toolkit for defining and rendering error responses for REST API"""
from typing import Callable, Optional, Union, Dict

_DEFAULT_ERROR_CODE_NUM = -1
_DEFAULT_STATUS_CODE = 400

ExtraFormatterFunc = Callable[[str, 'APIError'], str]


class APIError(Exception):
    """API error object with detailed code and description

    :param code: required, An english identifier
    :param message: required, Detailed error message, may contains templated variables
    :param code_num: A numeric error code
    :param extra_formatter: an extra function for formatting message
    :param status_code: desired HTTP status code for representing current Error
    """

    delimiter = ': '

    def __init__(
        self,
        code: str,
        message: str,
        code_num: int = _DEFAULT_ERROR_CODE_NUM,
        extra_formatter: Optional[ExtraFormatterFunc] = None,
        status_code: int = _DEFAULT_STATUS_CODE,
    ):
        self.code = code
        self.code_num = code_num
        self.extra_formatter = extra_formatter
        self.status_code = status_code
        # Save message as private field to expose it as an property
        self._message = message

    def format(self, message: Optional[str] = None, replace: bool = False, **kwargs) -> 'APIError':
        """Try to mutate and render the original error message, return a cloned `APIError` object

        :param message: if not given, default message will be used
        :param replace: if True, relace the default message, otherwise append the incoming message
        """
        if not message:
            return self._clone(message=self._render(self._message, kwargs))

        if replace:
            obj_message = message
        else:
            obj_message = ''.join([self._message, self.delimiter, message])
        obj_message = self._render(obj_message, kwargs)
        return self._clone(message=obj_message)

    # Shortcut method name
    f = format

    @property
    def message(self) -> str:
        """Get detailed error message, it the `extra_formatter` was defined, it will be used for formatting"""
        if self.extra_formatter:
            return self.extra_formatter(self._message, self)
        return self._message

    def _clone(self, message: Optional[str] = None) -> 'APIError':
        """Clone a new APIError object

        :param message: if given, the cloned object will use this message instead of current `self._message`
        """
        obj_message = message if message is not None else self._message
        return APIError(
            code=self.code,
            code_num=self.code_num,
            extra_formatter=self.extra_formatter,
            status_code=self.status_code,
            message=obj_message,
        )

    @staticmethod
    def _render(message: str, kwargs: Dict) -> str:
        """Render message template with variables, using standard python string template syntax"""
        if kwargs:
            return message.format(**kwargs)
        return message

    def __str__(self) -> str:
        return f'<{self.__class__.__name__}: {self.code}({self.code_num})-{self.message}>'


class ErrorCode:
    """A descriptor object for defining error codes

    :param error_cls: the class for making exception instance, default to `APIError`
    """

    def __init__(self, *args, **kwargs):
        self._code = None
        self._error_cls = kwargs.pop('error_cls', APIError)
        # Save arguments for making error object later
        self._error_args = args
        self._error_kwargs = kwargs

    def __get__(self, obj, obj_type) -> Union['ErrorCode', APIError]:
        """When retriving `ErrorCode` via object attribute, always making a brand new `APIError`
        exception object
        """
        if obj is None:
            return self
        if not self._code:
            raise RuntimeError('must be initialized as a descriptor')

        try:
            return self._error_cls(self._code, *self._error_args, **self._error_kwargs)
        except TypeError as e:
            raise TypeError(f'Unable to make {self._error_cls.__name__} object: {e}')

    def __set_name__(self, obj_type, name):
        """Set field name as error code object's code"""
        self._code = name
