<script lang="ts" setup>
  import { ref, watch } from 'vue'
  import { IPackagesCitedByApps } from '../../../../../../types/template';
  import { getUnNamedVersionAppsBoundByTemplateVersion } from '../../../../../api/template'

  const props = defineProps<{
    show: boolean;
    spaceId: string;
    templateSpaceId: number;
    templateId: number;
    versionId: number;
    pending: boolean;
  }>()

  const emits = defineEmits(['update:show', 'confirm'])

  const listLoading = ref(false)
  const appList = ref<IPackagesCitedByApps[]>([])

  watch(() => props.show, val => {
    if (val) {
      getBoundApps()
    }
  })

  const getBoundApps = async() => {
    listLoading.value = true
    const res = await getUnNamedVersionAppsBoundByTemplateVersion(props.spaceId, props.templateSpaceId, props.templateId, props.versionId, { start: 0, all: true })
    appList.value = res.details
    listLoading.value = false
  }

  const close = () => {
    emits('update:show', false)
  }
</script>
<template>
  <bk-dialog
    ext-cls="create-version-confirm-dialog"
    title="确认更新配置项版本？"
    header-align="center"
    footer-align="center"
    :width="400"
    :is-show="props.show"
    :esc-close="false"
    :quick-close="false"
    @closed="close">
    <p class="tips">以下套餐及服务未命名版本中引用的此配置项也将更新</p>
    <bk-table :data="appList">
      <bk-table-column label="所在模板套餐" prop="template_set_name"></bk-table-column>
      <bk-table-column label="使用此套餐的服务" prop="app_name"></bk-table-column>
    </bk-table>
    <template #footer>
      <div class="actions-wrapper">
        <bk-button theme="primary" :loading="pending" @click="emits('confirm')">确定</bk-button>
        <bk-button @click="close">取消</bk-button>
      </div>
    </template>
  </bk-dialog>
</template>
<style lang="scss" scoped>
  .actions-wrapper {
    padding-bottom: 20px;
    .bk-button:not(:last-of-type) {
      margin-right: 8px;
    }
  }
</style>
<style lang="scss">
  .create-version-confirm-dialog.bk-modal-wrapper.bk-dialog-wrapper {
    .bk-modal-footer {
      padding: 32px 0 48px;
      background: #ffffff;
      border-top: none;
      .bk-button {
        min-width: 88px;
      }
    }
  }
</style>
