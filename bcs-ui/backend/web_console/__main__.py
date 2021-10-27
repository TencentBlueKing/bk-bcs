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
import logging
import signal
import time
from functools import partial

import tornado.gen
import tornado.httpserver
import tornado.ioloop
from django.conf import settings

from backend.web_console import constants
from backend.web_console.handlers import WEBSOCKET_HANDLER_SET
from backend.web_console.pod_life_cycle import PodLifeCycle
from backend.web_console.urls import handlers
from backend.web_console.utils import _setup_logging

logger = logging.getLogger(__name__)


class Application(tornado.web.Application):
    def __init__(self):
        settings = {
            "template_path": "backend/web_console",
            "static_path": "backend/web_console/static",
            "websocket_ping_interval": constants.WEBSOCKET_PING_INTERVAL,
        }
        super(Application, self).__init__(handlers, **settings)


def exit_handler(server, sig, frame):
    io_loop = tornado.ioloop.IOLoop.instance()

    async def shutdown():
        logger.info("Stopping http server")
        server.stop()

        logger.info("Stopping websocket handler")
        for handler in WEBSOCKET_HANDLER_SET:
            handler.close()
        await tornado.gen.sleep(0.5)

        logger.info("Stopping io_loop")
        io_loop.stop()

        logger.info("Exited")

    logger.warning("Caught signal: %s", sig)
    io_loop.add_callback_from_signal(shutdown)


def main():
    _setup_logging(verbose=True)
    ws_app = Application()
    tornado.locale.load_gettext_translations(settings.LOCALE_PATHS[0], domain="django")
    server = tornado.httpserver.HTTPServer(ws_app)
    server.listen(settings.WEB_CONSOLE_PORT)

    # graceful_shutdown
    signal.signal(signal.SIGTERM, partial(exit_handler, server))
    signal.signal(signal.SIGINT, partial(exit_handler, server))

    pod_life_cycle = PodLifeCycle()
    pod_life_cycle.start()
    tornado.ioloop.IOLoop.instance().start()


if __name__ == "__main__":
    main()
