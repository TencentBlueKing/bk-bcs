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
import json
import logging

from django.conf import settings
from django.db import transaction
from django.utils import timezone
from django.utils.translation import ugettext_lazy as _
from rest_framework.exceptions import ValidationError

from backend.apps.whitelist import enabled_hpa_feature
from backend.components import paas_cc
from backend.components.bcs.k8s import K8SClient
from backend.container_service.projects.base.constants import LIMIT_FOR_ALL_DATA
from backend.uniapps.application.constants import FUNC_MAP
from backend.utils.errcodes import ErrorCode
from backend.utils.error_codes import error_codes

from ..configuration.constants import K8sResourceName
from ..configuration.models import CATE_ABBR_NAME, ShowVersion, Template, VersionedEntity
from .constants import InsState
from .drivers import get_scheduler_driver
from .generator import GENERATOR_DICT, get_bcs_context
from .models import InstanceConfig, VersionInstance
from .utils_pub import get_cluster_version

logger = logging.getLogger(__name__)


# TODO mark refactor
def check_template_available(tem, username):
    """检查模板集是否可操作（即未被加锁）"""
    is_locked = tem.is_locked
    if not is_locked:
        return True
    locker = tem.locker
    # 加锁者为当前用户，则可以操作;否则不可以操作
    if locker != username:
        raise ValidationError(
            '{locker}{prefix_msg},{link_msg}{locker}{suffix_msg}'.format(
                locker=locker, prefix_msg=_("正在操作"), link_msg=_("您如需操作请联系"), suffix_msg=_("解锁")
            )
        )
    return True


# TODO mark refactor
def validate_template_id(project_id, template_id, is_return_tempalte=False):
    if not project_id:
        raise ValidationError(_("请选择项目"))
    if not template_id:
        raise ValidationError(_("请选择模板集"))

    try:
        template = Template.objects.get(id=template_id)
        real_project_id = template.project_id
    except Exception:
        raise ValidationError(_("模板集(id:{})不存在").format(template_id))

    if project_id != real_project_id:
        raise ValidationError(_("模板集(id:{})不属于该项目").format(template_id))
    if is_return_tempalte:
        return template
    return True


# TODO mark refactor
def validate_version_id(
    project_id, version_id, is_version_entity_retrun=False, show_version_id=None, is_return_all=False
):
    if not project_id:
        raise ValidationError(_("请选择项目"))
    if not version_id:
        raise ValidationError(_("请选择模板版本"))

    try:
        version_entity = VersionedEntity.objects.get(id=version_id)
        template_id = version_entity.template_id
        template = Template.objects.get(id=template_id)
        real_project_id = template.project_id
    except Exception:
        raise ValidationError('{}(id:{}){}'.format(_("模板集版本"), version_id, _("不存在")))
    if project_id != real_project_id:
        raise ValidationError('{}(id:{}){}'.format(_("模板集版本"), version_id, _("不属于该项目")))
    # 验证用户可见版本号
    if show_version_id:
        is_show_version_id = ShowVersion.objects.filter(
            id=show_version_id, real_version_id=version_id, template_id=template_id
        ).exists()
        if not is_show_version_id:
            raise ValidationError('{}[{}]{}'.format(_("模板集"), template.name, _("内容已经被更新,请刷新页面后重试")))

    if is_return_all:
        return template, version_entity

    if is_version_entity_retrun:
        return version_entity
    return True


