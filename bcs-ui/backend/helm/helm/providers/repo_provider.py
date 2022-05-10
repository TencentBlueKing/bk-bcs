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
from urllib.parse import ParseResult, urlparse

import jinja2
from django.conf import settings
from rest_framework.exceptions import APIException

from backend.helm.app.models import App
from backend.helm.app.utils import compose_url_with_scheme
from backend.helm.helm.providers.constants import PUBLIC_REPO_URL
from backend.utils.error_codes import error_codes

from ..models.chart import Chart, ChartVersion
from ..models.repo import Repository, RepositoryAuth
from ..utils.auth import BasicAuthGenerator
from .constants import CURATOR_VALUES_TEMPLATE
from .storage_provider import RGWProvider

logger = logging.getLogger(__name__)


class ChartRepoProvider:
    """
    Base Provider
    Chart Repo Provider, represented different repo backends, such as ChartMuseum、JFrog、Git
    """

    def provision(self, create_info):
        """
        transform ChartRepo model into real instance
        :return:
        """
        raise NotImplementedError


class PlainChartMuseumProvider:
    """
    Plain ChartMuseum Provider
    Support adding a third-part chartmuseum provider
    """

    def provision(self, create_info):
        """
        transform ChartRepo model into real instance
        :return:
        """
        create_info["provider"] = "plain_chartmuseum"
        create_info["storage_info"] = {}
        repo_fields = ['url', 'name', 'project_id', 'provider', 'storage_info']
        params = {key: create_info.get(key) for key in repo_fields}
        return Repository.objects.create(**params)


class ChartMuseumProvider(ChartRepoProvider):
    """
    Chart Museum Provider
    Aim to 2 jobs:
    - manage the Chart Repository & Auth model
    - deploying Chart Museum into cluster
    """

    def __init__(self):
        self.chart_repo_instance = None

    @property
    def storage_provider(self):
        return RGWProvider(settings.RGW_CONFIG)

    def provide_s3_storage(self, repo_info):
        try:
            # chartmuseum use s3 as storage backend as default
            storage_info = self.storage_provider.provision(repo_info)
        except Exception as e:
            message = "Storage provision failed: {}".format(e)
            logger.exception(message)
            raise APIException(message)
            # raise error_codes.ComponentError.f(message)
        return storage_info

    def provision(self, create_info):
        create_info["storage_info"] = self.provide_s3_storage(create_info)

        # fill with platform repo url
        url = "http://{platform_repo_domain}/{repo_env}/{repo_name}".format(
            platform_repo_domain=settings.PLATFORM_REPO_DOMAIN,
            repo_env=settings.HELM_REPO_ENV,
            repo_name=create_info["name"],
        )
        create_info["url"] = url
        try:
            self.chart_repo_instance = self._write_info(create_info)
            self._attach_auth()

            self._provision(create_info.get('operator'))
        except Exception as e:
            logger.exception("chartmuseum provider write info failed, %s" % e)

            # Q: Should we clean repo related every failure?
            # A: There is no exact answer in fact,
            # and we indeed should when create and provision is binding
            try:
                self.storage_provider.delete(create_info)
            except Exception as e:
                logger.exception("Delete storage backend failed, %s" % e)

            raise e

        return self.chart_repo_instance

    def delete(self, repo_info):
        """
        clean all resource including db record & k8s resource
        """
        # 1. delete k8s resource

        # 2. delete db record
        if self.chart_repo_instance:
            Repository.objects.get(id=self.chart_repo_instance.id).delete()

        logger.info("Chart repo instance<{}> deleted.".format(repo_info.get('name')))

    @staticmethod
    def _write_info(create_info):
        repo_fields = ['url', 'name', 'project_id', 'provider', 'storage_info']
        params = {key: create_info.get(key) for key in repo_fields}
        return Repository.objects.create(**params)

    def _attach_auth(self):
        """
        As default, provider needs 2 kind of auth
        one for pulling chart, the other for managing chart (pull & push)
        """
        role_list = ["admin"]
        # now only support basic auth
        for role in role_list:
            basic_auth = BasicAuthGenerator().generate_basic_auth_by_role(role)
            params = {"type": "basic", "role": role, "repo": self.chart_repo_instance, "credentials": basic_auth}
            RepositoryAuth.objects.create(**params)

    def _provision(self, operator=None):
        """
        deploy ChartMuseum to cluster
        """

        values = {}
        # 1. get storage info from repo instance
        values.update(self.chart_repo_instance.storage_info)

        # 2. get auth info
        for auth in self.chart_repo_instance.auths.all():
            if auth.type != "basic":
                logger.error(u"chartmuseum only support auth now.")
                return

            values.update(
                {
                    auth.role + "_username": auth.credentials.get('username'),
                    auth.role + "_password": auth.credentials.get('password'),
                }
            )

        # FIXME 需要去掉 chartmuseum 中的 readonly 账户之后移除这部分代码
        values.update(
            {
                "readonly_username": "disabled",
                "readonly_password": "disabled",
            }
        )

        # 3. insert url info
        url_result = urlparse(self.chart_repo_instance.url)
        repo_url_without_scheme = self.chart_repo_instance.url.lstrip(url_result.scheme + "://")
        values.update(
            {
                "repo_url": self.chart_repo_instance.url,
                "repo_name": self.chart_repo_instance.name,
                "repo_url_without_scheme": repo_url_without_scheme,
                "platform_repo_domain": settings.PLATFORM_REPO_DOMAIN,
                "repo_env": settings.HELM_REPO_ENV,
            }
        )

        # 4. render context
        rendered = jinja2.Template(CURATOR_VALUES_TEMPLATE).render(values)

        # 5. initialize k8s app
        chartmusum_curator_version = self._fetch_chart_museum()
        App.objects.initialize_app(
            name="%s-repo" % self.chart_repo_instance.name.lower(),
            project_id=settings.DEFAULT_MANAGE_CLUSTER["project_id"],
            cluster_id=settings.DEFAULT_MANAGE_CLUSTER.get('id'),
            namespace=settings.DEFAULT_REPO_NAMESPACE_INFO.get('name'),
            namespace_id=settings.DEFAULT_REPO_NAMESPACE_INFO.get('id'),
            chart_version=chartmusum_curator_version,
            answers=[],
            valuefile=rendered,
            creator=operator.username,
            updator=operator.username,
            customs=[],
            # 当用户点启用Helm时，这是用户的access_token, 无法用于部署repo到平台集群
            # 因此chartmuseum的部署需要使用独立的bke_token
            access_token=None,
            deploy_options={
                "kubeconfig_content": settings.KUBECOFNIG4REPO_DEPLOY,
                "extra_inject_source": settings.INJECTED_DATA_FOR_REPO,
                "ignore_empty_access_token": True,
            },
            unique_ns=App.objects.new_unique_ns(),  # escape unique restriction
        )

        self.chart_repo_instance.is_provisioned = True
        self.chart_repo_instance.save()

    @staticmethod
    def _fetch_chart_museum():
        # 1. fetch platform repo
        repo = Repository.objects.get(
            name=settings.PLATFORM_REPO_INFO.get('name'), project_id=settings.DEFAULT_MANAGE_CLUSTER.get('project_id')
        )

        # 2. get chart instance
        chart = Chart.objects.get(name=settings.DEFAULT_CURATOR_CHART.get('name'), repository=repo)

        # 3. return chart version
        chart_version = ChartVersion.objects.get(chart=chart, version=settings.DEFAULT_CURATOR_CHART.get('version'))

        return chart_version


