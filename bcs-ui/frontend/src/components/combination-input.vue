<template>
  <div class="key-value">
    <template v-if="keyValueData.length">
      <div class="key-value-item" v-if="showHeader">
        <slot name="label">
          <span class="font-500">{{$t('cluster.create.aws.cidrWhitelist')}}</span>
        </slot>
      </div>
      <div
        v-for="(item, index) in keyValueData"
        :key="index"
        class="key-value-item"
      >
        <Validate
          class="key"
          :rules="rules"
          :value="item.key"
          :meta="index"
          :required="keyRequired"
          ref="validateRefs">
          <bcs-input
            :placeholder="$t(placeHolder)"
            :disabled="item.disabled"
            v-model="item.key">
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
  </div>
</template>
<script lang="ts">
import { computed, defineComponent, onMounted, PropType, ref, toRefs, watch } from 'vue';

import Validate from './validate.vue';

import $i18n from '@/i18n/i18n-setup';

export interface IData {
  key: string;
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
  props: {
    list: {
      type: Array as PropType<Array<IData>>,
      default: () => [],
    },
    showHeader: {
      type: Boolean,
      default: true,
    },
    keyRules: {
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
    placeHolder: {
      type: String,
      default: 'generic.placeholder.input',
    },
  },
  setup(props, ctx) {
    const { keyRules, minItems, uniqueKey } = toRefs(props);
    const keyValueData = ref<IData[]>([]);
    const disabledDelete = computed(() => keyValueData.value.length <= minItems.value);
    // watch(modelValue, () => {
    //   if (Array.isArray(modelValue.value)) {
    //     keyValueData.value = modelValue.value.map((item: any) => ({
    //       ...item,
    //       disabled: true,
    //     }));
    //   } else {
    //     keyValueData.value = Object.keys(modelValue.value).map(key => ({
    //       key,
    //       value: modelValue.value[key],
    //       disabled: true,
    //     }));
    //   }
    //   // 添加一组空值
    //   if (!keyValueData.value.length && minItems.value) {
    //     keyValueData.value.push({
    //       key: '',
    //     });
    //   }
    // }, { immediate: true });
    if (!keyValueData.value.length && minItems.value) {
      keyValueData.value.push({
        key: '',
      });
    }
    watch(keyValueData, () => {
      ctx.emit('data-change', keyValueData.value);
    }, { deep: true });

    const handleAddKeyValue = (index) => {
      keyValueData.value.splice(index + 1, 0, {
        key: '',
      });
    };
    const handleDeleteKeyValue = (index) => {
      if (disabledDelete.value) return;
      keyValueData.value.splice(index, 1);
    };
    const labels = computed(() => keyValueData.value.filter(item => !!item.key).reduce((pre, curLabelItem) => {
      pre[curLabelItem.key] = curLabelItem.key;
      return pre;
    }, {}));
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
      keyValueData.value.forEach((item) => {
        keys.push(item.key);
      });
      if (uniqueKey.value) {
        const removeDuplicateData = new Set(keys);
        if (keys.length !== removeDuplicateData.size) {
          return false;
        }
      }
      return keys.every(key => keyRules.value.every(rule => new RegExp(rule.validator).test(key)));
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
