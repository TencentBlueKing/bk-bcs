<script setup lang="ts">
  import { ref } from 'vue'
  import { ArrowsLeft, AngleRight } from 'bkui-vue/lib/icon'
  import InfoBox from "bkui-vue/lib/info-box";
  import { storeToRefs } from 'pinia'
  import { useServingStore } from '../../../../../../../store/serving'
  import { useConfigStore } from '../../../../../../../store/config'
  import VersionLayout from '../../../components/version-layout.vue'
  import ConfirmDialog from './confirm-dialog.vue'
  import { IConfigItem } from '../../../../../../../../types/config'
  import SelectGroup from './select-group/index.vue'
  import VersionDiff from '../../../components/version-diff/index.vue';

  const servingStore = useServingStore()
  const versionStore = useConfigStore()
  const { appData } = storeToRefs(servingStore)
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
  const groups = ref([])

  const handleConfirm = () => {
    InfoBox({
    // @ts-ignore
      infoType: "success",
      title: '版本已上线',
      dialogType: 'confirm',
      onConfirm () {
        emit('confirm')
        handlePanelClose()
      }
    })
  }

  const handlePanelClose = () => {
    openSelectGroupPanel.value = false
  }

</script>
<template>
    <section class="create-version">
        <bk-button theme="primary" @click="openSelectGroupPanel = true">上线版本</bk-button>
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
            <select-group></select-group>
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
        <VersionDiff v-model:show="isDiffSliderShow" :current-version="versionData" :show-publish-btn="true" />
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
