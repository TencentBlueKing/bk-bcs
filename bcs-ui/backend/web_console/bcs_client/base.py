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
import abc
import base64
import logging
import shlex
from urllib.parse import urlencode

import arrow
import tornado.gen
from django.utils.encoding import smart_bytes, smart_str
from django.utils.translation import ugettext_lazy as _
from tornado.httpclient import HTTPRequest
from tornado.ioloop import IOLoop
from tornado.websocket import websocket_connect

from backend.web_console import constants
from backend.web_console.utils import clean_bash_escape, hello_message

logger = logging.getLogger(__name__)


class BCSClientBase(abc.ABC):
    def __init__(self, url, rows, cols, msg_handler):
        self.init_rows = rows
        self.init_cols = cols

        self.url = HTTPRequest(url, validate_cert=False)

        self.msg_handler = msg_handler
        self.ws = None

        self.output_record = []
        self.output_buffer = ""
        self.last_output_ts = IOLoop.current().time()

    @tornado.gen.coroutine
    def connect(self):
        logger.info("trying to connect %s, %s", self.url.url, self.url.headers)
        try:
            self.ws = yield websocket_connect(self.url, ping_interval=constants.WEBSOCKET_PING_INTERVAL)
        except Exception as e:
            logger.exception("connection error, %s" % e)
            self.msg_handler.close()
        else:
            self.post_connected()
            self.run()

    def encode_console_msg(self, msg):
        """前后端统一使用base64编码"""
        encode_msg = base64.b64encode(smart_bytes(msg))
        return encode_msg

    def post_connected(self):
        logger.info("bcs client connected, %s", self.msg_handler.user_pod_name)
        msg = hello_message(self.msg_handler.source)
        self.msg_handler.write_message(self.encode_console_msg(msg))
        self.msg_handler.start_record()
        self.msg_handler.tick_timeout()
        self.set_pty_size(self.init_rows, self.init_cols)

    def flush_output_record(self):
        """获取输出记录"""
        record = self.output_record[:]
        self.output_record = []
        return record

    def close_transmission(self):
        """结束通讯, 发送CTRL-D
        ASCII编码 EOT 04
        """
        try:
            self.write_message(chr(4))
        except Exception as error:
            logger.warning("close_transmission %s error: %s", self.msg_handler.user_pod_name, error)

    def handle_message(self, message):
        """消息格式转换"""
        return message

    def write_message(self, message):
        """写入消息"""
        self.ws.write_message(message)

    @classmethod
    def get_command_params(cls, context):
        """获取k8s标准的命令参数"""
        if context.get("command") and context["command"] != "sh":
            command_list = shlex.split(context["command"])
        else:
            command_list = constants.DEFAULT_COMMAND

        command_list = [("command", i) for i in command_list]
        command = urlencode(command_list)
        return command

    @abc.abstractmethod
    def set_pty_size(self, rows: int, cols: int):
        """自动宽度适应"""

    @tornado.gen.coroutine
    def run(self):
        while True:
            msg = yield self.ws.read_message()
            if msg is None:
                logger.info("bcs client connection closed, %s", self.msg_handler.user_pod_name)
                message = str(_("BCS Console 服务端连接断开，请重新登录"))
                self.msg_handler.close(reason=message)
                break

            if self.msg_handler.stream.closed():
                logger.info("msg_handler connection closed, %s", self.msg_handler.user_pod_name)
                self.ws.close()
                break

            try:
                self.last_output_ts = IOLoop.current().time()

                # 不同类型, 子类继承处理 message
                raw_msg = self.handle_message(msg)
                if not raw_msg:
                    continue

                try:
                    msg = smart_str(raw_msg)
                except Exception:
                    msg = smart_str(raw_msg, "latin1")

                self.output_buffer += msg

                if constants.OUTPUT_LINE_BREAKER in self.output_buffer:
                    line_msg = self.output_buffer.split(constants.OUTPUT_LINE_BREAKER)
                    for i in line_msg[:-1]:
                        record = "%s: %s" % (arrow.now().strftime("%Y-%m-%d %H:%M:%S.%f"), clean_bash_escape(i))
                        self.output_record.append(record)
                    # 前面多行已经赋值到record, 最后一行可能剩余未换行的数据
                    self.output_buffer = line_msg[-1]

                # 前端对\r不会换行处理，在后台替换，规则是前后没有\n的\r字符，都会添加\n
                # msg = re.sub(r'(?<!\n)\r(?!\n)', '\r\n', msg)

                # 删除异常回车键
                # msg = re.sub(r'[\b]+$', '\b', msg)

                self.msg_handler.write_message(self.encode_console_msg(raw_msg))
            except Exception as e:
                logger.exception(e)
                self.ws.close()
