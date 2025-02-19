<template>
  <div class="text-[12px]">
    <bk-checkbox
      v-model="showDataDisk"
      :disabled="disabled || isEdit"
      @change="handleShowDataDisksChange">
      {{ $t('tke.button.purchaseDataDisk') }}
    </bk-checkbox>
    <template v-if="showDataDisk">
      <div
        class="relative bg-[#F5F7FA] py-[16px] px-[24px] mt-[10px]"
        v-for="item, index in cloudDataDisks"
        :key="index">
        <Validate
          :rules="[
            {
              validator: () => firstTrigger
                || (!!item.diskType && (Number(item.diskSize) % 10 === 0) && (Number(item.diskSize) >= 50)),
              message: !item.diskType ? $t('cluster.ca.nodePool.create.instanceTypeConfig.validate.dataDiskType')
                : $t('cluster.ca.nodePool.create.instanceTypeConfig.validate.dataDisks')
            }
          ]"
          :value="item"
          :meta="index"
          error-display-type="normal"
          ref="validateRefs"
          @validate="handleValidate">
          <div class="flex items-center">
            <span :class="['prefix', { disabled: isEdit }]">{{ $t('tke.label.dataDisk') }}</span>
            <bcs-select
              :disabled="isEdit"
              :clearable="false"
              class="ml-[-1px] w-[140px] mr-[16px] bg-[#fff]"
              v-model="item.diskType"
              :loading="loading">
              <bcs-option
                v-for="diskItem in list"
                :key="diskItem.id"
                :id="diskItem.id"
                :name="diskItem.name">
              </bcs-option>
            </bcs-select>
            <bcs-input
              class="max-w-[120px]"
              type="number"
              :disabled="isEdit"
              :min="50"
              :max="16380"
              v-model="item.diskSize">
            </bcs-input>
            <span class="suffix ml-[-1px]">GB</span>
          </div>
        </Validate>
        <div class="flex items-center mt-[16px]">
          <bk-checkbox :disabled="isEdit" v-model="item.autoFormatAndMount" class="mr-[8px]">
            {{ $t('tke.button.autoFormatAndMount') }}
          </bk-checkbox>
          <template v-if="item.autoFormatAndMount">
            <bcs-select
              :disabled="isEdit"
              :clearable="false"
              class="w-[80px] mr-[8px] bg-[#fff]"
              v-model="item.fileSystem">
              <bcs-option v-for="file in fileSystem" :key="file" :name="file" :id="file"></bcs-option>
            </bcs-select>
            <bk-input class="flex-1" :disabled="isEdit" v-model="item.mountTarget"></bk-input>
          </template>
        </div>
        <p
          class="bcs-form-error-tip text-[12px] text-[#ea3636] is-error"
          v-if="showRepeatMountTarget(index)">
          {{$t('cluster.ca.nodePool.create.instanceTypeConfig.validate.repeatPath')}}
        </p>
        <!-- 删除 -->
        <span
          :class="[
            'absolute right-0 top-0 text-[12px]',
            'inline-flex items-center justify-center w-[24px] h-[24px]',
            'cursor-pointer text-[#979ba5] hover:text-[#3a84ff]',
            { disabled: isEdit }
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
          'hover:text-[#3a84ff] hover:border-[#3a84ff]',
          { disabled: isEdit || cloudDataDisks.length > maxNum - 1 }
        ]"
        v-bk-tooltips="{
          content: $t('cluster.ca.nodePool.create.instanceTypeConfig.validate.maxDataDisks', [maxNum]),
          disabled: cloudDataDisks.length < maxNum
        }"
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
import Validate from '@/components/validate.vue';
import { IDataDisk } from '@/views/cluster-manage/types/types';

const props = defineProps({
  value: {
    type: Array as PropType<IDataDisk[]>,
    default: () => [],
  },
  disabled: {
    type: Boolean,
    default: false,
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
  maxNum: {
    type: Number,
    default: 5,
  },
});
const emits = defineEmits(['change', 'validate']);

// 数据盘
const fileSystem = ref(['ext3', 'ext4', 'xfs']);

const cloudDataDisks = ref<Array<IDataDisk>>(props.value);
const showDataDisk = ref(!!cloudDataDisks.value.length);
const validateRefs = ref();

function handleValidate(result?: boolean) {
  emits('validate', result);
}

// 文件目录不能重复
const validateDataDiskMountTarget = () => {
  const mountTargetList = cloudDataDisks.value?.map(item => item.mountTarget);
  return new Set(mountTargetList).size === mountTargetList.length;
};

async function validate() {
  handleValidate();
  const validateDiskMountTarget = validateDataDiskMountTarget();
  const data = validateRefs.value?.map($ref => $ref?.validate('blur')) || [];
  const results = await Promise.all(data);

  return validateDiskMountTarget && results.every(result => result);
};

const handleDeleteDiskData = (index) => {
  if (props.isEdit) return;
  cloudDataDisks.value.splice(index, 1);
  showDataDisk.value = !!cloudDataDisks.value.length;
};

const defaultDiskItem = {
  diskType: '', // 类型
  diskSize: '100', // 大小
  fileSystem: 'ext4', // 文件系统
  autoFormatAndMount: true, // 是否格式化
  mountTarget: '/data', // 挂载路径
};
const handleShowDataDisksChange = (show) => {
  cloudDataDisks.value = show ? [JSON.parse(JSON.stringify(defaultDiskItem))] : [];
};
const handleAddDiskData = () => {
  if (props.isEdit || cloudDataDisks.value.length > props.maxNum - 1) return;
  cloudDataDisks.value.push(JSON.parse(JSON.stringify(defaultDiskItem)));
};

const showRepeatMountTarget = (index) => {
  const disk = cloudDataDisks.value[index];
  return disk.autoFormatAndMount
            && disk.mountTarget
            && cloudDataDisks.value
              .filter((item, i) => i !== index && item.autoFormatAndMount)
              .some(item => item.mountTarget === disk.mountTarget);
};

watch(() => props.value, (newValue, oldValue) => {
  if (JSON.stringify(newValue) === JSON.stringify(oldValue)) return;

  cloudDataDisks.value = [
    ...props.value,
  ];
});
watch(() => cloudDataDisks.value, () => {
  emits('change', cloudDataDisks.value);
}, { deep: true });

defineExpose({
  validate,
});

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

.disabled {
  color: #C4C6CC;
  cursor: not-allowed;
  border-color: #C4C6CC;
}
</style>
