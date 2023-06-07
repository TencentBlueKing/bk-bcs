<script setup lang="ts">
  import { ref } from 'vue'
  import { Plus } from 'bkui-vue/lib/icon'
  import { storeToRefs } from 'pinia'
  import { useGlobalStore } from '../../../../store/global'
  import { IScriptVersion } from '../../../../../types/script'
  import { getScriptVersionList } from '../../../../api/script'

  const { spaceId } = storeToRefs(useGlobalStore())

  const props = withDefaults(defineProps<{
    edit?: boolean; // 是否编辑当前未上线版本
    scriptId: number
  }>(), {
    edit: false
  })

  const emits = defineEmits(['create'])

  const popoverShow = ref(false)
  const dialogShow = ref(false)
  const list = ref<IScriptVersion[]>([])
  const listLoading = ref(false)
  const selectedScript = ref<number|string>('')
  const formRef = ref()

  const afterDialogShow = async(val: boolean) => {
    if (val) {
      selectedScript.value = ''
      listLoading.value = true
      const res = await getScriptVersionList(spaceId.value, props.scriptId, { start: 0, limit: 10 })
      list.value = res.details
      listLoading.value = false
    }
  }

  const afterPopoverHidden = () => {
    popoverShow.value = false
  }

  const handleCreateClick = () => {
    if (props.edit) {
      popoverShow.value = true
      return
    }
    dialogShow.value = true
  }

  const handleLoadScript = async() => {
    await formRef.value.validate()
    const script = list.value.find(item => item.id === selectedScript.value)
    if (script) {
      dialogShow.value = false
      emits('create', script.spec.content)
    }
  }
</script>
<template>
  <bk-popover
    ext-cls="create-tips-popover"
    theme="light"
    trigger="click"
    placement="bottom-start"
    :disabled="!props.edit"
    :is-show="popoverShow"
    @after-hidden="afterPopoverHidden">
    <bk-button
      theme="primary"
      @click="handleCreateClick">
      <Plus class="button-icon" />
      新建版本
    </bk-button>
    <template #content>
      <h3 class="tips">当前已有「未上线」版本</h3>
      <div class="actions">
        <bk-button theme="primary" size="small">前往编辑</bk-button>
        <bk-button size="small" @click="popoverShow = false">取消</bk-button>
      </div>
    </template>
  </bk-popover>
  <bk-dialog
    title="创建版本"
    confirmText="创建"
    head-align="left"
    footer-align="right"
    width="480"
    :is-show="dialogShow"
    @value-change="afterDialogShow"
    @confirm="handleLoadScript"
    @closed="dialogShow = false">
    <bk-form ref="formRef" form-type="vertical" :model="{ selectedScript }">
      <bk-form-item label="选择载入脚本" required property="selectedScript">
        <bk-select v-model="selectedScript" :loading="listLoading" :clearable="false">
          <bk-option v-for="option in list" :key="option.id" :value="option.id" :label="option.spec.name"></bk-option>
        </bk-select>
      </bk-form-item>
    </bk-form>
  </bk-dialog>
</template>
<style lang="scss" scoped>
  .button-icon {
    font-size: 18px;
  }
  .tips {
    margin: 0 0 16px;
    line-height: 24px;
    font-size: 16px;
    font-weight: normal;
    color: #313238;
  }
  .actions {
    text-align: right;
    .bk-button {
      margin-left: 8px;
    }
  }
</style>
<style lang="scss">
  .create-tips-popover.bk-popover {
    padding: 16px;
  }
</style>