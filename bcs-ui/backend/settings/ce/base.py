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
import sys
from urllib import parse

from ..base import *  # noqa
from ..base import BASE_DIR, REST_FRAMEWORK

EDITION = COMMUNITY_EDITION

# TODO 统一 APP_ID 和 BCS_APP_CODE 为 APP_CODE, 统一 APP_TOKEN 和 BCS_APP_SECRET 为 APP_SECRET
APP_ID = "bk_bcs_app"
APP_TOKEN = os.environ.get("APP_TOKEN")

APP_CODE = APP_ID
APP_SECRET = APP_TOKEN

# drf鉴权, 权限控制配置
REST_FRAMEWORK["DEFAULT_AUTHENTICATION_CLASSES"] = ("backend.utils.authentication.BKTokenAuthentication",)
REST_FRAMEWORK["DEFAULT_PERMISSION_CLASSES"] = (
    "rest_framework.permissions.IsAuthenticated",
    "backend.utils.permissions.HasIAMProject",
    "backend.utils.permissions.ProjectHasBCS",
)

ALLOWED_HOSTS = ["*"]

INSTALLED_APPS += [
    "backend.uniapps.apis",
    "backend.bcs_web.apis.apps.APIConfig",
    "iam.contrib.iam_migration",
    "backend.iam.bcs_iam_migration.apps.BcsIamMigrationConfig",
]

# 统一登录页面
LOGIN_FULL = ""
LOGIN_SIMPLE = ""

# 设置 session 过期时间为 12H
SESSION_COOKIE_AGE = 12 * 60 * 60

# bkpaas_auth 模块会通过用户的 AccessToken 获取用户基本信息，因为这个 API 调用比较昂贵。
# 所以最好设置 Django 缓存来避免不必要的请求以提高效率。
CACHES = {
    "default": {
        "BACKEND": "django.core.cache.backends.filebased.FileBasedCache",
        "LOCATION": "/tmp/paas-backend-django_cache",
    },
}

REDIS_HOST = os.environ.get("REDIS_HOST", "127.0.0.1")
REDIS_PORT = os.environ.get("REDIS_PORT", 6379)
REDIS_DB = os.environ.get("REDIS_DB", 0)
REDIS_PASSWORD = os.environ.get("REDIS_PASSWORD", "")
REDIS_URL = os.environ.get("REDIS_URL", f"redis://:{REDIS_PASSWORD}@{REDIS_HOST}:{REDIS_PORT}/{REDIS_DB}")

# apigw 环境
APIGW_ENV = "test"
APIGW_PAAS_CC_ENV = "uat"
# ci部分apigw 环境
APIGW_CI_ENV = "prod"

# BK 环境的账号要单独申请
BK_JFROG_ACCOUNT_DOMAIN = "bk.artifactory.bking.com"
BK_JFROG_ACCOUNT_AUTH = ""

# ############### 模板开启后台参数验证
IS_TEMPLATE_VALIDATE = True
IS_CUP_LIMIT = False

# 不同集群对应的apigw环境, 正式环境暂时没有
# key 是 cc 中的environment变量, value 是bcs API 环境
BCS_API_ENV = {
    "stag": "uat",
    "debug": "debug",
    "prod": "prod",
}

# 针对集群的环境
CLUSTER_ENV = {
    "stag": "debug",
    "prod": "prod",
}

# 创建CC module环境
CC_MODDULE_ENV = {
    "stag": "test",
    "prod": "pro",
    "debug": "debug",
}

# 返回给前端的cluster环境
CLUSTER_ENV_FOR_FRONT = {"debug": "stag", "prod": "prod"}

# 查询事务时的bcs环境
BCS_EVENT_ENV = ["prod"]

APIGW_PUBLIC_KEY = ""

# 是否开启K8S
OPEN_K8S = True
# 使用使用 k8s 直连地址
IS_K8S_DRIVER_NO_APIGW = False

# cache invalid
CACHE_VERSION = "v1"

# OP SYSTEM ENV
APIGW_OP_ENV = "test"

RUN_ENV = "prod"

# ############### 仓库API
# 默认使用正式环境
DEPOT_STAG = "prod"
# 镜像地址前缀
DEPOT_PREFIX = ""

# CI系统API地址
DEVOPS_CI_API_HOST = ""

