<template>
  <BcsContent :title="$t('cluster.nodeTemplate.text')">
    <template #header-right>
      <a
        :href="PROJECT_CONFIG.nodeTemplate"
        target="_blank"
        class="bk-text-button"
      >{{$t('cluster.nodeTemplate.button.useLink')}}</a>
    </template>
    <div class="node-template">
      <div class="node-template-header">
        <bk-button
          theme="primary"
          icon="plus"
          @click="handleAddTemplate">
          {{$t('cluster.nodeTemplate.title.create')}}
        </bk-button>
        <bk-input
          class="search-input"
          v-model="searchValue"
          right-icon="bk-icon icon-search"
          :placeholder="$t('cluster.nodeTemplate.placeholder.searchTemplate')"
          clearable>
        </bk-input>
      </div>
      <bcs-table
        :data="curPageData"
        :pagination="pagination"
        v-bkloading="{ isLoading: loading }"
        @page-change="pageChange"
        @page-limit-change="pageSizeChange">
        <!-- <bcs-table-column label="ID" prop="nodeTemplateID"></bcs-table-column> -->
        <bcs-table-column :label="$t('cluster.nodeTemplate.label.templateName')" prop="name">
          <template #default="{ row }">
            <bcs-button
              text
              @click="handleShowDetail(row)"
            >
              <span class="bcs-ellipsis">{{row.name}}</span>
            </bcs-button>
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('cluster.create.label.desc')" prop="desc" show-overflow-tooltip>
          <template #default="{ row }">
            {{row.desc || '--'}}
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('generic.label.createdBy1')" prop="creator">
          <template #default="{ row }">
            <bk-user-display-name :user-id="row.creator"></bk-user-display-name>
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('generic.label.updator1')" prop="updater">
          <template #default="{ row }">
            <bk-user-display-name :user-id="row.updater"></bk-user-display-name>
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('cluster.labels.createdAt')" prop="createTime"></bcs-table-column>
        <bcs-table-column :label="$t('cluster.labels.updatedAt')" prop="updateTime"></bcs-table-column>
        <bcs-table-column :label="$t('generic.label.action')" width="180">
          <template #default="{ row }">
            <bk-button text @click="handleEdit(row)">{{$t('generic.button.edit')}}</bk-button>
            <bk-button text class="ml10" @click="handleDelete(row)">{{$t('generic.button.delete')}}</bk-button>
          </template>
        </bcs-table-column>
        <template #empty>
          <BcsEmptyTableStatus :type="searchValue ? 'search-empty' : 'empty'" @clear="searchValue = ''" />
        </template>
      </bcs-table>
      <bcs-sideslider
        :is-show.sync="showDetail"
        :title="currentRow.name"
        quick-close
        :width="800">
        <template #content>
          <NodeTemplateDetail
            :data="currentRow"
            operate
            class="h-[calc(100vh-60px)] overflow-auto"
            @delete="handleDeleteDetail"
            @cancel="showDetail = false" />
        </template>
      </bcs-sideslider>
    </div>
  </BcsContent>
</template>
<script lang="ts">
import { defineComponent, onMounted, ref } from 'vue';

import NodeTemplateDetail from './node-template-detail.vue';

import $bkMessage from '@/common/bkmagic';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import BcsContent from '@/components/layout/Content.vue';
import usePage from '@/composables/use-page';
import useSearch from '@/composables/use-search';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $store from '@/store/index';

export default defineComponent({
  name: 'NodeTemplateConfig',
  components: { NodeTemplateDetail, BcsContent },
  setup() {
    const loading = ref(false);
    const data = ref<any[]>([]);
    const handleNodeTemplateList = async () => {
      loading.value = true;
      data.value = await $store.dispatch('clustermanager/nodeTemplateList');
      loading.value = false;
    };
    const keys = ref(['name', 'creator', 'updater']);
    const { searchValue, tableDataMatchSearch } = useSearch(data, keys);
    const {
      pagination,
      curPageData,
      pageChange,
      pageSizeChange,
    } = usePage(tableDataMatchSearch);

    // 添加模板
    const handleAddTemplate = () => {
      $router.push({
        name: 'addNodeTemplate',
      });
    };
    const handleEdit = (row) => {
      $router.push({
        name: 'editNodeTemplate',
        params: {
          nodeTemplateID: row.nodeTemplateID,
        },
      });
    };
    const handleDelete = (row) => {
      $bkInfo({
        type: 'warning',
        clsName: 'custom-info-confirm',
        subTitle: row.name,
        title: $i18n.t('cluster.nodeTemplate.title.confirmDelete'),
        defaultInfo: true,
        confirmFn: async () => {
          loading.value = true;
          const result = await $store.dispatch('clustermanager/deleteNodeTemplate', {
            $nodeTemplateId: row.nodeTemplateID,
          });
          if (result) {
            $bkMessage({
              theme: 'success',
              message: $i18n.t('generic.msg.success.delete'),
            });
            handleNodeTemplateList();
          }
          loading.value = false;
        },
      });
    };
    const showDetail = ref(false);
    const currentRow = ref<any>({});
    const handleShowDetail = (row) => {
      currentRow.value = row;
      showDetail.value = true;
    };

    function handleDeleteDetail() {
      showDetail.value = false;
      handleNodeTemplateList();
    }
    onMounted(() => {
      handleNodeTemplateList();
    });
    return {
      showDetail,
      currentRow,
      loading,
      searchValue,
      pagination,
      curPageData,
      pageChange,
      pageSizeChange,
      handleEdit,
      handleDelete,
      handleAddTemplate,
      handleShowDetail,
      handleDeleteDetail,
    };
  },
});
</script>
<style lang="postcss" scoped>
.node-template {
    &-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        margin-bottom: 16px;
        .search-input {
            max-width: 400px;
        }
    }
}
</style>
