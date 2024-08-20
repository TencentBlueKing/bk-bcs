<template>
  <div>
    <div class="key-value" v-for="(item, index) in taints" :key="index">
      <span v-if="required" class="text-[#ea3636] mr-[12px]">*</span>
      <!-- 修改为新的校验规则 -->
      <Validate
        :rules="[
          {
            message: $i18n.t('generic.validate.SpeciaLabelKey'),
            validator: LABEL_KEY_MAXL,
          },
          {
            message: $i18n.t('generic.validate.SpeciaLabelKey'),
            validator: LABEL_KEY_DOMAIN,
          },
          {
            message: $i18n.t('generic.validate.SpeciaLabelKey'),
            validator: LABEL_KEY_PATH,
          },
          {
            message: $i18n.t('generic.validate.repeatKey'),
            validator: (value, meta) => taints.filter((_, i) => i !== meta).every(d => d.key !== value),
          },
        ]"
        :value="item.key"
        :required="required"
        :meta="index"
        class="flex-1">
        <bcs-input
          v-model="item.key"
          :placeholder="keyPlaceholder"
          @change="handleLabelKeyChange">
        </bcs-input>
      </Validate>
      <span class="ml8 mr8">=</span>
      <Validate
        :rules="[
          {
            message: $i18n.t('generic.validate.labelKey'),
            validator: TAINT_VALUE,
          },
        ]"
        :value="item.value"
        :meta="index"
        class="flex-1">
        <bcs-input
          v-model.trim="item.value"
          :placeholder="valuePlaceholder"
          @change="handleLabelValueChange">
        </bcs-input>
      </Validate>
      <bcs-select
        v-model="item.effect"
        class="effect ml15 flex-1"
        :placeholder="$t('generic.label.effect')"
        :clearable="false">
        <bcs-option
          v-for="effect in effectOptions"
          :key="effect"
          :id="effect"
          :name="effect">
        </bcs-option>
      </bcs-select>
      <i class="bk-icon icon-plus-circle ml15" @click="handleAddLabel(index)"></i>
      <i
        :class="['bk-icon icon-minus-circle ml10', { disabled: disabledDelete }]"
        @click="handleDeleteLabel(index)">
      </i>
    </div>
  </div>
</template>
<script lang="ts">
import { computed, defineComponent, ref, watch } from 'vue';

import { KEY_REGEXP, LABEL_KEY_DOMAIN, LABEL_KEY_MAXL, LABEL_KEY_PATH, TAINT_VALUE, VALUE_REGEXP } from '@/common/constant';
import Validate from '@/components/validate.vue';
import $i18n from '@/i18n/i18n-setup';

export interface ITaint {
  key: string;
  value: string;
  effect: 'PreferNoSchedule'| 'NoExecute'| 'NoSchedule';
}
export default defineComponent({
  name: 'NewTaints',
  components: { Validate },
  model: {
    prop: 'value',
    event: 'change',
  },
  props: {
    value: {
      type: Array,
      default: (): ITaint[] => ([]),
    },
    effectOptions: {
      type: Array,
      default: () => ['PreferNoSchedule', 'NoExecute', 'NoSchedule'],
    },
    minItem: {
      type: Number,
      default: 0,
    },
    required: {
      type: Boolean,
      default: false,
    },
    keyPlaceholder: {
      type: String,
      default:  $i18n.t('generic.label.key'),
    },
    valuePlaceholder: {
      type: String,
      default:  $i18n.t('generic.label.value'),
    }
  },
  setup(props, ctx) {
    const taints = ref<ITaint[]>([]);
    watch(() => props.value, (data: any[]) => {
      taints.value = data;
      if (taints.value.length < (props.minItem || 0)) {
        const reset: ITaint[] = new Array<ITaint>(props.minItem - taints.value.length)
          .fill({ key: '', value: '', effect: 'PreferNoSchedule' });
        taints.value.push(...reset);
      }
    }, { immediate: true, deep: true });
    watch(taints.value, () => {
      emitChange();
    }, { deep: true });
    const disabledDelete = computed(() => taints.value.length <= props.minItem);
    const emitChange = () => {
      ctx.emit('change', taints.value);
    };
    const handleLabelKeyChange = (newValue, oldValue) => {
      ctx.emit('key-change', newValue, oldValue);
    };
    const handleLabelValueChange = (newValue, oldValue) => {
      ctx.emit('value-change', newValue, oldValue);
    };
    const handleAddLabel = (index) => {
      taints.value.splice(index + 1, 0, { key: '', value: '', effect: 'PreferNoSchedule' });
    };
    const handleDeleteLabel = (index) => {
      if (disabledDelete.value) return;
      taints.value.splice(index, 1);
    };
    const validate = () => {
      const keys: string[] = [];
      const values: string[] = [];
      taints.value.forEach((item) => {
        keys.push(item.key);
        values.push(item.value);
      });
      const removeDuplicateData = new Set(keys);
      if (keys.length !== removeDuplicateData.size) {
        return false;
      }

      return keys.every(key => new RegExp(LABEL_KEY_DOMAIN).test(key))
        && keys.every(key => new RegExp(LABEL_KEY_MAXL).test(key))
        && keys.every(key => new RegExp(LABEL_KEY_PATH).test(key))
        && values.every(value => new RegExp(TAINT_VALUE).test(value)) ;
    };
    return {
      taints,
      disabledDelete,
      handleLabelKeyChange,
      handleLabelValueChange,
      handleAddLabel,
      handleDeleteLabel,
      validate,
      KEY_REGEXP,
      VALUE_REGEXP,
      LABEL_KEY_DOMAIN,
      LABEL_KEY_MAXL,
      LABEL_KEY_PATH,
      TAINT_VALUE,
    };
  },
});
</script>
<style lang="postcss" scoped>
.key-value {
  display: flex;
  align-items: center;
  margin-bottom: 16px;
  width: 100%;
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