# 应用访问路径
SITE_URL = "/"
ENVIRONMENT = os.environ.get("BK_ENV", "development")

# 运行模式， DEVELOP(开发模式)， TEST(测试模式)， PRODUCT(正式模式)
RUN_MODE = "DEVELOP"
if ENVIRONMENT.endswith("production"):
    RUN_MODE = "PRODUCT"
    DEBUG = False
    SITE_URL = f"/o/{APP_ID}/"
elif ENVIRONMENT.endswith("testing"):
    RUN_MODE = "TEST"
    DEBUG = False
    SITE_URL = f"/t/{APP_ID}/"
else:
    RUN_MODE = "DEVELOP"
    DEBUG = True

# 是否使用容器服务自身的TLS证书
IS_USE_BCS_TLS = True

# CELERY 配置
IS_USE_CELERY = True
# 本地使用Redis做broker
BROKER_URL_DEV = REDIS_URL

WEB_CONSOLE_PORT = int(os.environ.get("WEB_CONSOLE_PORT", 28800))

if IS_USE_CELERY:
    try:
        import djcelery

        INSTALLED_APPS += ("djcelery",)  # djcelery
        djcelery.setup_loader()
        CELERY_ENABLE_UTC = False
        CELERYBEAT_SCHEDULER = "djcelery.schedulers.DatabaseScheduler"
        if "celery" in sys.argv:
            DEBUG = False
        # celery 的消息队列（RabbitMQ）信息
        BROKER_URL = os.environ.get("BK_BROKER_URL", BROKER_URL_DEV)
        if RUN_MODE == "DEVELOP":
            from celery.signals import worker_process_init

            @worker_process_init.connect
            def configure_workers(*args, **kwargs):
                import django

                django.setup()

        from celery.schedules import crontab

        CELERY_BEAT_SCHEDULE = {
            # 为防止出现资源注册权限中心失败的情况，每天定时同步一次
            'bcs_perm_tasks': {
                'task': 'backend.accounts.bcs_perm.tasks.sync_bcs_perm',
                'schedule': crontab(minute=0, hour=2),
            },
            # 每天三点进行一次强制同步
            'helm_force_sync_repo_tasks': {
                'task': 'backend.helm.helm.tasks.force_sync_all_repo',
                'schedule': crontab(minute=0, hour=3),
            },
        }
    except Exception as error:
        print("use celery error: %s" % error)

# ******************************** Helm Config Begin ********************************
# kubectl 只有1.12版本
HELM_BASE_DIR = os.environ.get("HELM_BASE_DIR", BASE_DIR)
HELM_BIN = os.path.join(HELM_BASE_DIR, "bin/helm")  # helm bin filename
HELM3_BIN = os.path.join(HELM_BASE_DIR, "bin/helm3")
YTT_BIN = os.path.join(HELM_BASE_DIR, "bin/ytt")
KUBECTL_BIN = os.path.join(HELM_BASE_DIR, "bin/kubectl-v1.12.3")  # default kubectl bin filename
DASHBOARD_CTL_BIN = os.path.join(HELM_BASE_DIR, "bin/dashboard-ctl")  # default dashboard ctl filename
KUBECTL_BIN_MAP = {
    "1.8.3": os.path.join(HELM_BASE_DIR, "bin/kubectl-v1.12.3"),
    "1.12.3": os.path.join(HELM_BASE_DIR, "bin/kubectl-v1.12.3"),
    "1.14.9": os.path.join(HELM_BASE_DIR, "bin/kubectl-v1.14.9"),
    "1.16.3": os.path.join(HELM_BASE_DIR, "bin/kubectl-v1.16.3"),
    "1.18.12": os.path.join(HELM_BASE_DIR, "bin/kubectl-v1.18.12"),
    "1.20.13": os.path.join(HELM_BASE_DIR, "bin/kubectl-v1.20.13"),
}

# BKE企业版证书
BKE_CACERT = os.path.join(HELM_BASE_DIR, "etc/prod-server.crt")

BK_CC_HOST = os.environ.get("BK_CC_HOST", "")

STATIC_URL = "/staticfiles/"
SITE_STATIC_URL = SITE_URL + STATIC_URL.strip("/")

# 是否在中间件中统一输出异常信息
IS_COMMON_EXCEPTION_MSG = False
COMMON_EXCEPTION_MSG = ""

