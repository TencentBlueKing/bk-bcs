<script setup lang="ts">
  import { ref, onMounted } from 'vue'
  import { useRouter, useRoute } from 'vue-router'
  import { InfoBox } from 'bkui-vue'
  import { Search, AngleDoubleRightLine } from 'bkui-vue/lib/icon'
  import { storeToRefs } from 'pinia'
  import { useGlobalStore } from '../../../../store/global'
  import { useScriptStore } from '../../../../store/script'
  import { IScriptVersion } from '../../../../../types/script'
  import { getScriptDetail, getScriptVersionList, deleteScriptVersion, publishVersion } from '../../../../api/script'
  import DetailLayout from '../components/detail-layout.vue'
  import VersionListFullTable from './version-list-full-table.vue'
  import VersionListSimpleTable from './version-list-simple-table.vue'
  import CreateVersion from './create-version.vue'
  import VersionEdit from './version-edit.vue'
  import ScriptVersionDiff from './script-version-diff.vue'

  const { spaceId } = storeToRefs(useGlobalStore())
  const {versionListPageShouldOpenEdit } = storeToRefs(useScriptStore())
  const router = useRouter()
  const route = useRoute()

  const scriptId = ref(Number(route.params.scriptId))
  const detailLoading = ref(true)
  const scriptDetail = ref({ spec: { name: '', type: '' } })
  const versionLoading = ref(true)
  const versionList = ref<IScriptVersion[]>([])
  const unPublishVersion = ref<IScriptVersion|null>(null) // 未发布版本
  const showVersionDiff = ref(false)
  const crtVersion = ref<IScriptVersion|null>(null)
  const versionEditData = ref({
    panelOpen: false,
    editable: true,
    form: { // 版本编辑、新建、查看数据
      id: 0,
      name: '',
      memo: '',
      content: ''
    }
  })
  const searchStr = ref('')
  const pagination = ref({
    current: 1,
    count: 0,
    limit: 10,
  })

  onMounted(async() => {
    getScriptDetailData()
    await getVersionList()
    if (versionListPageShouldOpenEdit.value) {
      versionListPageShouldOpenEdit.value = false
      if (unPublishVersion.value) {
        handleEditVersionClick()
      } else {
        handleCreateVersionClick('')
      }
    }
  })

  // 获取脚本详情
  const getScriptDetailData = async() => {
    detailLoading.value = true
    scriptDetail.value = await getScriptDetail(spaceId.value, scriptId.value)
    detailLoading.value = false
  }

  // 获取版本列表
  const getVersionList = async() => {
    versionLoading.value = true
    const params: { start: number; limit: number; searchKey?: string } = {
      start: (pagination.value.current - 1) * pagination.value.limit,
      limit: pagination.value.limit
    }
    if (searchStr.value) {
      params.searchKey = searchStr.value
    }
    const res = await getScriptVersionList(spaceId.value, scriptId.value, params)
    versionList.value = res.details
    pagination.value.count = res.count
    if (pagination.value.current === 1) {
      const version = versionList.value.find(item => item.spec.state === 'not_deployed')
      if (version) {
        unPublishVersion.value = version
      }
    }
    versionLoading.value = false
  }

  // 点击新建版本
  const handleCreateVersionClick = (content: string) => {
    versionEditData.value = {
      panelOpen: true,
      editable: true,
      form: {
        id: 0,
        name: '',
        memo: '',
        content
      }
    }
  }

  // 编辑未上线版本
  const handleEditVersionClick = () => {
    if (unPublishVersion.value) {
      const { name, memo, content } = unPublishVersion.value.spec
      versionEditData.value = {
        panelOpen: true,
        editable: true,
        form: {
          id: unPublishVersion.value?.id,
          name,
          memo,
          content
        }
      }
    }
  }

  // 查看版本
  const handleViewVersionClick = (version: IScriptVersion) => {
    const { name, memo, content } = version.spec
    versionEditData.value = {
      panelOpen: true,
      editable: false,
      form: {
        id: version.id,
        name,
        memo,
        content
      }
    }
  }

  // 上线版本
  const handlePublishClick = (version: IScriptVersion) => {
    InfoBox({
      title: '确定上线此版本？',
      subTitle: '上线后，之前的线上版本将被置为「已下线」状态',
      // infoType: 'warning',
      confirmText: '确定',
      onConfirm: async () => {
        await publishVersion(spaceId.value, scriptId.value, version.id)
        refreshList()
        unPublishVersion.value = null
        
      },
    } as any)
  }

  // 删除版本
  const handleDelClick = (version: IScriptVersion) => {
    InfoBox({
      title: `确认是否删除版本【${version.spec.name}?】`,
      infoType: "danger",
      headerAlign: "center" as const,
      footerAlign: "center" as const,
      onConfirm: async () => {
        await deleteScriptVersion(spaceId.value, scriptId.value, version.id)
        if (versionList.value.length === 1 && pagination.value.current > 1) {
          pagination.value.current = pagination.value.current - 1
        }
        getVersionList()
      },
    } as any)
  }

  // 宽窄表视图下选择脚本
  const handleSelectVersion = (version: IScriptVersion) => {
    const { name, memo, content, state } = version.spec
    versionEditData.value = {
      panelOpen: true,
      editable: state === 'not_deployed',
      form: {
        id: version.id,
        name,
        memo,
        content
      }
    }
  }

  // 新建、编辑脚本后回调
  const handleVersionEditDataUpdate = (data: { id: number; name: string; memo: string; content: string; }) => {
    versionEditData.value.form = { ...data }
    refreshList()
  }

  // 版本对比
  const handleVersionDiff = (version: IScriptVersion) => {
    crtVersion.value = version
    showVersionDiff.value = true
  }

  const handleSearchInputChange = (val: string) => {
    if (!val) {
      refreshList()
    }
  }

  const refreshList = () => {
    pagination.value.current = 1
    getVersionList()
  }

  const handlePageLimitChange = (val: number) => {
    pagination.value.limit = val
    refreshList()
  }

  const handleClose = () => {
    router.push({ name: 'script-list', params: { spaceId: spaceId.value } })
  }

