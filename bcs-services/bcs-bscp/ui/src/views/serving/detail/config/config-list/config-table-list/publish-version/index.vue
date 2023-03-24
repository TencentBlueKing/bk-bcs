<script setup lang="ts">
  import { ref, watch } from 'vue'
  import { ArrowsLeft, AngleRight } from 'bkui-vue/lib/icon'
  import InfoBox from "bkui-vue/lib/info-box";
  import VersionLayout from '../../../components/version-layout.vue'
  import ConfirmDialog from './confirm-dialog.vue'
  import ConfigDiff from '../../../components/config-diff.vue'
  import { getConfigVersionList, getConfigList } from '../../../../../../../api/config'
  import { FilterOp, RuleOp } from '../../../../../../../types'
  import { IConfigItem, IConfigVersion, IConfigListQueryParams } from '../../../../../../../../types/config'

  const props = defineProps<{
    bkBizId: string,
    appId: number,
    releaseId: number,
    appName: string,
    versionName: string,
    configList: IConfigItem[]
  }>()

  const emit = defineEmits(['confirm'])

  const showDiffPanel = ref(false)
  const isConfirmDialogShow = ref(false)
  const groups = ref([])
  const versionListLoading = ref(true)
  const versionList = ref<IConfigVersion[]>([])
  const selectedVersion = ref<number>()
  const baseConfigList = ref<IConfigItem[]>([])
  const baseConfigLoading = ref(false)
  const filter = {
    op: FilterOp.AND,
    rules: [{
      field: "deprecated",
      op: RuleOp.eq,
      value: false
    }]
  }
  const page = {
    count: false,
    start: 0,
    limit: 200 // @todo 分页条数待确认
  }

  watch(showDiffPanel, (val) => {
    if (val) {
      getVersionList()
    }
  })

  // 获取所有版本
  const getVersionList = async() => {
    try {
      versionListLoading.value = true
      const res = await getConfigVersionList(props.bkBizId, props.appId, filter, page)
      versionList.value = res.data.details.filter((item: IConfigVersion) => item.id !== props.releaseId)
    } catch (e) {
      console.error(e)
    } finally {
      versionListLoading.value = false
    }
  }

  // 获取某个版本下配置项列表
  const getConfigsForVersion = async () => {
    baseConfigLoading.value = true
    try {
      const params: IConfigListQueryParams = {
        release_id: selectedVersion.value,
        start: 0,
        limit: 200 // @todo 分页条数待确认
      }

      const res = await getConfigList(props.appId, params)
      baseConfigList.value = res.details
    } catch (e) {
      console.error(e)
    } finally {
      baseConfigLoading.value = false
    }
  }

  const handleSelectVersion = (val: number) => {
    selectedVersion.value = val
    getConfigsForVersion()
  }

  const handleConfirm = () => {
    InfoBox({
    // @ts-ignore
      infoType: "success",
      title: '版本已上线',
      dialogType: 'confirm',
      onConfirm () {
        emit('confirm')
        handleClose()
      }
    })
  }

  const handleClose = () => {
    showDiffPanel.value = false
  }

</script>
<template>
    <section class="create-version">
        <bk-button theme="primary" :disabled="props.releaseId === 0"  @click="showDiffPanel = true">上线版本</bk-button>
        <VersionLayout v-if="showDiffPanel">
            <template #header>
                <section class="header-wrapper">
                    <span class="header-name" @click="handleClose">
                        <ArrowsLeft class="arrow-left" />
                        <span class="service-name">{{ props.appName }}</span>
                    </span>
                    <AngleRight class="arrow-right" />
                    上线版本：{{ props.versionName }}
                </section>
            </template>
            <config-diff
                :current-version="props.releaseId"
                :base-version="selectedVersion"
                :current-config-list="props.configList"
                :base-config-list="baseConfigList">
                <template #baseHead>
                    <div class="version-selector">
                        对比版本：
                        <bk-select
                            :model-value="selectedVersion"
                            style="width: 320px;"
                            size="small"
                            :loading="versionListLoading"
                            :clearable="false"
                            @change="handleSelectVersion">
                            <bk-option
                                v-for="version in versionList"
                                :key="version.id"
                                :label="version.spec.name"
                                :value="version.id">
                            </bk-option>
                        </bk-select>
                    </div>
                </template>
                <template #currentHead>
                    <div class="current-panel-head">
                        <span class="version-status">待上线</span>
                        {{ props.versionName }}
                    </div>
                </template>
            </config-diff>
            <template #footer>
                <section class="actions-wrapper">
                    <bk-button theme="primary" style="margin-right: 8px;" @click="isConfirmDialogShow = true">上线版本</bk-button>
                    <bk-button @click="handleClose">取消</bk-button>
                </section>
            </template>
        </VersionLayout>
        <ConfirmDialog
            v-model:show="isConfirmDialogShow"
            :bk-biz-id="props.bkBizId"
            :app-id="props.appId"
            :release-id="props.releaseId"
            :groups="groups"
            @confirm="handleConfirm"/>
    </section>
</template>
<style lang="scss" scoped>
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
    .current-panel-head {
        display: flex;
        align-items: center;
        padding: 0 24px;
        font-size: 12px;
        height: 100%;
        .version-status {
            margin-right: 4px;
            padding: 4px 10px;
            line-height: 1;
            color: #14a568;
            background: #e4faf0;
            border-radius: 2px;
        }
    }
    .actions-wrapper {
        display: flex;
        align-items: center;
        padding: 0 24px;
        height: 100%;
    }
    .version-selector {
        display: flex;
        align-items: center;
        height: 100%;
        padding: 0 24px;
        font-size: 12px;
    }
</style>
