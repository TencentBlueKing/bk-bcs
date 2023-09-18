<script setup lang="ts">
  import { ref, computed, watch, onMounted } from 'vue'
  import { ArrowsLeft, AngleRight } from 'bkui-vue/lib/icon'
  import InfoBox from 'bkui-vue/lib/info-box';
  import BkMessage from 'bkui-vue/lib/message';
  import { storeToRefs } from 'pinia'
  import { useGlobalStore } from '../../../../../store/global'
  import { useServiceStore } from '../../../../../store/service'
  import { useConfigStore } from '../../../../../store/config'
  import { IGroupToPublish } from '../../../../../../types/group';
  import { permissionCheck } from '../../../../../api/index'
  import VersionLayout from '../config/components/version-layout.vue'
  import ConfirmDialog from './publish-version/confirm-dialog.vue'
  import SelectGroup from './publish-version/select-group/index.vue'
  import VersionDiff from '../config/components/version-diff/index.vue';

  const { permissionQuery, showApplyPermDialog } = storeToRefs(useGlobalStore())
  const serviceStore = useServiceStore()
  const versionStore = useConfigStore()
  const { appData } = storeToRefs(serviceStore)
  const { versionData } = storeToRefs(versionStore)

  const props = defineProps<{
    bkBizId: string,
    appId: number
  }>()

  const emit = defineEmits(['confirm'])

  const permCheckLoading = ref(false)
  const hasPublishVersionPerm = ref(false)
  const isSelectGroupPanelOpen = ref(false)
  const isDiffSliderShow = ref(false)
  const isConfirmDialogShow = ref(false)
  const groupType = ref('select')
  const groups = ref<IGroupToPublish[]>([])
  const baseVersionId = ref(0)

  const currentSelectedGroups = computed(() => {
    return versionData.value.status.released_groups.map(group => group.id)
  })

  const permissionQueryResource = computed(() => {
    return [{
      biz_id: props.bkBizId,
      basic: {
        type: 'app',
        action: 'publish',
        resource_id: props.appId
      }
    }]
  })

  watch(() => versionData.value.status.publish_status, val => {
    if (val === 'partial_released') {
      checkPublishVersionPerm()
    }
  })

  onMounted(() => {
    if (versionData.value.status.publish_status === 'partial_released') {
      checkPublishVersionPerm()
    }
  })

  const checkPublishVersionPerm = async () => {
    permCheckLoading.value = true
    const res = await permissionCheck({ resources: permissionQueryResource.value })
    hasPublishVersionPerm.value = res.is_allowed
    permCheckLoading.value = false
  }

  const handleBtnClick = () => {
    if (hasPublishVersionPerm.value) {
      openSelectGroupPanel()
    } else {
      permissionQuery.value = { resources: permissionQueryResource.value }
      showApplyPermDialog.value = true
    }
  }

// 打开选择分组面板
  const openSelectGroupPanel = () => {
    isSelectGroupPanelOpen.value = true
    groups.value = versionData.value.status.released_groups.map(group => {
      const { id, name} = group
      const selector = group.new_selector
      const rules = selector.labels_and || []
      return {
        id,
        name,
        release_id: versionData.value.id,
        release_name: versionData.value.spec.name,
        disabled: true,
        rules: rules
      }
    })
  }

  // 打开上线版本确认弹窗
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

  // 上线确认
  const handleConfirm = () => {
    isDiffSliderShow.value = false
    handlePanelClose()
    emit('confirm')
    InfoBox({
    // @ts-ignore
      infoType: "success",
      title: '调整分组上线成功',
      dialogType: 'confirm'
    })
  }

  // 关闭选择分组面板
  const handlePanelClose = () => {
    groupType.value = 'select'
    isSelectGroupPanelOpen.value = false
    groups.value = []
  }

</script>
<template>
    <section class="create-version">
        <bk-button
          v-if="versionData.status.publish_status === 'partial_released'"
          v-cursor="{ active: !permCheckLoading && hasPublishVersionPerm }"
          theme="primary"
          :class="['trigger-button', { 'bk-button-with-no-perm': !permCheckLoading && hasPublishVersionPerm }]"
          :disabled="permCheckLoading"
          @click="handleBtnClick">
          调整分组上线
        </bk-button>
        <Teleport to="body">
          <VersionLayout v-if="isSelectGroupPanelOpen">
              <template #header>
                  <section class="header-wrapper">
                      <span class="header-name" @click="handlePanelClose">
                          <ArrowsLeft class="arrow-left" />
                          <span class="service-name">{{ appData.spec.name }}</span>
                      </span>
                      <AngleRight class="arrow-right" />
                      调整分组上线：{{ versionData.spec.name }}
                  </section>
              </template>
              <select-group
                :group-type="groupType"
                :groups="groups"
                :disabled="currentSelectedGroups"
                @openPreviewVersionDiff="openPreviewVersionDiff"
                @groupTypeChange="groupType = $event"
                @change="groups = $event">
              </select-group>
              <template #footer>
                  <section class="actions-wrapper">
                      <bk-button class="publish-btn" theme="primary" @click="isDiffSliderShow = true">对比并上线</bk-button>
                      <bk-button @click="handlePanelClose">取消</bk-button>
                  </section>
              </template>
          </VersionLayout>
        </Teleport>
        <ConfirmDialog
            v-model:show="isConfirmDialogShow"
            :bk-biz-id="props.bkBizId"
            :app-id="props.appId"
            :release-id="versionData.id"
            :group-type="groupType"
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
