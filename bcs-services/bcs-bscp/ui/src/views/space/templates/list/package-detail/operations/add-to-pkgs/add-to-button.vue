<template>
  <bk-button :disabled="props.configs.length === 0" @click="isDialogShow = true">批量添加至</bk-button>
  <AddToDialog v-model:show="isDialogShow" :value="props.configs" @added="emits('refresh')" />
</template>
<script lang="ts" setup>
import { ref } from 'vue';
import { ITemplateConfigItem } from '../../../../../../../../types/template';
import AddToDialog from './add-to-dialog.vue';

const props = defineProps<{
  configs: ITemplateConfigItem[];
}>();

const emits = defineEmits(['refresh']);

const isDialogShow = ref(false);
</script>
<style lang="scss" scoped></style>
