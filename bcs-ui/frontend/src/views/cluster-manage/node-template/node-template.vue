<template>
  <div>
    <ContentHeader :title="$t('节点模板')" hide-back>
      <template #right>
        <a
          :href="PROJECT_CONFIG.nodetemplate"
          target="_blank"
          class="bk-text-button"
        >{{$t('如何使用节点模板？')}}</a>
      </template>
    </ContentHeader>
    <div class="node-template bcs-content-wrapper">
      <div class="node-template-header">
        <bk-button theme="primary" icon="plus" @click="handleAddTemplate">{{$t('新建节点模板')}}</bk-button>
        <bk-input
          class="search-input"
          v-model="searchValue"
          right-icon="bk-icon icon-search"
          :placeholder="$t('输入名称、创建者、更新者搜索')"
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
        <bcs-table-column :label="$t('模板名称')" prop="name">
          <template #default="{ row }">
            <bcs-button
              text
              @click="handleShowDetail(row)"
            >
              <span class="bcs-ellipsis">{{row.name}}</span>
            </bcs-button>
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('描述')" prop="desc" show-overflow-tooltip>
          <template #default="{ row }">
            {{row.desc || '--'}}
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('创建者')" prop="creator"></bcs-table-column>
        <bcs-table-column :label="$t('更新者')" prop="updater"></bcs-table-column>
        <bcs-table-column :label="$t('创建时间')" prop="createTime"></bcs-table-column>
        <bcs-table-column :label="$t('更新时间')" prop="updateTime"></bcs-table-column>
        <bcs-table-column :label="$t('操作')" width="180">
          <template #default="{ row }">
            <bk-button text @click="handleEdit(row)">{{$t('编辑')}}</bk-button>
            <bk-button text class="ml10" @click="handleDelete(row)">{{$t('删除')}}</bk-button>
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
        <div slot="content">
          <NodeTemplateDetail
            :data="currentRow"
            operate
            @delete="handleDeleteDetail"
            @cancel="showDetail = false">
          </NodeTemplateDetail>
        </div>
      </bcs-sideslider>
    </div>
  </div>
</template>
<script lang="ts">
import { defineComponent, onMounted, ref } from 'vue';
import $i18n from '@/i18n/i18n-setup';
import usePage from '@/composables/use-page';
import useSearch from '@/composables/use-search';
import $router from '@/router';
import $store from '@/store/index';
import NodeTemplateDetail from './node-template-detail.vue';
import ContentHeader from '@/components/layout/Header.vue';
import $bkMessage from '@/common/bkmagic';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';

export default defineComponent({
  name: 'NodeTemplateConfig',
  components: { NodeTemplateDetail, ContentHeader },
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
        title: $i18n.t('确认删除配置模版？'),
        defaultInfo: true,
        confirmFn: async () => {
          loading.value = true;
          const result = await $store.dispatch('clustermanager/deleteNodeTemplate', {
            $nodeTemplateId: row.nodeTemplateID,
          });
          if (result) {
            $bkMessage({
              theme: 'success',
              message: $i18n.t('删除成功'),
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
    padding: 20px 24px;
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
