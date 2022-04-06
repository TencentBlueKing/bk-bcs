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
import threading
from typing import Collection

import MySQLdb
from celery.signals import worker_process_init
from django.conf import settings
from opentelemetry import trace
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.instrumentation import dbapi
from opentelemetry.instrumentation.celery import CeleryInstrumentor
from opentelemetry.instrumentation.django import DjangoInstrumentor
from opentelemetry.instrumentation.instrumentor import BaseInstrumentor
from opentelemetry.instrumentation.logging import LoggingInstrumentor
from opentelemetry.instrumentation.redis import RedisInstrumentor
from opentelemetry.instrumentation.requests import RequestsInstrumentor
from opentelemetry.sdk.resources import Resource
from opentelemetry.sdk.trace import ReadableSpan, TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.trace import Span, Status, StatusCode
from requests import Request, Response


class BluekingInstrumentor(BaseInstrumentor):
    has_instrument = False

    def _uninstrument(self, **kwargs):
        pass

    def _instrument(self, **kwargs):
        """Instrument the library"""
        # 判断是否使用
        if not settings.OPEN_OTLP:
            return
        if self.has_instrument:
            return
        # 初始化SDK 配置服务名称、bk_data_id和采样配置
        tracer_provider = TracerProvider(
            resource=Resource.create(
                {
                    "service.name": settings.OTLP_SERVICE_NAME,
                    "bk_data_id": int(settings.OTLP_DATA_ID),
                }
            )
        )
        # 配置grpc上报exporter配置
        otlp_exporter = OTLPSpanExporter(endpoint=settings.OTLP_GRPC_HOST)
        span_processor = LazyBatchSpanProcessor(otlp_exporter)
        tracer_provider.add_span_processor(span_processor)
        # 注入trace配置
        trace.set_tracer_provider(tracer_provider)
        # 安装插件
        DjangoInstrumentor().instrument(response_hook=django_response_hook)
        RedisInstrumentor().instrument()
        RequestsInstrumentor().instrument(tracer_provider=tracer_provider, span_callback=requests_callback)
        CeleryInstrumentor().instrument(tracer_provider=tracer_provider)
        LoggingInstrumentor().instrument()
        dbapi.wrap_connect(
            __name__,
            MySQLdb,
            "connect",
            "mysql",
            {
                "database": "db",
                "port": "port",
                "host": "host",
                "user": "user",
            },
            tracer_provider=tracer_provider,
        )
        self.has_instrument = True

    def instrumentation_dependencies(self) -> Collection[str]:
        return []


class LazyBatchSpanProcessor(BatchSpanProcessor):
    def __init__(self, *args, **kwargs):
        super(LazyBatchSpanProcessor, self).__init__(*args, **kwargs)
        # 停止默认线程
        self.done = True
        with self.condition:
            self.condition.notify_all()
        self.worker_thread.join()
        self.done = False
        self.worker_thread = None

    def on_end(self, span: ReadableSpan) -> None:
        if self.worker_thread is None:
            self.worker_thread = threading.Thread(target=self.worker, daemon=True)
            self.worker_thread.start()
        super(LazyBatchSpanProcessor, self).on_end(span)


@worker_process_init.connect(weak=False)
def init_celery_tracing(*args, **kwargs):
    """celery 初始化"""
    if not settings.OPEN_OTLP:
        return
    BluekingInstrumentor().instrument()


def requests_callback(span: Span, response: Response):
    """request 的 callback 处理"""
    try:
        resp = response.json()
    except Exception:
        return
    if not isinstance(resp, dict):
        return
    code = resp.get("code", 0)
    span.set_attribute("code", code)
    span.set_attribute("request_id", resp.get("request_id", ""))
    span.set_attribute("message", resp.get("message", ""))
    if code:
        span.set_status(Status(StatusCode.OK))
        return
    span.set_status(Status(StatusCode.ERROR))


def django_response_hook(span: Span, request: Request, response: Response):
    """Django 请求处理"""
    # 获取真正的path
    span.update_name(request.path)
    # 解析data
    if hasattr(response, "data"):
        resp = response.data
    else:
        try:
            resp = json.loads(response.content)
        except Exception:
            return
    if not isinstance(resp, dict):
        return
    code = resp.get("code", 0)
    span.set_attribute("code", code)
    span.set_attribute("message", resp.get("message", ""))
    if code:
        span.set_status(Status(StatusCode.OK))
        return
    span.set_status(Status(StatusCode.ERROR))
