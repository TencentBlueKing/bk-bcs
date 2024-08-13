<template>
  <div class="key-value">
    <template v-if="keyValueData.length">
      <div class="key-value-item" v-if="showHeader">
        <span class="key desc">
          {{$t('generic.label.key')}}:
          <i
            v-bk-tooltips="keyDesc"
            class="ml10 bcs-icon bcs-icon-question-circle"
            v-if="keyDesc"
          ></i>
        </span>
        <span class="value desc">
          {{$t('generic.label.value')}}:
          <i
            v-bk-tooltips="valueDesc"
            class="ml10 bcs-icon bcs-icon-question-circle"
            v-if="valueDesc"
          ></i>
        </span>
      </div>
      <div
        v-for="(item, index) in keyValueData"
        :key="index"
        class="key-value-item"
      >
        <bcs-dropdown-menu
          class="key" trigger="click"
          v-if="keyAdvice.length > 0 && !item.disabled">
          <template #dropdown-trigger>
            <Validate
              :rules="rules"
              :value="item.key"
              :meta="index"
              :required="keyRequired"
              ref="validateRefs">
              <bcs-input
                :placeholder="$t('generic.label.key')"
                :disabled="item.disabled"
                v-model="item.key">
              </bcs-input>
            </Validate>
          </template>
          <template #dropdown-content>
            <ul class="bk-dropdown-list">
              <li
                v-for="(advice, i) in keyAdvice" :key="i"
                @click="handleAdvice(advice, item)">
                <a
                  href="javascript:;"
                  v-bk-tooltips="{
                    content: advice.desc,
                    disabled: !advice.desc,
                    placement: 'right',
                    boundary: 'window'
                  }"
                >{{advice.name}}</a>
              </li>
            </ul>
          </template>
        </bcs-dropdown-menu>
        <Validate
          class="key"
          :rules="rules"
          :value="item.key"
          :meta="index"
          :required="keyRequired"
          ref="validateRefs"
          v-else>
          <bcs-input
            :placeholder="$t('generic.label.key')"
            :disabled="item.disabled"
            v-model="item.key">
          </bcs-input>
        </Validate>
        <span class="equals-sign">=</span>
        <Validate
          class="value"
          :rules="valueRules"
          :value="item.value"
          :meta="index"
          ref="validateRefs">
          <bcs-input
            :placeholder="item.placeholder || $t('generic.label.value')"
            class="value"
            v-model="item.value">
          </bcs-input>
        </Validate>
        <template v-if="showOperate">
          <i class="bk-icon icon-plus-circle ml10 mr5" @click="handleAddKeyValue(index)"></i>
          <i
            :class="['bk-icon icon-minus-circle', { disabled: disabledDelete }]"
            @click="handleDeleteKeyValue(index)"
          ></i>
        </template>
      </div>
    </template>
    <span
      class="add-btn mb15"
      v-else
      @click="handleAddKeyValue(-1)">
      <i class="bk-icon icon-plus-circle-shape mr5"></i>
      {{$t('generic.button.add')}}
    </span>
    <div class="mt15" v-if="showFooter">
      <bcs-button
        class="bcs-btn"
        theme="primary"
        :loading="loading"
        :disalbed="loading"
        @click="confirmSetLabel"
      >
        {{$t('generic.button.save')}}
      </bcs-button>
      <bcs-button class="bcs-btn" :disalbed="loading" @click="hideSetLabel">
        {{$t('generic.button.cancel')}}
      </bcs-button>
    </div>
  </div>
</template>
<script lang="ts">
import { computed, defineComponent, onMounted, PropType, ref, toRefs, watch } from 'vue';

import Validate from './validate.vue';

import $i18n from '@/i18n/i18n-setup';