</script>
<template>
  <DetailLayout
    :name="`版本管理 - ${scriptDetail.spec.name}`"
    :show-footer="false"
    @close="handleClose">
    <template #content>
      <div class="script-version-manage">
        <div class="operation-area">
          <CreateVersion
            :script-id="scriptId"
            :creatable="!unPublishVersion"
            @create="handleCreateVersionClick"
            @edit="handleEditVersionClick" />
          <bk-input
            v-model.trim="searchStr"
            class="search-input"
            placeholder="版本号/版本说明/更新人"
            :clearable="true"
            @enter="refreshList"
            @clear="refreshList"
            @change="handleSearchInputChange">
              <template #suffix>
                <Search class="search-input-icon" />
              </template>
          </bk-input>
        </div>
        <div :class="['version-data-container', { 'script-panel-open': versionEditData.panelOpen }]">
          <div class="table-data-area">
            <VersionListFullTable
              v-if="!versionEditData.panelOpen"
              :list="versionList"
              :pagination="pagination"
              @view="handleViewVersionClick"
              @page-change="refreshList"
              @page-limit-change="handlePageLimitChange">
              <template #operations="{ data }">
                <div v-if="data.spec" class="action-btns">
                  <bk-button v-if="data.spec.state === 'not_deployed'" text theme="primary" @click="handlePublishClick(data)">上线</bk-button>
                  <bk-button v-if="data.spec.state === 'not_deployed'" text theme="primary" @click="handleEditVersionClick">编辑</bk-button>
                  <bk-button text theme="primary" @click="handleVersionDiff(data)">版本对比</bk-button>
                  <bk-button
                    v-if="data.spec.state !== 'not_deployed'"
                    text
                    theme="primary"
                    :disabled="!!unPublishVersion"
                    @click="handleCreateVersionClick(data.spec.content)">
                    复制并新建
                  </bk-button>
                  <bk-button v-if="data.spec.state === 'not_deployed'" text theme="primary" @click="handleDelClick(data)">删除</bk-button>
              </div>
              </template>
            </VersionListFullTable>
            <template v-else>
              <bk-button
                class="back-table-btn"
                text
                theme="primary"
                @click="versionEditData.panelOpen = false">
                展开列表
                <AngleDoubleRightLine class="arrow-icon" />
              </bk-button>
              <VersionListSimpleTable
                :version-id="versionEditData.form.id"
                :list="versionList"
                :pagination="pagination"
                @select="handleSelectVersion"
                @page-change="refreshList" />
            </template>
          </div>
          <div v-if="versionEditData.panelOpen" class="script-edit-area">
            <VersionEdit
              :type="scriptDetail.spec.type"
              :version-data=versionEditData.form
              :script-id="scriptId"
              :editable="versionEditData.editable"
              @update="handleVersionEditDataUpdate"
              @close="versionEditData.panelOpen = false" />
          </div>
        </div>
      </div>
    </template>
  </DetailLayout>
  <ScriptVersionDiff
    v-if="showVersionDiff"
    v-model:show="showVersionDiff"
    :crt-version="(crtVersion as IScriptVersion)"
    :space-id="spaceId"
    :script-id="scriptId" />
</template>
<style lang="scss" scoped>
  .script-version-manage {
    padding: 24px;
    height: 100%;
    background: #f5f7fa;
  }
  .operation-area {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;
    .search-input {
      width: 320px;
    }
    .search-input-icon {
      padding-right: 10px;
      color: #979ba5;
      background: #ffffff;
    }
  }
  .action-btns {
    .bk-button {
      margin-right: 8px;
    }
  }
  .version-data-container {
    height: calc(100% - 48px);
    border-radius: 2px;
    &.script-panel-open {
      display: flex;
      align-items: flex-start;
      background: #ffffff;
      .table-data-area {
        width: 216px;
        border: 1px solid #dcdee5;
      }
    }
    .table-data-area {
      position: relative;
      height: 100%;
      .back-table-btn {
        position: absolute;
        top: 18px;
        right: 10px;
        font-size: 12px;
        z-index: 1;
        .arrow-icon {
          margin-left: 4px;
        }
      }
    }
    .script-edit-area {
      width: calc(100% - 216px);
      height: 100%;
    }
  }
</style>