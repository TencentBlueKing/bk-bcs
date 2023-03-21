<script setup lang="ts">
  import { ref, onMounted, computed } from 'vue'
  import { Search } from 'bkui-vue/lib/icon'
  import InfoBox from "bkui-vue/lib/info-box";
  import { getConfigVersionList } from '../../../../../api/config';
  import { IConfigVersionItem, IRequestFilter ,IPageFilter, FilterOp, RuleOp } from '../../../../../types'

  const props = defineProps<{
    bkBizId: string,
    appId: number
  }>()

  const listLoading = ref(true)
  const versionList = ref<Array<IConfigVersionItem>>([])
  const currentTab = ref('available')
  const pagination = ref({
    current: 1,
    count: 0,
    limit: 10,
  })
  const filter = ref<IRequestFilter>({
    op: FilterOp.AND,
    rules: [{
      field: "deprecated",
      op: RuleOp.eq,
      value: false
    }]
  })

  const page = computed(():IPageFilter => {
    return {
      count: false,
      start: (pagination.value.current - 1) * pagination.value.limit,
      limit: pagination.value.limit,
    }
  })

  onMounted(() => {
    getVersionList()
  })

  const getVersionList = async() => {
    listLoading.value = true
    const res = await getConfigVersionList(props.bkBizId, props.appId, filter.value, page.value)
    versionList.value = res.data.details
    listLoading.value = false
  }

  const handleTabChange = (tab: string) =>  {
    currentTab.value = tab
    filter.value.rules[0].value = tab === 'deprecate'
    refreshConfigList()
  }

  // 版本对比
  const handleOpenDiff = (version: IConfigVersionItem) => {
    console.log(version)
  }

  // 废弃
  const handleDeprecate = (id: number) => {
    InfoBox({
      title: '确认废弃此版本？',
      subTitle: '废弃操作无法撤回，请谨慎操作！',
      headerAlign: "center" as const,
      footerAlign: "center" as const,
      onConfirm: () => {
      },
    } as any);
  }

  const handlePageLimitChange = (limit: number) => {
    pagination.value.limit = limit
    refreshConfigList()
  }

  const refreshConfigList = (current: number = 1) => {
    pagination.value.current = current
    getVersionList()
  }

</script>
<template>
  <section class="version-detail-table">
    <div class="head-operate-wrapper">
      <div class="type-tabs">
        <div :class="['tab-item', { active: currentTab === 'available' }]" @click="handleTabChange('available')">可用版本</div>
        <div class="split-line"></div>
        <div :class="['tab-item', { active: currentTab === 'deprecate' }]" @click="handleTabChange('deprecate')">废弃版本</div>
      </div>
      <bk-input class="version-search-input" placeholder="版本名称/版本说明/修改人">
        <template #suffix>
            <Search class="search-input-icon" />
         </template>
      </bk-input>
    </div>
    <bk-loading :loading="listLoading">
        <bk-table :border="['outer']" :data="versionList">
          <bk-table-column label="版本" prop="spec.name"></bk-table-column>
          <bk-table-column label="版本描述" prop="spec.memo"></bk-table-column>
          <bk-table-column label="上线次数">xx</bk-table-column>
          <bk-table-column label="最后修改人">xx</bk-table-column>
          <bk-table-column label="最后修改时间">xx</bk-table-column>
          <bk-table-column label="状态">
            <div class="status-tag unpublished">未上线</div>
            <!-- <div class="status-tag published">已上线上线</div> -->
          </bk-table-column>
          <bk-table-column label="操作">
            <template v-slot="{ row }">
              <bk-button text theme="primary" @click="handleOpenDiff(row)">版本对比</bk-button>
              <bk-button style="margin-left: 16px;" text theme="primary" @click="handleDeprecate(row.id)">废弃</bk-button>
            </template>
          </bk-table-column>
        </bk-table>
        <bk-pagination
          class="table-list-pagination"
          v-model="pagination.current"
          location="left"
          :layout="['total', 'limit', 'list']"
          :count="pagination.count"
          :limit="pagination.limit"
          @change="refreshConfigList($event)"
          @limit-change="handlePageLimitChange"/>
    </bk-loading>
  </section>
</template>
<style lang="scss" scoped>
  .version-detail-table {
    padding: 24px;
    height: 100%;
    background: #ffffff;
  }
  .head-operate-wrapper {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;
  }
  .type-tabs {
    display: flex;
    align-items: center;
    padding: 3px 4px;
    background: #f0f1f5;
    border-radius: 4px;
    .tab-item {
      padding: 6px 14px;
      font-size: 12px;
      line-height: 14px;
      color: #63656e;
      border-radius: 4px;
      cursor: pointer;
      &.active {
        color: #3a84ff;
        background: #ffffff;
      }
    }
    .split-line {
      margin: 0 4px;
      width: 1px;
      height: 14px;
      background: #dcdee5;
    }
  }
  .version-search-input {
    width: 320px;
  }
  .search-input-icon {
    padding-right: 10px;
    color: #979ba5;
  }
  .status-tag {
    display: inline-block;
    padding: 0 10px;
    line-height: 20px;
    font-size: 12px;
    border: 1px solid #cccccc;
    border-radius: 11px;
    text-align: center;
    &.unpublished {
      color: #fe9000;
      background: #ffe8c3;
      border-color: rgba(254, 156, 0, 0.3);
    }
    &.published {
      color: #14a568;
      background: #e4faf0;
      border-color: rgba(20, 165, 104, 0.3);
    }
  }
  .table-list-pagination {
    padding: 12px;
    border: 1px solid #dcdee5;
    border-top: none;
    border-radius: 0 0 2px 2px;
    :deep(.bk-pagination-list.is-last) {
      margin-left: auto;
    }
  }
</style>
