<template>
  <bk-button class="reset-default-value-btn" text theme="primary" @click="triggerReset">
    {{ t('恢复默认值') }}
    <Help class="help-icon" v-bk-tooltips="{ content: t('服务变量值默认继承上个版本'), placement: 'top' }" />
  </bk-button>
</template>
<script lang="ts" setup>
import { ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { Help } from 'bkui-vue/lib/icon';
import { cloneDeep } from 'lodash';
import { IVariableEditParams } from '../../../../../../../../../types/variable';

const { t } = useI18n();
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
