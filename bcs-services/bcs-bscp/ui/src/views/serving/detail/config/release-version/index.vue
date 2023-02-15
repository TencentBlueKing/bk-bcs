<script setup lang="ts">
  import { defineProps, ref } from 'vue'
  import { ArrowsLeft, AngleRight } from 'bkui-vue/lib/icon'
  import InfoBox from "bkui-vue/lib/info-box";
  import VersionLayout from '../components/version-layout.vue'
  import ConfirmDialog from './confirm-dialog.vue'
  import ConfigDiff from '../config-diff.vue'

  const props = defineProps<{
    bkBizId: number,
    appId: number,
    appName: string,
    versionName: string
  }>()
  const showDiffPanel = ref(false)
  const isConfirmDialogShow = ref(false)

  const handleSuccess = () => {
    InfoBox({
    // @ts-ignore
      infoType: "success",
      title: '版本已上线',
      dialogType: 'confirm',
    })
  }

  const handleFail = (message: string = '') => {
    InfoBox({
    // @ts-ignore
      infoType: "danger",
      title: '版本上线失败',
      subTitle: message || 'fasdfgdsfgsdfgertewrt',
      confirmText: '重试',
      onConfirm () {
        isConfirmDialogShow.value = true
      }
    })
  }

  const handleClose = () => {
    showDiffPanel.value = false
  }

</script>
<template>
    <section class="create-version">
        <bk-button theme="primary" @click="showDiffPanel = true">上线版本</bk-button>
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
            <config-diff>
                <template #head>
                    <div class="diff-left-panel-head">
                        <span class="version-status">待上线</span>
                        {{ props.versionName }}
                        <!-- @todo 待确定这里展示什么名称 -->
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
            @successed="handleSuccess"
            @failed="handleFail" />
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
    .diff-left-panel-head {
        padding: 0 24px;
        font-size: 12px;
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
</style>
