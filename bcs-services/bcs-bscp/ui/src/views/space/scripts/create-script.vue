<script setup lang="ts">
  import { ref } from 'vue'
  import { storeToRefs } from 'pinia'
  import { useGlobalStore } from '../../../store/global'
  import { EScriptType, IScriptEditingForm } from '../../../../types/script'
  import { createScript } from '../../../api/script'
  import DetailLayout from './components/detail-layout.vue'
  import CodeEditor from '../../../components/code-editor/index.vue'

  const { spaceId } = storeToRefs(useGlobalStore())

  const emits = defineEmits(['update:show', 'created'])

  const SCRIPT_TYPE = [
    { id: EScriptType.Shell, name: 'Shell' },
    { id: EScriptType.Python, name: 'Python' }
  ]

  const formData = ref<IScriptEditingForm>({
    name: '',
    tag: '',
    release_name: '',
    memo: '',
    type: EScriptType.Shell,
    content: '',
  })
  const formRef = ref()
  const pending = ref(false)

  const handleCreate = async() => {
    await formRef.value.validate()
    try {
      pending.value = true
      await createScript(spaceId.value, formData.value)
      handleClose()
      emits('created')
    } catch (e) {
      console.error(e)
    } finally {
      pending.value = false
    }    
  }

  const handleClose = () => {
    emits('update:show', false)
  }
</script>
<template>
  <DetailLayout name="新建脚本" @close="handleClose">
    <template #content>
      <div class="create-script-forms">
        <bk-form ref="formRef" form-type="vertical" :model="formData">
          <bk-form-item class="fixed-width-form" label="脚本名称" property="name" required>
            <bk-input v-model="formData.name"/>
          </bk-form-item>
          <bk-form-item class="fixed-width-form"  property="tag" label="分类标签">
            <bk-input v-model="formData.tag" />
          </bk-form-item>
          <bk-form-item class="fixed-width-form"  property="memo" label="脚本描述">
            <bk-input v-model="formData.memo" type="textarea" :rows="3" :maxlength="200" />
          </bk-form-item>
          <bk-form-item class="fixed-width-form"  property="release_name" label="版本号" required>
            <bk-input v-model="formData.release_name" />
          </bk-form-item>
          <bk-form-item label="脚本内容"  property="content" required>
            <div class="script-content">
              <div class="language-tabs">
                <div
                  v-for="item in SCRIPT_TYPE"
                  :key="item.id"
                  :class="['tab', { actived: formData.type === item.id }]"
                  @click="formData.type = item.id">
                  {{ item.name }}
                </div>
              </div>
              <div class="content-editor">
                <CodeEditor v-model="formData.content" :language="formData.type" />
              </div>
            </div>
          </bk-form-item>
        </bk-form>
      </div>
    </template>
    <template #footer>
      <div class="actions-wrapper">
        <bk-button theme="primary" :loading="pending" @click="handleCreate">创建</bk-button>
        <bk-button @click="handleClose">取消</bk-button>
      </div>
    </template>
  </DetailLayout>
</template>
<style scoped lang="scss">
  .create-script-forms {
    padding: 24px 48px;
    height: 100%;
    background: #f5f7fa;
    overflow: auto;
  }
  .fixed-width-form {
    width: 520px;
  }
  .script-content {
    .language-tabs {
      display: flex;
      align-items: center;
      background: #2e2e2e;
      .tab {
        padding: 10px 24px;
        line-height: 20px;
        font-size: 14px;
        color: #8a8f99;
        border-top: 3px solid #2e2e2e;
        cursor: pointer;
        &.actived {
          color: #c4c6cc;
          font-weight: 700;
          background: #1a1a1a;
          border-color: #3a84ff;
        }
      }
    }
    .content-editor {
      height: 600px;
    }
  }
  .actions-wrapper {
    .bk-button {
      margin-right: 8px;
      min-width: 80px;
    }
  }
</style>