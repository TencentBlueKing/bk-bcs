<template>
  <div>
    <div class="key-value" v-for="(item, index) in taints" :key="index">
      <Validate :rules="rules" :value="item.key" :meta="index">
        <bcs-input
          v-model="item.key"
          class="key"
          :placeholder="$t('键')"
          @change="handleLabelKeyChange">
        </bcs-input>
      </Validate>
      <span class="ml8 mr8">=</span>
      <bcs-input
        v-model="item.value"
        :placeholder="$t('值')"
        class="value"
        @change="handleLabelValueChange"
      ></bcs-input>
      <bcs-select
        v-model="item.effect"
        class="effect ml15 flex"
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
import { computed, defineComponent, ref, watch } from '@vue/composition-api';
import Validate from '@/components/validate.vue';
import $i18n from '@/i18n/i18n-setup';
import { LABEL_KEY_REGEXP } from '@/common/constant';

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
    keyRules: {
      type: Array,
    },
  },
  setup(props, ctx) {
    const taints = ref<ITaint[]>([]);
    const rules = ref(props.keyRules
      ? props.keyRules
      : [
        {
          message: $i18n.t('有效的标签键有两个段：可选的前缀和名称，用斜杠（/）分隔。 名称段是必需的，必须小于等于 63 个字符，以字母数字字符（[a-z0-9A-Z]）开头和结尾， 可带有破折号（-），下划线（_），点（ .）和之间的字母数字。 前缀是可选的。如果指定，前缀必须是 DNS 子域：由点（.）分隔的一系列 DNS 标签，总共不超过 253 个字符， 后跟斜杠（/）。'),
          validator: LABEL_KEY_REGEXP,
        },
        {
          message: $i18n.t('重复键'),
          validator: (value, meta) => taints.value.filter((_, i) => i !== meta).every(d => d.key !== value),
        },
      ]);
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
      const data = taints.value.reduce<string[]>((pre, item) => {
        if (item.key) {
          pre.push(item.key);
        }
        return pre;
      }, []);
      const removeDuplicateData = new Set(data);
      if (data.length !== removeDuplicateData.size) {
        return false;
      }

      return data.every(key => new RegExp(LABEL_KEY_REGEXP).test(key));
    };
    return {
      rules,
      taints,
      disabledDelete,
      handleLabelKeyChange,
      handleLabelValueChange,
      handleAddLabel,
      handleDeleteLabel,
      validate,
    };
  },
});
</script>
<style lang="postcss" scoped>
.mr8 {
  margin-right: 8px;
}
.ml8 {
  margin-left: 8px;
}
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
  .flex {
      flex: 1;
  }
}
</style>
