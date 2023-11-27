<template>
  <div>
    <div
      class="flex flex-1 items-center mb-[10px]"
      v-for="subnet, index in subnetSourceNew"
      :key="index">
      <span
        class="prefix"
        v-bk-tooltips="$t('tke.tips.subnetZone')">
        <span class="bcs-border-tips">{{ $t('tke.label.zone') }}</span>
      </span>
      <bcs-select
        class="flex-1 ml-[-1px] mr-[8px]"
        searchable
        :clearable="false"
        v-model="subnet.zone">
        <bcs-option
          v-for="zone in zoneList"
          :key="zone.zoneID"
          :id="zone.zone"
          :name="zone.zoneName">
        </bcs-option>
      </bcs-select>
      <span class="prefix">{{ $t('tke.label.ipNum') }}</span>
      <bcs-select
        class="flex-1 ml-[-1px]"
        searchable
        :clearable="false"
        v-model="subnet.ipCnt">
        <bcs-option v-for="item in nodePodNumList" :id="item" :key="item" :name="item"></bcs-option>
      </bcs-select>
      <span class="flex items-center text-[#979BA5] ml-[8px] text-[14px]">
        <i
          class="bk-icon icon-plus-circle-shape mr-[5px] cursor-pointer"
          @click="addSubnetSource()"></i>
        <i
          :class="[
            'bk-icon icon-minus-circle-shape cursor-pointer',
            subnetSourceNew.length <= 1
              ? '!cursor-not-allowed !text-[#DCDEE5]' : ''
          ]"
          @click="removeSubnetSource(index)"></i>
      </span>
    </div>
  </div>
</template>
<script setup lang="ts">
import { PropType, ref, watch } from 'vue';

import { IZoneItem } from './types';

const nodePodNumList = ref([128, 256, 512, 1024, 2048, 4096]);

const props = defineProps({
  subnets: {
    type: Array as PropType<Array<{
      zone: string
      ipCnt: number
    }>>,
    default: () => [],
  },
  zoneList: {
    type: Array as PropType<IZoneItem[]>,
    default: () => [],
  },
});

const emits = defineEmits(['change']);

const subnetSourceNew = ref<Array<{
  zone: string
  ipCnt: number
}>>([]);

watch(() => props.subnets, () => {
  if (!props.subnets.length) {
    subnetSourceNew.value = [{
      ipCnt: 256,
      zone: '',
    }];
    return;
  }

  const newSubnets = JSON.stringify(props.subnets);
  const oldSubnets = JSON.stringify(subnetSourceNew.value);
  if (newSubnets === oldSubnets) return;

  subnetSourceNew.value = JSON.parse(newSubnets);
}, { immediate: true });

watch(subnetSourceNew, () => {
  emits('change', subnetSourceNew.value);
}, { deep: true });

// vpc-cni 模式ip数量
const addSubnetSource = () => {
  subnetSourceNew.value.push({
    zone: '',
    ipCnt: 256,
  });
};
const removeSubnetSource = (index) => {
  if (subnetSourceNew.value.length <= 1) return;
  subnetSourceNew.value.splice(index, 1);
};
</script>
<style scoped lang="postcss">
>>> .prefix {
  display: inline-block;
  height: 32px;
  line-height: 32px;
  background: #F0F1F5;
  border: 1px solid #C4C6CC;
  border-radius: 2px 0 0 2px;
  padding: 0 8px;
  font-size: 12px;
  &.disabled {
    border-color: #dcdee5;
  }
}
</style>
