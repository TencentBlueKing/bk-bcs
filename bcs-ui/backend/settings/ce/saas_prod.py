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
import os
from urllib import parse

import redis

from .base import *  # noqa
from .base import INSTALLED_APPS
from .cors import get_cors_allowed_origins

INSTALLED_APPS += [
    "backend.celery_app.CeleryConfig",
]

# 请求官方 API 默认版本号，可选值为："v2" 或 ""；其中，"v2"表示规范化API，""表示未规范化API
DEFAULT_BK_API_VER = "v2"

# 是否启用celery任务
IS_USE_CELERY = True

CELERY_IMPORTS = ("backend.celery_app",)

# ==============================================================================
# logging
# ==============================================================================
# 应用日志配置
BK_LOG_DIR = os.environ.get("BK_LOG_DIR", "/data/paas/apps/logs/")
LOGGING_DIR = os.path.join(BK_LOG_DIR, "logs", APP_ID)
LOG_CLASS = "logging.handlers.RotatingFileHandler"
if RUN_MODE == "DEVELOP":
    LOG_LEVEL = "DEBUG"
elif RUN_MODE == "TEST":
    LOGGING_DIR = os.path.join(BK_LOG_DIR, APP_ID)
    LOG_LEVEL = "INFO"
elif RUN_MODE == "PRODUCT":
    LOGGING_DIR = os.path.join(BK_LOG_DIR, APP_ID)
    LOG_LEVEL = "ERROR"

# 兼容企业版
LOGGING_DIR = os.environ.get("LOGGING_DIR", LOGGING_DIR)

# 自动建立日志目录
if not os.path.exists(LOGGING_DIR):
    try:
        os.makedirs(LOGGING_DIR)
    except:
        pass

DATABASES["default"] = {
    "ENGINE": "django.db.backends.mysql",
    "NAME": os.environ.get("DB_NAME"),
    "USER": os.environ.get("DB_USERNAME"),
    "PASSWORD": os.environ.get("DB_PASSWORD"),
    "HOST": os.environ.get("DB_HOST"),
    "PORT": os.environ.get("DB_PORT"),
}

LOG_LEVEL = "INFO"
LOG_FILE = os.path.join(LOGGING_DIR, f"{APP_ID}.log")
LOGGING = get_logging_config(LOG_LEVEL, None, LOG_FILE)
# don't need stdout
LOGGING["handlers"]["console"]["class"] = "logging.NullHandler"

REDIS_URL = os.environ.get("BKAPP_REDIS_URL")
# 解析url
_rpool = redis.from_url(REDIS_URL).connection_pool
REDIS_HOST = _rpool.connection_kwargs["host"]
REDIS_PORT = _rpool.connection_kwargs["port"]
REDIS_PASSWORD = _rpool.connection_kwargs["password"]
REDIS_DB = _rpool.connection_kwargs["db"]

# web-console配置需要，后台去除
RDS_HANDER_SETTINGS = {
    "level": "INFO",
    "class": "backend.utils.log.LogstashRedisHandler",
    "redis_url": REDIS_URL,
    "queue_name": "paas_backend_log_list",
    "message_type": "python-logstash",
    "tags": ["sz", "stag", "paas-backend"],
}

CACHES["default"] = {
    "BACKEND": "django_redis.cache.RedisCache",
    "LOCATION": REDIS_URL,
    "OPTIONS": {
        "CLIENT_CLASS": "django_redis.client.DefaultClient",
    },
}

# 针对BCS区分环境, backend的staging环境，连接bcs的uat和debugger，默认使用uat
DEFAULT_BCS_API_ENV = "prod"
# paas-cc环境
APIGW_PAAS_CC_ENV = "prod"
APIGW_ENV = ""

# 测试环境先禁用掉集群创建时的【prod】环境
DISABLE_PROD = True

# PaaS域名，发送邮件链接需要
PAAS_HOST = BK_PAAS_HOST
PAAS_ENV = "dev"
# 权限跳转URL, 添加console前缀
AUTH_REDIRECT_URL = "%s/console" % PAAS_HOST

# 项目地址
DEVOPS_HOST = os.environ.get("BKAPP_DEVOPS_HOST")
# PaaS Devops域名, 静态连接使用
PAAS_HOST_BCS = DEVOPS_HOST

# 统一登录页面
LOGIN_FULL = f"{BK_PAAS_HOST}/login/?c_url={DEVOPS_HOST}/console/bcs/"
LOGIN_SIMPLE = f"{BK_PAAS_HOST}/login/plain"

# 容器服务地址
DEVOPS_BCS_HOST = f"{BK_PAAS_HOST}/o/{APP_ID}"
# 容器服务 API 地址
DEVOPS_BCS_API_URL = f"{BK_PAAS_HOST}/o/{APP_ID}"
DEVOPS_ARTIFACTORY_HOST = os.environ.get("BKAPP_ARTIFACTORY_HOST")

# 企业版/社区版 helm没有平台k8s集群时，无法为项目分配chart repo服务
# 为解决该问题，容器服务会绑定一个chart repo服务使用，所有项目公用这个chart repo
HELM_MERELY_REPO_USERNAME = os.environ.get("BKAPP_HARBOR_CHARTS_USERNAME")
HELM_MERELY_REPO_PASSWORD = os.environ.get("BKAPP_HARBOR_CHARTS_PASSWORD")

# BKE 配置
# note：BKE_SERVER_HOST 配置为None时表示不使用bke，而是直接用本地kubectl
BKE_CACERT = ""

BCS_APIGW_DOMAIN = {"prod": os.environ.get("BKAPP_BCS_API_DOMAIN")}

HELM_INSECURE_SKIP_TLS_VERIFY = True

WEB_CONSOLE_KUBECTLD_IMAGE_PATH = f"{DEVOPS_ARTIFACTORY_HOST}/public/bcs/k8s/kubectld"

# web_console监听地址
WEB_CONSOLE_PORT = int(os.environ.get("WEB_CONSOLE_PORT", 28800))

THANOS_HOST = os.environ.get("BKAPP_THANOS_HOST")
# 默认指标数据来源，现在支持bk-data, prometheus
DEFAULT_METRIC_SOURCE = "prometheus"
# 普罗米修斯项目白名单
DEFAULT_METRIC_SOURCE_PROM_WLIST = []

WEB_CONSOLE_MODE = "internal"

# 初始化时渲染K8S版本
K8S_VERSION = os.environ.get("BKAPP_K8S_VERSION")

# cors settings
CORS_ALLOWED_ORIGINS = get_cors_allowed_origins([DEVOPS_HOST, BK_PAAS_HOST, DEVOPS_BCS_HOST, DEVOPS_BCS_API_URL])
CORS_ALLOW_CREDENTIALS = True

BCS_API_HOST = BCS_APIGW_DOMAIN['prod']
