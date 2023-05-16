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
# 本地开发配置文件
from .base import *  # noqa

SECRET_KEY = os.environ.get('SECRET_KEY')

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

# 本地开发先去除权限中心v3的数据初始逻辑
INSTALLED_APPS.remove("backend.iam.bcs_iam_migration.apps.BcsIamMigrationConfig")

# ******************************** 日志 配置 ********************************
LOG_LEVEL = "DEBUG"
LOGGING = get_logging_config(LOG_LEVEL)
