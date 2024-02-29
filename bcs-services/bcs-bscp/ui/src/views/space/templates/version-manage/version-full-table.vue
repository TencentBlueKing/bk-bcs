<template>
  <bk-table
    class="version-table"
    :border="['outer']"
    :data="props.list"
    :pagination="pagination"
    @page-value-change="emits('page-value-change', $event)"
    @page-limit-change="emits('page-limit-change', $event)">
    <bk-table-column :label="t('版本号')" prop="spec.revision_name">
      <template #default="{ row }">
        <bk-button v-if="row.spec" text theme="primary" @click="emits('select', row.id)">{{
          row.spec.revision_name
        }}</bk-button>
      </template>
    </bk-table-column>
    <bk-table-column :label="t('版本说明')">
      <template #default="{ row }">
        <span v-if="row.spec">{{ row.spec.revision_memo || '--' }}</span>
      </template>
    </bk-table-column>
    <bk-table-column :label="t('被引用')">
      <template #default="{ row, index }">
        <template v-if="boundByAppsCountLoading"><Spinner /></template>
        <template v-else-if="boundByAppsCountList[index]">
          <bk-button
            v-if="boundByAppsCountList[index].bound_unnamed_app_count > 0"
            text
            theme="primary"
            @click="handleOpenBoundDetailSlider(row)">
            {{ boundByAppsCountList[index].bound_unnamed_app_count }}
          </bk-button>
          <span v-else @click="handleOpenBoundDetailSlider(row)">0</span>
        </template>
      </template>
    </bk-table-column>
    <bk-table-column :label="t('创建人')" prop="revision.creator"></bk-table-column>
    <bk-table-column :label="t('创建时间')" prop="revision.create_at">
      <template #default="{ row }">
        <template v-if="row.revision">
          {{ datetimeFormat(row.revision.create_at) }}
        </template>
      </template>
    </bk-table-column>
    <bk-table-column :label="t('操作')" width="220">
      <template #default="{ row }">
        <div class="actions-wrapper">
          <bk-button text theme="primary" @click="handleOpenDiffSlider(row)">{{ t('版本对比') }}</bk-button>
          <bk-button text theme="primary" @click="handleDownload(row)">{{ t('下载') }}</bk-button>
          <!-- <bk-button text theme="primary" :disabled="pagination.count === 1" @click="handleDeleteVersion(row)"
            >删除</bk-button
          > -->
        </div>
      </template>
    </bk-table-column>
  </bk-table>
  <VersionBoundByAppsDetail
    v-model:show="boundDetailSliderData.open"
    :space-id="spaceId"
    :current-template-space="templateSpaceId"
    :config="boundDetailSliderData.data" />
  <TemplateVersionDiff
    v-model:show="diffSliderData.open"
    :space-id="spaceId"
    :template-space-id="templateSpaceId"
    :crt-version="diffSliderData.data" />
</template>
<script lang="ts" setup>
  import { ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Spinner } from 'bkui-vue/lib/icon';
  import { IPagination } from '../../../../../types/index';
  import {
    ITemplateVersionItem,
    ITemplateCitedCountDetailItem,
    DiffSliderDataType,
  } from '../../../../../types/template';
  import { datetimeFormat } from '../../../../utils/index';
  import { fileDownload } from '../../../../utils/file';
  import { downloadTemplateContent } from '../../../../api/template';
  import VersionBoundByAppsDetail from './version-bound-by-apps-detail.vue';
  import TemplateVersionDiff from './template-version-diff.vue';

  const { t } = useI18n();
  const props = defineProps<{
    spaceId: string;
    templateSpaceId: number;
    templateId: number;
    list: ITemplateVersionItem[];
    boundByAppsCountLoading: boolean;
    boundByAppsCountList: ITemplateCitedCountDetailItem[];
    pagination: IPagination;
  }>();

  const emits = defineEmits(['page-value-change', 'page-limit-change', 'openVersionDiff', 'select', 'deleted']);

  const boundDetailSliderData = ref<{ open: boolean; data: { id: number; versionId: number; name: string } }>({
    open: false,
    data: { id: 0, versionId: 0, name: '' },
  });
  const diffSliderData = ref<{ open: boolean; data: DiffSliderDataType }>({
    open: false,
    data: { id: 0, versionId: 0, name: '' },
  });

  const handleOpenBoundDetailSlider = (version: ITemplateVersionItem) => {
    const { id, spec } = version;
    boundDetailSliderData.value = {
      open: true,
      data: { id: props.templateId, versionId: id, name: spec.revision_name },
    };
  };

  const handleOpenDiffSlider = (version: ITemplateVersionItem) => {
    const { id, spec } = version;
    diffSliderData.value = {
      open: true,
      data: { id: props.templateId, versionId: id, name: spec.revision_name, permission: spec.permission },
    };
  };

  const handleDownload = async (version: ITemplateVersionItem) => {
    const { name, revision_name, content_spec } = version.spec;
    const content = await downloadTemplateContent(props.spaceId, props.templateSpaceId, content_spec.signature);
    fileDownload(content, `${name}_${revision_name}`);
  };

  // const handleDeleteVersion = (version: ITemplateVersionItem) => {
  //   InfoBox({
  //     title: `确认彻底删除版本【${version.spec.revision_name}】？`,
  //     confirmText: '确认删除',
  //     infoType: 'warning',
  //     onConfirm: async () => {
  //       pending.value = true;
  //       await deleteTemplateVersion(props.spaceId, props.templateSpaceId, props.templateId, version.id);
  //       pending.value = false;
  //       emits('deleted');
  //     },
  //   });
  // };
</script>
<style lang="scss" scoped>
  .version-table {
    width: 100%;
    background: #ffffff;
  }
  .actions-wrapper {
    .bk-button {
      margin-right: 8px;
    }
  }
</style>
