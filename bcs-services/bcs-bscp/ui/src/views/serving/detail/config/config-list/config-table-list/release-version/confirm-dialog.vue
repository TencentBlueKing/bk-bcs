<script setup lang="ts">
  import { defineProps, defineEmits, ref } from 'vue'
  import { createVersion } from '../../../../../../../api/config'

  const props = defineProps<{
    show: boolean,
    bkBizId: string,
    appId: number
  }>()

  const emits = defineEmits(['update:show', 'confirm'])

  const localVal = ref({
    name: '',
    memo: '',
    group: '',
    publishDesc: ''
  })
  const isPublish= ref(false)
  const pending = ref(false)
  const formRef = ref()
  const rules = {
    memo: [
      {
        validator: (value: string) => value.length < 100,
        message: '最大长度100个字符'
      }
    ],
    group: [],
    publishDesc: [
      {
        validator: (value: string) => value.length < 100,
        message: '最大长度100个字符'
      }
    ]
  }

  const handleClose = () => {
    emits('update:show', false)
    localVal.value = {
      name: '',
      memo: '',
      group: '',
      publishDesc: ''
    }
    isPublish.value = false
  }

  const handleConfirm = async() => {
    try {
      pending.value = true
      await formRef.value.validate()
      await createVersion(props.bkBizId, props.appId, localVal.value.name, localVal.value.memo)
      emits('confirm')
      handleClose()
    } catch (e) {
      console.error(e)
    } finally {
      pending.value = false
    }
  }

</script>
<template>
  <bk-dialog
    title="生成版本"
    ext-cls="release-version-dialog"
    :is-show="props.show"
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
        <template v-if="isPublish">
          <bk-form-item label="上线分组" property="group">
            <bk-input></bk-input>
          </bk-form-item>
          <bk-form-item label="上线说明" property="publishDesc">
            <bk-input v-model="localVal.publishDesc" type="textarea" :maxlength="100"></bk-input>
          </bk-form-item>
        </template>
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
  .release-version-dialog.bk-dialog-wrapper .bk-dialog-header {
    padding-bottom: 20px;
  }
</style>