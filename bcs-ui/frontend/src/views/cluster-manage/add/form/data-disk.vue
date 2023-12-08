<template>
  <div class="text-[12px]">
    <bk-checkbox v-model="showDataDisk" :disabled="disabled">{{ $t('tke.button.purchaseDataDisk') }}</bk-checkbox>
    <template v-if="showDataDisk">
      <div
        class="relative bg-[#F5F7FA] py-[16px] px-[24px] mt-[10px]"
        v-for="item, index in cloudDataDisks"
        :key="index">
        <div class="flex items-center">
          <span class="prefix">{{ $t('tke.label.dataDisk') }}</span>
          <bcs-select :clearable="false" class="ml-[-1px] w-[140px] mr-[16px] bg-[#fff]" v-model="item.diskType">
            <bcs-option
              v-for="diskItem in diskEnum"
              :key="diskItem.id"
              :id="diskItem.id"
              :name="diskItem.name">
            </bcs-option>
          </bcs-select>
          <bcs-input class="max-w-[120px]" type="number" v-model="item.diskSize">
            <span slot="append" class="group-text !px-[4px]">GB</span>
          </bcs-input>
        </div>
        <div class="flex items-center mt-[16px]">
          <bk-checkbox v-model="item.autoFormatAndMount" class="mr-[8px]">
            {{ $t('tke.button.autoFormatAndMount') }}
          </bk-checkbox>
          <template v-if="item.autoFormatAndMount">
            <bcs-select :clearable="false" class="w-[80px] mr-[8px] bg-[#fff]" v-model="item.fileSystem">
              <bcs-option v-for="file in fileSystem" :key="file" :name="file" :id="file"></bcs-option>
            </bcs-select>
            <bk-input class="flex-1" v-model="item.mountTarget"></bk-input>
          </template>
        </div>
        <!-- 删除 -->
        <span
          :class="[
            'absolute right-0 top-0 text-[12px]',
            'inline-flex items-center justify-center w-[24px] h-[24px]',
            'cursor-pointer text-[#979ba5] hover:text-[#3a84ff]'
          ]"
          v-if="!disabled"
          @click="handleDeleteDiskData(index)">
          <i class="bk-icon icon-close3-shape"></i>
        </span>
      </div>
      <!-- 添加 -->
      <div
        :class="[
          'flex items-center justify-center h-[32px] text-[12px] bg-[#fafbfd] mt-[16px]',
          'rounded border-dashed border-[#c4c6cc] border-[1px] cursor-pointer',
          'hover:text-[#3a84ff] hover:border-[#3a84ff]'
        ]"
        v-if="!disabled"
        @click="handleAddDiskData">
        <i class="bk-icon left-icon icon-plus"></i>
        <span>{{$t('cluster.ca.nodePool.create.instanceTypeConfig.button.addDataDisks')}}</span>
      </div>
    </template>
  </div>
</template>
<script setup lang="ts">
import { PropType, ref, watch } from 'vue';

import { diskEnum } from '@/common/constant';
import { IDataDisk } from '@/views/cluster-manage/add/tencent/types';

const props = defineProps({
  value: {
    type: Array as PropType<IDataDisk[]>,
    default: () => [],
  },
  disabled: {
    type: Boolean,
    default: false,
  },
});
const emits = defineEmits(['change']);

// 数据盘
const showDataDisk = ref(true);
const fileSystem = ref(['ext3', 'ext4', 'xfs']);

const cloudDataDisks = ref<Array<IDataDisk>>(props.value);

watch(() => props.value, (newValue, oldValue) => {
  if (JSON.stringify(newValue) === JSON.stringify(oldValue)) return;

  cloudDataDisks.value = [
    ...props.value,
  ];
});
watch(cloudDataDisks.value, () => {
  emits('change', cloudDataDisks.value);
});

const handleDeleteDiskData = (index) => {
  cloudDataDisks.value.splice(index, 1);
  showDataDisk.value = !!cloudDataDisks.value.length;
};

const handleAddDiskData = () => {
  cloudDataDisks.value.push({
    diskType: 'CLOUD_PREMIUM', // 类型
    diskSize: '100', // 大小
    fileSystem: 'ext4', // 文件系统
    autoFormatAndMount: true, // 是否格式化
    mountTarget: '/data', // 挂载路径
  });
};
</script>
