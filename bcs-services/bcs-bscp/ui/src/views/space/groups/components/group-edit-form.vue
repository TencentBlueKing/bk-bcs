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
        @change="change">
        <bk-option
          v-for="service in serviceList"
          :key="service.id"
          :label="service.spec.name"
          :value="service.id"></bk-option>
      </bk-select>
    </bk-form-item>
    <bk-form-item class="radio-group-form" :label="t('标签选择器')" required property="rules">
      <template #label>
        <span class="label-text">{{ t('标签选择器') }}</span>
        <span
          ref="nodeRef"
          v-bk-tooltips="{
            content: t(
              '标签选择器由key、操作符、value组成，筛选符合条件的客户端拉取服务配置，一般用于灰度发布服务配置',
            ),
          }"
          class="bk-tooltips-base">
          <Info />
        </span>
      </template>
      <div v-for="(rule, index) in formData.rules" class="rule-config" :key="index">
        <div style="max-width: 174px; min-width: 174px">
          <bk-popover
            :is-show="isShowPopover[index]"
            ref="popoverRef"
            theme="light"
            trigger="manual"
            ext-cls="group-selector-popover"
            placement="bottom">
            <bk-input
              v-model="rule.key"
              ref="keyInputRef"
              :class="[{ 'is-error': showErrorKeyValidation[index] }, 'key-input']"
              :placeholder="t('请输入或选择key')"
              @click="isShowPopover[index] = true"
              @enter="handleKeyInputEnter(index)">
              <template #suffix>
                <angle-down :class="['suffix-icon', { 'show-popover': isShowPopover[index] }]" />
              </template>
            </bk-input>
            <template #content>
              <div class="selector-list" v-click-outside="() => (isShowPopover[index] = false)">
                <div
                  v-for="item in BuiltInTag"
                  :key="item"
                  class="selector-item"
                  @click="handleSelectBuiltinTag(index, item)">
                  {{ item }}
                </div>
              </div>
            </template>
          </bk-popover>
          <div v-show="showErrorKeyValidation[index]" class="error-msg is--key">
            {{ $t("仅支持字母，数字，'-'，'_'，'.' 及 '/' 且需以字母数字开头和结尾") }}
          </div>
        </div>
        <bk-select
          :model-value="rule.op"
          style="width: 82px"
          :clearable="false"
          @change="handleLogicChange(index, $event)">
          <bk-option v-for="op in GROUP_RULE_OPS" :key="op.id" :value="op.id" :label="op.name"></bk-option>
        </bk-select>
        <div class="value-input">
          <bk-tag-input
            v-if="['in', 'nin'].includes(rule.op)"
            v-model="rule.value"
            :class="{ 'is-error': showErrorValueValidation[index] }"
            :allow-create="true"
            :collapse-tags="true"
            :has-delete-icon="true"
            :show-clear-only-hover="true"
            :allow-auto-match="true"
            :list="[]"
            placeholder="value"
            @change="validateValue(index)"
            @blur="validateValue(index)">
          </bk-tag-input>
          <bk-input
            v-else
            v-model="rule.value"
            placeholder="value"
            :class="{ 'is-error': showErrorValueValidation[index] }"
            :type="['gt', 'ge', 'lt', 'le'].includes(rule.op) ? 'number' : 'text'"
            @change="() => (['gt', 'ge', 'lt', 'le'].includes(rule.op) ? validateValue(index) : ruleChange)"
            @blur="validateValue(index)">
          </bk-input>
          <div v-show="showErrorValueValidation[index]" class="error-msg is--value">
            {{ $t("需以字母、数字开头和结尾，可包含 '-'，'_'，'.' 和字母数字及负数") }}
          </div>
        </div>
        <div class="action-btns">
          <i
            v-if="index > 0 || formData.rules.length > 1"
            class="bk-bscp-icon icon-reduce"
            @click="handleDeleteRule(index)"></i>
          <i
            v-if="index === formData.rules.length - 1"
            class="bk-bscp-icon icon-add"
            v-bk-tooltips="{ content: $t('分组最多支持 5 个标签选择器'), disabled: formData.rules.length < 5 }"
            @click="handleAddRule(index)"></i>
        </div>
      </div>
      <!-- <div v-if="!rulesValid" class="bk-form-error">{{ t('分组规则表单不能为空') }}</div> -->
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
  import { Info, AngleDown } from 'bkui-vue/lib/icon';

  const getDefaultRuleConfig = (): IGroupRuleItem => ({ key: '', op: 'eq', value: '' });

  const route = useRoute();
  const { t } = useI18n();
  // const { userInfo } = storeToRefs(useUserStore());

  const props = defineProps<{
    group: IGroupEditing;
  }>();

  const emits = defineEmits(['change']);

  const keyValidateReg = new RegExp(
    '^[a-z0-9A-Z]([-_a-z0-9A-Z]*[a-z0-9A-Z])?((\\.|\\/)[a-z0-9A-Z]([-_a-z0-9A-Z]*[a-z0-9A-Z])?)*$',
  );
  const valueValidateReg = new RegExp(/^(?:-?\d+(\.\d+)?|[A-Za-z0-9]([-A-Za-z0-9_.]*[A-Za-z0-9])?)$/);

  const serviceLoading = ref(false);
  const serviceList = ref<IAppItem[]>([]);
  const formData = ref(cloneDeep(props.group));
  const formRef = ref();
  // const rulesValid = ref(true);
  const showErrorKeyValidation = ref<boolean[]>([]);
  const showErrorValueValidation = ref<boolean[]>([]);
  const popoverRef = ref();
  const isShowPopover = ref<boolean[]>([]);
  const keyInputRef = ref();

  // 内置标签
  const BuiltInTag = ['ip', 'podname'];

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

  const handleSelectBuiltinTag = (index: number, val: string) => {
    formData.value.rules[index].key = val;
    isShowPopover.value[index] = false;
  };

  const handleKeyInputEnter = (index: number) => {
    keyInputRef.value[index].blur();
    isShowPopover.value[index] = false;
  };

  // 增加规则
  const handleAddRule = (index: number) => {
    if (formData.value.rules.length === 5) {
      return;
    }
    const rule = getDefaultRuleConfig();
    formData.value.rules.splice(index + 1, 0, rule);
    showErrorKeyValidation.value.push(false);
    showErrorValueValidation.value.push(false);
  };

  // 删除规则
  const handleDeleteRule = (index: number) => {
    formData.value.rules.splice(index, 1);
    showErrorKeyValidation.value.splice(index, 1);
    showErrorValueValidation.value.splice(index, 1);
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
    let allValid = true;
    formData.value.rules.forEach((item, index) => {
      const { op } = item;
      if (op === '') return (allValid = false);
      // 批量检测时，展示先校验失败的错误信息
      keyValidateReg.test(item.key) ? validateValue(index) : validateKey(index);
    });
    allValid = !showErrorKeyValidation.value.includes(true) && !showErrorValueValidation.value.includes(true);
    return allValid;
  };

  // 验证key
  const validateKey = (index: number) => {
    showErrorKeyValidation.value[index] = !keyValidateReg.test(formData.value.rules[index].key);
    if (showErrorValueValidation.value[index]) {
      showErrorValueValidation.value[index] = false;
    }
  };
  // 验证value
  const validateValue = (index: number) => {
    if (Array.isArray(formData.value.rules[index].value)) {
      const valueArrValidation = (formData.value.rules[index].value as string[]).every((item: string) => {
        return valueValidateReg.test(item);
      });
      showErrorValueValidation.value[index] =
        !valueArrValidation || !((formData.value.rules[index].value as string[]).length > 0);
    } else {
      showErrorValueValidation.value[index] = !valueValidateReg.test(`${formData.value.rules[index].value}`);
    }
    if (showErrorKeyValidation.value[index]) {
      showErrorKeyValidation.value[index] = false;
    }
    change();
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
    position: relative;
    display: flex;
    align-items: flex-start;
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
    height: 32px;
    font-size: 14px;
    color: #979ba5;
    .bk-bscp-icon {
      cursor: pointer;
    }
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
    vertical-align: middle;
  }
  .is-error {
    border-color: #ea3636;
    &:focus-within {
      border-color: #3a84ff;
    }
    &:hover:not(.is-disabled) {
      border-color: #ea3636;
    }
    :deep(.bk-tag-input-trigger) {
      border-color: #ea3636;
    }
  }
  .error-msg {
    font-size: 12px;
    line-height: 14px;
    white-space: normal;
    word-wrap: break-word;
    color: #ea3636;
    animation: form-error-appear-animation 0.15s;
    margin-top: 8px;
    &.is--key {
      white-space: nowrap;
    }
  }
  @keyframes form-error-appear-animation {
    0% {
      opacity: 0;
      transform: translateY(-30%);
    }
    100% {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .key-input {
    .suffix-icon {
      width: 20px;
      font-size: 14px;
      &.show-popover {
        transform: rotate(180deg);
      }
    }
  }

  .selector-list {
    width: 174px;
    padding: 4px 0;
    .selector-item {
      height: 32px;
      line-height: 32px;
      padding: 0 12px;
      cursor: pointer;
      align-items: center;
      &:hover {
        background-color: #f5f7fa;
        color: #63656e;
      }
    }
  }
</style>

<style>
  .bk-popover.bk-pop2-content.group-selector-popover {
    padding: 0;
  }
</style>
