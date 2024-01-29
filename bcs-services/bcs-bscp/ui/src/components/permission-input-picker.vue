<template>
  <div class="permission-input-picker">
    <bk-popover
      ext-cls="privilege-tips-wrap"
      theme="light"
      trigger="manual"
      placement="top"
      :is-show="showPrivilegeErrorTips"
    >
      <bk-input
        v-model="privilegeInputVal"
        type="number"
        :placeholder="t('请输入三位权限数字')"
        :disabled="props.disabled"
        @blur="handleInputBlur"
      />
      <template #content>
        <div>{{ t('只能输入三位 0~7 数字') }}</div>
        <div class="privilege-tips-btn-area">
          <bk-button text theme="primary" @click="showPrivilegeErrorTips = false">{{ t('我知道了') }}</bk-button>
        </div>
      </template>
    </bk-popover>
    <bk-popover ext-cls="privilege-select-popover" theme="light" trigger="click" placement="bottom">
      <div class="perm-panel-trigger">
        <i class="bk-bscp-icon icon-configuration-line"></i>
      </div>
      <template #content>
        <div class="privilege-select-panel">
          <div v-for="(item, index) in PRIVILEGE_GROUPS" class="group-item" :key="index" :label="item">
            <div class="header">{{ item }}</div>
            <div class="checkbox-area">
              <bk-checkbox-group
                class="group-checkboxs"
                :model-value="privilegeGroupsValue[index]"
                @change="handleSelect(index, $event)"
              >
                <bk-checkbox size="small" :label="4">{{ t('读') }}</bk-checkbox>
                <bk-checkbox size="small" :label="2">{{ t('写') }}</bk-checkbox>
                <bk-checkbox size="small" :label="1">{{ t('执行') }}</bk-checkbox>
              </bk-checkbox-group>
            </div>
          </div>
        </div>
      </template>
    </bk-popover>
  </div>
</template>
<script lang="ts" setup>
import { ref, computed, watch } from 'vue';
import { useI18n } from 'vue-i18n';
const { t } = useI18n();

const PRIVILEGE_GROUPS = [t('属主（own）'), t('属组（group）'), t('其他人（other）')];
const PRIVILEGE_VALUE_MAP = {
  0: [],
  1: [1],
  2: [2],
  3: [1, 2],
  4: [4],
  5: [1, 4],
  6: [2, 4],
  7: [1, 2, 4],
};

const props = defineProps<{
  disabled?: boolean;
  modelValue: string;
}>();

const emits = defineEmits(['update:modelValue', 'change']);

const localVal = ref('');
const privilegeInputVal = ref('');
const showPrivilegeErrorTips = ref(false);

watch(
  () => props.modelValue,
  (val) => {
    privilegeInputVal.value = val;
    localVal.value = val;
  },
  { immediate: true },
);

// 将权限数字拆分成三个分组配置
const privilegeGroupsValue = computed(() => {
  const data: { [index: string]: number[] } = { 0: [], 1: [], 2: [] };
  if (typeof localVal.value === 'string' && localVal.value.length > 0) {
    const valArr = localVal.value.split('').map(i => parseInt(i, 10));
    valArr.forEach((item, index) => {
      data[index as keyof typeof data] = PRIVILEGE_VALUE_MAP[item as keyof typeof PRIVILEGE_VALUE_MAP];
    });
  }
  return data;
});

// 权限输入框失焦后，校验输入是否合法，如不合法回退到上次输入
const handleInputBlur = () => {
  const val = String(privilegeInputVal.value);
  if (/^[0-7]{3}$/.test(val)) {
    localVal.value = val;
    showPrivilegeErrorTips.value = false;
    change();
  } else {
    privilegeInputVal.value = localVal.value as string;
    showPrivilegeErrorTips.value = true;
  }
};

const handleSelect = (index: number, val: number[]) => {
  const groupsValue = { ...privilegeGroupsValue.value };
  groupsValue[index] = val;
  const digits = [];
  for (let i = 0; i < 3; i++) {
    let sum = 0;
    if (groupsValue[i].length > 0) {
      sum = groupsValue[i].reduce((acc, crt) => acc + crt, 0);
    }
    digits.push(sum);
  }
  const newVal = digits.join('');
  privilegeInputVal.value = newVal;
  localVal.value = newVal;
  showPrivilegeErrorTips.value = false;
  change();
};

const change = () => {
  emits('update:modelValue', localVal.value);
  emits('change', localVal.value);
};
</script>
<style lang="scss" scoped>
.permission-input-picker {
  display: flex;
  align-items: center;
  width: 100%;
  :deep(.bk-input) {
    width: calc(100% - 32px);
    border-right: none;
    border-top-right-radius: 0;
    border-bottom-right-radius: 0;
    .bk-input--number-control {
      display: none;
    }
  }
  .perm-panel-trigger {
    width: 32px;
    height: 32px;
    text-align: center;
    background: #e1ecff;
    color: #3a84ff;
    border: 1px solid #3a84ff;
    cursor: pointer;
  }
}
.privilege-tips-btn-area {
  margin-top: 8px;
  text-align: right;
}
.privilege-select-panel {
  display: flex;
  align-items: top;
  border: 1px solid #dcdee5;
  .group-item {
    .header {
      padding: 0 16px;
      height: 42px;
      line-height: 42px;
      color: #313238;
      font-size: 12px;
      background: #fafbfd;
      border-bottom: 1px solid #dcdee5;
    }
    &:not(:last-of-type) {
      .header,
      .checkbox-area {
        border-right: 1px solid #dcdee5;
      }
    }
  }
  .checkbox-area {
    padding: 10px 16px 12px;
    background: #ffffff;
    &:not(:last-child) {
      border-right: 1px solid #dcdee5;
    }
  }
  .group-checkboxs {
    font-size: 12px;
    .bk-checkbox ~ .bk-checkbox {
      margin-left: 16px;
    }
    :deep(.bk-checkbox-label) {
      font-size: 12px;
    }
  }
}
</style>
