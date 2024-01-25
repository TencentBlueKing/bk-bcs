<template>
  <bk-button :disabled="props.configs.length === 0" @click="isDialogShow = true">{{t('批量删除')}}</bk-button>
  <DeleteConfigDialog
    v-model:show="isDialogShow"
    :is-batch-delete="true"
    :configs="props.configs"
    @deleted="emits('deleted')"/>
</template>
<script lang="ts" setup>
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { ITemplateConfigItem } from '../../../../../../../../types/template';
import DeleteConfigDialog from './delete-config-dialog.vue';

const { t } = useI18n();
const props = defineProps<{
  spaceId: string;
  currentTemplateSpace: number;
  configs: ITemplateConfigItem[];
}>();

const emits = defineEmits(['deleted']);

const isDialogShow = ref(false);

// const pending = ref(false)

// const handleDeleteTemplate = async () => {
//   try {
//     pending.value = true
//     const result = await useDeleteTemplateConfigs(props.spaceId, props.currentTemplateSpace, props.configs)
//     if (result) {
//       emits('deleted')
//     }
//   } catch (e) {
//     console.log(e)
//   } finally {
//     pending.value = false
//   }
// }
</script>
