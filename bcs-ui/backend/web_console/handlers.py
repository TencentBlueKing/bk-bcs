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
import base64
import json
import logging

import arrow
import tornado.web
import tornado.websocket
from django.conf import settings
from django.utils import translation
from django.utils.encoding import smart_str
from django.utils.translation import ugettext_lazy as _
from django.utils.translation.trans_real import get_supported_language_variant
from tornado import locale
from tornado.ioloop import IOLoop, PeriodicCallback

from backend.utils.funutils import remove_url_domain
from backend.utils import FancyDict
from backend.web_console import bcs_client, constants, utils
from backend.web_console.auth import authenticated
from backend.web_console.pod_life_cycle import PodLifeCycle
from backend.web_console.utils import clean_bash_escape, get_auditor

WEBSOCKET_HANDLER_SET = set()
logger = logging.getLogger(__name__)


class LocaleHandlerMixin:
    """国际化Mixin"""

    def get_user_locale(self):
        bk_lang = self.get_cookie(settings.LANGUAGE_COOKIE_NAME)
        try:
            lang_code = get_supported_language_variant(bk_lang)
        except LookupError:
            lang_code = settings.LANGUAGE_CODE
        translation.activate(lang_code)
        return locale.get(lang_code)


class IndexPageHandler(LocaleHandlerMixin, tornado.web.RequestHandler):
    """首页处理"""

    def get(self, project_id, cluster_id):
        session_url = f"{settings.DEVOPS_BCS_API_URL}/api/projects/{project_id}/clusters/{cluster_id}/web_console/session/"  # noqa

        # mesos集群会带具体信息
        query = self.request.query
        if query:
            session_url += f"?{query}"

        session_url = remove_url_domain(session_url)

        data = {"settings": settings, "session_url": session_url, "title": cluster_id}
        self.render("templates/index.html", **data)


class SessionPageHandler(LocaleHandlerMixin, tornado.web.RequestHandler):
    """开放的页面WebConsole页面"""

    def get(self):
        # session_id通过参数获取
        session_id = self.get_argument("session_id", "")
        title = self.get_argument("container_name", "--")

        session_url = f"{settings.DEVOPS_BCS_API_URL}/api/web_console/sessions/?session_id={session_id}"
        session_url = remove_url_domain(session_url)

        data = {"settings": settings, "session_url": session_url, "title": title}
        self.render("templates/index.html", **data)


class MgrHandler(LocaleHandlerMixin, tornado.web.RequestHandler):
    """管理页"""

    def get(self, project_id):
        domain_settings = {
            "SITE_STATIC_URL": settings.SITE_STATIC_URL,
            "DEVOPS_BCS_API_URL": remove_url_domain(settings.DEVOPS_BCS_API_URL),
        }
        data = {"settings": FancyDict(domain_settings), "project_id": project_id}
        self.render("templates/mgr.html", **data)


