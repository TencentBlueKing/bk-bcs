<script setup lang="ts">
  import { ref, onMounted, defineExpose } from 'vue'
  import { storeToRefs } from 'pinia'
  import { useConfigStore } from '../../../../../store/config'
  import { Ellipsis } from 'bkui-vue/lib/icon'
  import { InfoBox } from "bkui-vue/lib";
  import { getConfigVersionList } from '../../../../../api/config'
  import { IConfigVersion } from '../../../../../../types/config'
  import VersionDiff from '../components/version-diff/index.vue'

  const versionStore = useConfigStore()
  const { versionData } = storeToRefs(versionStore)

  const props = defineProps<{
    bkBizId: string,
    appId: number,
  }>()

  const emits = defineEmits(['loaded'])

  const currentConfig: IConfigVersion = {
    id: 0,
    attachment: {
      app_id: 0,
      biz_id: 0
    },
    revision: {
      create_at: '',
      creator: ''
    },
    spec: {
      name: '未命名版本',
      memo: ''
    },
    status: {
      publish_status: 'editing',
      released_groups: []
    }
  }
  const versionListLoading = ref(false)
  const versionList = ref<IConfigVersion[]>([])
  const showDiffPanel = ref(false)
  const diffVersion = ref()
  const pagination = ref({
    current: 1,
    limit: 10,
    count: 0
  })

  onMounted(async() => {
    await getVersionList()
    // 默认选中未命名版本
    handleSelectVersion(currentConfig)
  })

  const getVersionList = async() => {
    try {
      versionListLoading.value = true
      const params = {
        start: pagination.value.current === 1 ? 0 : (pagination.value.current - 1) * pagination.value.limit - 1,
        limit: pagination.value.current === 1 ? pagination.value.limit - 1 : pagination.value.limit,
      }
      const res = await getConfigVersionList(props.appId, params)
      if (pagination.value.current === 1) {
        versionList.value = [currentConfig, ...res.data.details]
      } else {
        versionList.value = res.data.details
      }
      pagination.value.count = res.data.count + 1
      emits('loaded')
    } catch (e) {
      console.error(e)
    } finally {
      versionListLoading.value = false
    }
  }

  const handleSelectVersion = (version: IConfigVersion) => {
    versionStore.$patch(state => {
      state.versionData = version
    })
  }

  // @todo 切换页码时，组件会调用两次change事件，待确认
  const handlePageChange = (val: number) => {
    pagination.value.current = val
    getVersionList()
  }

  const handleDiffDialogShow = (version: IConfigVersion) => {
    diffVersion.value = version
    showDiffPanel.value = true
  }

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

  defineExpose({
    getVersionList
  })

</script>
<template>
  <section class="version-container">
    <bk-loading :loading="versionListLoading">
      <section class="versions-wrapper">
        <section
          v-for="(version, index) in versionList"
          :key="version.id"
          :class="['version-item', { active: versionData.id === version.id }]"
          @click="handleSelectVersion(version)">
          <div :class="['dot', version.status.publish_status]"></div>
          <div class="version-name">{{ version.spec.name }}</div>
          <bk-dropdown class="action-area">
            <Ellipsis class="action-more-icon" />
            <template #content>
              <bk-dropdown-menu placement="bottom-end">
                <bk-dropdown-item @click="handleDiffDialogShow(version)">版本对比</bk-dropdown-item>
                <bk-dropdown-item @click="handleDeprecate(version.id)">废弃</bk-dropdown-item>
              </bk-dropdown-menu>
            </template>
          </bk-dropdown>
        </section>
      </section>
    </bk-loading>
      <bk-pagination
          class="list-pagination"
          v-model="pagination.current"
          small
          align="right"
          :show-limit="false"
          :show-total-count="false"
          :count="pagination.count"
          :limit="pagination.limit"
          @change="handlePageChange"/>
    <VersionDiff
      v-model:show="showDiffPanel"
      :current-version="diffVersion" />
  </section>
</template>

<style lang="scss" scoped>
  .version-container {
    padding: 16px 0;
    height: 100%;
  }
  .versions-wrapper {
    overflow: auto;
  }
  .version-steps {
    padding: 16px 0;
    overflow: auto;
  }
  .version-item {
    position: relative;
    padding: 0 40px 0 48px;
    cursor: pointer;
    &.active {
      background: #e1ecff;
    }
    &:hover {
      background: #e1ecff;
    }
    .dot {
      position: absolute;
      left: 28px;
      top: 16px;
      width: 8px;
      height: 8px;
      border-radius: 50%;
      border: 1px solid #c4c6cc;
      background: #f0f1f5;
      &.not_released {
        border: 1px solid #ff9c01;
        background: #ffe8c3;
      }
      &.full_released,
      &.partial_released {
        border: 1px solid #3fc06d;
        background: #e5f6ea;
      }
    }
  }
  .version-name {
    height: 40px;
    line-height: 40px;
    font-size: 12px;
    color: #313238;
    text-align: left;
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
  }
  .action-area {
    position: absolute;
    right: 13px;
    top: 9px;
    line-height: 22px;
    .action-more-icon {
      display: flex;
      align-items: center;
      justify-content: center;
      transform: rotate(90deg);
      width: 22px;
      height: 22px;
      color: #979ba5;
      border-radius: 50%;
      cursor: pointer;
      &:hover {
        background: rgba(99, 101, 110, 0.1);
        color: #3a84ff;
      }
    }
  }
  .list-pagination {
    margin-top: 16px;
  }
</style>