<script setup lang="ts">
  import { ref, computed, watch, onMounted, defineExpose } from 'vue'
  import { useStore } from 'vuex'
  import { Ellipsis, ArrowsLeft, AngleRight } from 'bkui-vue/lib/icon'
  import InfoBox from "bkui-vue/lib/info-box";
  import { getConfigVersionList } from '../../../../api/config'
  import { IVersionItem, FilterOp } from '../../../../types'
  import VersionLayout from './components/version-layout.vue'
  import ConfigDiff from './components/config-diff.vue';

  const store = useStore()

  const emit = defineEmits(['updateReleaseId'])

  const props = defineProps<{
    bkBizId: string,
    appId: number,
    releaseId: number|null
  }>()

  const currentConfig: IVersionItem = {
    id: 0,
    attachment: {},
    revision: {},
    spec: {
      name: '未命名版本'
    }
  }
  const appName = store.getters['config/appName']
  const versionListLoading = ref(false)
  const versionList = ref<IVersionItem[]>([])
  const showDiffPanel = ref(false)
  const diffVersion = ref()
  const filter = { op: FilterOp.AND, rules: [] }
  const page = {
    count: false,
    start: 0,
    limit: 200 // @todo 分页条数待确认
  }

  const listData = computed(() => {
    return [currentConfig, ...versionList.value]
  })


  watch(() => props.appId, () => {
    getVersionList()
  })

  onMounted(() => {
    getVersionList()
  })

  const getVersionList = async() => {
    try {
      versionListLoading.value = true
      const res = await getConfigVersionList(props.bkBizId, props.appId, filter, page)
      versionList.value = res.data.details
      handleSelectVersion(currentConfig)
    } catch (e) {
      console.error(e)
    } finally {
      versionListLoading.value = false
    }
  }

  const handleSelectVersion = (version: IVersionItem) => {
    store.commit('config/setCurrentVersion', version)
    emit('updateReleaseId', version.id)
  }

  const handleDiffDialogShow = (version: IVersionItem) => {
    console.log(version)
    diffVersion.value = version
    showDiffPanel.value = true
  }

  const handleDiffClose = () => {
    showDiffPanel.value = false
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
      <section
        v-for="(version, index) in listData"
        :key="version.id"
        :class="['version-item', { active: props.releaseId === version.id }]"
        @click="handleSelectVersion(version)">
        <div class="dot-line">
          <div :class="['dot', { first: index === 0, last: index === listData.length - 1 }]"></div>
        </div>
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
    </bk-loading>
    <VersionLayout v-if="showDiffPanel" :show-footer="false">
      <template #header>
        <section class="header-wrapper">
          <span class="header-name" @click="handleDiffClose">
            <ArrowsLeft class="arrow-left" />
            <span class="service-name">{{ appName }}</span>
          </span>
          <AngleRight class="arrow-right" />
          版本对比
        </section>
      </template>
      <config-diff :config-list="[]">
        <template #head>
          <div class="diff-left-panel-head">
            {{ diffVersion.spec.name }}
            <!-- @todo 待确定这里展示什么名称 -->
          </div>
        </template>
      </config-diff>
    </VersionLayout>
  </section>
</template>

<style lang="scss" scoped>
  .version-container {
    padding: 16px 0;
    height: 100%;
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
    .dot-line {
      position: absolute;
      left: 28px;
      top: 0;
      display: flex;
      align-items: center;
      height: 100%;
      .dot {
        width: 8px;
        height: 8px;
        border-radius: 50%;
        border: 1px solid #c4c6cc;
        background: #f0f1f5;
        &:not(.first):before {
          position: absolute;
          top: 0;
          left: 4px;
          content: '';
          width: 1px;
          height: 16px;
          background: #dcdee5;
        }
        &:not(.last):after {
          position: absolute;
          bottom: 0;
          left: 4px;
          content: '';
          width: 1px;
          height: 16px;
          background: #dcdee5;
        }
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
  .header-wrapper {
    display: flex;
    align-items: center;
    padding: 0 24px;
    height: 100%;
    font-size: 12px;
    line-height: 1;
  }
  .header-name {
    display: flex;
    align-items: center;
    font-size: 12px;
    color: #3a84ff;
    cursor: pointer;
  }
  .arrow-left {
    font-size: 26px;
    color: #3884ff;
  }
  .arrow-right {
    font-size: 24px;
    color: #c4c6cc;
  }
  .diff-left-panel-head {
    padding-left: 16px;
    font-size: 12px;
    color: #313238;
  }
</style>