def fetch_provider(provider_name):
    provider_name_map = {
        'chartmuseum': ChartMuseumProvider,
        'plain_chartmuseum': PlainChartMuseumProvider,
    }

    return provider_name_map[provider_name]


def add_platform_public_repos(target_project_id, repo_auth=None):
    """将平台提供的公用集群信息加入 target_project_id 对应的项目中"""
    repo = add_plain_repo(
        url=PUBLIC_REPO_URL, name="public-repo", target_project_id=target_project_id, repo_auth=repo_auth
    )

    return [repo]


def add_plain_repo(target_project_id, name, url, repo_auth=None):
    """为 target_project_id 对应的项目添加一个 plain 仓库"""
    # TODO support add plain repo with auth info
    repo = add_repo(
        target_project_id=target_project_id,
        user=None,
        url=url,
        name=name,
        provider_name="plain_chartmuseum",
    )

    if repo_auth is not None and not repo.auths.exists():
        """
        repo_auth = {
            "type": "basic",
            "role": role,
            "repo": self.chart_repo_instance,
            "credentials": basic_auth
        }
        """
        repo_auth["repo"] = repo
        repo_auth.setdefault("role", "admin")
        repo_auth.setdefault("type", "basic")
        RepositoryAuth.objects.create(**repo_auth)

    return repo


def add_repo(target_project_id, provider_name, user, name, url):
    """add repo for target_project_id
    Provider_name can be chartmuseum and plain_chartmuseum.
    For plain_chartmueum type, it will just add one record to database.
    For provider of type `chartmuseum` it will deploy a chartmuseum instance for project_id,
    in platform project/cluster settings.DEFAULT_MANAGE_CLUSTER
    """
    repo_info = {"url": url, "name": name, "project_id": target_project_id, "provider": provider_name}

    if Repository.objects.filter(name=name, project_id=target_project_id).exists():
        return Repository.objects.get(name=name, project_id=target_project_id)

    try:
        provider_cls = fetch_provider(provider_name)
        provider_instance = provider_cls()
    except KeyError:
        logger.warning("Provider: {} not supported yet".format(provider_name))
        raise error_codes.CheckFailed.f("Provider: {} not supported yet".format(provider_name))

    repo_info["target_project_id"] = target_project_id
    repo_info.update({'operator': user})

    try:
        new_chart_repo = provider_instance.provision(create_info=repo_info)
    except Exception as e:
        logger.exception("Chart provision failed, %s" % e)

        try:
            provider_instance.delete(repo_info)
        except Exception as e:
            logger.exception("Delete chart repo failed, %s" % e)
            raise error_codes.ComponentError.f("Delete chart repo failed: {}".format(e))
        raise error_codes.ComponentError.f("Chart provision failed: {}".format(e))
    else:
        return new_chart_repo
