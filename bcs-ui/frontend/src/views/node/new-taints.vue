<template>
  <div>
    <div class="key-value" v-for="(item, index) in taints" :key="index">
      <bcs-input
        v-model="item.key"
        :placeholder="keyPlaceholder"
        class="flex"
        @change="handleLabelKeyChange">
      </bcs-input>
      <span class="ml8 mr8">=</span>
      <bcs-input
        v-model="item.value"
        :placeholder="valuePlaceholder"
        class="flex"
        @change="handleLabelValueChange"
      ></bcs-input>
      <bcs-select v-model="item.effect" class="ml15 flex" :placeholder="effectPlaceholder">
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

interface ITaint {
  key: string;
  value: string;
  effect: string;
}
export default defineComponent({
  name: 'BcsTaints',
  model: {
    prop: 'value',
    event: 'change',
  },
  props: {
    value: {
      type: Array,
      default: (): ITaint[] => ([]),
    },
    keyPlaceholder: {
      type: String,
      default: '',
    },
    valuePlaceholder: {
      type: String,
      default: '',
    },
    effectPlaceholder: {
      type: String,
      default: '',
    },
    effectOptions: {
      type: Array,
      default: () => ['NoSchedule', 'PreferNoSchedule', 'NoExecute'],
    },
    minItem: {
      type: Number,
      default: 0,
    },
  },
  setup(props, ctx) {
    const taints = ref<ITaint[]>([]);
    watch(props.value, (data: any[]) => {
      taints.value = data;
      if (taints.value.length < (props.minItem || 0)) {
        const reset: ITaint[] = new Array<ITaint>(props.minItem - taints.value.length)
          .fill({ key: '', value: '', effect: '' });
        taints.value.push(...reset);
      }
    }, { immediate: true, deep: true });
    watch(taints.value, () => {
      emitChange();
    }, { deep: true });
    const disabledDelete = computed(() => taints.value.length <= props.minItem);
    const emitChange = () => {
      ctx.emit('change', taints);
    };
    const handleLabelKeyChange = (newValue, oldValue) => {
      ctx.emit('key-change', newValue, oldValue);
    };
    const handleLabelValueChange = (newValue, oldValue) => {
      ctx.emit('value-change', newValue, oldValue);
    };
    const handleAddLabel = (index) => {
      taints.value.splice(index + 1, 0, { key: '', value: '', effect: '' });
      ctx.emit('add', index);
    };
    const handleDeleteLabel = (index) => {
      if (disabledDelete.value) return;
      taints.value.splice(index, 1);
      ctx.emit('delete', index);
    };
    return {
      taints,
      disabledDelete,
      handleLabelKeyChange,
      handleLabelValueChange,
      handleAddLabel,
      handleDeleteLabel,
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
    .flex {
        flex: 1;
    }
    .ml8 {
        margin-left: 8px;
    }
    .mr8 {
        margin-right: 8px;
    }
}
</style>
