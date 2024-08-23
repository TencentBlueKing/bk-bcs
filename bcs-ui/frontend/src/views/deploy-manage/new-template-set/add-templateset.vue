<template>
  <BcsContent>
    <template #title>
      <i class="bcs-icon bcs-icon-arrows-left back mr-[4px]" @click="back"></i>
      <span>{{ formData.name }}</span>
      <span class="bcs-icon-btn ml-[10px]" @click="showSetMetadata = true">
        <i class="bk-icon icon-edit-line"></i>
      </span>
    </template>
    <!-- 模板文件 -->
    <div class="px-[24px] py-[12px] bg-[#fff] shadow-[0_2px_4px_0_rgba(25,25,41,0.05)]">
      <div class="flex items-center justify-between">
        <span class="text-[14px] font-bold">{{ $t('templateSet.label.templateFile') }}</span>
        <bcs-button text>
          <span class="flex items-center text-[12px]">
            <i class="bcs-icon bcs-icon-md mr-[4px]"></i>
            README
          </span>
        </bcs-button>
      </div>
      <div class="flex items-center justify-between mt-[16px] mb-[16px]">
        <PopoverSelector offset="0, 8">
          <bcs-button theme="primary" outline>{{ $t('templateSet.button.addFile') }}</bcs-button>
          <template #content>
            <ul>
              <li class="bcs-dropdown-item" @click="chooseTemplateFile">{{ $t('templateSet.button.existFile') }}</li>
              <li class="bcs-dropdown-item">{{ $t('templateSet.button.createFile') }}</li>
            </ul>
          </template>
        </PopoverSelector>
        <bcs-input
          class="w-[360px]"
          :placeholder="$t('templateSet.placeholder.searchFile')"
          right-icon="bk-icon icon-search" />
      </div>
      <bcs-exception type="empty" scene="part" v-if="!fileListByKind.length">
        <p class="text-[14px]">{{ $t('templateSet.tips.emptyTemplateFile') }}</p>
        <bcs-button text class="mt-[8px]">
          <span class="text-[12px]">{{ $t('templateSet.button.addNow') }}</span>
        </bcs-button>
      </bcs-exception>
      <!-- 根据kind分组模板文件 -->
      <LayoutGroup
        v-for="item in fileListByKind"
        :key="item.kind"
        :title="item.kind"
        collapsible>
        <div class="grid grid-cols-4 gap-[16px]">
          <div
            :class="[
              'group flex items-center px-[16px] h-[56px] relative',
              'hover:bg-[#F5F7FA] cursor-pointer'
            ]"
            v-for="file in item.data"
            :key="file.id"
            @click="handleShowFileDetail(file)">
            <i
              :class="[
                'absolute right-[-2px] top-[-2px] opacity-0',
                'bcs-icon bcs-icon-close-circle-shape text-[14px] text-[#979BA5]',
                'group-hover:opacity-100'
              ]"
              @click.stop="removeFile(file.id)"></i>
            <i class="flex items-center justify-center w-[32px] h-[32px] bg-[#D8D8D8]"></i>
            <span class="flex flex-col ml-[8px] text-[12px]">
              <span class="leading-[20px]">{{ file.name }}</span>
              <span class="leading-[20px] text-[#979BA5]">{{ file.version }}</span>
            </span>
          </div>
        </div>
      </LayoutGroup>
    </div>
    <!-- 选择已有的模板文件 -->
    <bcs-dialog
      v-model="showFileDialog"
      header-position="left"
      :title="$t('templateSet.title.addTemplateFile')"
      @confirm="handleAddTemplateFile">
      <!-- bk-form改变了this指向，会导致ref获取不到，后续再修复 -->
      <bcs-form form-type="vertical">
        <bcs-form-item :label="$t('templateSet.label.templateFile')" required>
          <TemplateFileSelect :key="showFileDialog" :value="fileData" ref="fileSelectRef" />
        </bcs-form-item>
      </bcs-form>
    </bcs-dialog>
    <!-- 基本信息 -->
    <MetadataDialog
      :value="showSetMetadata"
      :data="formData"
      @cancel="showSetMetadata = false"
      @confirm="handleSetMetadata" />
    <!-- 模板文件详情 -->
    <bcs-sideslider
      :is-show.sync="showFileDetail"
      quick-close
      :title="fileMetadata?.name"
      :width="960">
      <template #content>
        <TemplateFileDetail
          :template-space="fileMetadata?.templateSpace"
          :id="fileMetadata?.id"
          :version="fileMetadata?.version"
          class="px-[24px] py-[20px]" />
      </template>
    </bcs-sideslider>
  </BcsContent>
</template>
<script setup lang="ts">
import { cloneDeep } from 'lodash';
import { computed, ref } from 'vue';

import TemplateFileDetail from '../template-file/file-detail.vue';

import MetadataDialog, { FormValue } from './set-metadata.vue';
import TemplateFileSelect from './template-file-select.vue';

import { IListTemplateMetadataItem } from '@/@types/cluster-resource-patch';
import BcsContent from '@/components/layout/Content.vue';
import PopoverSelector from '@/components/popover-selector.vue';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import LayoutGroup from '@/views/cluster-manage/components/layout-group.vue';

const showSetMetadata = ref(false);
const formData = ref<CreateTemplateSetReq>({
  name: `${$i18n.t('templateSet.label.set')}-${Math.floor(Date.now() / 1000)}`,
  description: '',
  version: '',
  category: '',
  keywords: [],
  readme: '',
  templates: [],
  values: '',
  force: false,
});

// 设置基本信息
const handleSetMetadata = (data: FormValue) => {
  formData.value = {
    ...formData.value,
    ...data,
  };
  showSetMetadata.value = false;
};

// 选择模板文件
const showFileDialog = ref(false);
const fileSelectRef = ref<InstanceType<typeof TemplateFileSelect>>();
const fileData = ref<IListTemplateMetadataItem[]>([]);
const fileListByKind = computed(() => fileData.value.reduce<Array<{
  kind: string
  data: IListTemplateMetadataItem[]
}>>((pre, item) => {
  const kind = item.resourceType;
  const kindData = pre.find(d => d.kind === kind);
  if (kindData) {
    kindData.data.push(item);
  } else {
    pre.push({
      kind,
      data: [item],
    });
  }
  return pre;
}, []));
const chooseTemplateFile = () => {
  showFileDialog.value = true;
};
const handleAddTemplateFile = () => {
  fileData.value = cloneDeep(fileSelectRef.value?.fileList || []);
};

// 移除模板文件
const removeFile = (id: string) => {
  const index = fileData.value.findIndex(item => item.id === id);
  if (index > -1) {
    fileData.value.splice(index, 1);
  }
};

// 回退
const back = () => {
  $router.back();
};

// 模板文件详情
const showFileDetail = ref(false);
const fileMetadata = ref<IListTemplateMetadataItem>();
const handleShowFileDetail = async (file: IListTemplateMetadataItem) => {
  fileMetadata.value = file;
  showFileDetail.value = true;
};
</script>
