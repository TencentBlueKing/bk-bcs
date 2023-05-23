<script setup lang="ts">
  import { ref, computed, watch, nextTick } from 'vue'
  import { useI18n } from "vue-i18n"
  import { storeToRefs } from 'pinia'
  import { useGlobalStore } from '../../../../../store/global'
  import { updateApp } from "../../../../../api/index"
  import { IAppItem } from '../../../../../../types/app'

  const { spaceList } = storeToRefs(useGlobalStore())

  const { t } = useI18n()

  const props = defineProps<{
    show: boolean,
    service: IAppItem
  }>()

  const emits = defineEmits(['update:show'])

  const isMemoEdit = ref(false)
  const formData = ref({
    memo: ''
  })
  const formRef = ref()
  const memoRef = ref()

  const spaceName = computed(() => {
    const space = spaceList.value.find(item => item.space_id === props.service.space_id)
    return space?.space_name
  })

  const rules = {
    memo: [
      {
        validator: (value: string) => {
          if (value.length > 0) {
            return /^[\u4e00-\u9fa5a-zA-Z0-9][\u4e00-\u9fa5a-zA-Z0-9_\-\s]*[\u4e00-\u9fa5a-zA-Z0-9]?$/.test(value)
          }
          return true
        },
        message: '仅允许使用中文、英文、数字、下划线、中划线、空格，且必须以中文、英文、数字开头和结尾'
      }
    ]
  }

  watch(() => props.show, (val) => {
    if (val) {
      formData.value.memo = props.service.spec.memo
    }
  })

  const handleEditMemo = () => {
    isMemoEdit.value = true
    nextTick(() => {
      memoRef.value.focus()
    })
  }

  const handleUpdateMemo = async() => {
    await formRef.value.validate()
    const { id, biz_id, spec } = props.service;
    const { name, mode, config_type, reload } = spec;
    const data = {
      id,
      biz_id,
      name,
      mode,
      config_type,
      reload_type: reload.reload_type,
      reload_file_path: reload.file_reload_spec.reload_file_path,
      deploy_type: "common",
      memo: formData.value.memo
    }
    await updateApp({ id, biz_id, data })
    isMemoEdit.value = false
  }

  const handleClose = () => {
    emits('update:show', false)
  }

</script>
<template>
      <bk-sideslider
      width="400"
      quick-close
      :is-Show="props.show"
      :title="t('服务属性')"
      :before-close="handleClose">
      <template #header>
        <div class="service-edit-head">
          <span class="title">{{ t("服务属性") }}</span>
          <router-link class="credential-btn" :to="{ name: 'credentials-management' }">服务密钥</router-link>
        </div>
      </template>
      <div class="service-edit-wrapper">
        <bk-form ref="formRef" :model="formData" label-width="100" :rules="rules">
          <bk-form-item :label="t('服务名称')">{{ props.service.spec.name }}</bk-form-item>
          <bk-form-item :label="t('所属业务')">{{ spaceName }}</bk-form-item>
          <bk-form-item :label="t('服务描述')" property="memo" error-display-type="tooltips">
            <div class="content-edit">
              <template v-if="isMemoEdit">
                <bk-input
                  ref="memoRef"
                  v-model="formData.memo"
                  type="textarea"
                  :show-word-limit="true"
                  :maxlength="255"
                  :rows="5"
                  @blur="handleUpdateMemo">
                </bk-input>
              </template>
              <template v-else>
                {{ formData.memo || '--' }}
                <i class="bk-bscp-icon icon-edit-small edit-icon" @click="handleEditMemo"></i>
              </template>
            </div>
          </bk-form-item>
          <bk-form-item :label="t('接入方式')">
            {{ props.service.spec.config_type }}-{{ props.service.spec.deploy_type }}
          </bk-form-item>
          <bk-form-item :label="t('创建者')">
            {{ props.service.revision.creator}}
          </bk-form-item>
          <bk-form-item :label="t('创建时间')">
            {{ props.service.revision.create_at }}
          </bk-form-item>
        </bk-form>
      </div>
    </bk-sideslider>
</template>
<style lang="scss" scoped>
  .service-edit-head {
    display: flex;
    align-content: center;
    justify-content: space-between;
    padding-right: 24px;
    .credential-btn {
      font-size: 12px;
      color: #3a84ff;
    }
  }
  .service-edit-wrapper {
    padding: 20px 24px;
    font-size: 12px;
    :deep(.bk-form-item) {
      margin-bottom: 16px;
      .bk-form-label,
      .bk-form-content {
        line-height: 16px;
        font-size: 12px;
      }
      .bk-form-label {
        color: #979ba5;
      }
      .bk-form-content {
        color: #63656e;
      }
    }
    .content-edit {
      position: relative;
      // padding-right: 16px;
      .edit-icon {
        position: absolute;
        right: -20px;
        top: -3px;
        font-size: 22px;
        color: #979ba5;
        cursor: pointer;
        &:hover {
          color: #3a84ff;
        }
      }
      
    }
  }
</style>