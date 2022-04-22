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
import redis

from .base import *  # noqa

SECRET_KEY = "jllc(^rzpe8_udv)oadny2j3ym#qd^x^3ns11_8kq(1rf8qpd2"

DATABASES["default"] = {
    "ENGINE": "django.db.backends.mysql",
    "NAME": "bcs-app",
    "USER": "root",
    "PASSWORD": os.environ.get("DB_PASSWORD", ""),
    "HOST": os.environ.get("DB_HOST", "127.0.0.1"),
    "PORT": "3306",
    "OPTIONS": {
        "init_command": "SET default_storage_engine=INNODB",
    },
}

INSTALLED_APPS += [
    "backend.celery_app.CeleryConfig",
]

# 本地开发先去除权限中心v3的数据初始逻辑
INSTALLED_APPS.remove("backend.iam.bcs_iam_migration.apps.BcsIamMigrationConfig")

LOG_LEVEL = "DEBUG"
LOGGING = get_logging_config(LOG_LEVEL)

# cors settings
CORS_ORIGIN_REGEX_WHITELIST = (r".*",)

PAAS_ENV = "local"

# 容器服务地址
DEVOPS_HOST = os.environ.get("DEV_DEVOPS_HOST", "")
DEVOPS_BCS_HOST = os.environ.get("DEV_BCS_APP_HOST", "")
# 容器服务 API 地址
DEVOPS_BCS_API_URL = os.environ.get("DEV_BCS_APP_HOST", "")
DEVOPS_ARTIFACTORY_HOST = os.environ.get("BKAPP_ARTIFACTORY_HOST")

BK_PAAS_INNER_HOST = os.environ.get("BK_PAAS_INNER_HOST", BK_PAAS_HOST)

REDIS_URL = os.environ.get("BKAPP_REDIS_URL", "redis://127.0.0.1/0")
# 解析url
_rpool = redis.from_url(REDIS_URL).connection_pool
REDIS_HOST = _rpool.connection_kwargs["host"]
REDIS_PORT = _rpool.connection_kwargs["port"]
REDIS_PASSWORD = _rpool.connection_kwargs["password"]
REDIS_DB = _rpool.connection_kwargs["db"]

# IAM 地址
BK_IAM_HOST = os.environ.get('BKAPP_IAM_HOST', 'http://dev.iam.com')

APIGW_HOST = BK_PAAS_INNER_HOST

DEPOT_API = f"{APIGW_HOST}/api/apigw/harbor_api/"

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

# BCS CC PATH
BCS_CC_CLUSTER_CONFIG = "/v1/clusters/{cluster_id}/cluster_version_config/"
BCS_CC_GET_CLUSTER_MASTERS = "/projects/{project_id}/clusters/{cluster_id}/manager_masters/"
BCS_CC_GET_PROJECT_MASTERS = "/projects/{project_id}/clusters/null/manager_masters/"
BCS_CC_GET_PROJECT_NODES = "/projects/{project_id}/clusters/null/nodes/"
BCS_CC_OPER_PROJECT_NODE = "/projects/{project_id}/clusters/null/nodes/{node_id}/"
BCS_CC_OPER_PROJECT_NAMESPACES = "/projects/{project_id}/clusters/null/namespaces/"
BCS_CC_OPER_PROJECT_NAMESPACE = "/projects/{project_id}/clusters/null/namespaces/{namespace_id}/"

HELM_REPO_DOMAIN = os.environ.get("BKAPP_HARBOR_CHARTS_DOMAIN")

BCS_APIGW_DOMAIN = {"prod": os.environ.get("BKAPP_BCS_API_DOMAIN")}
