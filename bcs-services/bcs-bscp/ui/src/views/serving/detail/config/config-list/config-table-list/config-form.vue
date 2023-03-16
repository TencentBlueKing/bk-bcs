<script setup lang="ts">
  import { defineProps, defineEmits, ref, computed, watch } from 'vue'
  import SHA256 from 'crypto-js/sha256'
  import WordArray from 'crypto-js/lib-typedarrays'
  import { TextFill, InfoLine, Upload, Done, FilliscreenLine } from 'bkui-vue/lib/icon'
  import CodeEditor from '../../../../../../components/code-editor/index.vue'
  import { IServingEditParams } from '../../../../../../types'
  import { updateConfigContent } from '../../../../../../api/config'
  import { transFileToObject } from '../../../../../../utils/file'

  const props = defineProps<{
    config: IServingEditParams,
    content: string|File,
    bkBizId: string,
    appId: number,
    submitFn: Function
  }>()

  const emit = defineEmits(['submit', 'cancel'])

  const localVal = ref({ ...props.config })
  const stringContent = ref('')
  const file = ref()
  const submitPending = ref(false)
  const uploadPending = ref(false)
  const formRef = ref()
  const rules = {
    name: [
      {
        validator: (value: string) => value.length < 64,
        message: '最大长度64个字符'
      },
      {
        validator: (value: string) => {
          return /^[a-zA-Z0-9][a-zA-Z0-9_\-\.]*[a-zA-Z0-9]?$/.test(value)
        },
        message: '请使用英文、数字、下划线、中划线、点，且必须以英文、数字开头和结尾'
      }
    ],
    path: [
      {
        validator: (value: string) => value.length < 256,
        message: '最大长度256个字符'
      }
    ],
  }

  // 传入到bk-upload组件的文件对象
  const fileList = computed(() => {
    return file.value ? [transFileToObject(file.value)] : []
  })

  watch(() => props.content, () => {
    if (props.config.file_type === 'binary') {
      console.log(props.content)
      file.value = props.content as File
    } else {
      stringContent.value = props.content as string
    }
  }, { immediate: true })

  // 选择文件后上传
  const handleFileUpload = (option: { file: File }) => {
    return new Promise(resolve => {
      uploadPending.value = true
      file.value = option.file
      uploadContent().then(res => {
        console.log('uploaded res: ', res)
        uploadPending.value = false
        resolve(res)
      })
    })
  }

  // 提交保存
  const handleSubmit = async() => {
    try {
      await formRef.value.validate()
      submitPending.value = true
      let sign = await generateSHA256()
      let size = 0
      if (localVal.value.file_type === 'binary') {
        size = file.value.size
      } else {
        size = new Blob([stringContent.value]).size
        await uploadContent()
      }
      const params = { ...localVal.value, ...{ sign, byte_size: size } }
      if (typeof props.submitFn === 'function') {
        await props.submitFn(params)
      }
      emit('submit')
      cancel()
    } catch (e) {
      console.error(e)
    } finally {
      submitPending.value = false
    }
  }

  // 上传配置内容
  const uploadContent =  async () => {
    const SHA256Str = await generateSHA256()
    const data = localVal.value.file_type === 'binary' ? file.value : stringContent.value
    // @ts-ignore
    return updateConfigContent(props.bkBizId, props.appId, data, SHA256Str)
  }

  // 生成文件或文本的sha256
  const generateSHA256 = async () => {
    if (localVal.value.file_type === 'binary') {
      return new Promise(resolve => {
        const reader = new FileReader()
        // @ts-ignore
        reader.readAsArrayBuffer(file.value)
        reader.onload = () => {
          const wordArray = WordArray.create(reader.result);
          resolve(SHA256(wordArray).toString())
        }
      })
    } else {
      return SHA256(stringContent.value).toString()
    }
  }

  const cancel = () => {
    emit('cancel')
  }

</script>
<template>
  <section class="form-content">
    <bk-form ref="formRef" :model="localVal" :rules="rules">
        <bk-form-item label="配置项名称" property="name" :required="true">
        <bk-input v-model="localVal.name"></bk-input>
        </bk-form-item>
        <bk-form-item label="配置格式">
        <bk-radio-group v-model="localVal.file_type" :required="true">
            <bk-radio label="text">Text</bk-radio>
            <bk-radio label="binary">二进制文件</bk-radio>
        </bk-radio-group>
        </bk-form-item>
        <template v-if="['binary', 'text'].includes(localVal.file_type)">
        <bk-form-item label="文件权限" property="privilege" :required="true">
            <bk-input v-model="localVal.privilege"></bk-input>
        </bk-form-item>
        <bk-form-item label="用户" property="user" :required="true">
            <bk-input v-model="localVal.user"></bk-input>
        </bk-form-item>
        <bk-form-item label="配置路径" property="path" :required="true">
            <bk-input v-model="localVal.path"></bk-input>
        </bk-form-item>
        </template>
        <bk-form-item v-if="localVal.file_type === 'binary'" label="配置内容" :required="true">
          <bk-upload
            class="config-uploader"
            url=""
            theme="button"
            tip="支持扩展名：.bin"
            :multiple="false"
            :files="fileList"
            :size="100"
            :custom-request="handleFileUpload">
            <template #file="{ file }">
              <div class="file-wrapper">
                <Done class="done-icon"/>
                <TextFill class="file-icon" />
                <div v-bk-ellipsis class="name">{{ file.name }}</div>
                ({{ file.size }})
              </div>
            </template>
          </bk-upload>
        </bk-form-item>
        <bk-form-item v-else label="配置内容" :required="true">
        <div class="code-editor-content">
            <div class="editor-operate-area">
              <div class="tip">
                  <InfoLine />
                  仅支持大小不超过 100M
              </div>
              <div class="btns">
                  <Upload style="font-size: 14px; margin-right: 10px;" />
                  <FilliscreenLine />
              </div>
            </div>
            <CodeEditor v-model="stringContent" />
        </div>
        </bk-form-item>
    </bk-form>
    <section class="actions-wrapper">
      <bk-button theme="primary" :loading="submitPending" @click="handleSubmit">保存</bk-button>
      <bk-button @click="cancel">取消</bk-button>
    </section>
  </section>
</template>
<style lang="scss" scoped>
  .form-content {
    height: 100%;
  }
  .bk-form {
    padding: 22px;
    height: calc(100% - 48px);
    overflow: auto;
  }
  .config-uploader {
    :deep(.bk-upload-list__item) {
      padding: 0;
      border: none;
    }
    .file-wrapper {
      display: flex;
      align-items: center;
      color: #979ba5;
      font-size: 12px;
      .done-icon {
        font-size: 20px;
        color: #2dcb56;
      }
      .file-icon {
        margin: 0 6px 0 0;
      }
      .name {
        max-width: 360px;
        margin-right: 4px;
        color: #63656e;
        white-space: nowrap;
        text-overflow: ellipsis;
        overflow: hidden;
      }
    }
  }
  .editor-operate-area {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 16px;
    height: 40px;
    color: #979ba5;
    background: #2e2e2e;
    border-radius: 2px 2px 0 0;
    .tip {
      font-size: 12px;
    }
    .btns {
      display: flex;
      align-items: center;
      & > span{
        cursor: pointer;
        &:hover {
          color: #3a84ff;
        }
      }
    }
  }
  .actions-wrapper {
    display: flex;
    align-items: center;
    padding-left: 24px;
    height: 48px;
    background: #fafbfd;
    border-top: 1px solid #dcdee5;
    .bk-button {
      margin-right: 8px;
      min-width: 88px;
    }
  }
</style>