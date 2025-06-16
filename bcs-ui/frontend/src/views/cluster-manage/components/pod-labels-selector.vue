<template>
  <div class="ml-[22px] mt-[8px]">
    <div
      v-for="(item, index) in keyValueData"
      :key="index"
      class="flex items-center mb-[10px]"
    >
      <bcs-select
        v-model="item.key"
        class="w-[160px]"
        placeholder="key"
        :clearable="false"
        @change="item.value = ''">
        <bcs-option
          v-for="key in nodeKeyList"
          :key="key"
          :id="key"
          :name="key">
        </bcs-option>
      </bcs-select>
      <span class="mx-[8px]">=</span>
      <bcs-select
        v-model="item.value"
        class="w-[160px]"
        placeholder="value"
        :clearable="false">
        <bcs-option
          v-for="value in labelValueMap[item.key]"
          :key="value"
          :id="value"
          :name="value">
        </bcs-option>
      </bcs-select>
      <div class="felx items-center" v-if="multiple">
        <i
          :class="[
            'bk-icon icon-plus-circle ml10 mr5 text-[24px] cursor-pointer',
            { 'opacity-50 hover:cursor-not-allowed': nodeKeyList.length === keyValueData.length }
          ]" @click="handleAddKeyValue(index)"></i>
        <i
          :class="[
            'bk-icon icon-minus-circle text-[24px] cursor-pointer',
            { 'opacity-50 hover:cursor-not-allowed': keyValueData.length === 1 }
          ]"
          @click="handleDeleteKeyValue(index)"
        ></i>
      </div>
    </div>
  </div>
</template>
<script lang="ts" setup>
import { PropType, ref, watch } from 'vue';

interface IData {
  key: string;
  value: string;
  placeholder?: any;
  disabled?: boolean;
}

const props = defineProps({
  nodeKeyList: {
    type: Array as PropType <string[]>,
    default() {
      return [];
    },
  },
  labelValueMap: {
    type: Object as PropType <Record<string, string[]>>,
    default() {
      return [];
    },
  },
  multiple: {
    type: Boolean,
    default: false,
  },
});

const emits = defineEmits(['change']);


const keyValueData = ref<IData[]>([
  {
    key: '',
    value: '',
  },
]);
const handleAddKeyValue = (index) => {
  if (props.nodeKeyList?.length === keyValueData.value?.length) return;
  keyValueData.value.splice(index + 1, 0, {
    key: '',
    value: '',
  });
};
const handleDeleteKeyValue = (index) => {
  if (keyValueData.value?.length === 1) return;
  keyValueData.value.splice(index, 1);
};

watch(keyValueData, () => {
  emits('change', keyValueData.value);
}, { deep: true });

</script>
