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

from rest_framework.exceptions import ValidationError

from .. import models
from ..constants import FileAction


def save_res_files(updator, template_id, resource_name, res_files):
    add_res_file_ids, remove_res_file_ids = [], []
    for f in res_files:
        if f["action"] == FileAction.UNCHANGE.value:
            continue
        if f["action"] == FileAction.DELETE.value:
            remove_res_file_ids.append(f["id"])
            continue

        if f["action"] == FileAction.UPDATE.value:
            remove_res_file_ids.append(f["id"])

        # UPDATE or CREATE action will create ResourceFile
        res_file_obj = models.ResourceFile.objects.create(
            name=f["name"], resource_name=resource_name, content=f["content"], creator=updator, template_id=template_id
        )
        add_res_file_ids.append(str(res_file_obj.id))

    return add_res_file_ids, remove_res_file_ids


def create_entity(creator, template_id, template_files):
    entity = {}
    for res_files in template_files:
        resource_name = res_files["resource_name"]
        add_res_file_ids, _ = save_res_files(creator, template_id, resource_name, res_files["files"])
        entity[resource_name] = ",".join(add_res_file_ids)
    return entity


def create_resources(template, show_version, template_files):
    creator = template.creator
    entity = create_entity(creator, template.id, template_files)
    ventity = models.VersionedEntity.objects.create(
        template_id=template.id, version=models.get_default_version(), entity=entity, creator=creator
    )
    models.ShowVersion.objects.create(
        template_id=template.id,
        name=show_version["name"],
        real_version_id=ventity.id,
        comment=show_version["comment"],
        history=json.dumps([ventity.id]),
        creator=creator,
        updator=creator,
    )


def update_entity(updator, template_id, entity, template_files):
    if not template_files:
        return entity

    for res_files in template_files:
        resource_name = res_files["resource_name"]
        add_res_file_ids, remove_res_file_ids = save_res_files(updator, template_id, resource_name, res_files["files"])
        if resource_name not in entity:
            entity[resource_name] = ",".join(add_res_file_ids)
            continue

        res_file_ids = entity[resource_name].split(",")
        res_file_ids = list((set(res_file_ids) | set(add_res_file_ids)) - set(remove_res_file_ids))
        if res_file_ids:
            entity[resource_name] = ",".join(res_file_ids)
        else:
            del entity[resource_name]

    return entity


def update_resources(template, show_version, template_files):
    updator = template.updator
    template_id = template.id
    show_version_name = show_version["name"]
    old_show_version_id = show_version["old_show_version_id"]

    try:
        old_show_version = models.ShowVersion.objects.get(id=old_show_version_id, template_id=template_id)
    except models.ShowVersion.DoesNotExist:
        raise ValidationError(
            f"show version(id:{old_show_version_id}) does not exist or not belong to template(id:{template_id})"
        )

    ventity = models.VersionedEntity.objects.get(id=old_show_version.real_version_id)
    entity = update_entity(updator, template_id, ventity.get_entity(), template_files)
    ventity = models.VersionedEntity.objects.create(
        template_id=template_id, version=models.get_default_version(), entity=entity, creator=updator
    )

    if old_show_version.name == show_version_name:
        old_show_version.update_real_version_id(ventity.id, updator=updator, comment=show_version["comment"])
        return

    try:
        old_show_version = models.ShowVersion.default_objects.get(template_id=template.id, name=show_version_name)
    except models.ShowVersion.DoesNotExist:
        models.ShowVersion.objects.create(
            template_id=template.id,
            name=show_version_name,
            real_version_id=ventity.id,
            comment=show_version["comment"],
            history=[ventity.id],
            creator=updator,
            updator=updator,
        )
    else:
        old_show_version.update_real_version_id(
            ventity.id, updator=updator, comment=show_version["comment"], is_deleted=False, deleted_time=None
        )
