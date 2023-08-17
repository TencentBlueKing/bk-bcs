<script lang="ts" setup>
  import { ref } from 'vue'
  import { ITemplateConfigItem } from '../../../../../../../../types/template';
  import BatchMoveOutFromPkgDialog from './batch-move-out-from-pkg-dialog.vue';

  const props = defineProps<{
    configs: ITemplateConfigItem[];
  }>()

  const isDialogShow = ref(false)
</script>
<template>
  <bk-button :disabled="props.configs.length === 0" @click="isDialogShow = true">批量移出</bk-button>
  <BatchMoveOutFromPkgDialog v-model:show="isDialogShow" :value="props.configs" />
</template>
<style lang="scss" scoped>

</style>
