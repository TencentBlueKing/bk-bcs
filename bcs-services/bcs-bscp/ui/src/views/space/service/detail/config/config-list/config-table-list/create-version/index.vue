<script setup lang="ts">
  import { ref } from 'vue'
  import { storeToRefs } from 'pinia'
  import {  assign } from 'lodash'
	import BkMessage from 'bkui-vue/lib/message';
  import { GET_UNNAMED_VERSION_DATE } from '../../../../../../../../constants/config'
  import { useConfigStore } from '../../../../../../../../store/config'
  import { createVersion } from '../../../../../../../../api/config'

  const { versionData } = storeToRefs(useConfigStore())

  const props = defineProps<{
    bkBizId: string,
    appId: number,
    configCount: number
  }>()

  const emits = defineEmits(['confirm'])

  const isConfirmDialogShow = ref(false)
  const localVal = ref({
    name: '',
    memo: '',
  })
  const isPublish= ref(false)
  const pending = ref(false)
  const formRef = ref()
  const rules = {
    name: [
      {
        validator: (value: string) => {
          if (value.length > 0) {
            return /^[\u4e00-\u9fa5a-zA-Z0-9][\u4e00-\u9fa5a-zA-Z0-9_\-\.]*[\u4e00-\u9fa5a-zA-Z0-9]?$/.test(value)
          }
          return true
        },
        message: '仅允许使用中文、英文、数字、下划线、中划线、点，且必须以中文、英文、数字开头和结尾'
      }
    ],
    memo: [
      {
        validator: (value: string) => value.length < 100,
        message: '最大长度100个字符'
      }
    ],
  }

  const handleCreateDialogOpen = () => {
    localVal.value = {
      name: '',
      memo: '',
    }
    isConfirmDialogShow.value = true
  }

  const handleConfirm = async() => {
    try {
      pending.value = true
      await formRef.value.validate()
      const res = await createVersion(props.bkBizId, props.appId, localVal.value.name, localVal.value.memo)
      // 创建接口未返回完整的版本详情数据，在前端拼接最新版本数据，加载完版本列表后再更新
      const version = assign({}, GET_UNNAMED_VERSION_DATE(), { id: res.data.id, spec: { name: localVal.value.name, memo: localVal.value.memo } })
			BkMessage({ theme: 'success', message: '新版本已生成' })
      emits('confirm', version, isPublish.value)
      handleClose()
    } catch (e) {
      console.error(e)
    } finally {
      pending.value = false
    }
  }

  const handleClose = () => {
    isConfirmDialogShow.value = false
    isPublish.value = false
  }

</script>
<template>
  <bk-button
    v-if="versionData.id === 0"
    class="trigger-button"
    theme="primary"
    :disabled="props.configCount === 0"
    @click="handleCreateDialogOpen">
    生成版本
  </bk-button>
  <bk-dialog
    title="生成版本"
    ext-cls="create-version-dialog"
    :is-show="isConfirmDialogShow"
    :esc-close="false"
    :quick-close="false"
    :is-loading="pending"
    @closed="handleClose"
    @confirm="handleConfirm">
      <bk-form class="form-wrapper" form-type="vertical" ref="formRef" :rules="rules" :model="localVal">
        <bk-form-item label="版本名称" property="name" :required="true">
          <bk-input v-model="localVal.name"></bk-input>
        </bk-form-item>
        <bk-form-item label="版本描述" property="memo">
          <bk-input v-model="localVal.memo" type="textarea" :maxlength="100"></bk-input>
        </bk-form-item>
        <bk-checkbox
          v-model="isPublish"
          style="margin-bottom: 15px;"
          :true-label="true"
          :false-label="false">
          同时上线版本
        </bk-checkbox>
      </bk-form>
      <template #footer>
        <div class="dialog-footer">
          <bk-button theme="primary" :loading="pending" @click="handleConfirm">确定</bk-button>
          <bk-button :disabled="pending" @click="handleClose">取消</bk-button>
        </div>
      </template>
  </bk-dialog>
</template>
<style lang="scss" scoped>
  .trigger-button {
    margin-left: 8px;
  }
  .form-wrapper {
    padding-bottom: 24px;
  }
  .dialog-footer {
    .bk-button {
      margin-left: 8px;
    }
  }
</style>
<style lang="scss">
  .create-version-dialog.bk-dialog-wrapper .bk-dialog-header {
    padding-bottom: 20px;
  }
</style>
