<script setup lang="ts">
  import { ref, watch, onMounted, nextTick } from 'vue'
  import { storeToRefs } from 'pinia'
  import { useServiceStore } from '../../../../../../../store/service'
  import { useConfigStore } from '../../../../../../../store/config'
  import { InfoBox } from "bkui-vue/lib";
  import { InfoLine, Search } from 'bkui-vue/lib/icon';
  import { IConfigItem, IConfigListQueryParams } from '../../../../../../../../types/config'
  import { CONFIG_STATUS_MAP, VERSION_STATUS_MAP } from '../../../../../../../constants/config'
  import { getConfigList, deleteServiceConfigItem } from '../../../../../../../api/config'
  import { getConfigTypeName } from '../../../../../../../utils/config'
  import EditConfig from './edit-config.vue'
  import InitScript from './init-script/index.vue'
  import CreateConfig from './create-config.vue'
  import PublishVersion from './publish-version/index.vue'
  import CreateVersion from './create-version/index.vue'
  import ModifyGroupPublish from './modify-group-publish.vue'
  import VersionDiff from '../../components/version-diff/index.vue'

  const serviceStore = useServiceStore()
  const configStore = useConfigStore()
  const { appData } = storeToRefs(serviceStore)
  const { versionData } = storeToRefs(configStore)

  const props = defineProps<{
    bkBizId: string,
    appId: number,
  }>()

  const loading = ref(false)
  const configList = ref<IConfigItem[]>([])
  const pagination = ref({
    current: 1,
    count: 0,
    limit: 10,
  })
  const editPanelShow = ref(false)
  const activeConfig = ref(0)
  const isDiffPanelShow = ref(false)
  const diffConfig = ref(0)
  const publishVersionRef = ref()

  watch(() => versionData.value.id, () => {
    getListData()
  })

  onMounted(() => {
    getListData()
  })

  const getListData = async () => {
    loading.value = true
    try {
      const params: IConfigListQueryParams = {
        start: (pagination.value.current - 1) * pagination.value.limit,
        limit: pagination.value.limit
      }
      if (versionData.value.id !== 0) {
        params.release_id = versionData.value.id
      }
      const res = await getConfigList(props.bkBizId, props.appId, params)
      // @ts-ignore
      configList.value = res.details
      pagination.value.count = res.count
    } catch (e) {
      console.error(e)
    } finally {
      loading.value = false
    }
  }

  const refreshConfigList = (current: number = 1) => {
    pagination.value.current = current
    getListData()
  }

  const handleEdit = (config: IConfigItem) => {
    activeConfig.value = config.id
    editPanelShow.value = true
  }

  const handleDiff = (config: IConfigItem) => {
    diffConfig.value = config.id
    isDiffPanelShow.value = true
  }

  const handleDel = (config: IConfigItem) => {
    InfoBox({
      title: `确认是否删除配置项 ${config.spec.name}?`,
      type: "danger",
      headerAlign: "center" as const,
      footerAlign: "center" as const,
      onConfirm: async () => {
        await deleteServiceConfigItem(config.id, config.attachment.biz_id, config.attachment.app_id)
        if (configList.value.length === 1 && pagination.value.current > 1) {
          pagination.value.current -= 1
        }
        getListData();
      },
    } as any);
  }

  // 创建版本成功后，刷新版本列表，若选择同时上线，则打开选择分组面板
  const handleVersionCreated = (versionId: number, isPublish: boolean) => {
    versionData.value.id = versionId
    refreshVesionList()
    nextTick(() => {
      if (isPublish && publishVersionRef.value) {
        publishVersionRef.value.handleOpenSelectGroupPanel()
      }
    })
  }

  const handlePageLimitChange = (limit: number) => {
    pagination.value.limit = limit
    refreshConfigList()
  }

  const refreshVesionList = () => {
    configStore.$patch((state) => {
      state.refreshVersionListFlag = true
    })
  }

  defineExpose({
    refreshConfigList
  })
