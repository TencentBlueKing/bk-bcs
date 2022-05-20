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
import sys

from ..base import *  # noqa

EDITION = COMMUNITY_EDITION

# ******************************** Django 原始配置 ********************************
APP_ID = os.environ.get('BKPAAS_APP_ID')
APP_TOKEN = os.environ.get('BKPAAS_APP_SECRET')

# 兼容老版本平台变量名
APP_CODE = APP_ID
SECRET_KEY = APP_TOKEN
APP_SECRET = APP_TOKEN

# 社区版特殊配置
BCS_APP_CODE = APP_CODE
BCS_APP_SECRET = SECRET_KEY

# drf 鉴权, 权限控制配置
REST_FRAMEWORK['DEFAULT_AUTHENTICATION_CLASSES'] = ('backend.utils.authentication.BKTokenAuthentication',)
REST_FRAMEWORK['DEFAULT_PERMISSION_CLASSES'] = (
    'rest_framework.permissions.IsAuthenticated',
    'backend.utils.permissions.HasIAMProject',
    'backend.utils.permissions.ProjectHasBCS',
)

ALLOWED_HOSTS = ['*']

INSTALLED_APPS += [
    'backend.uniapps.apis',
    'backend.bcs_web.apis.apps.APIConfig',
    'iam.contrib.iam_migration',
    'backend.iam.bcs_iam_migration.apps.BcsIamMigrationConfig',
]

# 应用访问路径
SITE_URL = '/'
# 静态文件相关
STATIC_URL = '/staticfiles/'
SITE_STATIC_URL = SITE_URL + STATIC_URL.strip('/')

# cors 相关配置
CORS_ORIGIN_REGEX_WHITELIST = (r'.*',)
CORS_ALLOW_CREDENTIALS = True

# ******************************** 容器服务 配置 ********************************

# 设置 session 过期时间为 12H
SESSION_COOKIE_AGE = 12 * 60 * 60

# BK 环境的账号要单独申请
BK_JFROG_ACCOUNT_DOMAIN = 'bk.artifactory.bking.com'

# 模板开启后台参数验证
IS_TEMPLATE_VALIDATE = True

# mesos 不同集群对应的apigw环境, 正式环境暂时没有
# key 是 cc 中的environment变量, value 是bcs API 环境
BCS_API_ENV = {
    'stag': 'uat',
    'debug': 'debug',
    'prod': 'prod',
}

# 针对集群的环境
CLUSTER_ENV = {
    'stag': 'debug',
    'prod': 'prod',
}

# 创建CC module环境
CC_MODDULE_ENV = {
    'stag': 'test',
    'prod': 'pro',
    'debug': 'debug',
}

# 返回给前端的cluster环境
CLUSTER_ENV_FOR_FRONT = {'debug': 'stag', 'prod': 'prod'}

# 查询事务时的bcs环境
BCS_EVENT_ENV = ['prod']

APIGW_PUBLIC_KEY = ''

RUN_ENV = 'prod'

# 仓库API
# 默认使用正式环境
DEPOT_STAG = 'prod'
# 镜像地址前缀
DEPOT_PREFIX = ''

# V3 部署环境
ENVIRONMENT = os.environ.get('BKPAAS_ENVIRONMENT', 'prod')
# 运行模式， DEVELOP(开发模式)， TEST(测试模式)， PRODUCT(正式模式)
RUN_MODE = 'DEVELOP'
if ENVIRONMENT == 'prod':
    RUN_MODE = 'PRODUCT'
    DEBUG = False
elif ENVIRONMENT == 'stag':
    RUN_MODE = 'TEST'
    DEBUG = False
else:
    RUN_MODE = 'DEVELOP'
    DEBUG = True

# 是否在中间件中统一输出异常信息
IS_COMMON_EXCEPTION_MSG = True
COMMON_EXCEPTION_MSG = '联系管理员处理'

# 是否使用容器服务自身的TLS证书
IS_USE_BCS_TLS = True

# 初始化时渲染K8S/MESOS版本
K8S_VERSION = os.environ.get('BKAPP_K8S_VERSION')
MESOS_VERSION = os.environ.get('BKAPP_MESOS_VERSION')

# admin 权限用户
ADMIN_USERNAME = 'admin'
# BCS 默认业务
BCS_APP_ID = 1

# 项目功能白名单Code
PROJECT_FUNC_CODES = ['ServiceMonitor']

# 覆盖上层base中的DIRECT_ON_FUNC_CODE: 直接开启的功能开关，不需要在db中配置
DIRECT_ON_FUNC_CODE = ['HAS_IMAGE_SECRET', 'ServiceMonitor', 'BCS_API_GATEWAY_FOR_CLUSTER']

