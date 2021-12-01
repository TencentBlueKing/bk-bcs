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

import arrow
import jwt
from django.conf import settings
from django.contrib.auth import get_user_model
from jwt import exceptions as jwt_exceptions
from rest_framework.authentication import BaseAuthentication, SessionAuthentication

from backend.components import ssm
from backend.components.utils import http_get
from backend.utils import FancyDict
from backend.utils.cache import region

logger = logging.getLogger(__name__)

User = get_user_model()


class JWTUser(User):
    @property
    def is_authenticated(self):
        return True

    class Meta(object):
        app_label = 'bkpaas_auth'


class JWTClient(object):
    def __init__(self, content):
        self.content = content
        self.payload = {}
        self.headers = {}

    @property
    def project(self):
        return FancyDict(self.payload.get('project') or {})

    @property
    def user(self):
        return FancyDict(self.payload.get('user') or {})

    @property
    def app(self):
        return FancyDict(self.payload.get('app') or {})

    def is_valid(self, apigw_public_key=None):
        if not self.content:
            return False

        try:
            if apigw_public_key is None:
                apigw_public_key = settings.APIGW_PUBLIC_KEY

            self.headers = jwt.get_unverified_header(self.content)
            self.payload = jwt.decode(self.content, apigw_public_key, issuer='APIGW')
            return True
        except jwt_exceptions.InvalidTokenError as error:
            logger.error("check jwt error, %s", error)
            return False
        except Exception:
            logger.exception("check jwt exception")
            return False

    def __str__(self):
        return '<%s, %s>' % (self.headers, self.payload)


class JWTAuthentication(BaseAuthentication):
    JWT_KEY_NAME = 'HTTP_X_BKAPI_JWT'

    def authenticate(self, request):
        client = JWTClient(request.META.get(self.JWT_KEY_NAME, ''))
        if not client.is_valid():
            return None

        user = JWTUser(username=client.user.username)
        return (user, None)


class CsrfExceptSessionAuthentication(SessionAuthentication):
    def enforce_csrf(self, request):
        return


class NoAuthError(Exception):
    pass


def get_access_token_by_credentials(bk_token):
    """Request a new request token by credentials"""
    cache_key = f'BK_BCS:USER_ACCESS_TOKEN_INFO:{bk_token}'
    # 每过【一小时】必定失效，需要重新获取
    token_info = region.get(cache_key, expiration_time=60 * 60)
    # 获取不到 access_token 信息 或 被标记为过期 都需要重新获取
    if not token_info or token_info['expires_at'] < arrow.now():
        resp = ssm.get_bk_login_access_token(bk_token)
        token_info = {
            'access_token': resp['access_token'],
            'expires_at': arrow.now().shift(seconds=resp['expires_in']),
        }
        region.set(cache_key, token_info)
    return token_info['access_token']


class SSMAccessToken(object):
    def __init__(self, credentials):
        self.credentials = credentials
        self.bk_token = credentials["bk_token"]

    @property
    def access_token(self):
        return get_access_token_by_credentials(self.bk_token)

    def is_valid(self) -> bool:
        """
        当 access_token 缓存失效时，会发起重新获取，如果未获取到正确的 access_token, 则校验不通过，返回 False
        """
        try:
            _ = self.access_token
        except Exception as e:
            logger.error('no valid access_token: %s', e)
            return False
        else:
            return True


class BKTokenAuthentication(BaseAuthentication):
    """企业版bk_token校验"""

    def verify_bk_token(self, bk_token):
        """校验是否"""
        url = f"{settings.BK_PAAS_INNER_HOST}/login/accounts/is_login/"
        params = {"bk_token": bk_token}
        resp = http_get(url, params=params)
        if resp.get("result") is not True:
            raise NoAuthError(resp.get("message", ""))

        return resp["data"]["username"]

    def get_credentials(self, request):
        return {
            "bk_token": request.COOKIES.get("bk_token"),
        }

    def get_user(self, username):
        user_model = get_user_model()
        defaults = {"is_active": True, "is_staff": False, "is_superuser": False}
        user, _ = user_model.objects.get_or_create(username=username, defaults=defaults)
        return user

    def authenticate(self, request):
        auth_credentials = self.get_credentials(request)
        if not auth_credentials["bk_token"]:
            return None

        credentials = request.session.get("auth_credentials")
        if not credentials or credentials["bk_token"] != auth_credentials["bk_token"]:
            try:
                username = self.verify_bk_token(**auth_credentials)
            except NoAuthError as e:
                logger.info("%s authentication error: %s", auth_credentials["bk_token"], e)
                return None
            except Exception as e:
                logger.exception("ticket authentication error: %s", e)
                return None

            # 缓存auth_credentials
            auth_credentials["username"] = username
            request.session["auth_credentials"] = auth_credentials
        else:
            username = credentials["username"]

        user = self.get_user(username)
        user.token = SSMAccessToken(auth_credentials)
        # 增加校验 access_token 的有效性
        if not user.token.is_valid():
            return None

        return (user, None)


try:
    from .authentication_ext import BKTicketAuthentication, BKTicketAuthenticationBackend
except ImportError as e:
    logger.debug('Load extension failed: %s', e)
