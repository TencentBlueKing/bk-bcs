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

import dataclasses
from typing import Any, Dict, Optional, List, Tuple, Type
from enum import Enum as OrigEnum
from enum import EnumMeta, auto
from collections import OrderedDict


@dataclasses.dataclass(init=False)
class FeatureFlagField:
    label: str
    default: bool
    name: str

    def __init__(
        self,
        name: Optional[str] = None,
        label: Optional[str] = None,
        default: bool = False,
    ):
        """FeatureFlag 中的字段, 记录了 label、default 等属性

        :param label: 对这个 feature flag 的描述语句
        :param default: feature flag 的默认状态, 对于新引入的 feature flag, 该值建议为 False.
        :param name: 当前 Feature Flag 的名字, 当使用 register_ext_feature_flag 进行注册时, 必须提供该字段.
        """
        self.name = name or ""
        self.label = label or name or ""
        self.default = default

    def __set_name__(self, owner, name):
        """利用描述符协议, 往 FeatureFlagField 注入 FeatureFlag 的名字."""
        self.name = name
        if not self.label:
            self.label = self.name

    def __get__(self, instance, owner):
        """返回 FeatureFlag 的名称. 因为 FeatureFlag 的值就是他自身的名称."""
        return self.name

    def __str__(self):
        return self.name


class FeatureFlagMeta(type):
    _feature_flag_fields_: Dict[str, FeatureFlagField]

    def __new__(mcs, cls_name: str, bases, dct: Dict):
        _feature_flag_fields_ = {}
        for base in bases:
            _feature_flag_fields_.update(getattr(base, "_feature_flag_fields_", dict()))

        for attr, field in dct.items():
            if not isinstance(field, FeatureFlagField):
                continue

            if _feature_flag_fields_.get(attr, field).label != field.label:
                raise ValueError("Two Feature Flags cannot be set to the same value.")
            if not attr:
                raise ValueError("Feature flag's name can not be empty.")

            _feature_flag_fields_[attr] = field

        dct["_feature_flag_fields_"] = _feature_flag_fields_
        return super().__new__(mcs, cls_name, bases, dct)

    def _get_feature_fields_(cls) -> Dict[str, FeatureFlagField]:
        return cls._feature_flag_fields_

    def __iter__(cls):
        for feature in cls._get_feature_fields_():
            yield feature


class FeatureFlag(str, metaclass=FeatureFlagMeta):
    def __new__(cls, value):
        """Cast a string into a predefined feature flag."""
        for field in cls._get_feature_fields_().values():
            if field.name == value:
                return value
        return cls._missing_(value)

    @classmethod
    def _missing_(cls, value) -> str:
        raise ValueError("%r is not a valid %s" % (value, cls.__name__))

    @classmethod
    def get_default_flags(cls) -> Dict[str, bool]:
        """Get the default user feature flags, client is safe to modify the result because it's a copy"""
        features = {field.name: field.default for field in cls._get_feature_fields_().values()}
        return features.copy()

    @classmethod
    def get_django_choices(cls) -> List[Tuple[str, str]]:
        """Get Django-Like Choices for this FeatureFlag Collection."""
        return [(field.name, field.label) for field in cls._get_feature_fields_().values()]

    @classmethod
    def get_feature_label(cls, feature: str) -> str:
        """Get the label of provided feature flag"""
        return cls._get_feature_fields_()[cls(feature)].label

    @classmethod
    def register_feature_flag(cls, field: FeatureFlagField):
        """注册额外的FeatureFlagField"""
        name = field.name
        if cls._feature_flag_fields_.get(name, field).label != field.label:
            raise ValueError("Two Feature Flags cannot be set to the same value.")
        if not name:
            raise ValueError("Feature flag's name can not be empty.")

        cls._feature_flag_fields_[name] = field
        setattr(cls, name, field)


class EnumField:
    """Use it with `StructuredEnum` type

    :param real_value: the real value of enum member
    :param label: the label text of current enum value
    :param is_reserved: if current member was reserved, it will not be included in choices
    """

    def __init__(self, real_value: Any, label: Optional[str] = None, is_reserved: bool = False):
        self.real_value = real_value
        self.label = label
        self.is_reserved = is_reserved

    def set_label_if_empty(self, key: str):
        """Set field's label if not provided"""
        if not self.label:
            self.label = key.lower().replace("_", " ").capitalize()


class StructuredEnumMeta(EnumMeta):
    """The metaclass of StructuredEnum"""

    __field_members__: Dict[Type, Dict]

    def __new__(metacls, cls, bases, classdict):
        field_members = metacls.process_enum_fields(classdict)
        classdict["__field_members__"] = field_members
        return super().__new__(metacls, cls, bases, classdict)

    @staticmethod
    def process_enum_fields(classdict) -> Dict:
        """Iterate all enum members, transform them into `EnumField` objects and return as a dict.

        `EnumField` members in `classdict` will be replaced with their `real_value` attribute so `EnumMeta` can
        continue the initialization.
        """
        fields = OrderedDict()
        # Find out all `EnumField` instance, store them into class so we can use them later
        for key, member in classdict.items():
            # Ignore all private members
            if key.startswith("_"):
                continue

            # Turn regular enum member into EnumField instance
            if not isinstance(member, EnumField) and isinstance(member, (int, str, auto)):
                member = EnumField(member)
            if not isinstance(member, EnumField) or member.is_reserved:
                continue

            member.set_label_if_empty(key)
            fields[key] = member
            # Use dict's setitem method because setting value with `classdict[key]` is forbidden
            dict.__setitem__(classdict, key, member.real_value)
        return fields

    def get_field_members(cls) -> Dict:
        return cls.__field_members__


class StructuredEnum(OrigEnum, metaclass=StructuredEnumMeta):
    """Structured Enum type, providing extra features such as getting enum members as choices tuple"""

    @classmethod
    def get_django_choices(cls) -> List[Tuple[Any, str]]:
        """Get Django-Like Choices for all field members."""
        return cls.get_choices()

    @classmethod
    def get_choice_label(cls, value: Any) -> str:
        """Get the label of field member by value"""
        if isinstance(value, cls):
            value = value.value
        return dict(cls.get_choices()).get(value, value)

    @classmethod
    def get_labels(cls) -> List[str]:
        """Get the label list for all field members."""
        return [item[1] for item in cls.get_choices()]

    @classmethod
    def get_values(cls) -> List[Any]:
        """Get the value list for all field members."""
        return [item[0] for item in cls.get_choices()]

    @classmethod
    def get_choices(cls) -> List[Tuple[Any, str]]:
        """Get Choices for all field members."""
        members = cls.get_field_members()
        return [(field.real_value, field.label) for field in members.values()]
