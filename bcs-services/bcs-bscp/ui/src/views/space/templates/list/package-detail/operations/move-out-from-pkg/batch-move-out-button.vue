<template>
  <bk-button :disabled="props.configs.length === 0" @click="handleClick">{{ t('批量移出') }}</bk-button>
  <BatchMoveOutFromPkgDialog
    v-model:show="isBatchMoveDialogShow"
    :current-pkg="props.currentPkg"
    :value="props.configs"
    @moved-out="emits('movedOut')" />
  <MoveOutFromPkgsDialog
    v-model:show="isSingleMoveDialogShow"
    :id="props.configs.length > 0 ? props.configs[0].id : 0"
    :name="props.configs.length > 0 ? props.configs[0].spec.name : ''"
    :current-pkg="props.currentPkg"
    @moved-out="emits('movedOut')" />
</template>
<script lang="ts" setup>
  import { ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { ITemplateConfigItem } from '../../../../../../../../types/template';
  import BatchMoveOutFromPkgDialog from './batch-move-out-from-pkg-dialog.vue';
  import MoveOutFromPkgsDialog from './move-out-from-pkgs-dialog.vue';

  const { t } = useI18n();
  const props = defineProps<{
    configs: ITemplateConfigItem[];
    currentPkg: number;
  }>();

  const emits = defineEmits(['movedOut']);

  const isBatchMoveDialogShow = ref(false);
  const isSingleMoveDialogShow = ref(false);

  const handleClick = () => {
    if (props.configs.length === 1) {
      isSingleMoveDialogShow.value = true;
    } else {
      isBatchMoveDialogShow.value = true;
    }
  };
</script>