def validate_ns_by_tempalte_id(template_id, ns_list, access_token, project_id, instance_entity={}):
    """实例化，参数 ns_list 不能与 db 中已经实例化过的 ns 重复"""
    namespace = paas_cc.get_namespace_list(access_token, project_id, limit=LIMIT_FOR_ALL_DATA)
    namespace = namespace.get('data', {}).get('results') or []
    namespace_dict = {str(i['id']): i['name'] for i in namespace}

    # 查看模板下已经实例化过的 ns
    exist_instance_id = VersionInstance.objects.filter(template_id=template_id, is_deleted=False).values_list(
        'id', flat=True
    )
    filter_ns = (
        InstanceConfig.objects.filter(instance_id__in=exist_instance_id, is_deleted=False, is_bcs_success=True)
        .exclude(ins_state=InsState.NO_INS.value)
        .values_list('namespace', flat=True)
    )
    exist_ns = []
    # 查询每类资源已经实例化的ns，求合集，这些已经实例化过的ns不能再被实例化
    for cate in instance_entity:
        # HPA 编辑以模板集为准, 可以重复实例化
        if cate == K8sResourceName.K8sHPA.value:
            continue
        cate_data = instance_entity[cate]
        cate_name_list = [i.get('name') for i in cate_data if i.get('name')]
        cate_ns = filter_ns.filter(category=cate, name__in=cate_name_list).values_list('namespace', flat=True)
        exist_ns.extend(list(cate_ns))

    new_ns_list = [str(_i) for _i in ns_list]
    # 列表的交集
    intersection_list = list(set(exist_ns).intersection(set(new_ns_list)))

    # hpa白名单控制
    cluster_id_list = list(set([i['cluster_id'] for i in namespace if str(i['id']) in new_ns_list]))
    if K8sResourceName.K8sHPA.value in instance_entity:
        if not enabled_hpa_feature(cluster_id_list):
            raise error_codes.APIError(_("当前实例化包含HPA资源，{}").format(settings.GRAYSCALE_FEATURE_MSG))

    if not intersection_list:
        return True, [], namespace_dict

    ns_name_list = []
    for _n_id in intersection_list:
        _ns_name = namespace_dict.get(_n_id, _n_id)
        ns_name_list.append(_ns_name)
    return False, ns_name_list, namespace_dict


def validate_update_ns_by_tempalte_id(template_id, ns_list, access_token, project_id):
    """更新，参数 ns_list 必须全部为 db 中已经实例化过的 ns"""
    namespace = paas_cc.get_namespace_list(access_token, project_id, limit=LIMIT_FOR_ALL_DATA)
    namespace = namespace.get('data', {}).get('results') or []
    namespace_dict = {str(i['id']): i['name'] for i in namespace}

    # 查看模板下已经实例化过的 ns
    exist_ns = VersionInstance.objects.filter(template_id=template_id).values_list('ns_id', flat=True)
    new_ns_list = [int(_i) for _i in ns_list]

    wrong_list = [_n for _n in new_ns_list if _n not in exist_ns]
    if not wrong_list:
        return True, [], namespace_dict

    ns_name_list = []
    for _n_id in wrong_list:
        _ns_name = namespace_dict.get(_n_id, _n_id)
        ns_name_list.append(_ns_name)
    return False, ns_name_list, namespace_dict


def validate_instance_entity(req_instance_entity, tem_instance_entity):
    """验证前端传过了的预览资源是否在该版本的资源"""
    # 前端不传参数，则查询模板版本所有的资源
    if not req_instance_entity:
        instance_entity = tem_instance_entity
    else:
        instance_entity = {}
        for _cate in req_instance_entity:
            for _data in req_instance_entity[_cate]:
                if _data['id'] not in tem_instance_entity[_cate]:
                    raise ValidationError(_('{}[{}]不在当前选择的模板中').format(_cate, _data['name']))
            instance_entity[_cate] = [_i['id'] for _i in req_instance_entity[_cate]]
    return instance_entity


def get_ns_variable(access_token, project_id, namespace_id):
    """获取命名空间相关的变量信息"""
    context = {}
    # 获取命名空间的信息
    resp = paas_cc.get_namespace(access_token, project_id, namespace_id)
    if resp.get('code') != 0:
        raise ValidationError('{}(namespace_id:{}):{}'.format(_("查询命名空间的信息出错"), namespace_id, resp.get('message')))
    data = resp.get('data')
    cluster_id = data.get('cluster_id')
    context['SYS_CLUSTER_ID'] = cluster_id
    context['SYS_NAMESPACE'] = data.get('name')
    has_image_secret = data.get('has_image_secret')
    # 获取镜像地址
    context['SYS_JFROG_DOMAIN'] = paas_cc.get_jfrog_domain(access_token, project_id, context['SYS_CLUSTER_ID'])
    context['SYS_IMAGE_REGISTRY_LIST'] = paas_cc.get_image_registry_list(access_token, cluster_id)
    bcs_context = get_bcs_context(access_token, project_id)
    context.update(bcs_context)
    # k8s 集群获取集群版本信息
    cluster_version = get_cluster_version(access_token, project_id, cluster_id)
    return has_image_secret, cluster_version, context


