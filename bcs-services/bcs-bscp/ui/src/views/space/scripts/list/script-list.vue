<script setup lang="ts">
  import { ref, onMounted } from 'vue'
  import { useRouter } from 'vue-router'
  import { InfoBox } from 'bkui-vue'
  import { Plus, Search } from 'bkui-vue/lib/icon'
  import { storeToRefs } from 'pinia'
  import { useGlobalStore } from '../../../../store/global'
  import { getScriptList, getScriptTagList, deleteScript } from '../../../../api/script'
  import { IScriptItem, IScriptTagItem } from '../../../../../types/script'
  import CreateScript from '../create-script.vue'
  import ScriptCited from '../script-cited.vue'

  const { spaceId } = storeToRefs(useGlobalStore())
  const router = useRouter()

  const showCreateScript = ref(false)
  const showCiteSlider = ref(false)
  const scriptsData = ref<IScriptItem[]>([])
  const scriptsLoading = ref(false)
  const tagsData = ref<IScriptTagItem[]>([])
  const tagsLoading = ref(false)
  const currentId = ref(0)
  const pagination = ref({
    current: 1,
    count: 0,
    limit: 10,
  })

  onMounted(() => {
    getScripts()
    getTags()
  })

  // 获取脚本列表
  const getScripts = async () => {
    scriptsLoading.value = true
    const params = {
      start: (pagination.value.current - 1) * pagination.value.limit,
      limit: pagination.value.limit,
      all: true
    }
    const res = await getScriptList(spaceId.value, params)
    scriptsData.value = res.details
    pagination.value.count = res.count
    scriptsLoading.value = false
  }

  // 获取标签列表
  const getTags = async () => {
    tagsLoading.value = true
    const res = await getScriptTagList(spaceId.value)
    scriptsData.value = res.details
    pagination.value.count = res.count
    scriptsLoading.value = false
  }

  // 删除分组
  const handleDeleteScript = (script: IScriptItem) => { 
    InfoBox({
      title: `确认是否删除脚本【${script.spec.name}?】`,
      infoType: "danger",
      headerAlign: "center" as const,
      footerAlign: "center" as const,
      onConfirm: async () => {
        await deleteScript(spaceId.value, script.id)
        if (scriptsData.value.length === 1 && pagination.value.current > 1) {
          pagination.value.current = pagination.value.current - 1
        }
        getScripts()
      },
    } as any)
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
      <div class="group-wrapper">
        <li v-for="(item, index) in tagsData" :key="index"><i class="bk-bscp-icon icon-tags"></i>{{ item.tag }}</li>
      </div>
    </div>
    <div class="script-list-wrapper">
      <div class="operate-area">
        <bk-button theme="primary" @click="showCreateScript = true"><Plus class="button-icon" />新建脚本</bk-button>
        <bk-button theme="primary" @click="showCiteSlider = true">被引用</bk-button>
        <bk-input class="search-script-input" placeholder="脚本名称">
           <template #suffix>
              <Search class="search-input-icon" />
           </template>
        </bk-input>
      </div>
      <bk-table :border="['outer']" :data="scriptsData">
        <bk-table-column label="脚本名称" prop="spec.name"></bk-table-column>
        <bk-table-column label="脚本语言" prop="spec.type" width="120"></bk-table-column>
        <bk-table-column label="分类标签" prop="spec.tag"></bk-table-column>
        <bk-table-column label="被引用" width="100">
          <template #default="{ row }">
            <template v-if="row.spec">
              <bk-button v-if="row.spec.publish_num > 0" text theme="primary">{{ row.spec.publish_num }}</bk-button>
              <span v-else>0</span>
            </template>
          </template>
        </bk-table-column>
        <bk-table-column label="更新人" prop="revision.reviser"></bk-table-column>
        <bk-table-column label="更新时间" prop="revision.update_at" min-width="180"></bk-table-column>
        <bk-table-column label="操作">
          <template #default="{ row }" width="180">
            <div class="action-btns">
              <bk-button text theme="primary">编辑</bk-button>
              <bk-button text theme="primary" @click="router.push({ name: 'script-version-manage', params: { spaceId, scriptId: row.id } })">版本管理</bk-button>
              <bk-button text theme="primary" @click="handleDeleteScript(row)">删除</bk-button>
            </div>
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
  .action-btns {
    .bk-button {
      margin-right: 8px;
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