class BCSWebSocketHandler(LocaleHandlerMixin, tornado.websocket.WebSocketHandler):
    """WebSocket处理"""

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.input_record = []
        self.input_buffer = ""
        self.last_input_ts = IOLoop.current().time()
        self.login_ts = IOLoop.current().time()

        self.record_callback = None
        self.tick_callback = None
        self.record_interval = 10
        self.heartbeat_callback = None
        self.auditor = get_auditor()
        self.pod_life_cycle = PodLifeCycle()
        self.exit_buffer = ""
        self.exit_command = "exit"
        self.user_pod_name = None
        self.source = None

    def check_origin(self, origin):
        return True

    @authenticated
    def get(self, *args, **kwargs):
        """只鉴权使用"""
        return super().get(*args, **kwargs)

    def open(self, project_id, cluster_id, context):
        self.project_id = project_id
        self.cluster_id = cluster_id
        self.context = context
        self.user_pod_name = context["user_pod_name"]
        self.source = self.get_argument("source")

        rows = self.get_argument("rows")
        rows = utils.format_term_size(rows, constants.DEFAULT_ROWS)

        cols = self.get_argument("cols")
        cols = utils.format_term_size(cols, constants.DEFAULT_COLS)

        mode = context.get("mode")
        self.bcs_client = bcs_client.factory.create(mode, self, context, rows, cols)

        WEBSOCKET_HANDLER_SET.add(self)

    def is_exit_command(self, message):
        """判断是否主动退出"""
        # 去除空格
        message = message.strip()

        # 分号表示多个命令执行, 任一一个有exit命令即退出
        for i in message.split(";"):
            if self.exit_command == i.strip():
                return True

        # 空格表示按顺序执行, 第一个是exit命令即退出
        for i in message.split():
            if self.exit_command == i.strip():
                return True
            break

        return False

    def on_message(self, message):
        self.last_input_ts = IOLoop.current().time()
        channel = int(message[0])
        message = base64.b64decode(message[1:])
        if channel == constants.RESIZE_CHANNEL:
            size = json.loads(message)
            self.bcs_client.set_pty_size(size["rows"], size["cols"])
        else:
            self.send_message(smart_str(message))

    def on_close(self):
        if self.tick_callback:
            logger.info("stop tick callback, %s", self.user_pod_name)
            self.tick_callback.stop()

        if self.record_callback:
            logger.info("stop record_callback, %s", self.user_pod_name)
            self.record_callback.stop()

        if self.heartbeat_callback:
            logger.info("stop heartbeat_callback, %s", self.user_pod_name)
            self.heartbeat_callback.stop()

        self.bcs_client.close_transmission()
        WEBSOCKET_HANDLER_SET.remove(self)

        logger.info("on_close, code: %s, reason: %s, pod: %s", self.close_code, self.close_reason, self.user_pod_name)

    def flush_input_record(self):
        """获取输出记录"""
        record = self.input_record[:]
        self.input_record = []
        return record

    def tick_timeout(self):
        """主动停止掉session"""
        self.tick_callback = PeriodicCallback(self.periodic_tick, self.record_interval * 1000)
        self.tick_callback.start()

    def periodic_tick(self):
        now = IOLoop.current().time()
        idle_time = now - max(self.bcs_client.last_output_ts, self.last_input_ts)
        if idle_time > constants.TICK_TIMEOUT:
            tick_timeout_min = constants.TICK_TIMEOUT // 60
            message = _("BCS Console 已经{}分钟无操作").format(tick_timeout_min)
            self.close_reason = message
            self.close(reason=message)
            logger.info("tick timeout, close session %s, idle time, %.2f", self.user_pod_name, idle_time)
        logger.info("tick active %s, idle time, %.2f", self.user_pod_name, idle_time)

        login_time = now - self.login_ts
        if login_time > constants.LOGIN_TIMEOUT:
            login_timeout = constants.LOGIN_TIMEOUT // (60 * 60)
            message = _("BCS Console 使用已经超过{}小时，请重新登录").format(login_timeout)
            self.close_reason = message
            self.close(reason=message)
            logger.info("tick timeout, close session %s, login time, %.2f", self.user_pod_name, login_time)
        logger.info("tick active %s, login time, %.2f", self.user_pod_name, login_time)

    def heartbeat(self):
        """每秒钟上报心跳"""
        self.heartbeat_callback = PeriodicCallback(lambda: self.pod_life_cycle.heartbeat(self.user_pod_name), 1000)
        self.heartbeat_callback.start()

    def start_record(self):
        """操作审计"""
        self.record_callback = PeriodicCallback(self.periodic_record, self.record_interval * 1000)
        self.record_callback.start()

    def periodic_record(self):
        """周期上报操作记录"""
        input_record = self.flush_input_record()
        output_record = self.bcs_client.flush_output_record()

        if not input_record and not output_record:
            return

        # 上报的数据
        data = {
            "input_record": "\r\n".join(input_record),
            "output_record": "\r\n".join(output_record),
            "session_id": self.context["session_id"],
            "context": self.context,
            "project_id": self.project_id,
            "cluster_id": self.cluster_id,
            "user_pod_name": self.user_pod_name,
            "username": self.context["username"],
        }
        self.auditor.emit(data)
        logger.info(data)

    def send_message(self, message):
        if not self.bcs_client.ws or self.bcs_client.ws.stream.closed():
            logger.info("session %s, close, message just ignore", self)
            return

        self.input_buffer += message

        if self.input_buffer.endswith(constants.INPUT_LINE_BREAKER):
            # line_msg = ['command', '']
            line_msg = self.input_buffer.split(constants.INPUT_LINE_BREAKER)
            for i in line_msg[:-1]:
                record = "%s: %s" % (arrow.now().strftime("%Y-%m-%d %H:%M:%S.%f"), clean_bash_escape(i))
                logger.debug(record)
                self.input_record.append(record)
            # empty input_buffer
            self.input_buffer = line_msg[-1]

        try:
            self.bcs_client.write_message(message)
        except Exception as e:
            logger.exception(e)
