<template>
  <div>
    <div
      class="flex items-center mb-[8px]"
      v-for="item, index in data"
      :key="index">
      <bcs-input
        :disabled="disabled"
        class="flex-1 mr-[4px]"
        v-model="item.key">
      </bcs-input>
      <PopoverSelector
        offset="0, 6"
        :on-hide="onHide"
        :on-show="onShow"
        ref="popoverRef">
        <span
          :class="[
            'flex items-center justify-between px-[8px] h-[32px] min-w-[68px]',
            'text-[#FF9C01] bg-[#EAEBF0] rounded-sm cursor-pointer'
          ]"
          @click="activeOpIndex = index">
          <span class="flex flex-1 items-center justify-center">
            {{ item.operator }}
          </span>
          <i :class="[
            'bk-icon icon-angle-down relative transition-all text-[18px] text-[#979BA5]',
            {
              'bcs-rotate': !isHide && activeOpIndex === index
            }
          ]"></i>
        </span>
        <template #content>
          <li
            :class="['bcs-dropdown-item', { active: op === item.operator }]"
            v-for="op in opList"
            :key="op"
            @click="handleChangeOp(index, op)">
            {{ op }}
          </li>
        </template>
      </PopoverSelector>
      <bcs-input
        :disabled="disabled"
        :placeholder="$t('logCollector.placeholder.multiLabel')"
        class="flex-1 ml-[4px]"
        v-model="item.value"
        v-if="!['Exists', 'DoesNotExist'].includes(item.operator)">
      </bcs-input>
      <i
        :class="[
          'text-[16px] text-[#C4C6CC] bk-icon icon-plus-circle-shape ml-[16px] cursor-pointer',
          { '!cursor-not-allowed !text-[#EAEBF0]': disabled },
        ]"
        @click="handleAddLabel">
      </i>
      <i
        :class="[
          'text-[16px] text-[#C4C6CC] bk-icon icon-minus-circle-shape ml-[10px] cursor-pointer',
          { '!cursor-not-allowed !text-[#EAEBF0]': data.length === 1 || disabled },
        ]"
        @click="handleDeleteLabel(index)">
      </i>
    </div>
  </div>
</template>
<script setup lang="ts">
import { cloneDeep } from 'lodash';
import { onBeforeMount, PropType, ref, watch } from 'vue';

import PopoverSelector from '@/components/popover-selector.vue';

interface ILabel {
  key: string
  value: string
  operator: string
}

const props = defineProps({
  value: {
    type: Array as PropType<Array<ILabel>>,
    default: () => [],
  },
  disabled: {
    type: Boolean,
    default: false,
  },
});

const emits = defineEmits(['change']);

const activeOpIndex = ref(-1);
const data = ref<Array<ILabel>>([]);

const watchOnce = watch(() => props.value, () => {
  handleSetData();
  watchOnce();
});

watch(data, () => {
  emits('change', data.value);
});

// 运算符
const isHide = ref(true);
const onHide = () => {
  isHide.value = true;
};
const onShow = () => {
  isHide.value = false;
};
const popoverRef = ref();
const opList = ref(['=', 'In', 'NotIn', 'Exists', 'DoesNotExist']);
const handleChangeOp = (index: number, op: string) => {
  const item = data.value[index];
  item.operator = op;
  if (['Exists', 'DoesNotExist'].includes(item.operator)) {
    item.value = '';
  }
  popoverRef.value?.forEach((item) => {
    item?.hide();
  });
};

const handleSetData = () => {
  if (props.value.length) {
    data.value = cloneDeep(props.value);
  } else {
    data.value = [
      {
        key: '',
        value: '',
        operator: '=',
      },
    ];
  }
};

const handleAddLabel = () => {
  if (props.disabled) return;
  data.value.push({
    key: '',
    value: '',
    operator: '=',
  });
};

const handleDeleteLabel = (index: number) => {
  if (data.value.length === 1 || props.disabled) return;
  data.value.splice(index, 1);
};

onBeforeMount(() => {
  handleSetData();
});
</script>