# 集群及节点metric功能白名单
CLUSTER_FUNC_CODES = ['MesosResource']

INSTALLED_APPS += [
    'backend.celery_app.CeleryConfig',
]

# ******************************** BCS 或 依赖服务 URL / ADDR ********************************
# 容器服务地址
DEVOPS_BCS_HOST = os.environ.get('BKAPP_BCS_UI_URL')
# 容器化版本值 与 DEVOPS_BCS_HOST 相同，后续统一合并
DEVOPS_HOST = DEVOPS_BCS_HOST

# CI系统API地址 TODO 私有化不再需要
DEVOPS_CI_API_HOST = os.environ.get('DEVOPS_CI_API_URL')

# 容器服务 API 地址
DEVOPS_BCS_API_URL = os.environ.get('BKAPP_BCS_UI_API_URL')
DEVOPS_ARTIFACTORY_HOST = os.environ.get('BKAPP_ARTIFACTORY_ADDR')

# TODO 旧蓝鲸网关地址, 先废弃置空
BCS_API_PRE_URL = ''

BK_PAAS_HOST = os.environ.get('BK_PAAS_URL', "http://bk-paas.example.com")
BK_PAAS_INNER_HOST = BK_PAAS_HOST

# 统一登录页面
LOGIN_SIMPLE = f"{BK_PAAS_HOST}/login/plain"
LOGIN_FULL = f"{BK_PAAS_HOST}/login/?c_url={DEVOPS_BCS_HOST}"
LOGIN_URL = f"{BK_PAAS_HOST}/login/"

# APIGW API 访问地址
APIGW_HOST = os.environ.get('BK_APIGW_URL', 'http://apigw.example.com')
APIGW_URL_TMPL = os.environ.get('bkApiUrlTmpl', 'http://bkapi.example.com/api/{api_name}')

# paas-cc 服务，后续接入 cmdb
BK_CC_HOST = os.environ.get('BKAPP_CC_URL', 'http://bcs-cc.example.com')
# BCS CC HOST
BCS_CC_API_PRE_URL = BK_CC_HOST

# 组件API地址（ESB）默认与 APIGW_HOST 相同
COMPONENT_HOST = APIGW_HOST
DEPOT_API = f'{APIGW_HOST}/api/harbor_api/'

# 数据平台清洗URL
_URI_DATA_CLEAN = '%2Fs%2Fdata%2Fdataset%2Finfo%2F{data_id}%2F%23data_clean'
URI_DATA_CLEAN = f'{BK_PAAS_HOST}?app=data&url=' + _URI_DATA_CLEAN

# SOPS API HOST TODO 接入新版 clustermanager 后废弃
SOPS_API_HOST = os.environ.get('BKAPP_SOPS_URL')

# 默认 BKCC 设备供应方，社区版默认 '0'
BKCC_DEFAULT_SUPPLIER_ACCOUNT = os.environ.get('BKCC_DEFAULT_SUPPLIER_ACCOUNT', '0')

# 可能有带端口的情况，需要去除
SESSION_COOKIE_DOMAIN = '.' + os.environ.get('ROOT_DOMAIN', 'example.com')
CSRF_COOKIE_DOMAIN = SESSION_COOKIE_DOMAIN

# ******************************** IAM & SSM 配置 ********************************
BK_SSM_HOST = os.environ.get('BKAPP_SSM_URL')

# BCS IAM 配置信息
BK_IAM_HOST = os.environ.get('BKAPP_IAM_URL', 'http://bkiam.example.com')
BK_IAM_SYSTEM_ID = APP_ID
BK_IAM_MIGRATION_APP_NAME = 'bcs_iam_migration'
BK_IAM_INNER_HOST = BK_IAM_HOST
# 参数说明 https://github.com/TencentBlueKing/iam-python-sdk/blob/master/docs/usage.md#22-config
# 如果通过网关访问, BK_IAM_APIGATEWAY_URL 将替代 BK_IAM_HOST
BK_IAM_USE_APIGATEWAY = True
BK_IAM_APIGATEWAY_URL = os.environ.get('BK_IAM_APIGATEWAY_URL', f"{APIGW_URL_TMPL.format(api_name='bk-iam')}/prod")
# 权限中心前端地址
BK_IAM_APP_URL = os.environ.get('BKAPP_IAM_APP_URL', f"{BK_PAAS_HOST}/o/bk_iam")
# 注册到权限中心的资源地址
BK_IAM_RESOURCE_API_HOST = os.environ.get('BK_IAM_RESOURCE_API_HOST', '')
BK_IAM_PROVIDER_PATH_PREFIX = os.environ.get('BK_IAM_PROVIDER_PATH_PREFIX', '/bcsapi/v4/iam-provider')

