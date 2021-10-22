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

import boto

from .rgw_admin import RGWAdminClient

logger = logging.getLogger(__name__)


class StorageProvider:
    """
    Base Provider
    Storage Provider, now mainly for repo storage backend
    """

    def provision(self, param):
        raise NotImplementedError


class RGWProvider(StorageProvider):

    """
    provider bucket
    the name of repo will be named to bucket and user
    """

    def __init__(self, config):
        self.access_key = config['access_key']
        self.secret_key = config['secret_key']
        self.admin_host = config['admin_host']
        self.admin_endpoint = config['admin_endpoint']
        self.tenant = config['tenant']
        self.default_policy = config['default_policy']
        self.max_size = config['max_size']

    def _get_client(self):
        client = RGWAdminClient(
            access_key=self.access_key,
            secret_key=self.secret_key,
            admin_host=self.admin_host,
            admin_endpoint=self.admin_endpoint,
            tenant=self.tenant,
        )
        return client

    def provision(self, params):
        """
        申请存储空间，并创建 Bucket
        """
        project_id = params.get("project_id")
        if not project_id:
            raise Exception("Lack project id when provision ceph rgw storage")
        # use repo name and project id as unique key
        # uuid is too long to display for bucket name
        repo_name = "%s-%s-chart-repo" % (project_id[:7], params.get("name"))

        bucket = repo_name
        client = self._get_client()
        user = client.get_user_info(repo_name)
        if user:
            client.modify_user(repo_name, email='', display_name='Chart Repo: %s', max_buckets=1)
        else:
            client.create_user(repo_name, email='', display_name='Chart Repo: %s', max_buckets=1)

        # 设置用户容量限制
        client.set_user_quota(repo_name, enabled=True, max_size_kb=self.max_size, max_objects=-1)

        # 使用新建用户的身份创建 buckets
        user = client.get_user_info(repo_name)
        conn = boto.connect_s3(
            aws_access_key_id=user['keys'][0]['access_key'],
            aws_secret_access_key=user['keys'][0]['secret_key'],
            host=self.admin_host,
            is_secure=False,
            calling_format=boto.s3.connection.OrdinaryCallingFormat(),
        )

        conn.create_bucket(bucket, policy=self.default_policy)
        raw_credentials = {
            'aws_access_key_id': user['keys'][0]['access_key'],
            'aws_secret_access_key': user['keys'][0]['secret_key'],
            'rgw_host': self.admin_host,
            'rgw_url': "http://%s/" % self.admin_host,
            'bucket': bucket,
            'max_size': self.max_size,
        }

        return raw_credentials

    def delete(self, params):
        project_id = params.get("project_id")
        if not project_id:
            raise Exception("Lack project id when delete ceph rgw storage")
        repo_name = "%s-%s-chart-repo" % (project_id[:7], params.get("name"))
        bucket = repo_name
        client = self._get_client()
        # user must exist
        user = client.get_user_info(repo_name)
        conn = boto.connect_s3(
            aws_access_key_id=user['keys'][0]['access_key'],
            aws_secret_access_key=user['keys'][0]['secret_key'],
            host=self.admin_host,
            is_secure=False,
            calling_format=boto.s3.connection.OrdinaryCallingFormat(),
        )
        conn.delete_bucket(bucket)
        logger.info("S3 bucket<%s> delete successfully." % bucket)
        return True
