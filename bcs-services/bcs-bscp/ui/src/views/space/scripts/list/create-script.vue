<script setup lang="ts">
  import { ref, onMounted } from 'vue'
  import { storeToRefs } from 'pinia'
  import BkMessage from 'bkui-vue/lib/message'
  import { useGlobalStore } from '../../../../store/global'
  import { EScriptType, IScriptEditingForm, IScriptTagItem } from '../../../../../types/script'
  import { createScript, getScriptTagList } from '../../../../api/script'
  import DetailLayout from '../components/detail-layout.vue'
  import ScriptEditor from '../components/script-editor.vue'

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
  const tagsLoading = ref(false)
  const tagsData = ref<IScriptTagItem[]>([])

  onMounted(() => {
    getTags()
  })

  // 获取标签列表
  const getTags = async () => {
    tagsLoading.value = true
    const res = await getScriptTagList(spaceId.value)
    tagsData.value = res.details
    tagsLoading.value = false
  }

  const handleCreate = async() => {
    await formRef.value.validate()
    try {
      pending.value = true
      await createScript(spaceId.value, formData.value)
      BkMessage({
        theme: 'success',
        message: '脚本创建成功'
      })
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
            <!-- <bk-input v-model="formData.tag" /> -->
            <bk-select
              v-model="formData.tag"
              placeholder="请选择标签或输入新标签按Enter结束"
              :loading="tagsLoading"
              :allow-create="true"
              :filterable="true">
              <bk-option v-for="option in tagsData" :key="option.tag" :value="option.tag" :label="option.tag"></bk-option>
            </bk-select>
          </bk-form-item>
          <bk-form-item class="fixed-width-form"  property="memo" label="脚本描述">
            <bk-input v-model="formData.memo" type="textarea" :rows="3" :maxlength="200" />
          </bk-form-item>
          <bk-form-item class="fixed-width-form"  property="release_name" label="版本号" required>
            <bk-input v-model="formData.release_name" />
          </bk-form-item>
          <bk-form-item label="脚本内容"  property="content" required>
            <ScriptEditor v-model="formData.content" class="script-content-wrapper" :language="formData.type">
              <template #header>
                <div class="language-tabs">
                  <div
                    v-for="item in SCRIPT_TYPE"
                    :key="item.id"
                    :class="['tab', { actived: formData.type === item.id }]"
                    @click="formData.type = item.id">
                    {{ item.name }}
                  </div>
                </div>
              </template>
            </ScriptEditor>
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
  .script-content-wrapper {
    min-width: 520px;
  }
  
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
  .actions-wrapper {
    .bk-button {
      margin-right: 8px;
      min-width: 88px;
    }
  }
</style>