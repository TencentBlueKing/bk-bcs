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

A rgw admin API example using boto
"""
import json
import urllib.parse

import boto
import boto.s3.connection
from boto.connection import AWSAuthConnection
from boto.exception import StorageResponseError


class RGWAdminClient(object):
    def __init__(self, access_key, secret_key, admin_host, admin_endpoint, tenant=None):
        self.access_key = access_key
        self.secret_key = secret_key
        self.admin_host = admin_host
        self.admin_endpoint = admin_endpoint
        self.tenant = tenant

        self.conn = boto.connect_s3(
            aws_access_key_id=access_key,
            aws_secret_access_key=secret_key,
            host=admin_host,
            is_secure=False,
            calling_format=boto.s3.connection.OrdinaryCallingFormat(),
        )

    def _handle_response(self, response):
        body = response.read()
        if response.status == 200:
            return json.loads(body) if body else body
        else:
            raise StorageResponseError(response.status, response.reason, body)

    def get_user_info(self, uid, tenant=None):
        """根据 uin 获取用户信息"""
        tenant = tenant or self.tenant
        if tenant:
            uid = '%s$%s' % (tenant, uid)

        parameters = {'uid': uid}
        response = AWSAuthConnection.make_request(
            self.conn,
            'GET',
            self.admin_endpoint + 'user?' + urllib.parse.urlencode(parameters),
        )
        body = response.read()
        if response.status == 200:
            return json.loads(body)
        elif response.status == 404:
            return None
        else:
            raise StorageResponseError(response.status, response.reason, body)

    def update_or_modify_user(self, method, uid, tenant=None, display_name='', email='', max_buckets=100):
        tenant = tenant or self.tenant
        if tenant:
            uid = '%s$%s' % (tenant, uid)

        parameters = {'uid': uid, 'display-name': display_name, 'email': email, 'max-buckets': max_buckets}
        response = AWSAuthConnection.make_request(
            self.conn,
            method,
            self.admin_endpoint + 'user?' + urllib.parse.urlencode(parameters),
            # data=urllib.urlencode(parameters)
        )
        return self._handle_response(response)

    def create_user(self, *args, **kwargs):
        """创建一个新用户

        - 当 uid 一样，而 display_name 不一样时，会返回用户已经存在错误。反之，如果 uid 和
          display_name 都一样，那么调用该接口将会为用户添加一组 key
        """
        return self.update_or_modify_user('PUT', *args, **kwargs)

    def modify_user(self, *args, **kwargs):
        """修改一个用户"""
        return self.update_or_modify_user('POST', *args, **kwargs)

    def link_bucket(self, uid, bucket, tenant=''):
        """将某个 bucket 绑定到指定 uid"""
        tenant = tenant or self.tenant
        if tenant:
            uid = '%s$%s' % (tenant, uid)

        parameters = {
            'uid': uid,
            'bucket': bucket,
        }
        response = AWSAuthConnection.make_request(
            self.conn,
            'PUT',
            self.admin_endpoint + 'bucket?' + urllib.parse.urlencode(parameters),
        )
        return self._handle_response(response)

    def get_user_quota(self, uid, tenant=''):
        """获取用户容量限制"""
        tenant = tenant or self.tenant
        if tenant:
            uid = '%s$%s' % (tenant, uid)

        parameters = {'uid': uid, 'quota-type': 'user'}
        response = AWSAuthConnection.make_request(
            self.conn,
            'GET',
            self.admin_endpoint + 'user?quota&' + urllib.parse.urlencode(parameters),
        )
        return self._handle_response(response)

    def set_user_quota(self, uid, tenant='', enabled=False, max_objects=-1, max_size_kb=-1):
        """获取用户容量限制"""
        tenant = tenant or self.tenant
        if tenant:
            uid = '%s$%s' % (tenant, uid)

        parameters = {'uid': uid, 'quota-type': 'user', 'quota-scope': 'user'}
        data = {
            'uid': uid,
            'quota-scope': 'user',
            'enabled': enabled,
            'max_objects': max_objects,
            'max_size_kb': max_size_kb,
        }
        response = AWSAuthConnection.make_request(
            self.conn,
            'PUT',
            self.admin_endpoint + 'user?quota&' + urllib.parse.urlencode(parameters),
            data=json.dumps(data),
        )
        return self._handle_response(response)