# ******************************** Helm 配置 ********************************
# kubectl 只有1.12版本
HELM_BASE_DIR = os.environ.get('HELM_BASE_DIR', BASE_DIR)
HELM_BIN = os.path.join(HELM_BASE_DIR, 'bin/helm')  # helm bin filename
HELM3_BIN = os.path.join(HELM_BASE_DIR, 'bin/helm3')
YTT_BIN = os.path.join(HELM_BASE_DIR, 'bin/ytt')
KUBECTL_BIN = os.path.join(HELM_BASE_DIR, 'bin/kubectl-v1.20.13')  # default kubectl bin filename
DASHBOARD_CTL_BIN = os.path.join(HELM_BASE_DIR, 'bin/dashboard-ctl')  # default dashboard ctl filename
KUBECTL_BIN_MAP = {
    '1.8.3': os.path.join(HELM_BASE_DIR, 'bin/kubectl-v1.12.3'),
    '1.12.3': os.path.join(HELM_BASE_DIR, 'bin/kubectl-v1.12.3'),
    '1.14.9': os.path.join(HELM_BASE_DIR, 'bin/kubectl-v1.14.9'),
    '1.16.3': os.path.join(HELM_BASE_DIR, 'bin/kubectl-v1.16.3'),
    '1.18.12': os.path.join(HELM_BASE_DIR, 'bin/kubectl-v1.18.12'),
    '1.20.13': os.path.join(HELM_BASE_DIR, 'bin/kubectl-v1.20.13'),
}
# 查询 helm release 状态需要的二进制版本
DASHBOARD_CTL_VERSION_MAP = {
    "v1": os.path.join(HELM_BASE_DIR, 'bin/dashboard-ctl'),
    "v2": os.path.join(HELM_BASE_DIR, 'bin/dashboard-ctl-v2'),
}

# 企业版/社区版 helm没有平台k8s集群时，无法为项目分配chart repo服务
# 为解决该问题，容器服务会绑定一个chart repo服务使用，所有项目公用这个chart repo
HELM_REPO_DOMAIN = os.environ.get("BKAPP_HARBOR_CHARTS_URL")
HELM_MERELY_REPO_USERNAME = os.environ.get("BKAPP_HARBOR_CHARTS_USERNAME")
HELM_MERELY_REPO_PASSWORD = os.environ.get("BKAPP_HARBOR_CHARTS_PASSWORD")

# BKE企业版证书
BKE_CACERT = os.path.join(HELM_BASE_DIR, 'etc/prod-server.crt')

# ******************************** 数据库 & 缓存 ********************************

DATABASES['default'] = {
    'ENGINE': 'django.db.backends.mysql',
    'NAME': os.environ.get('MYSQL_NAME', 'bcs-app'),
    'USER': os.environ.get('MYSQL_USER', 'root'),
    'PASSWORD': os.environ.get('MYSQL_PASSWORD', ''),
    'HOST': os.environ.get('MYSQL_HOST', '127.0.0.1'),
    'PORT': os.environ.get('MYSQL_PORT', '3306'),
    'OPTIONS': {
        'init_command': 'SET default_storage_engine=INNODB',
    },
}

REDIS_HOST = os.environ.get('REDIS_HOST', '127.0.0.1')
REDIS_PORT = os.environ.get('REDIS_PORT', 6379)
REDIS_DB = os.environ.get('REDIS_DB', 0)
REDIS_PASSWORD = os.environ.get('REDIS_PASSWORD', '')
REDIS_URL = os.environ.get('REDIS_URL', f'redis://:{REDIS_PASSWORD}@{REDIS_HOST}:{REDIS_PORT}/{REDIS_DB}')

# bkpaas_auth 模块会通过用户的 AccessToken 获取用户基本信息，因为这个 API 调用比较昂贵。
# 所以最好设置 Django 缓存来避免不必要的请求以提高效率。
CACHES = {
    'default': {
        'BACKEND': 'django_redis.cache.RedisCache',
        'LOCATION': REDIS_URL,
        'OPTIONS': {
            'CLIENT_CLASS': 'django_redis.client.DefaultClient',
        },
    }
}

# ******************************** 定时任务 & 消息队列 配置 ********************************
# CELERY 配置
IS_USE_CELERY = True

# 较高版本（>=4.0）的 celery 使用 CELERY_BROKER_URL
# ref: https://docs.celeryproject.org/en/latest/history/whatsnew-4.0.html#lowercase-setting-names
CELERY_BROKER_URL = os.environ.get('CELERY_BROKER_URL', REDIS_URL)

