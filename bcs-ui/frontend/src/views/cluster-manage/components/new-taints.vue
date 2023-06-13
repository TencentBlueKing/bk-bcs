<template>
  <div>
    <div class="key-value" v-for="(item, index) in taints" :key="index">
      <Validate
        :rules="[
          {
            message: $i18n.t('仅支持字母，数字和字符(-_./)'),
            validator: '^[A-Za-z0-9._/-]+$',
          },
          {
            message: $i18n.t('重复键'),
            validator: (value, meta) => taints.filter((_, i) => i !== meta).every(d => d.key !== value),
          },
        ]"
        :value="item.key"
        :meta="index"
        class="flex-1">
        <bcs-input
          v-model="item.key"
          :placeholder="$t('键')"
          @change="handleLabelKeyChange">
        </bcs-input>
      </Validate>
      <span class="ml8 mr8">=</span>
      <Validate
        :rules="[
          {
            message: $i18n.t('仅支持字母，数字和字符(-_./)'),
            validator: '^[A-Za-z0-9._/-]+$',
          },
        ]"
        :value="item.value"
        :meta="index"
        class="flex-1">
        <bcs-input
          v-model="item.value"
          :placeholder="$t('值')"
          @change="handleLabelValueChange">
        </bcs-input>
      </Validate>
      <bcs-select
        v-model="item.effect"
        class="effect ml15 flex-1"
        :placeholder="$t('影响')"
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
import Validate from '@/components/validate.vue';
import { KEY_REGEXP, VALUE_REGEXP } from '@/common/constant';

interface ITaint {
  key: string;
  value: string;
  effect: string;
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

      return keys.every(key => new RegExp(KEY_REGEXP).test(key))
        && values.every(value => new RegExp(VALUE_REGEXP).test(value)) ;
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
