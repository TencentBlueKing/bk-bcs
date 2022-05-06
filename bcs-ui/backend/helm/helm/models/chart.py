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
import json
import logging
from typing import Dict, Tuple

import yaml
from django.conf import settings
from django.db import models
from django.utils.crypto import get_random_string
from jsonfield import JSONField

from backend.helm.helm.bcs_variable import get_bcs_variables, merge_valuefile_with_bcs_variables
from backend.helm.toolkit.diff import parser
from backend.helm.toolkit.kubehelm.helm import KubeHelmClient
from backend.utils.basic import normalize_time
from backend.utils.models import BaseTSModel

from .. import constants
from ..utils.repo import download_icon_data, download_template_data
from ..utils.util import fix_chart_url, merge_rancher_answers
from .managers import ChartManager, ChartVersionManager, ChartVersionSnapshotManager

logger = logging.getLogger(__name__)
ALLOWED_CHARS = 'abcdefghijklmnopqrstuvwxyz0123456789'


class Chart(BaseTSModel):
    """
    Helm Chart
    unique: name + repository
    """

    name = models.CharField(max_length=50)
    repository = models.ForeignKey("Repository", on_delete=models.CASCADE)
    description = models.CharField(max_length=1000, blank=True, null=True, default="")
    defaultChartVersion = models.ForeignKey(
        "ChartVersion", related_name="default_chart_version", null=True, on_delete=models.SET_NULL
    )
    # base64, format: [data:{content-type};base64,b64string]
    icon = models.TextField()

    deleted = models.BooleanField(default=False)
    deleted_at = models.DateTimeField(null=True, blank=True)

    # field show when content changed
    changed_at = models.DateTimeField(default=datetime.datetime.now, blank=True)
    # chart.yaml支持annotations字段https://helm.sh/docs/topics/charts/#the-chartyaml-file
    annotations = JSONField(default={})

    objects = ChartManager()
    default_objects = models.Manager()

    class Meta:
        unique_together = ('name', 'repository')
        db_table = 'helm_chart'

    def __str__(self):
        return "{repo}/{name}".format(repo=self.repository, name=self.name)

    # TODO: change to SLZ
    def to_json(self):
        chart_version_fields = {}
        if self.defaultChartVersion is not None:
            chart_version_fields = self.defaultChartVersion.to_json()

            del chart_version_fields["id"]
            del chart_version_fields["chart"]

        fields = chart_version_fields
        fields.update(
            {
                "name": self.name,
                "repository_id": self.repository.id,
                "icon": self.icon,
                "annotations": self.annotations,
            }
        )
        return fields

    def update_icon(self, icon_url):
        ok, icon_data = download_icon_data(icon_url, auths=self.repository.plain_auths)
        if ok:
            self.icon = icon_data
            self.save()
            return True
        return False

    def do_delete(self):
        self.deleted = True
        self.deleted_at = datetime.datetime.now()
        self.save()

    def clean_deleted_status(self):
        self.deleted = False
        self.deleted_at = None
        self.save()


class BaseChartVersion(BaseTSModel):
    """
    refer: https://github.com/kubernetes/helm/blob/master/pkg/proto/hapi/chart/metadata.pb.go#L75:6
    """

    name = models.CharField(max_length=50)
    home = models.CharField(max_length=200, null=True)
    description = models.CharField(max_length=1000, blank=True, null=True, default="")
    engine = models.CharField(max_length=20)

    created = models.DateTimeField()

    # all json fields, most of them are just for show in html
    maintainers = JSONField(null=True, default=[])
    sources = JSONField(null=True, default=[])
    urls = JSONField(default=[])
    files = JSONField(default={})
    questions = JSONField(default={})

    def to_json(self):
        from django.forms.models import model_to_dict

        values = model_to_dict(
            self, fields=[field.name for field in self._meta.fields if not field.name.startswith("_")]
        )
        values.update(
            {
                "maintainers": self.maintainers,
                "sources": self.sources,
                "urls": self.urls,
                "files": self.files,
                "questions": self.questions,
            }
        )
        return values

    @property
    def chart_info(self):
        default = {"description": "", "name": self.name, "version": self.version}
        for key in ["Chart.yaml", "%s/Chart.yaml" % self.name]:
            content = self.files.get(key)
            if content:
                break

        if not content:
            return default

        try:
            return yaml.load(content)
        except Exception as e:
            logger.exception("BaseChartVersion:%s load Chart.yaml failed, %s" % (self.id, e))
            return default

    class Meta:
        abstract = True


