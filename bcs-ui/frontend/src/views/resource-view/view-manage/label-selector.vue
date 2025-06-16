<template>
  <div>
    <div
      v-for="item, index in labelSelector"
      :key="index"
      :class="[
        'flex items-center',
        {
          'mb-[8px]': index < (labelSelector.length - 1)
        }
      ]">
      <div class="flex items-center flex-1">
        <span
          class="flex items-center justify-center w-[26px] h-[32px] text-[#3A84FF] mr-[8px] bcs-border"
          v-if="index > 0 && type === 'normal'">
          &
        </span>
        <bcs-select
          :popover-min-width="300"
          :loading="keyLoading"
          searchable
          v-model="item.key"
          class="flex-1 bg-[#fff] w-0 mr-[4px]"
          v-bk-tooltips="{
            content: item.key,
            disabled: !item.key
          }">
          <bcs-option v-for="label in labelSuggestList" :key="label" :id="label" :name="label" />
        </bcs-select>
      </div>
      <PopoverSelector
        v-if="type === 'normal'"
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
            {{ item.op }}
          </span>
          <i
            :class="[
              'bk-icon icon-angle-down relative transition-all text-[18px] text-[#979BA5]',
              {
                'bcs-rotate': !isHide && activeOpIndex === index
              }
            ]"></i>
        </span>
        <template #content>
          <li
            :class="['bcs-dropdown-item', { active: op === item.op }]"
            v-for="op in opList"
            :key="op"
            @click="handleChangeOp(index, op)">
            {{ op }}
          </li>
        </template>
      </PopoverSelector>
      <div
        v-else
        :class="[
          'px-[8px] h-[32px] min-w-[68px] text-center leading-[32px]',
          'text-[#FF9C01] bg-[#EAEBF0] rounded-sm'
        ]">=</div>
      <LabelValues
        v-model="item.values"
        :label="item.key"
        :cluster-namespaces="clusterNamespaces"
        :multiple="item.op !== '='"
        :key="item.op"
        class="ml-[4px]"
        v-bk-tooltips="{
          content: Array.isArray(item.values) ? item.values.join(', ') : item.values,
          disabled: Array.isArray(item.values) ? !item.values.length : !item.values
        }"
        v-if="!['Exists', 'DoesNotExist'].includes(item.op)" />
      <i
        :class="[
          'bk-icon icon-plus-circle-shape cursor-pointer',
          'text-[16px] text-[#C4C6CC] ml-[16px]'
        ]"
        @click="handleAddLabelSelector">
      </i>
      <i
        :class="[
          'bk-icon icon-minus-circle-shape',
          'text-[16px] text-[#C4C6CC] ml-[10px] cursor-pointer',
          labelSelector.length === 1 ? '!cursor-not-allowed !text-[#EAEBF0]' : ''
        ]"
        @click="handleMinusLabelSelector(index)">
      </i>
    </div>
  </div>
</template>
<script setup lang="ts">
import { onBeforeMount, PropType, ref, watch } from 'vue';

import LabelValues from './label-values.vue';
import useViewConfig from './use-view-config';

import PopoverSelector from '@/components/popover-selector.vue';

interface ILabelSelector {
  key: string
  op: string
  values: string[]
}

const props = defineProps({
  value: {
    type: Array as PropType<ILabelSelector[]>,
    default: () => [],
  },
  clusterNamespaces: {
    type: Array as PropType<IClusterNamespace[]>,
    default: () => [],
  },
  type: {
    type: String as PropType<'normal' | 'simple'>,
    default: 'normal',
  },
});
const emits = defineEmits(['input', 'change']);

const activeOpIndex = ref(-1);
const isHide = ref(true);
const onHide = () => {
  isHide.value = true;
};
const onShow = () => {
  isHide.value = false;
};

const labelSelector = ref<ILabelSelector[]>([{
  key: '',
  op: '=',
  values: [],
}]);
const initLabelSelector = () => {
  if (!props.value.length) {
    labelSelector.value = [{
      key: '',
      op: '=',
      values: [],
    }];
  } else {
    labelSelector.value = JSON.parse(JSON.stringify(props.value));
  }
};
const watchOnce = watch(() => props.value, () => {
  initLabelSelector();
  watchOnce();
});
watch(labelSelector, () => {
  const filterLabelSelector = labelSelector.value.filter(item => !!item.key);

  emits('change', filterLabelSelector);
  emits('input', filterLabelSelector);
}, { deep: true });

const popoverRef = ref();
const opList = ref(['=', 'In', 'NotIn', 'Exists', 'DoesNotExist']);
const handleChangeOp = (index: number, op: string) => {
  const item = labelSelector.value[index];
  item.op = op;
  if (['Exists', 'DoesNotExist'].includes(item.op)) {
    item.values = [];
  }
  popoverRef.value?.forEach((item) => {
    item?.hide();
  });
};

const handleAddLabelSelector = () => {
  labelSelector.value.push({
    key: '',
    op: '=',
    values: [],
  });
};
const handleMinusLabelSelector = (index: number) => {
  if (labelSelector.value.length <= 1) return;
  labelSelector.value.splice(index, 1);
};

const keyLoading = ref(false);
const labelSuggestList = ref([]);
const { labelSuggest } = useViewConfig();
watch(() => props.clusterNamespaces, async () => {
  // if (isEqual(newValue, oldValue)) return;
  keyLoading.value = true;
  labelSuggestList.value = await labelSuggest({
    clusterNamespaces: props.clusterNamespaces,
  });
  keyLoading.value = false;
}, { immediate: true, deep: true });

onBeforeMount(() => {
  initLabelSelector();
});
</script>
