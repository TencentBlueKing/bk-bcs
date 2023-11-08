<template>
  <bk-button class="reset-default-value-btn" text theme="primary" @click="triggerReset">
    恢复默认值
    <Help class="help-icon" v-bk-tooltips="{ content: '服务变量值默认继承上个版本', placement: 'top' }" />
  </bk-button>
</template>
<script lang="ts" setup>
import { ref, watch } from 'vue';
import { Help } from 'bkui-vue/lib/icon';
import { cloneDeep } from 'lodash';
import { IVariableEditParams } from '../../../../../../../../../types/variable';

const props = defineProps<{
  list: IVariableEditParams[];
}>();

const emits = defineEmits(['reset']);

const initialList = ref(cloneDeep(props.list));

watch(
  () => props.list,
  () => {
    initialList.value = cloneDeep(props.list);
  },
);

const triggerReset = () => {
  emits('reset', cloneDeep(initialList.value));
};
</script>
<style lang="scss" scoped>
.reset-default-value-btn {
  font-size: 12px;
  .help-icon {
    margin-left: 4px;
  }
}
</style>
