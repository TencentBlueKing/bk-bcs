<script setup lang="ts">
  import { ref, computed, watch, onMounted } from 'vue'
  import { useStore } from 'vuex'
  import { IVersionItem, IConfigVersionItem } from '../../../../types'
  import { IConfigListQueryParams } from '../../../../../types/config'
  import { getConfigList } from '../../../../api/config'

  const store = useStore()

  const props = defineProps<{
    bkBizId: string,
    appId: number,
  }>()

  const loading = ref(false)
  const configList = ref<Array<IConfigVersionItem>>([])

  const currentVersion = computed((): IVersionItem => {
    return store.state.config.currentVersion
  })

  const versionName = computed(() => {
    return store.state.config.currentVersion.spec?.name || ''
  })

  watch(() => currentVersion.value.id, () => {
    getListData()
  } )

  onMounted(() => {
    getListData()
  })

  const getListData = async () => {
    // 拉取到版本列表之前不加在列表数据
    if (typeof currentVersion.value.id !== 'number') {
      return
    }

    loading.value = true
    try {
      const params: IConfigListQueryParams = {
        start: 0,
        limit: 200 // @todo 分页条数待确认
      }
      if (currentVersion.value.id !== 0) {
        params.release_id = <number>currentVersion.value.id
      }
      const res = await getConfigList(props.appId, params)
      configList.value = res.details
    } catch (e) {
      console.error(e)
    } finally {
      loading.value = false
    }
  }

</script>
<template>
  <section class="current-config-list">
    <bk-loading :loading="loading">
      <h4 class="version-name">{{ versionName }}</h4>
      <div class="config-list-wrapper">
        <div v-for="config in configList" class="config-item" :key="config.id">
          <div class="config-name">{{ config.spec.name }}</div>
          <div class="config-type">二进制文件</div>
        </div>
      </div>
    </bk-loading>
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
