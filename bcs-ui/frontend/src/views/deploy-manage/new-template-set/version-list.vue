<template>
  <div>
    <bcs-input
      right-icon="bk-icon icon-search"
      :placeholder="$t('templateSet.placeholder.searchVersion')"
      clearable
      v-model.trim="searchValue">
    </bcs-input>
    <bcs-table
      class="mt-[16px]"
      v-bkloading="{ isLoading: loading }"
      :data="curPageData"
      :pagination="pagination"
      @page-change="pageChange"
      @page-limit-change="pageSizeChange">
      <bcs-table-column :label="$t('templateSet.label.versionNo')" prop="version"></bcs-table-column>
      <!-- 被引用 -->
      <!-- <bcs-table-column></bcs-table-column> -->
      <bcs-table-column :label="$t('generic.label.updator')" prop="updateBy">
        <template #default="{ row }">
          <bk-user-display-name :user-id="row.updateBy"></bk-user-display-name>
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('generic.label.updatedAt')" prop="updateTime"></bcs-table-column>
      <bcs-table-column :label="$t('generic.label.action')" width="120">
        <template #default="{ row }">
          <bcs-button text @click="cloneVersion(row)">{{ $t('generic.button.clone') }}</bcs-button>
          <bcs-button class="ml-[5px]" text @click="deleteVersion(row)">{{ $t('generic.button.delete') }}</bcs-button>
        </template>
      </bcs-table-column>
    </bcs-table>
  </div>
</template>
<script setup lang="ts">
import { ref, watch } from 'vue';

import { HelmManagerService } from '@/api/modules/new-helm-manager';
import $bkMessage from '@/common/bkmagic';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import usePage from '@/composables/use-page';
import useSearch from '@/composables/use-search';
import $i18n from '@/i18n/i18n-setup';

const props = defineProps({
  name: String,
  repoName: String,
});

// 版本列表
const loading = ref(false);
const versionList = ref<ChartVersion[]>([]);
const keys = ref(['version', 'updateBy']);
const { searchValue, tableDataMatchSearch } = useSearch(versionList, keys);
const {
  pagination,
  curPageData,
  pageChange,
  pageSizeChange,
} = usePage<ChartVersion>(tableDataMatchSearch);
async function getVersionList() {
  if (!props.name || !props.repoName) return;

  loading.value = true;
  const { data = [] } = await HelmManagerService.ListChartVersionV1({
    $name: props.name,
    $repoName: props.repoName,
    page: 1,
    size: 9999,
  }).catch(() => ({ data: [] }));
  versionList.value = data;
  loading.value = false;
}

// 克隆版本
function cloneVersion(row: ChartVersion) {
  console.log(row);
}

// 删除版本
function deleteVersion(row: ChartVersion) {
  $bkInfo({
    type: 'warning',
    clsName: 'custom-info-confirm',
    title: $i18n.t('generic.title.confirmDelete1', { name: row.version }),
    defaultInfo: true,
    okText: $i18n.t('generic.button.delete'),
    confirmFn: async () => {
      if (!props.name || !props.repoName || !row.version) return;
      const result = await HelmManagerService.DeleteChartVersion({
        $name: props.name,
        $repoName: props.repoName,
        $version: row.version,
      }).then(() => true)
        .catch(() => false);
      if (result) {
        $bkMessage({
          theme: 'success',
          message: $i18n.t('generic.msg.success.delete'),
        });
        getVersionList();
      }
    },
  });
}
watch(
  () => [
    props.name,
    props.repoName,
  ],
  () => {
    getVersionList();
  },
  { immediate: true },
);
</script>
