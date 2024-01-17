<template>
  <div class="flex items-center">
    <span class="prefix">{{ $t('tke.label.systemDisk') }}</span>
    <bcs-select :clearable="false" class="ml-[-1px] w-[140px]" v-model="systemDisk.diskType">
      <bcs-option
        v-for="diskItem in diskEnum"
        :key="diskItem.id"
        :id="diskItem.id"
        :name="diskItem.name">
      </bcs-option>
    </bcs-select>
    <bcs-input
      class="w-[88px] bg-[#fff] ml10"
      type="number"
      :min="50"
      :max="1000"
      v-model="systemDisk.diskSize">
    </bcs-input>
    <span class="suffix ml-[-1px]">GB</span>
    <p
      class="bcs-form-error-tip text-[12px] text-[#ea3636] ml-[6px]"
      v-if="Number(systemDisk.diskSize || 0) % 10 !== 0">
      {{$t('cluster.ca.nodePool.create.instanceTypeConfig.validate.systemDisk')}}
    </p>
  </div>
</template>
<script setup lang="ts">
import { ref, watch } from 'vue';

import { diskEnum } from '@/common/constant';

const props = defineProps({
  value: {
    type: Object,
    default: () => ({}),
  },
});
const emits = defineEmits(['change']);

const systemDisk = ref({
  diskType: 'CLOUD_PREMIUM',
  diskSize: '50',
});

watch(() => props.value, (newValue, oldValue) => {
  if (JSON.stringify(newValue) === JSON.stringify(oldValue)) return;

  systemDisk.value = Object.assign({
    diskType: 'CLOUD_PREMIUM',
    diskSize: '50',
  }, props.value);
}, { immediate: true });

watch(systemDisk, () => {
  emits('change', {
    ...systemDisk.value,
    diskSize: String(systemDisk.value.diskSize),
  });
}, { deep: true });
</script>
<style lang="postcss" scoped>
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
.suffix{
  line-height: 30px;
  font-size: 12px;
  display: inline-block;
  min-width: 30px;
  padding: 0 4px 0 4px;
  height: 32px;
  border: 1px solid #C4C6CC;
  text-align: center;
  border-left: none;
  background-color: #fafbfd;
  &.disabled {
    border-color: #dcdee5;
  }
}
</style>
