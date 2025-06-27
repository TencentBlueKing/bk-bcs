<template>
  <div>
    <span
      :class="['add-btn', { '!text-[#DCDEE5] !cursor-not-allowed': disabled }]"
      v-if="!labels.length"
      @click="handleAdd(0)">
      <i class="bk-icon icon-plus-circle-shape mr5"></i>
      {{$t('generic.button.add')}}
    </span>
    <div
      v-for="(item, index) in labels"
      :key="index"
      :class="['key-value', { '!mb-0': index === (labels.length - 1) }]">
      <span v-if="required" class="text-[#ea3636] mr-[12px]">*</span>
      <Validate class="w-[100%]" :value="item.value" :rules="valueRules" :required="required">
        <bcs-input
          :disabled="disabled"
          v-model="item.value"
          ref="inputRef">
        </bcs-input>
      </Validate>
      <div class="flex" v-if="!disabled">
        <i class="bk-icon icon-plus-circle ml15" @click="handleAdd(index)"></i>
        <i
          :class="['bk-icon icon-minus-circle ml10', { disabled: disabledDelete }]"
          @click="handleDelete(index)">
        </i>
      </div>
    </div>
  </div>
</template>
<script lang="ts">
import { computed, defineComponent, PropType, ref, toRefs, watch } from 'vue';

import Validate from '@/components/validate.vue';
import $i18n from '@/i18n/i18n-setup';

interface ILabel {
  value: string;
}
export default defineComponent({
  name: 'Values',
  components: { Validate },
  model: {
    prop: 'value',
    event: 'change',
  },
  props: {
    value: {
      type: Array as PropType<Array<string>>,
      default: () => [],
    },
    minItem: {
      type: Number,
      default: 1,
    },
    disableDeleteItem: {
      type: Boolean,
      default: true,
    },
    valueRules: {
      type: Array as PropType<Array<any>>,
      default: () => [
        {
          message: $i18n.t('generic.validate.labelKey'),
          validator: '^[A-Za-z0-9._/-]+$',
        },
      ],
    },
    required: {
      type: Boolean,
      default: false,
    },
    placeholder: {
      type: String,
      default: $i18n.t('generic.label.value'),
    },
    disabled: {
      type: Boolean,
      default: false,
    },
  },
  setup(props, ctx) {
    const labels = ref<ILabel[]>([]);
    const { value, disableDeleteItem } = toRefs(props);
    const unWatchValue = watch(value, () => {
      labels.value = props.value.map(v => ({
        value: v,
      }));
      if (labels.value.length < (props.minItem || 0)) {
        const reset: ILabel[] = new Array(props.minItem - labels.value.length).fill({ key: '', value: '' });
        labels.value.push(...reset);
      }
    }, { immediate: true, deep: true });
    watch(labels, () => {
      unWatchValue();
      emitChange();
    }, { deep: true });
    const disabledDelete = computed(() => labels.value.length <= props.minItem && disableDeleteItem.value);
    function emitChange() {
      const values = labels.value.reduce<string[]>((pre, item) => {
        if (!item.value) return pre;
        pre.push(item.value);
        return pre;
      }, []);
      ctx.emit('validate', validate());
      ctx.emit('change', values);
    };
    function handleAdd(index = 0) {
      if (props.disabled) return;
      labels.value.splice(index + 1, 0, { value: '' });
    };
    function handleDelete(index) {
      if (disabledDelete.value) return;
      labels.value.splice(index, 1);
    };
    // 聚焦
    const inputRef = ref<any[]>([]);
    function focus() {
      const [firstInput] = inputRef.value || [];
      firstInput?.focus();
    };
    // 数据校验
    function validate() {
      return labels.value.every(value => props.valueRules.every(rule => new RegExp(rule.validator).test(value.value)));
    };
    return {
      focus,
      labels,
      disabledDelete,
      handleAdd,
      handleDelete,
      validate,
    };
  },
});
</script>
<style lang="postcss" scoped>
.add-btn {
    font-size: 14px;
    color: #3a84ff;
    cursor: pointer;
    display: flex;
    align-items: center;
    height: 32px;
    max-width: 100px;
}
.key-value {
    display: flex;
    align-items: center;
    margin-bottom: 16px;
    .bk-icon {
        font-size: 20px;
        color: #979bA5;
        cursor: pointer;
        &.disabled {
            color: #DCDEE5;
            cursor: not-allowed;
        }
    }
    .ml8 {
        margin-left: 8px;
    }
    .mr8 {
        margin-right: 8px;
    }
}
</style>
