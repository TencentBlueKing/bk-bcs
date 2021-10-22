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

import jwt
from django.conf import settings
from django.contrib.auth import get_user_model
from jwt import exceptions as jwt_exceptions
from rest_framework.authentication import BaseAuthentication, SessionAuthentication

from backend.components import ssm
from backend.components.utils import http_get
from backend.utils import FancyDict, cache

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


@cache.region.cache_on_arguments(expiration_time=240)
def get_access_token_by_credentials(bk_token):
    """Request a new request token by credentials"""
    return ssm.get_bk_login_access_token(bk_token)


class SSMAccessToken(object):
    def __init__(self, credentials):
        self.credentials = credentials
        self.bk_token = credentials["bk_token"]

    @property
    def access_token(self):
        data = get_access_token_by_credentials(self.bk_token)
        return data["access_token"]


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
        if not credentials or credentials != auth_credentials:
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
        return (user, None)


try:
    from .authentication_ext import BKTicketAuthentication, BKTicketAuthenticationBackend
except ImportError as e:
    logger.debug('Load extension failed: %s', e)
