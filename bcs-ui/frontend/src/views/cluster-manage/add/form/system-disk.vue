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
    <bcs-select
      class="w-[88px] bg-[#fff] ml10"
      :clearable="false"
      v-model="systemDisk.diskSize">
      <!-- 后端需要字符串类型 -->
      <bcs-option id="50" name="50"></bcs-option>
      <bcs-option id="100" name="100"></bcs-option>
    </bcs-select>
    <span class="company">GB</span>
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
  emits('change', systemDisk.value);
});
</script>
