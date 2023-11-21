<template>
  <bk-sideslider title="导入配置文件" :width="960" :is-show="isShow" :before-close="handleBeforeClose" @closed="close">
    <div class="slider-content-container">
      <bk-form form-type="vertical">
        <bk-form-item label="上传配置包" required property="package">
          <bk-upload
            v-show="!isTableChange"
            class="config-uploader"
            theme="button"
            tip="支持扩展名：.zip  .tar  .gz"
            :size="100"
            :multiple="false"
            accept=".zip, .tar, .gz"
            :custom-request="handleFileUpload"
          >
            <template #trigger>
              <div ref="buttonRef">
                <bk-button class="upload-button">
                  <upload />
                  <span class="text">上传文件</span>
                </bk-button>
              </div>
            </template>
          </bk-upload>
          <div v-show="isTableChange">
            <bk-pop-confirm
              title="确认放弃下方修改，重新上传配置项包？"
              trigger="click"
              @confirm="() => buttonRef.click()"
            >
              <bk-button class="upload-button">
                <upload />
                <span class="text">重新上传</span>
              </bk-button>
              <span class="upload-tips">支持扩展名：.zip .tar .gz</span>
            </bk-pop-confirm>
          </div>
        </bk-form-item>
      </bk-form>
      <span v-if="loading" style="color: #63656e">上传中...</span>
      <bk-loading :loading="loading">
        <div class="tips" v-if="loading">
          共将导入 <span>{{ importConfigList.length }}</span> 个配置项，其中
          <span>{{ existConfigList.length }}</span> 个已存在,导入后将<span style="color: #ff9c01">覆盖原配置</span>
        </div>
        <ConfigTable
          :table-data="nonExistConfigList"
          :is-exsit-table="false"
          v-if="nonExistConfigList.length"
          :expand="expandNonExistTable"
          @change-expand="expandNonExistTable = !expandNonExistTable"
          @change="handleTableChange($event,true)"
        />
        <ConfigTable
          :table-data="existConfigList"
          :is-exsit-table="true"
          v-if="existConfigList.length"
          :expand="!expandNonExistTable"
          @change-expand="expandNonExistTable = !expandNonExistTable"
          @change="handleTableChange($event,false)"
        />
      </bk-loading>
    </div>
    <div class="action-btns">
      <bk-button
        theme="primary"
        :loading="pending"
        :disabled="!importConfigList.length"
        @click="isSelectPkgDialogShow = true"
        >去导入</bk-button
      >
      <bk-button @click="close">取消</bk-button>
    </div>
  </bk-sideslider>
  <SelectPackage v-model:show="isSelectPkgDialogShow" :pending="pending" @confirm="handleImport" />
</template>
<script lang="ts" setup>
import { ref, watch, computed } from 'vue';
import { storeToRefs } from 'pinia';
import useGlobalStore from '../../../../../../../../store/global';
import useTemplateStore from '../../../../../../../../store/template';
import useModalCloseConfirmation from '../../../../../../../../utils/hooks/use-modal-close-confirmation';
import { IConfigImportItem } from '../../../../../../../../../types/config';
import { importTemplateFile, importTemplateBatchAdd, addTemplateToPackage } from '../../../../../../../../api/template';
import ConfigTable from './config-table.vue';
import SelectPackage from './select-package.vue';
import { Message } from 'bkui-vue';
import { Upload } from 'bkui-vue/lib/icon';
const props = defineProps<{
  show: boolean;
}>();

const emits = defineEmits(['update:show', 'added']);
const { spaceId } = storeToRefs(useGlobalStore());
const { currentTemplateSpace } = storeToRefs(useTemplateStore());
const isShow = ref(false);
const isTableChange = ref(false);
const pending = ref(false);
const existConfigList = ref<IConfigImportItem[]>([]);
const nonExistConfigList = ref<IConfigImportItem[]>([]);
const loading = ref(false);
const expandNonExistTable = ref(true);
const isSelectPkgDialogShow = ref(false);
const buttonRef = ref();

watch(
  () => props.show,
  (val) => {
    isShow.value = val;
    isTableChange.value = false;
  },
);

const importConfigList = computed(() => [...existConfigList.value, ...nonExistConfigList.value]);

const handleFileUpload = async (option: { file: File }) => {
  clearData();
  loading.value = true;
  const res = await importTemplateFile(spaceId.value, currentTemplateSpace.value, option.file);
  existConfigList.value = res.exist;
  nonExistConfigList.value = res.non_exist;
  nonExistConfigList.value.forEach((item: IConfigImportItem) => {
    item.privilege = '677';
    item.user = 'root';
    item.user_group = 'root';
  });
  if (nonExistConfigList.value.length === 0) expandNonExistTable.value = false;
  isTableChange.value = false;
  loading.value = false;
};

const handleBeforeClose = async () => {
  if (isTableChange.value) {
    const result = await useModalCloseConfirmation();
    return result;
  }
  return true;
};

const close = () => {
  clearData();
  emits('update:show', false);
};

const handleImport = async (pkgIds: number[]) => {
  try {
    const res = await importTemplateBatchAdd(spaceId.value, currentTemplateSpace.value, [
      ...existConfigList.value,
      ...nonExistConfigList.value,
    ]);
    // 选择未指定套餐时,不需要调用添加接口
    if (pkgIds.length > 1 || pkgIds[0] !== 0) {
      await addTemplateToPackage(spaceId.value, currentTemplateSpace.value, res.ids, pkgIds);
    }
    isSelectPkgDialogShow.value = false;
    emits('added');
    close();
    Message({
      theme: 'success',
      message: '导入配置文件成功',
    });
  } catch (e) {
    console.log(e);
  } finally {
    pending.value = false;
  }
};

const handleTableChange = (data: IConfigImportItem[], isNonExistData: boolean) => {
  if (isNonExistData) {
    nonExistConfigList.value = data;
  } else {
    existConfigList.value = data;
  }
  isTableChange.value = true;
};

const clearData = () => {
  nonExistConfigList.value = [];
  existConfigList.value = [];
};
</script>
<style lang="scss" scoped>
.slider-content-container {
  padding: 20px 40px;
  height: calc(100vh - 101px);
}
.upload-button {
  width: 100px;
  .text {
    margin-left: 5px;
  }
}
.upload-tips {
  margin-left: 8px;
  font-size: 12px;
  color: #63656e;
}
.action-btns {
  border-top: 1px solid #dcdee5;
  padding: 8px 24px;
  .bk-button {
    margin-right: 8px;
    min-width: 88px;
  }
}
.config-uploader {
  :deep(.bk-upload-list) {
    display: none;
  }
}
.tips {
  color: #63656e;
  margin-bottom: 16px;
  span {
    color: #313238;
  }
}
</style>
