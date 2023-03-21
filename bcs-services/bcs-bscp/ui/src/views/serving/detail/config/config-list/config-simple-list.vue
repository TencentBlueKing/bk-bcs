<script setup lang="ts">
  import { ref, watch, onMounted } from 'vue'
  import { storeToRefs } from 'pinia'
  import { useConfigStore } from '../../../../../store/config'
  import { IConfigVersionItem } from '../../../../../types'
  import { IConfigListQueryParams } from '../../../../../../types/config'
  import { getConfigList } from '../../../../../api/config'
  import EditConfig from './config-table-list/edit-config.vue'

  const store = useConfigStore()
  const { versionData } = storeToRefs(store)

  const props = defineProps<{
    bkBizId: string,
    appId: number,
  }>()

  const loading = ref(false)
  const configList = ref<Array<IConfigVersionItem>>([])
  const editDialogShow = ref(false)

  watch(() => versionData.value.id, () => {
    getListData()
  } )

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
      const res = await getConfigList(props.appId, params)
      configList.value = res.details
    } catch (e) {
      console.error(e)
    } finally {
      loading.value = false
    }
  }

  const handleEditConfirm = () => {
    editDialogShow.value = false
  }

</script>
<template>
  <section class="current-config-list">
    <bk-loading :loading="loading">
      <h4 class="version-name">{{ versionData.spec.name }}</h4>
      <div class="config-list-wrapper">
        <div v-for="config in configList" class="config-item" :key="config.id" @click="editDialogShow = true">
          <div class="config-name">{{ config.spec.name }}</div>
          <div class="config-type">二进制文件</div>
        </div>
      </div>
    </bk-loading>
    <EditConfig
      v-model:show="editDialogShow"
      :bk-biz-id="props.bkBizId"
      :config-id="props.appId"
      :app-id="props.appId" />
  </section>
</template>
<style lang="scss" scoped>
  .current-config-list {
    padding: 24px;
    height: 100%;
    background: #fafbfd;
  }
  .version-name {
    margin: 0 0 16px 0;
    font-size: 14px;
    color: #63656e;
    font-weight: 700;
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
