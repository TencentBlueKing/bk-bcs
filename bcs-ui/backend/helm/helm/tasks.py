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

import datetime
import logging
from typing import Dict

from celery import shared_task
from django.utils import timezone
from natsort import natsorted

from backend.apps.whitelist import enable_incremental_sync_chart_repo
from backend.helm.helm.utils.repo_bk import get_incremental_charts_and_hash_value
from backend.utils.basic import normalize_time

from .models.chart import Chart, ChartVersion
from .models.repo import Repository
from .utils.repo import InProcessSign, prepareRepoCharts

logger = logging.getLogger(__name__)


@shared_task
def sync_all_repo():
    for repo in Repository.objects.all():
        try:
            sync_helm_repo(repo.id, False)
        except Exception as e:
            logger.exception("sync_helm_repo %s failed %s" % (repo.id, e))


@shared_task
def force_sync_all_repo():
    for repo in Repository.objects.all():
        try:
            sync_helm_repo(repo.id, True)
        except Exception as e:
            logger.exception("force sync_helm_repo %s failed %s" % (repo.id, e))


def enable_increment(force, project_id):
    """判断是否增量功能
    NOTE: 当force为false，并且在白名单中才允许增量操作
    """
    if not force and enable_incremental_sync_chart_repo(project_id):
        return True

    return False


@shared_task
def sync_helm_repo(repo_id, force=False):
    # if in processing, then do nothing
    sign = InProcessSign(repo_id)
    # TODO: FIXME: uncomment
    if sign.exists():
        logger.error("the helm repo %s if in processing, this task will not be started", repo_id)
        return

    repo = Repository.objects.get(id=repo_id)
    repo_name = repo.name
    repo_url = repo.url
    plain_auths = repo.plain_auths

    try:
        # NOTE: 针对白名单中的项目先开启增量同步
        if enable_increment(force, repo.project_id):
            if not plain_auths:
                username, password = None, None
            else:
                credentials = plain_auths[0]["credentials"]
                username, password = credentials["username"], credentials["password"]
            start_time = normalize_time(repo.refreshed_at)
            charts_info, charts_info_hash = get_incremental_charts_and_hash_value(
                repo_name, username, password, start_time
            )
        else:
            charts_info, charts_info_hash = prepareRepoCharts(repo_url, repo_name, plain_auths)
    except Exception as e:
        logger.exception("prepareRepoCharts fail: repo_url=%s, repo_name=%s, error: %s", repo_url, repo_name, e)
        return

    logger.debug("prepareRepoCharts repo_url=%s, charts_info=%s", repo_url, charts_info)

    # 如果不存在或者为空，认为同步失败
    if not charts_info:
        logger.error("load chart info from repo fail![name=%s, url=%s]", repo_name, repo_url)
        sign.delete()
        return

    # if the index_hash is the same as the commit in db
    # 现阶段兼容先前逻辑，仍然比对MD5，判断是否需要更新
    if not force and charts_info_hash == repo.commit:
        logger.info(
            "the chart index commit [%s] of repo %s not been update since last refresh: %s",
            repo.commit,
            repo_id,
            repo.refreshed_at,
        )
        return

    try:
        # 增量获取的数据，直接添加到本地记录
        if enable_increment(force, repo.project_id):
            _add_charts(repo, sign, charts_info, charts_info_hash, force)
        else:
            _do_helm_repo_charts_update(repo, sign, charts_info, charts_info_hash, force)
    except Exception as e:
        logger.exception("_do_helm_repo_charts_update fail, error: %s", e)
        sign.delete()


def _add_charts(repo, sign, charts, index_hash, force=False):
    """添加charts"""
    for chart_name, versions in charts.items():
        chart, _created = Chart.objects.get_or_create(name=chart_name, repository=repo)
        # 开始添加版本
        full_chart_versions = {}
        for version in versions:
            # update sign every time
            sign.update()
            chart_version = ChartVersion()
            try:
                chart_version.update_from_import_version(chart, version, force)
            except Exception as e:
                logger.exception("save_chart_version fail: chart=%s, version=%s, error: %s", chart, version, e)
                continue
            # 更新chart icon
            icon_url = version.get("icon")
            if not chart.icon and icon_url:
                chart.update_icon(icon_url)
            # 记录版本，便于更新chart默认版本
            full_chart_versions[chart_version.id] = chart_version
        # 更新chart的变更时间
        chart.changed_at = timezone.now()
        chart.save(update_fields=["changed_at"])
        # 更新chart默认版本为最新推送的chart版本
        _update_default_chart_version(chart, full_chart_versions)
    # 更新hash
    repo.refreshed(index_hash)
    # delete sign
    sign.delete()


