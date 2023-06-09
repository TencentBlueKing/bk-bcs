<script setup lang="ts">
  import { ref } from 'vue'
  import { ArrowsLeft, AngleRight } from 'bkui-vue/lib/icon'
  import InfoBox from 'bkui-vue/lib/info-box';
  import BkMessage from 'bkui-vue/lib/message';
  import { storeToRefs } from 'pinia'
  import { useServiceStore } from '../../../../../../../../store/service'
  import { useConfigStore } from '../../../../../../../../store/config'
  import VersionLayout from '../../../components/version-layout.vue'
  import ConfirmDialog from './confirm-dialog.vue'
  import { IConfigItem } from '../../../../../../../../../types/config'
  import { IGroupToPublish } from '../../../../../../../../../types/group';
  import SelectGroup from './select-group/index.vue'
  import VersionDiff from '../../../components/version-diff/index.vue';

  const serviceStore = useServiceStore()
  const versionStore = useConfigStore()
  const { appData } = storeToRefs(serviceStore)
  const { versionData } = storeToRefs(versionStore)

  const props = defineProps<{
    bkBizId: string,
    appId: number,
    configList: IConfigItem[]
  }>()

  const emit = defineEmits(['confirm'])

  const openSelectGroupPanel = ref(false)
  const isDiffSliderShow = ref(false)
  const isConfirmDialogShow = ref(false)
  const groups = ref<IGroupToPublish[]>([])
  const baseVersionId = ref(0)

  const handleOpenSelectGroupPanel = () => {
    openSelectGroupPanel.value = true
  }

  const handleOpenPublishDialog = () => {
    if (groups.value.length === 0) {
      BkMessage({ theme: 'error', message: '请选择上线分组' })
      return
    }
    isConfirmDialogShow.value = true
  }

  // 选择分组面板上线预览版本对比
  const openPreviewVersionDiff = (id: number) => {
    baseVersionId.value = id
    isDiffSliderShow.value = true
  }

  // 版本上线成功
  const handleConfirm = () => {
    isDiffSliderShow.value = false
    handlePanelClose()
    emit('confirm')
    InfoBox({
    // @ts-ignore
      infoType: "success",
      title: '版本已上线',
      dialogType: 'confirm'
    })
  }

  const handlePanelClose = () => {
    openSelectGroupPanel.value = false
    groups.value = []
  }

  defineExpose({
    handleOpenSelectGroupPanel
  })

</script>
<template>
    <section class="create-version">
        <bk-button v-if="versionData.status.publish_status === 'not_released'" class="trigger-button" theme="primary" @click="handleOpenSelectGroupPanel">上线版本</bk-button>
        <VersionLayout v-if="openSelectGroupPanel">
            <template #header>
                <section class="header-wrapper">
                    <span class="header-name" @click="handlePanelClose">
                        <ArrowsLeft class="arrow-left" />
                        <span class="service-name">{{ appData.spec.name }}</span>
                    </span>
                    <AngleRight class="arrow-right" />
                    上线版本：{{ versionData.spec.name }}
                </section>
            </template>
            <select-group :groups="groups" @openPreviewVersionDiff="openPreviewVersionDiff" @change="groups = $event"></select-group>
            <template #footer>
                <section class="actions-wrapper">
                    <bk-button class="publish-btn" theme="primary" @click="isDiffSliderShow = true">对比并上线</bk-button>
                    <bk-button @click="handlePanelClose">取消</bk-button>
                </section>
            </template>
        </VersionLayout>
        <ConfirmDialog
            v-model:show="isConfirmDialogShow"
            :bk-biz-id="props.bkBizId"
            :app-id="props.appId"
            :release-id="versionData.id"
            :groups="groups"
            @confirm="handleConfirm"/>
        <VersionDiff
          v-model:show="isDiffSliderShow"
          :current-version="versionData"
          :base-version-id="baseVersionId"
          :show-publish-btn="true"
          @publish="handleOpenPublishDialog" />
    </section>
</template>
<style lang="scss" scoped>
    .trigger-button {
      margin-left: 8px;
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
    .actions-wrapper {
        display: flex;
        align-items: center;
        padding: 0 24px;
        height: 100%;
        .publish-btn {
          margin-right: 8px;
        }
        .bk-button {
          min-width: 88px;
        }
    }
    .version-selector {
        display: flex;
        align-items: center;
        height: 100%;
        padding: 0 24px;
        font-size: 12px;
    }
</style>
