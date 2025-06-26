<template>
  <Validate
    :rules="rules"
    :value="systemDisk"
    error-display-type="normal"
    ref="validateRef"
    @validate="handleValidate">
    <div class="flex items-center">
      <span :class="['prefix', { disabled: isEdit }]">{{ $t('tke.label.systemDisk') }}</span>
      <bcs-select
        :clearable="false"
        :disabled="isEdit"
        class="ml-[-1px] w-[140px]"
        :loading="loading"
        v-model="systemDisk.diskType">
        <template #trigger v-if="isEdit">
          <div class="relative">
            <div
              :title="diskMap[systemDisk.diskType] || systemDisk.diskType"
              class="pr-[36px] pl-[10px] overflow-hidden text-ellipsis text-nowrap">
              {{ diskMap[systemDisk.diskType] || systemDisk.diskType }}
            </div>
            <i
              :class="[
                'absolute top-[4px] right-[2px] text-[22px] text-[#979ba5]',
                'bk-icon icon-angle-down',
              ]"></i>
          </div>
        </template>
        <bcs-option
          v-for="diskItem in list"
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
        :disabled="isEdit"
        v-model="systemDisk.diskSize">
      </bcs-input>
      <span class="suffix ml-[-1px]">GB</span>
    </div>
  </Validate>
</template>
<script setup lang="ts">
import { computed, PropType, ref, watch } from 'vue';

import { diskMap } from '../create/tencent-public-cloud/use-disk';

import { diskEnum } from '@/common/constant';
import Validate from '@/components/validate.vue';
import $i18n from '@/i18n/i18n-setup';

const props = defineProps({
  value: {
    type: Object,
    default: () => ({}),
  },
  list: {
    type: Array as PropType<{id: string, name: string}[]>,
    default: () => [...diskEnum],
  },
  firstTrigger: {
    type: Boolean,
    default: true,
  },
  loading: {
    type: Boolean,
    default: false,
  },
  isEdit: {
    type: Boolean,
    default: false,
  },
});
const emits = defineEmits(['change', 'validate']);

const systemDisk = ref({
  diskType: '',
  diskSize: '50',
});
const validateRef = ref();

// 校验系统盘
const rules = computed<Array<{validator: () => boolean, message: string}>>(() => [
  {
    validator: validateSystemDisk,
    message: !systemDisk.value.diskType
      ? $i18n.t('cluster.ca.nodePool.create.instanceTypeConfig.validate.systemDiskType')
      : $i18n.t('cluster.ca.nodePool.create.instanceTypeConfig.validate.systemDisk'),
  },
]);
function validateSystemDisk() {
  const diskSize = Number(systemDisk.value.diskSize);
  return props.firstTrigger
    || (!!systemDisk.value.diskType
      && (diskSize % 10 === 0)
      && (diskSize) >= 50);
};

async function validate() {
  return await validateRef.value?.validate();
}

function handleValidate(result: boolean) {
  emits('validate', result);
}

watch(() => props.value, (newValue, oldValue) => {
  if (JSON.stringify(newValue) === JSON.stringify(oldValue)) return;

  systemDisk.value = Object.assign({
    diskType: '',
    diskSize: '50',
  }, props.value);
}, { immediate: true });

watch(systemDisk, () => {
  emits('change', {
    ...systemDisk.value,
    diskSize: String(systemDisk.value.diskSize),
  });
}, { deep: true });

defineExpose({
  validate,
});
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
