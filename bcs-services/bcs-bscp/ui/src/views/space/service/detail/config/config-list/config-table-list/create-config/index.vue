<script lang="ts" setup>
  import { ref } from 'vue'
  import { AngleDown } from 'bkui-vue/lib/icon'
  import ManualCreate from './manual-create.vue';
  import ImportFromTemplate from './import-from-templates.vue'

  const props = defineProps<{
    bkBizId: string,
    appId: number
  }>()

  const buttonRef = ref()
  const isPopoverOpen = ref(false)
  const isManualCreateSliderOpen = ref(false)
  const isImportTemplatesDialogOpen = ref(false)

  const handleManualCreateSlideOpen = () => {
    isManualCreateSliderOpen.value = true
    buttonRef.value.hide()
  }

  const handleImportTemplateDialogOpen = () => {
    isImportTemplatesDialogOpen.value = true
    buttonRef.value.hide()
  }

  const handleImported = () => {}

</script>
<template>
  <bk-popover
    ref="buttonRef"
    theme="light create-config-button-popover"
    placement="bottom-end"
    trigger="click"
    width="122"
    :arrow="false"
    @after-show="isPopoverOpen = true"
    @after-hidden="isPopoverOpen = false">
    <div
      theme="primary"
      :class="['create-config-btn', { 'popover-open': isPopoverOpen }]">
      新增配置文件
      <AngleDown class="angle-icon" />
    </div>
    <template #content>
      <div class="add-config-operations">
        <div class="operation-item" @click="handleManualCreateSlideOpen">手动新增</div>
        <div class="operation-item" @click="handleImportTemplateDialogOpen">从配置模板导入</div>
      </div>
    </template>
  </bk-popover>
  <ManualCreate
    v-model:show="isManualCreateSliderOpen"
    :bk-biz-id="props.bkBizId"
    :app-id="props.appId" />
  <ImportFromTemplate
    v-model:show="isImportTemplatesDialogOpen"
    :bk-biz-id="props.bkBizId"
    :app-id="props.appId"
    @imported="handleImported" />
</template>
<style lang="scss" scoped>
  .create-config-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    min-width: 122px;
    height: 32px;
    line-height: 32px;
    color: #3a84ff;
    border: 1px solid #3a84ff;
    border-radius: 2px;
    cursor: pointer;
    &.popover-open {
      .angle-icon {
        transform: rotate(-180deg);
      }
    }
    .angle-icon {
      font-size: 20px;
      transition: transform .3s cubic-bezier(.4, 0, .2, 1);
    }
  }
</style>
<style lang="scss">
  .create-config-button-popover.bk-popover.bk-pop2-content {
    padding: 4px 0;
    border: 1px solid #dcdee5;
    box-shadow: 0 2px 6px 0 #0000001a;
    .add-config-operations {
      .operation-item {
        padding: 0 12px;
        min-width: 58px;
        height: 32px;
        line-height: 32px;
        color: #63656e;
        font-size: 12px;
        cursor: pointer;
        &:hover {
          background: #f5f7fa;
        }
      }
    }
  }
</style>