def generate_namespace_config(namespace_id, instance_entity, is_save, is_validate=True, **params):
    """生成单个namespace下的所有配置文件"""
    # 将版本修改为可见版本
    show_version_id = params.get('show_version_id')
    show_version_name = ShowVersion.objects.get(id=show_version_id).name
    params['version'] = show_version_name

    # 查询命名空间相关的参数
    project_id = params.get('project_id')
    access_token = params.get('access_token')
    has_image_secret, cluster_version, context = get_ns_variable(access_token, project_id, namespace_id)
    params['has_image_secret'] = has_image_secret
    params['cluster_version'] = cluster_version
    params['context'] = context

    version_config = {}
    for item in instance_entity:
        item_id_list = instance_entity[item]
        item_config = []
        for item_id in item_id_list:
            # TODO 忽略metric, 后续从db中清理
            if item == 'metric':
                continue

            generator = GENERATOR_DICT.get(item)(item_id, namespace_id, is_validate, **params)
            file_content = generator.get_config_profile()
            file_name = generator.resource_show_name

            try:
                show_config = json.loads(file_content)
            except Exception:
                show_config = file_content

            _config_content = {
                'name': file_name,
                'config': show_config,
                'context': generator.context,
            }
            if is_save:
                save_kwargs = {
                    'instance_id': params.get('instance_id'),
                    'namespace': namespace_id,
                    'category': item,
                    'config': file_content,
                    'creator': params.get('username'),
                    'variables': json.dumps(params.get('variable_dict')),
                    # 更新时，需要将其他参数恢复为默认值
                    "updator": params.get('username'),
                    "updated": timezone.now(),
                    "created": timezone.now(),
                    "is_deleted": False,
                    "deleted_time": None,
                    "is_bcs_success": True,
                    "ins_state": InsState.NO_INS.value,
                }
                is_update_save_kwargs = False
                save_kwargs.update(
                    {
                        "oper_type": "create",
                        "status": "Running",
                        "last_config": "",
                    }
                )
                obj_module = InstanceConfig
                # 判断db中是否已经有记录，有则做更新操作
                _exist_ins_confg = obj_module.objects.filter(name=file_name, namespace=namespace_id, category=item)
                if _exist_ins_confg.exists():
                    # 更新第一条数据
                    _instance_config = _exist_ins_confg.first()
                    update_id = _instance_config.id
                    # 将其他数据设置为 is_deleted：True
                    _exist_ins_confg.exclude(id=update_id).update(is_deleted=True, deleted_time=timezone.now())
                    is_update_save_kwargs = True
                    # db中已经有记录，则实例化前不更新，实例化成功、失败后再更新
                    # _exist_ins_confg.filter(id=update_id).update(**save_kwargs)
                else:
                    _instance_config = obj_module.objects.create(**save_kwargs)

                _config_content['instance_config_id'] = _instance_config.id
                _config_content['save_kwargs'] = save_kwargs
                _config_content['is_update_save_kwargs'] = is_update_save_kwargs

            item_config.append(_config_content)
        version_config[item] = item_config
    return version_config


def preview_config_json(namespace_id, instance_entity, **params):
    """
    预览配置文件
    """
    config_data = generate_namespace_config(namespace_id, instance_entity, is_save=False, **params)
    config_data = {CATE_ABBR_NAME.get(x, x): config_data[x] for x in config_data}
    return config_data


