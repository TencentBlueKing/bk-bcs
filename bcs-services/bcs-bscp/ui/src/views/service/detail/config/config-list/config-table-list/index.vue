<script setup lang="ts">
  import { ref, watch, onMounted } from 'vue'
  import { storeToRefs } from 'pinia'
  import { useServiceStore } from '../../../../../../store/service'
  import { useConfigStore } from '../../../../../../store/config'
  import InfoBox from "bkui-vue/lib/info-box";
  import { IConfigItem, IConfigListQueryParams } from '../../../../../../../types/config'
  import { CONFIG_STATUS_MAP } from '../../../../../../constants/index'
  import { getConfigList, deleteServiceConfigItem } from '../../../../../../api/config'
  import { getConfigTypeName } from '../../../../../../utils/config'
  import EditConfig from './edit-config.vue'
  import CreateConfig from './create-config.vue'
  import PublishVersion from './publish-version/index.vue'
  import ReleaseVersion from './release-version/index.vue'
  import ModifyGroup from './modify-group.vue'
  import VersionDiff from '../../components/version-diff/index.vue'

  const serviceStore = useServiceStore()
  const versionStore = useConfigStore()
  const { appData } = storeToRefs(serviceStore)
  const { versionData } = storeToRefs(versionStore)

  const emit = defineEmits(['updateVersionList'])

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
      const res = await getConfigList(props.appId, params)
      // @ts-ignore
      configList.value = res.details
      pagination.value.count = 4
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

  const handlePageLimitChange = (limit: number) => {
    pagination.value.limit = limit
    refreshConfigList()
  }

  const handleUpdateStatus = () => {
    emit('updateVersionList')
  }

  defineExpose({
    refreshConfigList
  })
</script>
<template>
  <section class="config-list-wrapper">
    <section class="config-content-header">
      <section class="summary-wrapper">
        <div class="status-tag">编辑中</div>
        <div class="version-name">未命名版本</div>
      </section>
      <section class="actions-wrapper">
        <ReleaseVersion
          v-if="versionData.id === 0"
          :bk-biz-id="props.bkBizId"
          :app-id="props.appId"
          @confirm="handleUpdateStatus" />
        <PublishVersion
          style="margin-left: 8px"
          :bk-biz-id="props.bkBizId"
          :app-id="props.appId"
          :release-id="versionData.id"
          :app-name="appData.spec.name"
          :version-name="versionData.spec.name"
          :config-list="configList"
          @confirm="handleUpdateStatus" />
        <ModifyGroup
          style="margin-left: 8px"
          :bk-biz-id="props.bkBizId"
          :app-id="props.appId"
          :release-id="versionData.id"
          @confirm="handleUpdateStatus" />
      </section>
    </section>
    <CreateConfig v-if="versionData.id === 0" :bk-biz-id="props.bkBizId" :app-id="props.appId" @confirm="refreshConfigList" />
    <section class="config-list-table">
      <bk-loading :loading="loading">
        <bk-table v-if="!loading" :border="['outer']" :data="configList">
          <bk-table-column label="配置项名称" prop="spec.name" :sort="true"></bk-table-column>
          <bk-table-column label="配置格式">
            <template #default="{ row }">
              {{ getConfigTypeName(row.spec?.file_type) }}
            </template>
          </bk-table-column>
          <bk-table-column label="创建人" prop="revision.creator"></bk-table-column>
          <bk-table-column label="修改人" prop="revision.reviser"></bk-table-column>
          <bk-table-column label="修改时间" prop="revision.update_at" :sort="true"></bk-table-column>
          <bk-table-column v-if="versionData.id === 0" label="变更状态">
            <template #default="{ row }">
                <span v-if="row.file_state" :class="['status', row.file_state.toLowerCase()]">
                  {{ CONFIG_STATUS_MAP[row.file_state as keyof typeof CONFIG_STATUS_MAP] }}
                </span>
            </template>
          </bk-table-column>
          <bk-table-column label="操作">
            <template #default="{ row }">
              <div class="operate-action-btns">
                <bk-button text theme="primary" @click="handleEdit(row)">{{ versionData.id === 0 ? '编辑' : '查看' }}</bk-button>
                <bk-button text theme="primary" @click="handleDiff(row)">对比</bk-button>
                <bk-button v-if="versionData.id === 0" text theme="primary" @click="handleDel(row)">删除</bk-button>
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
        @confirm="refreshConfigList" />
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
    border-bottom: 1px solid #dcdee5;
    .actions-wrapper {
      display: flex;
      align-items: center;
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
    }
    .version-name {
      color: #63656e;
      font-size: 14px;
      font-weight: bold;
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
