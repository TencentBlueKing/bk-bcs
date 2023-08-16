<script lang="ts" setup>
  import { ref } from 'vue'
  import { InfoBox } from 'bkui-vue';
  import { IPagination } from '../../../../../types/index';
  import { ITemplateVersionItem } from '../../../../../types/template';
  import { deleteTemplateVersion } from '../../../../api/template'

  const props = defineProps<{
    spaceId: string;
    currentTemplateSpace: number;
    templateId: number;
    list: ITemplateVersionItem[];
    pagination: IPagination;
  }>()

  const emits = defineEmits(['page-value-change', 'page-limit-change', 'openVersionDiff', 'deleted'])

  const pending = ref(false)

  const handleDeleteVersion = (version: ITemplateVersionItem) => {
    InfoBox({
      title: `确认彻底删除版本【${version.spec.revision_name}】？`,
      confirmText: '确认删除',
      infoType: 'warning',
      onConfirm: async() => {
        pending.value = true
        await deleteTemplateVersion(props.spaceId, props.currentTemplateSpace, props.templateId, version.id)
        pending.value = false
      }
    })
  }

</script>
<template>
  <bk-table
    class="version-table"
    :border="['outer']"
    :data="props.list"
    :pagination="pagination"
    @page-value-change="emits('page-value-change', $event)"
    @page-limit-change="emits('page-limit-change', $event)">
    <bk-table-column label="版本号" prop="spec.revision_name"></bk-table-column>
    <bk-table-column label="版本说明" prop="spec.revision_memo"></bk-table-column>
    <bk-table-column label="被引用"></bk-table-column>
    <bk-table-column label="更新人" prop="revision.reviser"></bk-table-column>
    <bk-table-column label="更新时间" prop="revision.update_at"></bk-table-column>
    <bk-table-column label="操作" width="180">
      <template #default="{ row }">
        <div class="actions-wrapper">
          <bk-button text theme="primary">版本对比</bk-button>
          <bk-button text theme="primary" @click="handleDeleteVersion(row)">删除</bk-button>
        </div>
      </template>
    </bk-table-column>
  </bk-table>
</template>
<style lang="scss" scoped>
  .version-table {
    width: 100%;
    background: #ffffff;
    :deep(.bk-table-footer) {
      padding-left: 16px;
    }
  }
  .actions-wrapper {
    .bk-button {
      margin-right: 8px;
    }
  }
</style>
