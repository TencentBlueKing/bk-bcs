<script setup lang="ts">
  import { ref, onMounted } from 'vue'
  import { Plus, Search } from 'bkui-vue/lib/icon'
  import { storeToRefs } from 'pinia'
  import { useGlobalStore } from '../../../store/global'
  import { getScriptList } from '../../../api/script'
  import { IScriptItem } from '../../../../types/script'
  import CreateScript from './create-script.vue'
  import ScriptCited from './script-cited.vue'
  import VersionManage from './version-manage/index.vue'

  const { spaceId } = storeToRefs(useGlobalStore())

  const showCreateScript = ref(false)
  const showVersionManage = ref(false)
  const showCiteSlider = ref(false)
  const scriptsData = ref<IScriptItem[]>([])
  const scriptsLoading = ref(false)
  const currentId = ref(0)
  const pagination = ref({
    current: 1,
    count: 0,
    limit: 10,
  })

  onMounted(() => {
    getScripts()
  })

  // 获取脚本列表
  const getScripts = async () => {
    scriptsLoading.value = true
    const params = {
      start: (pagination.value.current - 1) * pagination.value.limit,
      limit: pagination.value.limit
    }
    const res = await getScriptList(spaceId.value, params)
    scriptsData.value = res.detail
    pagination.value.count = res.count
    scriptsLoading.value = false
  }

  const refreshList = (val: number = 1) => {
    pagination.value.current = 1
    getScripts()
  }

  const handlePageLimitChange = (val: number) => {
    pagination.value.limit = val
    refreshList()
  }
</script>
<template>
  <section class="scripts-manange-page">
    <div class="side-menu">
      <div class="group-wrapper">
        <li><i class="bk-bscp-icon icon-block-shape"></i>全部脚本</li>
        <li><i class="bk-bscp-icon icon-tags"></i>未分类</li>
      </div>
    </div>
    <div class="script-list-wrapper">
      <div class="operate-area">
        <bk-button theme="primary" @click="showCreateScript = true"><Plus class="button-icon" />新建脚本</bk-button>
        <bk-button theme="primary" @click="showVersionManage = true"><Plus class="button-icon" />版本管理</bk-button>
        <bk-button theme="primary" @click="showCiteSlider = true"><Plus class="button-icon" />被引用</bk-button>
        <bk-input class="search-script-input" placeholder="脚本名称">
           <template #suffix>
              <Search class="search-input-icon" />
           </template>
        </bk-input>
      </div>
      <bk-table :border="['outer']">
        <bk-table-column label="脚本名称"></bk-table-column>
        <bk-table-column label="脚本语言"></bk-table-column>
        <bk-table-column label="分类标签"></bk-table-column>
        <bk-table-column label="被引用"></bk-table-column>
        <bk-table-column label="更新人"></bk-table-column>
        <bk-table-column label="更新时间"></bk-table-column>
        <bk-table-column label="操作">
          <template #defaut="{ row }">
            <bk-button text theme="primary">编辑</bk-button>
            <bk-button text theme="primary">版本管理</bk-button>
            <bk-button text theme="primary">删除</bk-button>
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
        @change="refreshList"
        @limit-change="handlePageLimitChange"/>
    </div>
    <CreateScript v-if="showCreateScript" v-model:show="showCreateScript" @created="refreshList" />
    <ScriptCited v-model:show="showCiteSlider" :id="currentId" />
    <VersionManage v-if="showVersionManage" v-model:show="showVersionManage" />
  </section>
</template>
<style lang="scss" scoped>
  .scripts-manange-page {
    display: flex;
    align-items: center;
    height: 100%;
    background: #ffffff;
  }
  .side-menu {
    padding: 16px 0;
    width: 280px;
    height: 100%;
    background: #f5f7fa;
    box-shadow: 0 2px 2px 0 rgba(0, 0, 0, 0.15);
    z-index: 1;
    .group-wrapper {
      padding-bottom: 16px;
      border-bottom: 1px solid #dcdee5;
      & > li {
        padding: 8px 22px;
        color: #313238;
        font-size: 12px;
        cursor: pointer;
        &:hover {
          color: #348aff;
        }
        i {
          margin-right: 8px;
          font-size: 14px;
          color: #979ba5;
        }
      }
    }
  }
  .script-list-wrapper {
    padding: 24px;
    width: calc(100% - 280px);
    height: 100%;
    background: #ffffff;
  }
  .operate-area {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;
    .button-icon {
      font-size: 18px;
    }
  }
  .search-script-input {
    width: 320px;
  }
  .search-input-icon {
    padding-right: 10px;
    color: #979ba5;
    background: #ffffff;
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
