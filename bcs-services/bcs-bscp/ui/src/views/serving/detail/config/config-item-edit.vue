<script setup lang="ts">
  import { defineProps, defineEmits, ref, computed, watch } from 'vue'
  import { cloneDeep } from 'lodash'
  import { InfoLine, Upload, FilliscreenLine } from 'bkui-vue/lib/icon'
  import CodeEditor from '../../../../components/code-editor/index.vue'
  import { createServingConfigItem, updateServingConfigItem } from '../../../../api/config'
  import { IServingEditParams } from '../../../../types'

  const props = defineProps<{
    show: boolean,
    config: IServingEditParams,
    bkBizId: number
  }>()

  const emit = defineEmits(['update:show', 'confirm'])

  const isShow = ref(props.show)
  const localVal = ref(cloneDeep(props.config))
  const pending = ref(false)
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

  const type = computed(() => {
    return typeof props.config.id === 'number' ? 'edit' : 'new'
  })  

  const title = computed(() => {
    return type.value === 'edit' ? '编辑配置项' : '新增配置项'
  })

  watch(() => props.show, (val) => {
    isShow.value = val
    localVal.value = cloneDeep(props.config)
  })

  const handleSubmit = async() => {
    try {
      await formRef.value.validate()
      pending.value = true
      if (type.value === 'edit') {
        await updateServingConfigItem(localVal.value)
      } else {
        await createServingConfigItem(localVal.value)
      }
      close()
      emit('confirm')
    } catch (e) {
      console.error(e)
    } finally {
      pending.value = false
    }
  }

  const close = () => {
    isShow.value = false
    emit('update:show', false)
  }

</script>
<template>
    <bk-sideslider
      width="640"
      :is-show="isShow"
      :title="title"
      :before-close="close">
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
          <template v-if="localVal.file_type === 'binary'">
            <bk-form-item label="文件权限" property="privilege" :required="true">
              <bk-input v-model="localVal.privilege"></bk-input>
            </bk-form-item>
            <bk-form-item label="用户" property="user" :required="true">
              <bk-input v-model="localVal.user"></bk-input>
            </bk-form-item>
            <bk-form-item label="配置路径" property="path" :required="true">
              <bk-input v-model="localVal.path"></bk-input>
            </bk-form-item>
            <bk-form-item label="配置内容" :required="true">
              <bk-upload theme="button" tip="支持扩展名：.bin" accept=".bin" :multiple="false"></bk-upload>
            </bk-form-item>
          </template>
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
              <CodeEditor />
            </div>
          </bk-form-item>
        </bk-form>
      </section>
      <section class="actions-wrapper">
        <bk-button theme="primary" :loading="pending" @click="handleSubmit">保存</bk-button>
        <bk-button @click="close">取消</bk-button>
      </section>
    </bk-sideslider>
</template>
<style lang="scss" scoped>
  :deep(.bk-modal-content) {
    height: 100%;
  }
  .form-content {
    padding: 22px;
    height: calc(100% - 48px);
    overflow: auto;
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