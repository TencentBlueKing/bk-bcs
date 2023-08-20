<script lang="ts" setup>
  import { ref } from 'vue'
  import { ITemplateConfigItem } from '../../../../../../../../types/template';
  import BatchMoveOutFromPkgDialog from './batch-move-out-from-pkg-dialog.vue';

  const props = defineProps<{
    configs: ITemplateConfigItem[];
  }>()

  const emits = defineEmits(['movedOut'])

  const isDialogShow = ref(false)
</script>
<template>
  <bk-button :disabled="props.configs.length === 0" @click="isDialogShow = true">批量移出</bk-button>
  <BatchMoveOutFromPkgDialog v-model:show="isDialogShow" :value="props.configs" @moved-out="emits('movedOut')" />
</template>
<style lang="scss" scoped>

</style>