class ChartVersion(BaseChartVersion):
    """
    ChartVersion, 随着repo refresh, 随时同步更新, 是动态的(同一个版本的包/文件等, 都可能存在更新)
    project-repo唯一, 全局不唯一
    uniq: chart + version
    """

    chart = models.ForeignKey("Chart", on_delete=models.CASCADE, related_name='versions')
    keywords = models.CharField(max_length=200, null=True, blank=True)
    version = models.CharField(max_length=255)
    digest = models.CharField(max_length=64)

    objects = ChartVersionManager()

    class Meta:
        db_table = 'helm_chart_version'
        unique_together = ('chart', 'version', 'digest')

    def __str__(self):
        return "{chart}/{version}".format(chart=self.chart, version=self.version)

    @staticmethod
    def gen_key(name, version, digest):
        return "{name}-{version}-{digest}".format(name=name, version=version, digest=digest)

    def update_from_import_version(self, chart, version, force=False):
        """
        - read from index.yaml
        -> entries -> chart -> versions -> version
        """
        self.chart = chart
        self.version = version.get("version")
        self.name = version.get("name")
        self.home = version.get("home")
        self.description = version.get("description")
        self.engine = version.get("engine", "default")
        self.created = normalize_time(version.get("created"))
        maintainers = version.get("maintainers")
        if maintainers:
            self.maintainers = version.get("maintainers")
        sources = version.get("sources")
        if sources:
            self.sources = sources
        urls = version.get("urls")

        # fix url which don't contains repo url
        repo_url = self.chart.repository.url
        for idx, url in enumerate(urls):
            urls[idx] = fix_chart_url(url, repo_url)
        self.urls = urls

        keywords = version.get("keywords")
        if keywords:
            self.keywords = ','.join(keywords)

        old_digest = self.digest
        current_digest = version.get("digest")

        # 标识是否有变动
        # 当digest变动时，肯定chart有变动
        # 当repo server允许版本覆盖时，内容有变动但是digest相同，这时也认为有变动
        chart_version_changed = old_digest != current_digest
        if force or chart_version_changed:
            self.digest = version.get("digest")
            # donwload the tar.gz and update files and questions
            url = self.urls[0] if self.urls else None
            if not url:
                self.save()
                return chart_version_changed

            ok, files, questions = download_template_data(chart.name, url, auths=self.chart.repository.plain_auths)
            if not ok:
                self.save()
                return chart_version_changed

            if self.files != files:
                self.files = files
                chart_version_changed = True
            if self.questions != questions:
                self.questions = questions
                chart_version_changed = True

        self.save()
        return chart_version_changed

    @classmethod
    def update_or_create_version(cls, chart: Chart, version: Dict) -> Tuple["ChartVersion", bool]:
        chart_version, created = cls.objects.update_or_create(chart=chart, version=version.get("version"))
        chart_version.update_from_import_version(chart, version)
        return chart_version, created


class ChartVersionSnapshot(BaseChartVersion):
    """
    ChartVersionSnapshot, 一旦chart do release, 会生成一个snapshot, 不允许修改
    只要digest, 就能确定引用的是同一个包, 全局唯一
    uniq: digest
    """

    version = models.CharField(max_length=255)
    version_id = models.IntegerField(default=-1, help_text="record the chart version which this snapshot comes from")

    # digest shouldn't be unique, as multi app may use same version.
    # Make digest unique must add foreignkey chart and index_together.
    digest = models.CharField(max_length=64, unique=True)

    objects = ChartVersionSnapshotManager()

    def __str__(self):
        return "{name}/{version}".format(name=self.name, version=self.version)

    class Meta:
        db_table = 'helm_chart_version_snapshot'

    @property
    def state(self):
        try:
            chart_version = ChartVersion.objects.get(id=self.version_id)
        except ChartVersion.DoesNotExist:
            return "deleted"
        else:
            return "unchanged" if chart_version.digest == self.digest else "changed"

    @property
    def version_detail(self):
        from django.forms.models import model_to_dict

        values = model_to_dict(
            self,
            fields=[
                "version",
                "digest",
                "name",
                "home",
                "description",
                "engine",
                "created",
                "maintainers",
                "sources",
                "urls",
                "files",
                "questions",
            ],
        )

        return values


class ChartReleaseManager(models.Manager):
    def create(self, **kwargs):
        kwargs["short_name"] = get_random_string(constants.CHART_RELEASE_SHOT_NAME_LENGTH, allowed_chars=ALLOWED_CHARS)
        return super(ChartReleaseManager, self).create(**kwargs)

    def make_upgrade_release(
        self, app, chart_version_id, answers, customs, valuefile="", valuefile_name=constants.DEFAULT_VALUES_FILE_NAME
    ):
        # make upgrade for app
        # `chart_version_id` indicate the target chartverion for app,
        # it can also use KEEP_TEMPLATE_UNCHANGED to keep app template unchanged.
        customs = customs if customs else list()
        valuefile = "" if not valuefile else valuefile
        assert isinstance(answers, list)
        assert isinstance(customs, list)
        snapshot = app.release.chartVersionSnapshot
        if int(chart_version_id) != constants.KEEP_TEMPLATE_UNCHANGED:
            # when user choose a new version to upgrade, make sure it has a snapshot,
            # so that it and rollback easily, and it's the base of KEEP_TEMPLATE_UNCHANGED feature
            chart_version = ChartVersion.objects.get(id=chart_version_id)
            snapshot = ChartVersionSnapshot.objects.make_snapshot(chart_version)

        release = self.create(
            app_id=app.id,
            repository=app.release.repository,
            chart=app.release.chart,
            chartVersionSnapshot=snapshot,
            answers=answers,
            customs=customs,
            valuefile=valuefile,
            valuefile_name=valuefile_name,
        )
        return release

    def make_rollback_release(self, app, release):
        return self.create(
            app_id=app.id,
            repository=release.repository,
            chart=release.chart,
            chartVersionSnapshot=release.chartVersionSnapshot,
            answers=release.answers,
            customs=release.customs,
            release_type=constants.ChartReleaseTypes.ROLLBACK.value,
            valuefile=release.valuefile,
            valuefile_name=release.valuefile_name,
            revision=release.revision,
        )