BK_PAAS_HOST = os.environ.get("BK_PAAS_HOST", "http://dev.paas.com")
BK_PAAS_INNER_HOST = os.environ.get("BK_PAAS_INNER_HOST", BK_PAAS_HOST)
APIGW_HOST = BK_PAAS_INNER_HOST
# 组件API地址
COMPONENT_HOST = BK_PAAS_INNER_HOST

DEPOT_API = f"{APIGW_HOST}/api/apigw/harbor_api/"

# BCS API PRE URL
BCS_API_PRE_URL = f"{APIGW_HOST}/api/apigw/bcs_api"

BK_SSM_HOST = os.environ.get("BKAPP_SSM_HOST")

# BCS CC HOST
BCS_CC_API_PRE_URL = f"{APIGW_HOST}/api/apigw/bcs_cc/prod"

# iamv v3 migration 相关，用于初始资源数据到权限中心
# migrate 时，使用settings.APP_CODE, settings.SECRET_KEY
SECRET_KEY = APP_SECRET
BK_IAM_SYSTEM_ID = 'bk_bcs_app'
BK_IAM_MIGRATION_APP_NAME = "bcs_iam_migration"
BK_IAM_RESOURCE_API_HOST = os.environ.get(
    'BK_IAM_RESOURCE_API_HOST', BK_PAAS_INNER_HOST or "http://paas.service.consul"
)
BK_IAM_PROVIDER_PATH_PREFIX = os.environ.get('BK_IAM_PROVIDER_PATH_PREFIX', '/o/bk_bcs_app/apis/iam')
BK_IAM_HOST = os.environ.get("BKAPP_IAM_HOST")
BK_IAM_INNER_HOST = BK_IAM_HOST
# 参数说明 https://github.com/TencentBlueKing/iam-python-sdk/blob/master/docs/usage.md#22-config
BK_IAM_USE_APIGATEWAY = False
BK_IAM_APIGATEWAY_URL = os.environ.get('BK_IAM_APIGATEWAY_URL', None)
# 权限中心前端地址
BK_IAM_APP_URL = os.environ.get('BK_IAM_APP_URL', f"{BK_PAAS_HOST}/o/bk_iam")

# 数据平台清洗URL
_URI_DATA_CLEAN = '%2Fs%2Fdata%2Fdataset%2Finfo%2F{data_id}%2F%23data_clean'
URI_DATA_CLEAN = f'{BK_PAAS_HOST}?app=data&url=' + _URI_DATA_CLEAN

# 覆盖上层base中的DIRECT_ON_FUNC_CODE: 直接开启的功能开关，不需要在db中配置
DIRECT_ON_FUNC_CODE = ["HAS_IMAGE_SECRET", "ServiceMonitor"]

# SOPS API HOST
SOPS_API_HOST = os.environ.get("SOPS_API_HOST")

# admin 权限用户
ADMIN_USERNAME = "admin"
# BCS 默认业务
BCS_APP_ID = 1

# 社区版特殊配置
BCS_APP_CODE = APP_CODE
BCS_APP_SECRET = SECRET_KEY

# REPO 相关配置
HELM_REPO_DOMAIN = os.environ.get('HELM_REPO_DOMAIN')
BK_REPO_URL_PREFIX = os.environ.get('BK_REPO_URL_PREFIX')

# 默认 BKCC 设备供应方，社区版默认 '0'
BKCC_DEFAULT_SUPPLIER_ACCOUNT = os.environ.get('BKCC_DEFAULT_SUPPLIER_ACCOUNT', '0')

# clustermanager域名
CLUSTER_MANAGER_DOMAIN = os.environ.get("CLUSTER_MANAGER_DOMAIN", "")

# 可能有带端口的情况，需要去除
SESSION_COOKIE_DOMAIN = "." + parse.urlparse(BK_PAAS_HOST).netloc.split(":")[0]
CSRF_COOKIE_DOMAIN = SESSION_COOKIE_DOMAIN

# 蓝鲸 opentelemetry trace 配置
# 是否开启 OTLP, 默认不开启
OPEN_OTLP = False
# 上报的地址
OTLP_GRPC_HOST = os.environ.get("OTLP_GRPC_HOST", "")
# 上报的 data id
OTLP_DATA_ID = os.environ.get("OTLP_DATA_ID", "")
# 上报时, 使用的服务名称
OTLP_SERVICE_NAME = APP_ID
