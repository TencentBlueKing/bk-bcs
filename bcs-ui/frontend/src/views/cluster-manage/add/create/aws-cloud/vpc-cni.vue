<template>
  <div>
    <div
      class="flex flex-1 items-center mb-[10px]"
      v-for="subnet, index in subnetSourceNew"
      :key="index">
      <template v-if="showZone">
        <span class="prefix">{{ $t('tke.label.zone') }}</span>
        <Zone
          class="flex-1 ml-[-1px] mr-[8px]"
          :region="region"
          :cloud-account-i-d="cloudAccountID"
          :cloud-i-d="cloudID"
          :disabled-zone-list="getDisableZoneList(subnet.zone)"
          :disabled-tips="$t('tke.tips.hasSelected')"
          :init-data="index === 0"
          :value-id="valueId"
          v-model="subnet.zone" />
      </template>
      <span class="prefix">{{ $t('tke.label.ipNum') }}</span>
      <bcs-select
        class="flex-1 ml-[-1px]"
        searchable
        :clearable="false"
        v-model="subnet.mask">
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

import Zone from '@/views/cluster-manage/add/components/zone.vue';

const nodePodNumList = ref([12, 24, 32]);

const props = defineProps({
  subnets: {
    type: Array as PropType<Array<{
      zone: string
      mask: number
    }>>,
    default: () => [],
  },
  cloudAccountID: {
    type: String,
    default: '',
  },
  cloudID: {
    type: String,
    default: '',
  },
  region: {
    type: String,
    default: '',
  },
  showZone: {
    type: Boolean,
    default: true,
  },
  valueId: {
    type: String,
    default: 'zone',
  },
});

const emits = defineEmits(['change']);

const subnetSourceNew = ref<Array<{
  zone: string
  mask: number
}>>([]);

watch(() => props.subnets, () => {
  if (!props.subnets.length) {
    subnetSourceNew.value = [{
      mask: 24,
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

// 一个可用区只能选择一次
const getDisableZoneList = (excludeZone: string) => subnetSourceNew.value
  .filter(item => item.zone !== excludeZone)
  .map(item => item.zone);
// vpc-cni 模式ip数量
const addSubnetSource = () => {
  subnetSourceNew.value.push({
    zone: '',
    mask: 24,
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