class ChartRelease(BaseTSModel):
    """
    TODO: what is the relationship of relase and reversion, is or belong to
    repo -> chart -> release
    var priority: answers > values > values.yaml
    """

    # from which repo
    repository = models.ForeignKey("Repository", on_delete=models.PROTECT)
    # from which chart, maybe Null if the source chart has been deleted
    chart = models.ForeignKey("Chart", on_delete=models.SET_NULL, db_constraint=False, null=True)
    # the snapshot of the chart, with the chart content details
    chartVersionSnapshot = models.ForeignKey("ChartVersionSnapshot", on_delete=models.CASCADE)
    # base on questions => get answers
    answers = JSONField(null=True, default=[])
    customs = JSONField(null=True, default=[])
    valuefile = models.TextField(help_text="yaml format")
    valuefile_name = models.CharField(
        max_length=64, default=constants.DEFAULT_VALUES_FILE_NAME, help_text="helm value file name"
    )

    short_name = models.CharField(max_length=constants.CHART_RELEASE_SHOT_NAME_LENGTH)
    app_id = models.IntegerField("App ID", db_index=True, default=constants.TEMPORARY_APP_ID)
    release_type = models.CharField(
        max_length=10,
        choices=constants.ChartReleaseTypes.get_choices(),
        default=constants.ChartReleaseTypes.RELEASE.value,
    )

    # content generated when a release create
    # used for accelerate status query
    content = models.TextField(null=True, default="")
    # list of {"name": "", "kind": ""}
    structure = JSONField(null=True, default=[])
    # 记录release的revision
    revision = models.IntegerField(default=0, help_text="用来标识helm release的升级版本")

    objects = ChartReleaseManager()

    def generate_valuesyaml(self, project_id, namespace_id, cluster_id):
        """valuefile + bcs namespace variables"""
        sys_variables = self.app.sys_variables
        bcs_variables = get_bcs_variables(project_id, cluster_id, namespace_id)
        return merge_valuefile_with_bcs_variables(self.valuefile, bcs_variables, sys_variables)

    def refresh_structure(self, namespace):
        structure = []
        resources = parser.parse(self.content, namespace).values()
        for resource in resources:
            structure.append(
                {
                    "name": resource.name.split("/")[-1],
                    "kind": resource.kind,
                }
            )
        self.structure = structure
        self.save(update_fields=["structure"])

    def extract_structure(self, namespace):
        if not self.structure:
            self.refresh_structure(namespace)

        return self.structure

    @property
    def resources(self):
        return self.chartVersionSnapshot.version

    @property
    def version(self):
        return self.chartVersionSnapshot.version

    @property
    def parameters(self):
        parameters = merge_rancher_answers(self.answers, self.customs)
        return parameters

    @property
    def snapshot_state(self):
        return self.chartVersionSnapshot.state

    @property
    def app(self):
        if self.app_id == constants.TEMPORARY_APP_ID:
            return None
        else:
            from backend.helm.app.models import App

            return App.objects.get(id=self.app_id)

    def render(self, namespace="default", bcs_inject_data=None):
        client = KubeHelmClient(helm_bin=settings.HELM_BIN)
        # 针对rollback的diff，不比对平台注入的信息
        if not bcs_inject_data:
            content, notes = client.template(
                files=self.chartVersionSnapshot.files,
                name=self.app.name,
                namespace=namespace,
                parameters=self.parameters,
                valuefile=self.generate_valuesyaml(self.app.project_id, self.app.namespace_id, self.app.cluster_id),
                cluster_id=self.app.cluster_id,
            )
        else:
            content, notes = client.template_with_ytt_renderer(
                files=self.chartVersionSnapshot.files,
                name=self.app.name,
                namespace=namespace,
                parameters=self.parameters,
                valuefile=self.generate_valuesyaml(self.app.project_id, self.app.namespace_id, self.app.cluster_id),
                cluster_id=self.app.cluster_id,
                bcs_inject_data=bcs_inject_data,
                cmd_flags=json.loads(self.app.cmd_flags),
            )

        return content, notes

    class Meta:
        db_table = 'helm_chart_release'