if IS_USE_CELERY:
    try:
        import django_celery_beat

        INSTALLED_APPS += ('django_celery_beat',)  # django_celery_beat

        CELERY_ENABLE_UTC = False
        CELERYBEAT_SCHEDULER = 'django_celery_beat.schedulers:DatabaseScheduler'
        if 'celery' in sys.argv:
            DEBUG = False
        if RUN_MODE == 'DEVELOP':
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
        print('use celery error: %s' % error)

CELERY_IMPORTS = ("backend.celery_app",)

# ******************************** WebConsole 配置  ********************************

# web-console配置需要，后台去除
RDS_HANDER_SETTINGS = {
    'level': 'INFO',
    'class': 'backend.utils.log.LogstashRedisHandler',
    'redis_url': REDIS_URL,
    'queue_name': 'paas_backend_log_list',
    'message_type': 'python-logstash',
    'tags': ['sz', 'stag', 'paas-backend'],
}

# web_console监听地址
WEB_CONSOLE_PORT = int(os.environ.get('WEB_CONSOLE_PORT', 28800))

# web_console运行模式, 支持external(平台托管), internal（自己集群托管）
WEB_CONSOLE_MODE = 'internal'

# web_console 镜像地址
WEB_CONSOLE_KUBECTLD_IMAGE_PATH = os.environ.get(
    'WEB_CONSOLE_KUBECTLD_IMAGE', f'{DEVOPS_ARTIFACTORY_HOST}/public/bcs/k8s/kubectld'
)

# ******************************** 监控 & 指标配置  ********************************

THANOS_HOST = os.environ.get("BKAPP_THANOS_URL")

# 默认指标数据来源，现在支持bk-data, prometheus
DEFAULT_METRIC_SOURCE = "prometheus"

# 普罗米修斯项目白名单
DEFAULT_METRIC_SOURCE_PROM_WLIST = []

# 容器化部署版本，不走 APIGW，直连 Cluster Service，需要指定 URL
BCS_CC_CLUSTER_CONFIG = "/v1/clusters/{cluster_id}/cluster_version_config/"
BCS_CC_GET_CLUSTER_MASTERS = "/projects/{project_id}/clusters/{cluster_id}/manager_masters/"
BCS_CC_GET_PROJECT_MASTERS = "/projects/{project_id}/clusters/null/manager_masters/"
BCS_CC_GET_PROJECT_NODES = "/projects/{project_id}/clusters/null/nodes/"
BCS_CC_OPER_PROJECT_NODE = "/projects/{project_id}/clusters/null/nodes/{node_id}/"
BCS_CC_OPER_PROJECT_NAMESPACES = "/projects/{project_id}/clusters/null/namespaces/"
BCS_CC_OPER_PROJECT_NAMESPACE = "/projects/{project_id}/clusters/null/namespaces/{namespace_id}/"

# 容器化部署版本，暂不需要提供 APIGW API，默认 PUBLIC KEY 设置为空值
BCS_APP_APIGW_PUBLIC_KEY = None

# 蓝鲸制品库域名，支持镜像仓库和chart仓库
BK_REPO_DOMAIN = os.environ.get("BK_REPO_DOMAIN", "")
DOCKER_REPO_DOMAIN = os.environ.get("DOCKER_REPO_DOMAIN", "")
HELM_REPO_DOMAIN = os.environ.get('HELM_REPO_DOMAIN')
BK_REPO_AUTHORIZATION = os.environ.get("BK_REPO_AUTHORIZATION", "")
# 设置蓝鲸制品库的公共项目和仓库名称，默认为bcs-shared
BK_REPO_SHARED_PROJECT_NAME = os.environ.get("BK_REPO_SHARED_PROJECT_NAME", "bcs-shared-project")
BK_REPO_SHARED_IMAGE_DEPOT_NAME = os.environ.get("BK_REPO_SHARED_IMAGE_DEPOT_NAME", "image-repo")
BK_REPO_SHARED_CHART_DEPOT_NAME = os.environ.get("BK_REPO_SHARED_CHART_DEPOT_NAME", "chart-repo")

# 集群管理的代理
CLUSTER_MANAGER_DOMAIN = BCS_APIGW_DOMAIN["prod"]

# 蓝鲸监控域名
BKMONITOR_HOST = os.environ.get("BKAPP_BKMONITOR_URL", "")
# 蓝鲸日志平台域名
BKLOG_HOST = os.environ.get("BKAPP_BKLOG_URL", "")

# 蓝鲸 opentelemetry trace 配置
# 是否开启 OTLP, 默认不开启
OPEN_OTLP = False
# 上报的地址
OTLP_GRPC_HOST = os.environ.get("OTLP_GRPC_HOST", "")
# 上报的 data id
OTLP_DATA_ID = os.environ.get("OTLP_DATA_ID", "")
# 上报时, 使用的服务名称
OTLP_SERVICE_NAME = APP_ID
