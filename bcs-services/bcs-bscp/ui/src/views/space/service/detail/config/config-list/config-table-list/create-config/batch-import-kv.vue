<template>
  <bk-dialog
    :is-show="props.show"
    :title="'批量导入'"
    :theme="'primary'"
    width="960"
    height="720"
    ext-cls="variable-import-dialog"
    :esc-close="false"
    @closed="handleClose"
  >
    <bk-form>
      <!-- <bk-form-item label="导入方式">
        <bk-radio-group v-model="importType">
          <bk-radio label="text">文本导入</bk-radio>
          <bk-radio label="file">文件导入</bk-radio>
        </bk-radio-group>
        <div class="tips" v-if="importType === 'text'">只支持string、number类型,其他类型请使用文件导入</div>
      </bk-form-item> -->
      <bk-form-item label="配置文件内容" required>
        <KvContentEditor
          v-if="importType === 'text'"
          ref="editorRef"
          @trigger="confirmBtnPerm = $event"
        />
        <bk-upload v-else with-credentials>
          <template #tip>
            <div class="upload-tips">
              <span>支持 Excel、JSON、XML、YAML 等类型文件，后台会自动检测文件类型，配置项格式请参照</span>
              <span class="sample">示例文件包</span>
            </div>
          </template>
        </bk-upload>
      </bk-form-item>
    </bk-form>
    <template #footer>
      <bk-button theme="primary" style="margin-right: 8px" :disabled="!confirmBtnPerm" @click="handleConfirm"
        >导入</bk-button>
      <bk-button @click="handleClose">取消</bk-button>
    </template>
  </bk-dialog>
</template>

<script lang="ts" setup>
import { ref, watch } from 'vue';
import KvContentEditor from '../../../components/kv-import-editor.vue';
const props = defineProps<{
  show: boolean;
  bkBizId: string;
  appId: number;
}>();
const emits = defineEmits(['update:show', 'confirm']);

const editorRef = ref();
const isFormChange = ref(false);
const importType = ref('text');
const confirmBtnPerm = ref(false);
watch(
  () => props.show,
  () => {
    isFormChange.value = false;
  },
);
const handleClose = () => {
  emits('update:show', false);
};
const handleConfirm = async () => {
  if (importType.value === 'file') return;
  await editorRef.value.handleImport();
  emits('update:show', false);
  emits('confirm');
};
</script>

<style scoped lang="scss">
.tips {
  font-size: 12px;
  color: #979ba5;
}
.upload-tips {
  font-size: 12px;
  color: #63656e;
  .sample {
    margin-left: 2px;
    color: #3a84ff;
    cursor: pointer;
  }
}
</style>

<style lang="scss">
.variable-import-dialog {
  .bk-modal-content {
    overflow: hidden !important;
  }
}
</style>
