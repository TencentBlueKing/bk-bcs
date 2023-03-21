<script setup lang="ts">
  import { ref, watch, onMounted } from 'vue'
  import { storeToRefs } from 'pinia'
  import { useServingStore } from '../../../../../../store/serving'
  import { useConfigStore } from '../../../../../../store/config'
  import InfoBox from "bkui-vue/lib/info-box";
  import { IConfigItem, IConfigListQueryParams } from '../../../../../../../types/config'
  import { getConfigList, deleteServingConfigItem } from '../../../../../../api/config'
  import EditConfig from './edit-config.vue'
  import CreateConfig from './create-config.vue'
  import PublishVersion from './publish-version/index.vue'
  import ReleaseVersion from './release-version/index.vue'
  import ModifyGroup from './modify-group.vue'
  import VersionDiffDialog from '../../components/version-diff-dialog.vue';

  const servingStore = useServingStore()
  const versionStore = useConfigStore()
  const { appData } = storeToRefs(servingStore)
  const { versionData } = storeToRefs(versionStore)

  const emit = defineEmits(['updateVersionList'])

  const props = defineProps<{
    bkBizId: string,
    appId: number,
  }>()

  const loading = ref(false)
  const configList = ref<Array<IConfigItem>>([])
  const pagination = ref({
    current: 1,
    count: 0,
    limit: 10,
  })
  const editPanelShow = ref(false)
  const activeConfig = ref(0)
  const isDiffDialogShow = ref(false)
  const diffConfig = ref()

  watch(() => props.appId, () => {
    getListData()
  })

  watch(() => versionData.value.id, () => {
    getListData()
  })

  onMounted(() => {
    getListData()
  })

  const getListData = async () => {
    // 拉取到版本列表之前不加在列表数据
    if (typeof versionData.value.id !== 'number' || versionData.value.id === 0) {
      return
    }

    loading.value = true
    try {
      const params: IConfigListQueryParams = {
        start: 0,
        limit: 200 // @todo 分页条数待确认
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
    console.log(config)
    diffConfig.value = config
    isDiffDialogShow.value = true
  }

  const handleDel = (config: IConfigItem) => {
    InfoBox({
      title: `确认是否删除配置项 ${config.spec.name}?`,
      type: "danger",
      headerAlign: "center" as const,
      footerAlign: "center" as const,
      onConfirm: async () => {
        await deleteServingConfigItem(config.id, config.attachment.biz_id, config.attachment.app_id)
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
          :app-name="appData.spec.name"
          :config-list="configList"
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
        <bk-table :border="['outer']" :data="configList">
          <bk-table-column label="配置项名称" prop="spec.name" :sort="true"></bk-table-column>
          <bk-table-column label="配置预览">-</bk-table-column>
          <bk-table-column label="配置格式" prop="spec.file_type"></bk-table-column>
          <bk-table-column label="创建人" prop="revision.creator"></bk-table-column>
          <bk-table-column label="修改人" prop="revision.reviser"></bk-table-column>
          <bk-table-column label="修改时间" prop="revision.update_at" :sort="true"></bk-table-column>
          <bk-table-column label="变更状态">-</bk-table-column>
          <bk-table-column label="操作">
            <template #default="{ row }">
              <div class="operate-action-btns">
                <bk-button text theme="primary" @click="handleEdit(row)">编辑</bk-button>
                <bk-button text theme="primary" @click="handleDiff(row)">对比</bk-button>
                <bk-button text theme="primary" @click="handleDel(row)">删除</bk-button>
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
    <VersionDiffDialog v-model:show="isDiffDialogShow" version-name="未命名版本" :config="diffConfig" />
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