</script>
<template>
  <section class="config-list-wrapper">
    <section class="config-content-header">
      <section class="summary-wrapper">
        <div :class="['status-tag', versionData.status.publish_status]">{{ VERSION_STATUS_MAP[versionData.status.publish_status as keyof typeof VERSION_STATUS_MAP] || '编辑中' }}</div>
        <div class="version-name">{{ versionData.spec.name }}</div>
        <InfoLine
          v-if="versionData.spec.memo"
          v-bk-tooltips="{
            content: versionData.spec.memo,
            placement: 'right',
            theme: 'light'
          }"
          class="version-desc" />
      </section>
      <section class="version-operations">
        <InitScript :app-id="props.appId" />
        <CreateVersion
          v-if="versionData.status.publish_status === 'editing'"
          ref="publishVersionRef"
          :bk-biz-id="props.bkBizId"
          :app-id="props.appId"
          :config-count="pagination.count"
          @confirm="handleVersionCreated" />
        <PublishVersion
          v-if="versionData.status.publish_status === 'not_released'"
          :bk-biz-id="props.bkBizId"
          :app-id="props.appId"
          :release-id="versionData.id"
          :app-name="appData.spec.name"
          :version-name="versionData.spec.name"
          :config-list="configList"
          @confirm="refreshVesionList" />
        <ModifyGroupPublish
          v-if="versionData.status.publish_status === 'partial_released'"
          :bk-biz-id="props.bkBizId"
          :app-id="props.appId"
          :release-id="versionData.id"
          :app-name="appData.spec.name"
          :version-name="versionData.spec.name"
          :config-list="configList"
          @confirm="refreshVesionList" />
      </section>
    </section>
    <div class="operate-area">
      <CreateConfig v-if="versionData.status.publish_status === 'editing'" :bk-biz-id="props.bkBizId" :app-id="props.appId" @confirm="refreshConfigList" />
      <div class="groups-info" v-if="versionData.status.released_groups.length > 0">
        <div v-for="group in versionData.status.released_groups" class="group-item" :key="group.id">
          {{ group.name }}
        </div>
      </div>
      <bk-input class="search-config-input" placeholder="配置项名称/创建人/修改人">
        <template #suffix>
            <Search class="search-input-icon" />
        </template>
      </bk-input>
    </div>
    <section class="config-list-table">
      <bk-loading :loading="loading">
        <bk-table v-if="!loading" :border="['outer']" :data="configList">
          <bk-table-column label="配置项名称" prop="spec.name" :sort="true" :min-width="240"></bk-table-column>
          <bk-table-column label="配置格式">
            <template #default="{ row }">
              {{ getConfigTypeName(row.spec?.file_type) }}
            </template>
          </bk-table-column>
          <bk-table-column label="配置路径" prop="spec.path"></bk-table-column>
          <bk-table-column label="创建人" prop="revision.creator"></bk-table-column>
          <bk-table-column label="修改人" prop="revision.reviser"></bk-table-column>
          <bk-table-column label="修改时间" prop="revision.update_at" :sort="true" :width="180"></bk-table-column>
          <bk-table-column v-if="versionData.id === 0" label="变更状态">
            <template #default="{ row }">
                <span v-if="row.file_state" :class="['status', row.file_state.toLowerCase()]">
                  {{ CONFIG_STATUS_MAP[row.file_state as keyof typeof CONFIG_STATUS_MAP] }}
                </span>
            </template>
          </bk-table-column>
          <bk-table-column label="操作" fixed="right">
            <template #default="{ row }">
              <div class="operate-action-btns">
                <bk-button :disabled="row.file_state === 'DELETE'" text theme="primary" @click="handleEdit(row)">{{ versionData.id === 0 ? '编辑' : '查看' }}</bk-button>
                <bk-button v-if="versionData.status.publish_status !== 'editing'" text theme="primary" @click="handleDiff(row)">对比</bk-button>
                <bk-button v-if="versionData.id === 0" text theme="primary" :disabled="row.file_state === 'DELETE'" @click="handleDel(row)">删除</bk-button>
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
          @change="refreshConfigList($event)"
          @limit-change="handlePageLimitChange"/>
      </bk-loading>
      <edit-config
        v-model:show="editPanelShow"
        :config-id="activeConfig"
        :bk-biz-id="props.bkBizId"
        :app-id="props.appId"
        @confirm="getListData" />
    </section>
    <VersionDiff
      v-model:show="isDiffPanelShow"
      :current-version="versionData"
      :current-config="diffConfig" />
  </section>
</template>
<style lang="scss" scoped>
  .config-list-wrapper {
    padding: 0 24px;
  }
  .config-content-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    height: 64px;
    .version-operations {
      display: flex;
      align-items: center;
    }
  }
  .operate-area {
    display: flex;
    align-items: center;
    padding: 16px 0;
    border-top: 1px solid #dcdee5;
    .groups-info {
      display: flex;
      align-items: center;
      .group-item {
        padding: 0 8px;
        line-height: 22px;
        color: #63656e;
        font-size: 12px;
        background: #f0f1f5;
        border-radius: 2px;
        &:not(:last-of-type) {
          margin-right: 8px;
        }
      }
    }
    .search-config-input {
      margin-left: auto;
      width: 280px;
    }
    .search-input-icon {
      padding-right: 10px;
      color: #979ba5;
      background: #ffffff;
    }
  }
  .summary-wrapper {
    display: flex;
    align-items: center;
    justify-content: space-between;
    .status-tag {
      margin-right: 8px;
      padding: 0 10px;
      height: 22px;
      line-height: 20px;
      font-size: 12px;
      color: #63656e;
      border: 1px solid rgba(151,155,165,0.30);
      border-radius: 11px;
      &.not_released {
      color: #fe9000;
      background: #ffe8c3;
      border-color: rgba(254, 156, 0, 0.3);
      }
    &.full_released,
    &.partial_released {
      color: #14a568;
      background: #e4faf0;
      border-color: rgba(20, 165, 104, 0.3);
    }
    }
    .version-name {
      color: #63656e;
      font-size: 14px;
      font-weight: bold;
    }
    .version-desc {
      margin-left: 8px;
      font-size: 15px;
      color: #979ba5;
      cursor: pointer;
    }
  }
  .config-list-table {
    :deep(.bk-pagination) {
      padding-left: 15px;
    }
    .status{
      &:not(.nochange) {
        padding: 4px 10px;
        border-radius: 2px;
        font-size: 12px;
      }
      &.add {
        background: #edf4ff;
        color: #3a84ff;
      }
      &.delete {
        background: #feebea;
        color: #ea3536;
      }
      &.revise {
        background: #fff1db;
        color: #fe9c00;
      }
    }
  }
  .operate-action-btns {
    .bk-button:not(:last-of-type) {
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
