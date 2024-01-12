<template>
  <bk-form form-type="vertical" ref="formRef" :model="formData" :rules="rules">
    <bk-form-item :label="t('分组名称')" required property="name">
      <bk-input v-model="formData.name" :placeholder="t('请输入分组名称')" @blur="change"></bk-input>
    </bk-form-item>
    <bk-form-item class="radio-group-form" :label="t('服务可见范围')" required property="public">
      <bk-radio-group v-model="formData.public" @change="change">
        <bk-radio :label="true">{{ t('公开') }}</bk-radio>
        <bk-radio :label="false">{{ t('指定服务') }}</bk-radio>
      </bk-radio-group>
      <bk-select
        v-if="!formData.public"
        v-model="formData.bind_apps"
        class="service-selector"
        multiple
        filterable
        :placeholder="t('请选择服务')"
        :input-search="false"
        @change="change"
      >
        <bk-option
          v-for="service in serviceList"
          :key="service.id"
          :label="service.spec.name"
          :value="service.id"
        ></bk-option>
      </bk-select>
    </bk-form-item>
    <bk-form-item class="radio-group-form" :label="t('标签选择器')" required property="rules">
      <template #label>
        <span class="label-text">{{ t('标签选择器') }}</span>
        <span
          ref="nodeRef"
          v-bk-tooltips="{
            content: t('标签选择器由key、操作符、value组成，筛选符合条件的客户端拉取服务配置，一般用于灰度发布服务配置'),
          }"
          class="bk-tooltips-base">
          <Info />
        </span>
      </template>
      <div v-for="(rule, index) in formData.rules" class="rule-config" :key="index">
        <bk-input v-model="rule.key" style="width: 174px" placeholder="key" @change="ruleChange"></bk-input>
        <bk-select
          :model-value="rule.op"
          style="width: 72px"
          :clearable="false"
          @change="handleLogicChange(index, $event)"
        >
          <bk-option v-for="op in GROUP_RULE_OPS" :key="op.id" :value="op.id" :label="op.name"></bk-option>
        </bk-select>
        <div class="value-input">
          <bk-tag-input
            v-if="['in', 'nin'].includes(rule.op)"
            v-model="rule.value"
            :allow-create="true"
            :collapse-tags="true"
            :has-delete-icon="true"
            :show-clear-only-hover="true"
            :allow-auto-match="true"
            :list="[]"
            placeholder="value"
            @change="ruleChange"
          >
          </bk-tag-input>
          <bk-input
            v-else
            v-model="rule.value"
            placeholder="value"
            :type="['gt', 'ge', 'lt', 'le'].includes(rule.op) ? 'number' : 'text'"
            @change="ruleChange"
          >
          </bk-input>
        </div>
        <div class="action-btns">
          <i
            v-if="index > 0 || formData.rules.length > 1"
            class="bk-bscp-icon icon-reduce"
            @click="handleDeleteRule(index)"
          ></i>
          <i v-if="index === formData.rules.length - 1" class="bk-bscp-icon icon-add" @click="handleAddRule(index)"></i>
        </div>
      </div>
      <div v-if="!rulesValid" class="bk-form-error">{{ t('分组规则表单不能为空') }}</div>
    </bk-form-item>
  </bk-form>
</template>
<script setup lang="ts">
import { ref, watch, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute } from 'vue-router';
// import { storeToRefs } from 'pinia';
import { cloneDeep } from 'lodash';
// import useUserStore from '../../../../store/user';
import { IGroupEditing, EGroupRuleType, IGroupRuleItem } from '../../../../../types/group';
import GROUP_RULE_OPS from '../../../../constants/group';
import { getAppList } from '../../../../api/index';
import { IAppItem } from '../../../../../types/app';
import { Info } from 'bkui-vue/lib/icon';

const getDefaultRuleConfig = (): IGroupRuleItem => ({ key: '', op: 'eq', value: '' });

const route = useRoute();
const { t } = useI18n();
// const { userInfo } = storeToRefs(useUserStore());

const props = defineProps<{
  group: IGroupEditing;
}>();

const emits = defineEmits(['change']);

const serviceLoading = ref(false);
const serviceList = ref<IAppItem[]>([]);
const formData = ref(cloneDeep(props.group));
const formRef = ref();
const rulesValid = ref(true);

