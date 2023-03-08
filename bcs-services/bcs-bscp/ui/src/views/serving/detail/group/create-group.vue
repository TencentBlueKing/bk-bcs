<script setup lang="ts">
import { ref } from 'vue'
  import { IGroupEditing, ECategoryType, ICategoryGroup } from '../../../../../types/group'

  const props = defineProps<{
    show: boolean,
    categoryList: Array<ICategoryGroup>
  }>()

  const emits = defineEmits(['update:show'])

  const formRef = ref()
  const pending = ref(false)
  const mode = {
    custom: ECategoryType.Custom,
    debug: ECategoryType.Debug
  }
  const formData = ref<IGroupEditing>({
    app_id: '',
    group_category_id: '',
    name: '',
    mode: ECategoryType.Custom,
  });
  const rules = {
    name: [
      {
        validator: (value: string) => value.length < 128,
        message: '最大长度128个字符'
      },
      {
        validator: (value: string) => {
          if (value.length > 0) {
            return /^[\u4e00-\u9fa5a-zA-Z0-9][\u4e00-\u9fa5a-zA-Z0-9_\-]*[\u4e00-\u9fa5a-zA-Z0-9]?$/.test(value)
          }
          return true
        },
        message: '仅允许使用中文、英文、数字、下划线、中划线，且必须以中文、英文、数字开头和结尾'
      }
    ]
  }


  const handleConfirm = async () => {
    await formRef.value.validate()
    handleClose()
  }

  const handleClose = () => {
    emits('update:show', false)
  }

</script>
<template>
  <bk-dialog
    title="创建分组"
    ext-cls="create-group-dialog"
    :is-show="props.show"
    :width="960"
    :esc-close="false"
    :quick-close="false"
    :is-loading="pending"
    @closed="handleClose"
    @confirm="handleConfirm">
      <bk-alert theme="info" title="同分类下的不同分组之间需要保证没有交集，否则客户端将产生配置错误的风险！" />
      <div class="config-wrapper">
        <bk-form class="form-content-area" ref="formRef" :model="formData" :rules="rules">
          <bk-form-item label="分组名称" property="name" required>
            <bk-input v-model="formData.name" placeholder="请输入"></bk-input>
          </bk-form-item>
          <bk-form-item label="分组标签" property="group_category_id" required>
            <bk-select v-model="formData.group_category_id" placeholder="请选择">
              <bk-option
                v-for="category in categoryList"
                :key="category.config.id"
                :value="category.config.id"
                :label="category.config.spec.name">
              </bk-option>
            </bk-select>
          </bk-form-item>
          <bk-form-item label="调试用分组" required>
            <bk-switcher v-model="formData.mode" :true-value="mode.debug" :false-value="mode.custom"></bk-switcher>
            <p class="debug-tips">启用调试用分组后，仅可使用 UID 作为分组规则，且配置版本将不跟随主线</p>
          </bk-form-item>
          <bk-form-item label="分组规则" required>
            <div class="rule-config">
              <bk-input style="width: 80px;" placeholder=""></bk-input>
              <bk-select style="width: 72px;"></bk-select>
              <bk-input style="width: 120px;" placeholder=""></bk-input>
            </div>
          </bk-form-item>
        </bk-form>
        <div class="group-intersection-detect">
          <h4 class="title">分组交集检测</h4>
          <bk-table class="rule-table" :border="['outer']">
            <bk-table-column label="分组名称"></bk-table-column>
            <bk-table-column label="分组规则"></bk-table-column>
          </bk-table>
        </div>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <bk-button theme="primary" :loading="pending" @click="handleConfirm">提交</bk-button>
          <bk-button :disabled="pending" @click="handleClose">取消</bk-button>
        </div>
      </template>
  </bk-dialog>
</template>
<style lang="scss" scoped>
  .config-wrapper {
    display: flex;
    padding: 24px 0 14px;
    height: 530px;
  }
  .form-content-area {
    padding: 24px 24px 24px 0;
    width: 50%;
    height: 100%;
    overflow: auto;
  }
  .debug-tips {
    margin: 0;
    font-size: 12px;
    line-height: 18px;
    color: #979ba5;
  }
  .rule-config {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
  .group-intersection-detect {
    padding: 16px;
    width: 50%;
    background: #f5f7fa;
    .title {
      margin: 0;
    }
    .rule-table {
      margin-top: 16px;
    }
  }
  .dialog-footer {
    .bk-button {
      margin-left: 8px;
    }
  }
</style>
<style lang="scss">
  .create-group-dialog.bk-dialog-wrapper {
    .bk-dialog-header {
      padding-bottom: 20px;
    }
    .bk-modal-footer {
      height: auto;
      padding: 8px 24px;
    }
  }
    
</style>
