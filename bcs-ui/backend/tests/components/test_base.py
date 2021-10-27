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
import pytest
import requests

from backend.components.base import (
    BaseHttpClient,
    CompInternalError,
    CompParseBkCommonResponseError,
    CompRequestError,
    response_handler,
    update_request_body,
    update_url_parameters,
)
from backend.tests.testing_utils.base import nullcontext


class TestBaseHttpClient:
    def test_default_kwargs(self, requests_mock):
        requests_mock.get('http://test.com', json={})
        client = BaseHttpClient()
        client.request('GET', 'http://test.com/')

        req_history = requests_mock.request_history[0]
        assert req_history.headers.get('X-Request-Id') is not None
        assert req_history.timeout is not None
        assert req_history.verify is client._ssl_verify

    def test_request_json(self, requests_mock):
        requests_mock.get('http://test.com', json={'foo': 'bar'})
        client = BaseHttpClient()
        assert client.request_json('GET', 'http://test.com/') == {'foo': 'bar'}

    @pytest.mark.parametrize(
        'req_exc,exc',
        [
            (requests.exceptions.ConnectTimeout, CompRequestError),
            (ValueError, CompInternalError),
        ],
    )
    def test_request_error(self, req_exc, exc, requests_mock):
        requests_mock.get('http://test.com', exc=req_exc)
        with pytest.raises(exc):
            BaseHttpClient().request('GET', 'http://test.com/')

    @pytest.mark.parametrize(
        'status_code,exc',
        [
            (200, None),
            (404, CompRequestError),
            (500, CompRequestError),
        ],
    )
    def test_status_errors(self, status_code, exc, requests_mock):
        requests_mock.get('http://test.com', json={'foo': 'bar'}, status_code=status_code)
        exc_context = pytest.raises(exc) if exc else nullcontext()
        with exc_context:
            BaseHttpClient().request('GET', 'http://test.com/')


@pytest.mark.parametrize(
    'url,parameters,expected_result',
    [
        ('http://foo.com/path', {'foo': 'bar'}, 'http://foo.com/path?foo=bar'),
        ('https://foo.com/path/?bar=3', {'foo': 'bar'}, 'https://foo.com/path/?bar=3&foo=bar'),
        ('http://foo.com/path/?foo=3', {'foo': 'bar'}, 'http://foo.com/path/?foo=bar'),
        ('http://foo.com/path/?bar=3', {'foo': ['bar', 'baz']}, 'http://foo.com/path/?bar=3&foo=bar&foo=baz'),
    ],
)
def test_update_url_parameters(url, parameters, expected_result):
    result = update_url_parameters(url, parameters)
    assert result == expected_result


@pytest.mark.parametrize(
    "body,params,expected_body",
    [
        (None, {"test": "test"}, b'{"test": "test"}'),
        (b'{"raw_body": "raw_body"}', {"test": "test"}, b'{"raw_body": "raw_body", "test": "test"}'),
    ],
)
def test_update_request_body(body, params, expected_body):
    updated_body = update_request_body(body, params)
    assert updated_body == expected_body


class TestBkCommonResponseHandler:
    def func_data_ok(self):
        return {"code": 0, "data": {"status": "running"}}

    def func_null_ok(self):
        return {"code": 0, "message": "success"}

    def func_error(self):
        return {"code": 1, "message": "this is error test"}

    @pytest.mark.parametrize(
        "default_data,func,expected_data",
        [
            (None, "func_data_ok", {"status": "running"}),
            ({"default": "data"}, "func_data_ok", {"status": "running"}),
            (None, "func_null_ok", None),
            ({"default": "data"}, "func_null_ok", {"default": "data"}),
        ],
    )
    def test_response_ok_data_hander(self, default_data, func, expected_data):
        data = response_handler(default_data)(getattr(self, func))()
        assert data == expected_data

    def test_response_error_data_hander(self):
        with pytest.raises(CompParseBkCommonResponseError):
            response_handler()(self.func_error)()

    @pytest.mark.parametrize(
        "default_data,func,expected_data",
        [
            (None, "func_data_ok", {"code": 0, "data": {"status": "running"}}),
            ({"default": "data"}, "func_data_ok", {"code": 0, "data": {"status": "running"}}),
            (None, "func_null_ok", {"code": 0, "message": "success"}),
            ({"default": "data"}, "func_null_ok", {"code": 0, "message": "success"}),
        ],
    )
    def test_response_raw_hander(self, default_data, func, expected_data):
        data = response_handler(default_data)(getattr(self, func)).raw_request()
        assert data == expected_data