def _update_default_chart_version(chart, full_chart_versions):
    """更新chart对应的默认版本信息"""
    if not full_chart_versions:
        return
    # 以created逆序
    all_versions = list(full_chart_versions.values())
    # 转换为字符串做时间对比，以兼容出现string和datatime的格式时间对比
    all_versions.sort(key=lambda info: str(info.created), reverse=True)

    # 如果latest_chart_version和先前的版本一致，则无需更新
    latest_chart_version = all_versions[0]
    # 处理异常: backend.helm.helm.models.chart.DoesNotExist: ChartVersion matching query does not exist.
    try:
        if chart.defaultChartVersion and (chart.defaultChartVersion.version == latest_chart_version.version):
            return
    except Exception as err:
        logger.error("match chart verson failed, err: %s", err)

    chart.defaultChartVersion = latest_chart_version
    chart.description = latest_chart_version.description
    chart.save()


def _sync_delete_chart_versions(chart, old_chart_versions, full_chart_versions, to_delete_ids):
    # 1. delete chart version
    for vid in to_delete_ids:
        try:
            old_chart_versions.get(vid).delete()
            # del old_chart_versions[vid]
        except Exception as e:
            logger.exception("sync_helm_repo: delete old chartVersion fail![ChartVersionID=%s], error: %s", vid, e)


def _sync_delete_charts(charts, old_charts):
    current_chart_names = charts.keys()
    to_delete_chart_names = set(old_charts.keys()) - set(current_chart_names)

    for chart_name in to_delete_chart_names:
        c = old_charts[chart_name]
        c.do_delete()


def _get_old_charts(repo):
    charts = Chart.objects.filter(repository=repo).all()
    old_charts_dict = {c.name: c for c in charts}

    return old_charts_dict


def _get_old_chart_version_data(chart, created):
    # if created, no old chart versions
    if created:
        return {}, {}

    versions = ChartVersion.objects.filter(chart=chart).all()

    # {key => id}
    old_chart_version_key_ids = {ChartVersion.gen_key(v.name, v.version, v.digest): v.id for v in versions}
    # {id => object}
    old_chart_versions = {v.id: v for v in versions}

    return old_chart_version_key_ids, old_chart_versions


def update_chart_annotations(chart: Chart, version: Dict):
    """更新chart的annotations"""
    # NOTE: 仓库chart中annotations不为空，并且和bcs中存储不一致时，进行更新
    chart_annotations = version.get("annotations") or {}
    if chart_annotations and chart_annotations != chart.annotations:
        chart.annotations = chart_annotations
        chart.save()


def _do_helm_repo_charts_update(repo, sign, charts, index_hash, force=False):
    # for sync chart, some chart maybe delete
    old_charts = _get_old_charts(repo)

    for chart_name, versions in charts.items():
        chart, chart_created = Chart.default_objects.get_or_create(name=chart_name, repository=repo)
        chart.clean_deleted_status()

        # 1. prepare data
        old_chart_version_key_ids, old_chart_versions = _get_old_chart_version_data(chart, chart_created)

        # 2. add or update
        full_chart_versions = {}
        current_chart_version_ids = []
        chart_changed = False
        for version in versions:
            # 2.1 update sign every time
            sign.update()

            # 2.2 create chart version
            key = ChartVersion.gen_key(
                name=version.get("name"), version=version.get("version"), digest=version.get("digest")
            )
            # 如果数据库中已经存在记录，并且不是强制同步，则不进行其它信息的变动
            chart_version_id = old_chart_version_key_ids.get(key)
            if chart_version_id:
                # 记录相关数据供删除使用
                chart_version = old_chart_versions.get(chart_version_id)
                current_chart_version_ids.append(chart_version.id)
                full_chart_versions[chart_version.id] = chart_version
                if not force:
                    continue
            else:
                chart_version = ChartVersion()

            # 2.3 do update
            try:
                version_changed = chart_version.update_from_import_version(chart, version, force)
            except Exception as e:
                logger.exception(
                    "_save_or_update_chart_version fail: chart=%s, version=%s, error: %s", chart, version, e
                )
                continue
            else:
                chart_changed = chart_changed or version_changed

            # 2.4 update icon  NOTE: icon just add at the first time
            # 验证 版本号不变动时，icon，desc不会更新
            icon_url = version.get("icon")
            if not chart.icon and icon_url:
                chart.update_icon(icon_url)

            # 更新chart annotations
            update_chart_annotations(chart, version)

            # 针对新创建的chart version，记录chart全量版本，便于后续针对版本的处理
            if not chart_version_id:
                current_chart_version_ids.append(chart_version.id)
                full_chart_versions[chart_version.id] = chart_version

        if chart_changed:
            chart.changed_at = timezone.now()
            chart.save(update_fields=["changed_at"])

        # 3. chartVersion sync delete
        to_delete_ids = set(old_chart_versions.keys()) - set(current_chart_version_ids)
        _sync_delete_chart_versions(chart, old_chart_versions, full_chart_versions, to_delete_ids)

        # 更新chart默认版本为最新推送的chart版本
        _update_default_chart_version(chart, full_chart_versions)

    # sync chart
    _sync_delete_charts(charts, old_charts)

    # update refreshed_at and commit
    repo.refreshed(index_hash)

    # delete sign
    sign.delete()
