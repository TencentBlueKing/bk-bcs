<script setup lang="ts">
  import { ref, watch, onMounted } from 'vue'
  import { storeToRefs } from 'pinia'
  import { useConfigStore } from '../../../../../../store/config'
  import { IConfigItem, IConfigListQueryParams } from '../../../../../../../types/config'
  import { getConfigList } from '../../../../../../api/config'
  import { getConfigTypeName } from '../../../../../../utils/config'
  import EditConfig from './config-table-list/edit-config.vue'

  const store = useConfigStore()
  const { versionData } = storeToRefs(store)

  const props = defineProps<{
    bkBizId: string,
    appId: number,
  }>()

  const loading = ref(false)
  const configList = ref<Array<IConfigItem>>([])
  const configId = ref(0)
  const editDialogShow = ref(false)

  watch(() => versionData.value.id, () => {
    getListData()
  })

  onMounted(() => {
    getListData()
  })

  const getListData = async () => {
    // 拉取到版本列表之前不加在列表数据
    if (typeof versionData.value.id !== 'number') {
      return
    }

    loading.value = true
    try {
      const params: IConfigListQueryParams = {
        start: 0,
        limit: 200 // @todo 分页条数待确认
      }
      if (versionData.value.id !== 0) {
        params.release_id = <number>versionData.value.id
      }
      const res = await getConfigList(props.bkBizId, props.appId, params)
      configList.value = res.details
    } catch (e) {
      console.error(e)
    } finally {
      loading.value = false
    }
  }

  const handleEditConfigOpen = (id: number) => {
    editDialogShow.value = true
    configId.value = id
  }

</script>
<template>
  <section class="current-config-list">
    <bk-loading :loading="loading">
      <div v-if="configList.length > 0" class="config-list-wrapper">
        <div v-for="config in configList" class="config-item" :key="config.id" @click="handleEditConfigOpen(config.id)">
          <div class="config-name">{{ config.spec.name }}</div>
          <div class="config-type">{{ getConfigTypeName(config.spec.file_type) }}</div>
        </div>
      </div>
      <bk-exception v-else scene="part" type="empty" description="暂无数据"></bk-exception>
    </bk-loading>
    <EditConfig
      v-model:show="editDialogShow"
      :bk-biz-id="props.bkBizId"
      :app-id="props.appId"
      :config-id="configId" />
  </section>
</template>
<style lang="scss" scoped>
  .current-config-list {
    padding: 24px;
    height: 100%;
    background: #fafbfd;
    overflow: auto;
  }
  .config-item {
    display: flex;
    align-items: center;
    margin-bottom: 8px;
    font-size: 12px;
    background: #ffffff;
    box-shadow: 0 1px 1px 0 rgba(0, 0, 0, 0.06);
    border-radius: 2px;
    cursor: pointer;
    &:hover {
      background: #e1ecff;
    }
    .config-name {
      padding: 0 16px;
      width: 242px;
      height: 40px;
      line-height: 40px;
      color: #313238;
      white-space: nowrap;
      text-overflow: ellipsis;
      overflow: hidden;
    }
    .config-type {
      color: #979ba5;
    }
  }
</style>
