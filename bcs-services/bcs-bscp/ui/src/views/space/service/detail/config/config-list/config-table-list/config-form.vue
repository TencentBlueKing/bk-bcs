<script setup lang="ts">
  import { ref, computed, watch } from 'vue'
  import SHA256 from 'crypto-js/sha256'
  import WordArray from 'crypto-js/lib-typedarrays'
  import { TextFill, Done } from 'bkui-vue/lib/icon'
  import BkMessage from 'bkui-vue/lib/message'
  import { IAppEditParams } from '../../../../../../../../types/app'
  import { IFileConfigContentSummary } from '../../../../../../../../types/config'
  import { updateConfigContent, getConfigContent } from '../../../../../../../api/config'
  import { stringLengthInBytes } from '../../../../../../../utils/index'
  import { transFileToObject, fileDownload } from '../../../../../../../utils/file'
  import { CONFIG_FILE_TYPE } from '../../../../../../../constants/config'
  import ConfigContentEditor from '../../components/config-content-editor.vue'

  const props = withDefaults(defineProps<{
    config: IAppEditParams,
    editable: boolean,
    content: string|IFileConfigContentSummary,
    bkBizId: string,
    appId: number,
    submitFn: Function
  }>(), {
    editable: true
  })

  const emit = defineEmits(['confirm', 'cancel'])

  const localVal = ref({ ...props.config })
  const stringContent = ref('')
  const fileContent = ref<IFileConfigContentSummary|File>()
  const isFileChanged = ref(false) // 标识文件是否被修改，编辑配置项时若文件未修改，不重新上传文件
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
          return /^[a-zA-Z0-9_\-\.]+$/.test(value)
        },
        message: '请使用英文、数字、下划线、中划线或点'
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
    return fileContent.value ? [transFileToObject(<File>fileContent.value)] : []
  })

  watch(() => props.content, () => {
    if (props.config.file_type === 'binary') {
      fileContent.value = <IFileConfigContentSummary>props.content
    } else {
      stringContent.value = props.content as string
    }
  }, { immediate: true })

  // 选择文件后上传
  const handleFileUpload = (option: { file: File }) => {
    isFileChanged.value = true
    return new Promise(resolve => {
      uploadPending.value = true
      fileContent.value = option.file
      uploadContent().then(res => {
        uploadPending.value = false
        resolve(res)
      })
    })
  }

  // 提交保存
  const handleSubmit = async() => {
    try {
      await formRef.value.validate()
      if (localVal.value.file_type === 'binary'){
        if (fileList.value.length === 0) {
          BkMessage({ theme: 'error', message: '请上传文件' })
          return
        }
      } else if (localVal.value.file_type === 'text') {
        if (stringLengthInBytes(stringContent.value) > 1024 * 1024 * 40 ) {
          BkMessage({ theme: 'error', message: '配置内容不能超过40M' })
          return
        }
      }

      submitPending.value = true
      let sign = await generateSHA256()
      let size = 0
      if (localVal.value.file_type === 'binary') {
        size = Number((<IFileConfigContentSummary|File>fileContent.value).size)
      } else {
        size = new Blob([stringContent.value]).size
        await uploadContent()
      }
      const params = { ...localVal.value, ...{ sign, byte_size: size } }
      if (typeof props.submitFn === 'function') {
        await props.submitFn(params)
      }
      emit('confirm')
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
    const data = localVal.value.file_type === 'binary' ? fileContent.value : JSON.stringify(stringContent.value)
    // @ts-ignore
    return updateConfigContent(props.bkBizId, props.appId, data, SHA256Str)
  }

  // 生成文件或文本的sha256
  const generateSHA256 = async () => {
    if (localVal.value.file_type === 'binary') {
      if (isFileChanged.value) {
        return new Promise(resolve => {
          const reader = new FileReader()
          // @ts-ignore
          reader.readAsArrayBuffer(fileContent.value)
          reader.onload = () => {
            const wordArray = WordArray.create(reader.result);
            resolve(SHA256(wordArray).toString())
          }
        })
      }
      return (fileContent.value as IFileConfigContentSummary).signature
    }
    return SHA256(JSON.stringify(stringContent.value)).toString()
  }

  // 下载已上传文件
  const handleDownloadFile = async () => {
    const { signature, name } = <IFileConfigContentSummary>fileContent.value
    const res = await getConfigContent(props.bkBizId, props.appId, signature)
    fileDownload(res, `${name}.bin`)
  }

  // 打开全屏
  const handleOpenFullScreen = () => {}

  const cancel = () => {
    emit('cancel')
  }

</script>
<template>
  <section class="form-content">
    <bk-form ref="formRef" :model="localVal" :rules="rules">
        <bk-form-item label="配置项名称" property="name" :required="true">
          <bk-input v-model="localVal.name" :disabled="!editable"></bk-input>
        </bk-form-item>
        <bk-form-item label="配置格式">
        <bk-radio-group v-model="localVal.file_type" :required="true">
            <bk-radio v-for="typeItem in CONFIG_FILE_TYPE" :key="typeItem.id" :label="typeItem.id" :disabled="!editable">{{ typeItem.name }}</bk-radio>
        </bk-radio-group>
        </bk-form-item>
        <template v-if="['binary', 'text'].includes(localVal.file_type)">
        <bk-form-item label="文件权限" property="privilege" :required="true">
            <bk-input v-model="localVal.privilege" :disabled="!editable"></bk-input>
        </bk-form-item>
        <bk-form-item label="用户" property="user" :required="true">
            <bk-input v-model="localVal.user" :disabled="!editable"></bk-input>
        </bk-form-item>
        <bk-form-item label="配置路径" property="path" :required="true">
            <bk-input v-model="localVal.path" :disabled="!editable"></bk-input>
        </bk-form-item>
        </template>
        <bk-form-item v-if="localVal.file_type === 'binary'" label="配置内容" :required="true">
          <bk-upload
            class="config-uploader"
            url=""
            theme="button"
            tip="支持扩展名：.bin，文件大小100M以内"
            :size="100"
            :disabled="!editable"
            :multiple="false"
            :files="fileList"
            :custom-request="handleFileUpload">
            <template #file="{ file }">
              <div class="file-wrapper">
                <Done class="done-icon"/>
                <TextFill class="file-icon" />
                <div v-bk-ellipsis class="name" @click="handleDownloadFile">{{ file.name }}</div>
                ({{ file.size }})
              </div>
            </template>
          </bk-upload>
        </bk-form-item>
        <bk-form-item v-else label="配置内容" :required="true">
          <ConfigContentEditor
            :content="stringContent"
            :editable="editable"
            @change="stringContent = $event" />
        </bk-form-item>
    </bk-form>
    <section class="actions-wrapper">
      <bk-button v-if="props.editable" theme="primary" :disabled="uploadPending" :loading="submitPending" @click="handleSubmit">保存</bk-button>
      <bk-button @click="cancel">{{ props.editable ? '取消' : '关闭' }}</bk-button>
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
    :deep(.bk-upload-list--disabled .bk-upload-list__item) {
      pointer-events: inherit;
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
        cursor: pointer;
        &:hover {
          color: #3a84ff;
          text-decoration: underline;
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