<template>
  <div class="border border-[#DCDEE5] border-solid">
    <div class="flex items-center bg-[#fafbfd] h-[42px] px-[20px] config-border-b">
      <div class="w-[110px] mr-[10px]">{{ $t('logCollector.label.index') }}</div>
      <div class="flex-1 mr-[10px]">{{ $t('logCollector.label.eq') }}</div>
      <div class="w-[60px]">{{ $t('logCollector.label.addAndDele') }}</div>
    </div>
    <div class="pb-[14px] px-[20px] bg-[#fff] relative">
      <div v-for="item, index in data" :key="index" class="flex items-center pt-[12px]">
        <bcs-input
          class="w-[110px] mr-[10px]"
          type="number"
          :precision="0"
          :min="1"
          :disabled="fromOldRule"
          v-model="item.fieldindex">
        </bcs-input>
        <div class="flex-1 mr-[10px]">
          <bcs-input
            :disabled="fromOldRule"
            :class="[
              'w-[240px]',
              {
                'line-after': data.length > 1
              }
            ]"
            v-model="item.word">
          </bcs-input>
        </div>
        <div class="w-[60px]">
          <i
            class="text-[16px] text-[#C4C6CC] bk-icon icon-plus-circle-shape cursor-pointer"
            @click="handleAddConfig">
          </i>
          <i
            :class="[
              'text-[16px] text-[#C4C6CC] bk-icon icon-minus-circle-shape ml-[10px] cursor-pointer',
              { 'cursor-not-allowed text-[#dcdee5]': data.length === 1 }
            ]"
            @click="handleDeleteConfig(index)">
          </i>
        </div>
      </div>
      <div
        :class="[
          'flex items-center justify-center absolute top-0 right-[100px] h-full',
          'separator-config'
        ]"
        v-if="data.length > 1">
        <bcs-select :disabled="fromOldRule" :clearable="false" class="w-[80px]" v-model="logicOp">
          <bcs-option id="and" :name="$t('logCollector.label.matchContent.separator.conditions.and')"></bcs-option>
          <bcs-option id="or" :name="$t('logCollector.label.matchContent.separator.conditions.or')"></bcs-option>
        </bcs-select>
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
import { PropType, ref, watch } from 'vue';
import { ISeparatorFilter } from './use-log';

const props = defineProps({
  value: {
    type: Array as PropType<ISeparatorFilter[]>,
    default: () => [],
  },
  fromOldRule: {
    type: Boolean,
    default: false,
  },
});

const emits = defineEmits(['input']);

const logicOp = ref(props.value[0]?.logic_op || 'and');

const data = ref(props.value.length ? props.value : [{
  fieldindex: '',
  word: '',
  op: '=',
  logic_op: 'and',
}]);

watch([
  logicOp,
  data,
], () => {
  emits('input', data.value.map(item => ({
    ...item,
    logic_op: logicOp.value,
  })));
}, { deep: true });

const handleAddConfig = () => {
  data.value.push({
    fieldindex: '',
    word: '',
    op: '=',
    logic_op: 'and',
  });
};

const handleDeleteConfig = (index) => {
  if (data.value.length <= 1) return;
  data.value.splice(index, 1);
};
</script>
<style scoped>
.config-border-b {
  border-bottom: 1px solid #DCDEE5;
}
.line-after::after {
  border-top: 1px dashed #c4c6cc;
  content: "";
  height: 1px;
  left: 100%;
  position: absolute;
  top: 16px;
  width: 25px;
}
.separator-config::before {
  border-top: 1px dashed #c4c6cc;
  content: "";
  height: 1px;
  position: absolute;
  right: 80px;
  top: 50%;
  width: 20px;
}
.separator-config::after {
  border-left: 1px dashed #c4c6cc;
  content: "";
  height: calc(100% - 56px);
  position: absolute;
  right: 100px;
  width: 1px;
}
</style>