export interface IData {
  key: string;
  value: string;
  placeholder?: any;
  disabled?: boolean;
}
export interface IAdvice {
  desc: string
  name: string
}
export interface IRule {
  message: any
  validator: any
}
export default defineComponent({
  components: { Validate },
  model: {
    prop: 'modelValue',
    event: 'change',
  },
  props: {
    modelValue: {
      type: [Object, Array] as PropType<Object|Array<IData>>,
      default: () => [],
    },
    valueDesc: {
      type: String,
      default: '',
    },
    keyDesc: {
      type: String,
      default: '',
    },
    loading: {
      type: Boolean,
      default: false,
    },
    showFooter: {
      type: Boolean,
      default: true,
    },
    showHeader: {
      type: Boolean,
      default: true,
    },
    keyAdvice: {
      type: Array as PropType<Array<IAdvice>>,
      default: () => [],
    },
    keyRules: {
      type: Array as PropType<Array<IRule>>,
      default: () => [],
    },
    valueRules: {
      type: Array as PropType<Array<IRule>>,
      default: () => [],
    },
    minItems: {
      type: Number,
      default: 1,
    },
    uniqueKey: {
      type: Boolean,
      default: true,
    },
    showOperate: {
      type: Boolean,
      default: true,
    },
    keyRequired: {
      type: Boolean,
      default: false,
    },
  },
  setup(props, ctx) {
    const { modelValue, keyRules, valueRules, minItems, uniqueKey } = toRefs(props);
    const keyValueData = ref<IData[]>([]);
    const disabledDelete = computed(() => keyValueData.value.length <= minItems.value);
    watch(modelValue, () => {
      if (Array.isArray(modelValue.value)) {
        keyValueData.value = modelValue.value.map((item: any) => ({
          ...item,
          disabled: true,
        }));
      } else {
        keyValueData.value = Object.keys(modelValue.value).map(key => ({
          key,
          value: modelValue.value[key],
          disabled: true,
        }));
      }
      // 添加一组空值
      if (!keyValueData.value.length && minItems.value) {
        keyValueData.value.push({
          key: '',
          value: '',
        });
      }
    }, { immediate: true });
    watch(keyValueData, () => {
      ctx.emit('data-change', keyValueData.value);
    }, { deep: true });

    const handleAddKeyValue = (index) => {
      keyValueData.value.splice(index + 1, 0, {
        key: '',
        value: '',
      });
    };
    const handleDeleteKeyValue = (index) => {
      if (disabledDelete.value) return;
      keyValueData.value.splice(index, 1);
    };
    const labels = computed(() => keyValueData.value.filter(item => !!item.key).reduce((pre, curLabelItem) => {
      pre[curLabelItem.key] = curLabelItem.value;
      return pre;
    }, {}));
    const confirmSetLabel = async () => {
      if (!validate()) return;
      ctx.emit('confirm', labels.value, keyValueData.value);
    };
    const hideSetLabel = () => {
      ctx.emit('cancel', labels.value, keyValueData.value);
    };
    // key联想功能
    const handleAdvice = (advice, item) => {
      item.key = advice.name;
      item.value = advice.default;
    };
    const rules = ref<IRule[]>([
      ...keyRules.value,
    ]);
    // 数据校验
    const validate = () => {
      const keys: string[] = [];
      const values: string[] = [];
      keyValueData.value.forEach((item) => {
        keys.push(item.key);
        values.push(item.value);
      });
      if (uniqueKey.value) {
        const removeDuplicateData = new Set(keys);
        if (keys.length !== removeDuplicateData.size) {
          return false;
        }
      }
      return keys.every(key => keyRules.value.every(rule => new RegExp(rule.validator).test(key)))
      && values.every(value => valueRules.value.every(rule => new RegExp(rule.validator).test(value)));
    };
    // 组件校验（调用validate组件的方法）
    const validateRefs = ref<any[]>([]);
    const validateAll = async () => {
      const data = validateRefs.value.map($ref => $ref?.validate('blur'));
      const results = await Promise.all(data);
      return results.every(result => result);
    };

    onMounted(() => {
      if (uniqueKey.value) {
        rules.value.push({
          message: $i18n.t('generic.validate.repeatKey'),
          validator: (value, index) => keyValueData.value.filter((_, i) => i !== index).every(d => d.key !== value),
        });
      }
    });

    return {
      disabledDelete,
      rules,
      labels,
      keyValueData,
      validate,
      validateRefs,
      validateAll,
      confirmSetLabel,
      hideSetLabel,
      handleAddKeyValue,
      handleDeleteKeyValue,
      handleAdvice,
    };
  },
});
</script>
<style lang="postcss" scoped>
.add-btn {
  cursor: pointer;
  background: #fff;
  border: 1px dashed #c4c6cc;
  border-radius: 2px;
  display: flex;
  align-items: center;
  justify-content: center;
  height: 32px;
  font-size: 14px;
  &:hover {
      border-color: #3a84ff;
      color: #3a84ff;
  }
}
.key-value-item {
  display: flex;
  align-items: center;
  height: 32px;
  line-height: 32px;
  margin-bottom: 10px;
  font-size: 14px;
  .key {
      flex: 1;
  }
  .value {
      flex: 1;
  }
  .desc {
      display: flex;
      align-items: center;
  }
  .bk-icon {
      font-size: 24px;
      color: #979bA5;
      cursor: pointer;
  }
  .bk-icon.disabled {
      color: #DCDEE5;
      cursor: not-allowed;
  }
  .equals-sign {
      color: #c3cdd7;
      margin: 0 15px;
  }
}
.bcs-btn {
  width: 86px;
}
</style>
