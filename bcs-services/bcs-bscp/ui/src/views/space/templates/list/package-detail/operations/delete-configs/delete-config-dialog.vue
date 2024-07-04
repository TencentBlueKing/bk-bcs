<template>
  <bk-dialog
    v-if="props.isBatchDelete"
    :title="t('确认删除所选的 {n} 个配置文件?', { n: props.configs.length })"
    header-align="center"
    :is-show="props.show"
    ext-cls="delete-confirm-dialog">
    <div class="tips">{{ t('一旦删除，该操作将无法撤销，请谨慎操作') }}</div>
    <bk-table :data="props.configs" border="outer" max-height="200">
      <bk-table-column :label="t('配置文件绝对路径')">
        <template #default="{ row }">
          <span v-if="row.spec">{{ fileAP(row) }}</span>
        </template>
      </bk-table-column>
    </bk-table>
    <template #footer>
      <bk-button
        theme="primary"
        style="margin-right: 8px"
        :disabled="pending"
        :loading="pending"
        @click="handleConfirm">
        {{ t('删除') }}
      </bk-button>
      <bk-button @click="close">{{ t('取消') }}</bk-button>
    </template>
  </bk-dialog>
  <DeleteConfirmDialog
    v-else
    :title="t('确认删除该配置文件？')"
    :is-show="props.show"
    :pending="pending"
    @confirm="handleConfirm"
    @close="close">
    <div style="margin-bottom: 8px">
      {{ t('配置文件') }}:
      <span style="color: #313238; font-weight: 600">{{ props.configs[0] ? props.configs[0].spec.name : '' }}</span>
    </div>
    <div>{{ t('一旦删除，该操作将无法撤销，请谨慎操作') }}</div>
  </DeleteConfirmDialog>
</template>
<script lang="ts" setup>
  import { ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { storeToRefs } from 'pinia';
  import Message from 'bkui-vue/lib/message';
  import useGlobalStore from '../../../../../../../store/global';
  import useTemplateStore from '../../../../../../../store/template';
  import { ITemplateConfigItem } from '../../../../../../../../types/template';
  import { deleteTemplate } from '../../../../../../../api/template';
  import DeleteConfirmDialog from '../../../../../../../components/delete-confirm-dialog.vue';
  const { spaceId } = storeToRefs(useGlobalStore());
  const { currentTemplateSpace } = storeToRefs(useTemplateStore());
  const { t } = useI18n();

  const props = defineProps<{
    show: boolean;
    configs: ITemplateConfigItem[];
    isBatchDelete?: boolean;
  }>();

  const emits = defineEmits(['update:show', 'deleted']);

  const pending = ref(false);

  // 配置文件绝对路径
  const fileAP = (config: ITemplateConfigItem) => {
    const { path, name } = config.spec;
    if (path.endsWith('/')) {
      return `${path}${name}`;
    }
    return `${path}/${name}`;
  };

  const handleConfirm = async () => {
    try {
      pending.value = true;
      const ids = props.configs.map((config) => config.id);
      await deleteTemplate(spaceId.value, currentTemplateSpace.value, ids);
      close();
      setTimeout(() => {
        emits('deleted');
        Message({
          theme: 'success',
          message: t('删除配置文件成功'),
        });
      }, 300);
    } catch (e) {
      console.log(e);
    } finally {
      pending.value = false;
    }
  };

  const close = () => {
    emits('update:show', false);
  };
</script>

<style lang="scss" scoped>
  .list-wrap {
    margin-bottom: 8px;
    border: 1px solid #dbdbdb;
    border-radius: 2px;
  }

  :deep(.bk-table) {
    margin-bottom: 16px;
    margin-top: 8px;
  }
</style>