@transaction.atomic
def save_all_config(slz_data, access_token="", username="", is_update=False):
    """
    TODO ：后台任务（需要调用第三方API）
    """
    ns_list = slz_data['ns_list']
    instance_entity = slz_data['instance_entity']
    history = [
        {
            'version_id': slz_data['version_id'],
            'instance_entity': slz_data['instance_entity'],
            'creator': username,
            'created': timezone.now().strftime('%Y-%m-%d %H:%M:%S'),
        }
    ]
    # 添加用户可见的版本号
    show_version_id = slz_data['show_version_id']
    show_version_name = ShowVersion.objects.get(id=show_version_id).name
    configuration = {}
    for ns in ns_list:
        if is_update:
            instance = VersionInstance.objects.filter(
                ns_id=ns,
                template_id=slz_data['template_id'],
            ).first()
            old_history = instance.history
            try:
                old_history = json.loads(old_history)
            except Exception:
                old_history = []
            old_history.append(history[0])
            # 更新列表
            instance.version_id = slz_data['version_id']
            instance.show_version_id = show_version_id
            instance.show_version_name = show_version_name
            instance.instance_entity = json.dumps(slz_data['instance_entity'])
            instance.is_start = slz_data['is_start']
            instance.history = json.dumps(old_history)
            instance.save()
        else:
            instance = VersionInstance.objects.create(
                version_id=slz_data['version_id'],
                show_version_id=show_version_id,
                show_version_name=show_version_name,
                instance_entity=json.dumps(slz_data['instance_entity']),
                is_start=slz_data['is_start'],
                ns_id=ns,
                template_id=slz_data['template_id'],
                history=json.dumps(history),
            )
        variable_dict = slz_data.get('variable_info', {}).get(ns) or {}
        params = {
            "instance_id": instance.id,
            "version": show_version_name,
            "show_version_id": show_version_id,
            "version_id": instance.version_id,
            "template_id": instance.get_template_id,
            "project_id": instance.get_project_id,
            "access_token": access_token,
            "username": username,
            "lb_info": slz_data.get('lb_info', {}),
            "variable_dict": variable_dict,
        }
        configuration[ns] = generate_namespace_config(ns, instance_entity, is_save=True, **params)
    return configuration


def handle_all_config(slz_data, access_token="", username="", is_update=False, project_kind=None):
    """"""
    project_id = slz_data['project_id']
    is_start = slz_data['is_start']

    configuration = save_all_config(slz_data, access_token, username, is_update)
    # 调用 bcs API
    if is_start:
        driver = get_scheduler_driver(access_token, project_id, configuration, project_kind)
        instantiation_result = driver.instantiation(is_update)
        return instantiation_result


def get_k8s_app_status(access_token, project_id, cluster_id, instance_name, namespace, category):
    field = [
        "resourceName",
        "namespace",
    ]
    client = K8SClient(access_token, project_id, cluster_id, None)
    curr_func = getattr(client, "%s_with_post" % (FUNC_MAP[category] % "get"))
    params = {"name": instance_name, "namespace": namespace, "field": ",".join(field)}
    resp = curr_func(params)
    if resp.get("code") != ErrorCode.NoError:
        return []
    return resp.get("data") or []


def get_app_status(access_token, project_id, project_kind, cluster_id, instance_name, namespace, category):
    ret_data_dict = {}
    data = get_k8s_app_status(access_token, project_id, cluster_id, instance_name, namespace, category)
    for info in data:
        key = (info.get("resourceName"), info.get("namespace"), category)
        ret_data_dict[key] = True
    return ret_data_dict


def has_instance_of_show_version(template_id, show_version_id):
    """模板版本是否被实例化过"""
    ins_id_list = VersionInstance.objects.filter(
        template_id=template_id, show_version_id=show_version_id, is_bcs_success=True, is_deleted=False
    ).values_list('id', flat=True)
    is_exists = (
        InstanceConfig.objects.filter(instance_id__in=ins_id_list, is_deleted=False, is_bcs_success=True)
        .exclude(ins_state=InsState.NO_INS.value)
        .exists()
    )
    return is_exists
