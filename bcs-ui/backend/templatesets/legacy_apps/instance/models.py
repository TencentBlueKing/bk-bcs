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

模板实例化
"""
import json
import logging

from django.db import models
from django.utils.translation import ugettext_lazy as _

from ..configuration.models import BaseModel, Template, VersionedEntity
from .constants import EventType, InsState
from .manager import InstanceConfigManager, VersionInstanceManager

logger = logging.getLogger(__name__)


class VersionInstance(BaseModel):
    """版本实例化信息
    instance_entity 字段内容如下：表名：记录ID
    {
        'application': 'ApplicationID1,ApplicationID2,ApplicationID3',
        'service': 'ServiceID1,ServiceID2',
        ...
    }
    from backend.templatesets.legacy_apps.configuration.models import MODULE_DICT
    MODULE_DICT 记录 `表名` 和  `model` 的对应关系，并且所有的 `model` 都定义了 `get_name` 方法来查看名称
    """

    version_id = models.IntegerField(_("关联的VersionedEntity ID"))
    instance_entity = models.TextField(_("需要实例化的资源"), help_text=_("json格式数据"))
    is_start = models.BooleanField(default=False, help_text=_("false:生成配置文件；true:生成配置文件，调用bsc api实例化配置信息"))
    # add by 应用更新操作
    ns_id = models.IntegerField(_("命名空间ID"))
    template_id = models.IntegerField(_("关联的模板 ID"), help_text=_("该字段只在db中查看使用"))
    history = models.TextField(_("历史变更数据"), help_text=_("以json格式存储"))
    is_bcs_success = models.BooleanField(_("调用BCS API 是否成功"), default=True)
    # TODEL
    namespaces = models.TextField(_("命名空间ID"), help_text=_("该字段已经废弃"))
    # 添加用户可见版本
    show_version_id = models.IntegerField(_("用户可见版本ID"), default=0)
    show_version_name = models.CharField(_("用户可见版本Name"), max_length=255, default='')

    objects = VersionInstanceManager()

    @property
    def get_entity(self):
        return json.loads(self.instance_entity)

    @property
    def get_template_id(self):
        if self.template_id:
            return self.template_id
        return VersionedEntity.objects.get(id=self.version_id).template_id

    @property
    def get_project_id(self):
        template_id = self.get_template_id
        return Template.objects.get(id=template_id).project_id

    @property
    def get_version(self):
        return VersionedEntity.objects.get(id=self.version_id).version

    @property
    def get_ns_list(self):
        if not self.namespaces:
            return []
        return self.namespaces.split(',')


class InstanceConfig(BaseModel):
    """"""

    category_choice = (
        ('application', u"Application"),
        ('deployment', u"Deplpyment"),
        ('service', u"Service"),
        ('configmap', u"ConfigMap"),
        ('secret', u"Secret"),
        ('K8sSecret', u"K8sSecret"),
        ('K8sConfigMap', u"K8sConfigMap"),
        ('K8sService', u"K8sService"),
        ('K8sDeployment', u"K8sDeployment"),
        ('K8sDaemonSet', u"K8sDaemonSet"),
        ('K8sJob', u"K8sJob"),
        ('K8sStatefulSet', u"K8sStatefulSet"),
        ('K8sIngress', 'K8sIngress'),
    )
    instance_id = models.IntegerField(_("关联的 VersionInstance ID"), db_index=True)
    namespace = models.CharField(_("命名空间ID"), max_length=32)
    category = models.CharField(_("资源类型"), max_length=32, choices=category_choice)
    config = models.TextField(_("配置文件"), help_text=_('json格式数据'))
    is_bcs_success = models.BooleanField(_("调用BCS API 是否成功"), default=True)
    # 添加操作类型及状态，用于轮训任务记录
    oper_type = models.CharField(_("操作类型"), max_length=16, default="create")
    status = models.CharField(_("任务状态"), max_length=16, default="Running")

    # 实例化状态，解决appliation, deployment is_bcs_success为False，状态不一致问题
    # 0，未实例化
    # 1, 已实例化，但是实例化失败，需要再应用页面显示
    # 2, 已实例化，且实例化成功
    ins_state = models.IntegerField(_("实例化状态"), default=InsState.NO_INS.value, choices=InsState.get_choices())

    name = models.CharField(_("名称"), max_length=255, default='')
    # 添加一个字段用于记录滚动升级前的配置信息
    last_config = models.TextField(_("滚动升级前的配置"), default='', help_text=_('json格式'))
    # 保存变量信息
    variables = models.TextField(_("变量"), default='{}')

    objects = InstanceConfigManager()

    def save(self, *args, **kwargs):
        # 保存时,name字段单独保存
        config = json.loads(self.config)
        self.name = config.get('metadata', {}).get('name')

        super(InstanceConfig, self).save(*args, **kwargs)


class MetricConfig(BaseModel):
    # TODO 待模板集重构废弃
    category_choice = (('metric', u"Metric"),)
    instance_id = models.IntegerField(_("关联的 VersionInstance ID"))
    namespace = models.CharField(_("命名空间ID"), max_length=32)
    category = models.CharField(_("资源类型"), max_length=32, choices=category_choice)
    config = models.TextField(_("配置文件"), help_text=_('json格式数据'))
    is_bcs_success = models.BooleanField(_("调用BCS API 是否成功"), default=True)
    name = models.CharField(_("名称"), max_length=32, default='')

    # 关联ID, 没有取名metric_id
    ref_id = models.IntegerField(_("关联ID"), null=True, default=None)
    ins_state = models.IntegerField(_("实例化状态"), default=InsState.NO_INS.value, choices=InsState.get_choices())
    # 保存变量信息
    variables = models.TextField(_("变量"), default='{}')

    def save(self, *args, **kwargs):
        # 保存时,name字段单独保存
        config = json.loads(self.config)
        self.name = config.get('name')

        super(MetricConfig, self).save(*args, **kwargs)

    @classmethod
    def get_active_metric(cls, ref_id, ns_id_list=[]):
        """获取活跃metric,下发更新操作使用"""
        refs = cls.objects.filter(
            ref_id=ref_id,
            ins_state__in=[InsState.INS_SUCCESS.value, InsState.UPDATE_SUCCESS.value, InsState.METRIC_UPDATED.value],
        )
        if ns_id_list:
            refs = refs.filter(namespace__in=ns_id_list)
        return [i for i in refs]


class InstanceEvent(BaseModel):
    """实例化或者更新application时，能够记录错误信息"""

    category_choice = (
        ('application', u"Application"),
        ('deployment', u"Deplpyment"),
        ('service', u"Service"),
        ('configmap', u"ConfigMap"),
        ('secret', u"Secret"),
        ('metric', u"Metric"),
    )

    # 实例化的ID instance_versioninstance
    instance_id = models.IntegerField(_("关联的 VersionInstance ID"))
    instance_config_id = models.IntegerField(_("资源ID"))  # 对应InstanceConfig中的ID
    category = models.CharField(_("资源类型"), max_length=32, choices=category_choice)

    msg_type = models.IntegerField(_("消息类型"), choices=EventType.get_choices())
    msg = models.TextField(_("消息"))

    resp_snapshot = models.TextField(_("返回快照"))

    @classmethod
    def log(cls, instance_config_id, category, msg_type, result, context):
        """实例化记录错误信息，其他可以再封装下其他函数"""
        msg = result.get('message', '')
        ref = cls(
            instance_config_id=instance_config_id,
            category=category,
            msg_type=msg_type,
            instance_id=context['SYS_INSTANCE_ID'],
            creator=context['SYS_CREATOR'],
            updator=context['SYS_UPDATOR'],
            msg=msg,
            resp_snapshot=json.dumps(result),
        )
        ref.save()
        return ref