const rules = {
  name: [
    {
      validator: (value: string) => value.length <= 128,
      message: t('最大长度128个字符'),
    },
    {
      validator: (value: string) => {
        if (value.length > 0) {
          return /^[\u4e00-\u9fa5a-zA-Z0-9][\u4e00-\u9fa5a-zA-Z0-9_-]*[\u4e00-\u9fa5a-zA-Z0-9]?$/.test(value);
        }
        return true;
      },
      message: t('仅允许使用中文、英文、数字、下划线、中划线，且必须以中文、英文、数字开头和结尾'),
    },
  ],
  public: [
    {
      validator: (val: boolean) => {
        if (!val && formData.value.bind_apps.length === 0) {
          return false;
        }
        return true;
      },
      message: t('指定服务不能为空'),
    },
  ],
};

watch(
  () => props.group,
  (val) => {
    formData.value = cloneDeep(val);
  },
);

onMounted(() => {
  getServiceList();
});

const getServiceList = async () => {
  serviceLoading.value = true;
  try {
    const bizId = route.params.spaceId as string;
    const query = {
      all: true,
    };
    const resp = await getAppList(bizId, query);
    serviceList.value = resp.details;
  } catch (e) {
    console.error(e);
  } finally {
    serviceLoading.value = false;
  }
};

// 获取操作符对应操作值的数据类型
const getOpValType = (op: string) => {
  if (['in', 'nin'].includes(op)) {
    return 'array';
  }
  if (['gt', 'ge', 'lt', 'le'].includes(op)) {
    return 'number';
  }
  return 'string';
};

// 增加规则
const handleAddRule = (index: number) => {
  const rule = getDefaultRuleConfig();
  formData.value.rules.splice(index + 1, 0, rule);
};

// 删除规则
const handleDeleteRule = (index: number) => {
  formData.value.rules.splice(index, 1);
  change();
};

// 操作符修改后，string和number类型之间操作值可直接转换时自动转换，不能转换则设置为默认空值
const handleLogicChange = (index: number, val: EGroupRuleType) => {
  const rule = formData.value.rules[index];
  const newValType = getOpValType(val);
  const oldValType = getOpValType(rule.op);
  if (newValType !== oldValType) {
    if (newValType === 'array' && ['string', 'number'].includes(oldValType)) {
      rule.value = [];
    } else if (newValType === 'string' && oldValType === 'number') {
      rule.value = String(rule.value);
    } else if (newValType === 'number' && oldValType === 'string' && /\d+/.test(rule.value as string)) {
      rule.value = Number(rule.value);
    } else {
      rule.value = '';
    }
  }
  rule.op = val;
  ruleChange();
};

const ruleChange = () => {
  change();
};

const change = () => {
  emits('change', formData.value);
};

const validate = () => {
  const isRulesValid = validateRules();

  return formRef.value.validate().then(() => isRulesValid);
};

// 校验分组规则是否有表单项为空
const validateRules = () => {
  const inValid = formData.value.rules.some((item) => {
    const { key, op, value } = item;
    return key === '' || op === '' || (Array.isArray(value) ? (value as string[]).length === 0 : value === '');
  });

  rulesValid.value = !inValid;

  return !inValid;
};

defineExpose({
  validate,
});
</script>
<style lang="scss" scoped>
.bk-form {
  :deep(.bk-form-label) {
    font-size: 12px;
  }
  :deep(.radio-group-form .bk-form-content) {
    line-height: 1;
  }
}
.service-selector {
  margin-top: 10px;
}
.published-version {
  line-height: 16px;
  font-size: 12px;
  color: #313238;
}
.rule-config {
  display: flex;
  align-items: center;
  justify-content: space-between;
  position: relative;
  margin-top: 15px;
  .rule-logic {
    position: absolute;
    top: 3px;
    left: -48px;
    height: 26px;
    line-height: 26px;
    width: 40px;
    background: #e1ecff;
    color: #3a84ff;
    font-size: 12px;
    text-align: center;
    cursor: pointer;
  }
  .value-input {
    width: 280px;
  }
}
.action-btns {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 38px;
  font-size: 14px;
  color: #979ba5;
  cursor: pointer;
  i:hover {
    color: #3a84ff;
  }
}
.label-text {
  margin-right: 5px;
}
.bk-tooltips-base {
  font-size: 14px;
  color: #3a84ff;
  line-height: 19px;
}
</style>
