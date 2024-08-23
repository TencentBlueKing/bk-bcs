<template>
  <bcs-select
    v-model="localValue"
    searchable
    multiple
    :popover-min-width="320"
    :scroll-height="460"
    :loading="isLoading"
    @change="handleFileChange">
    <bcs-option-group
      v-for="item, index in spaceList"
      :key="item.id"
      :name="item.name"
      :is-collapse="collapseList.includes(item.id)"
      :class="[
        'mt-[8px]',
        index === (spaceList.length - 1) ? 'mb-[4px]' : ''
      ]">
      <template #group-name>
        <CollapseTitle
          :title="`${item.name} (${spaceFileListMap[item.id]?.length})`"
          :collapse="collapseList.includes(item.id)"
          @click="handleToggleCollapse(item.id)" />
      </template>
      <bcs-option
        v-for="file in (spaceFileListMap[item.id] || [])"
        :key="file.id"
        :id="file.id"
        :name="file.name" />
    </bcs-option-group>
  </bcs-select>
</template>
<script setup lang="ts">
import { isEqual } from 'lodash';
import { computed, onBeforeMount, ref, watch } from 'vue';

import { IListTemplateMetadataItem, ITemplateSpaceData } from '@/@types/cluster-resource-patch';
import { TemplateSetService } from '@/api/modules/new-cluster-resource';
import CollapseTitle from '@/components/cluster-selector/collapse-title.vue';

interface Props {
  value?: Array<IListTemplateMetadataItem>
}
type Emits = (e: 'input', v: Array<IListTemplateMetadataItem>) => void;
const props = defineProps<Props>();
const emits = defineEmits<Emits>();

// 同步值
const localValue = ref<string[]>([]);
watch(() => props.value, () => {
  const newValue = props.value?.map(v => v.id);
  if (isEqual(newValue, localValue.value)) {
    return;
  }
  localValue.value = newValue || [];
}, { immediate: true });
const fileList = computed(() => localValue.value.map(id => fileListMap.value[id]));

const collapseList = ref<string[]>([]);
const isLoading = ref(false);
// 获取命名空间列表
const spaceList = ref<ITemplateSpaceData[]>([]);
async function listTemplateSpace() {
  spaceList.value = await TemplateSetService.ListTemplateSpace().catch(() => []);
}

// 获取空间下的文件
const spaceFileListMap = ref <Record<string, IListTemplateMetadataItem[]>>({});
const fileListMap = computed(() => {
  const data: Record<string, IListTemplateMetadataItem> = {};
  Object.keys(spaceFileListMap.value).forEach((key) => {
    const fileList = spaceFileListMap.value[key];
    fileList.forEach((file) => {
      data[file.id] = file;
    });
  });
  return data;
});
async function getTemplateMetadata(spaceID: string) {
  if (!spaceID) return;

  const data = await TemplateSetService.ListTemplateMetadata({
    templateSpace: spaceID,
  }).catch(() => []);
  spaceFileListMap.value[spaceID] = data || [];

  return data;
}

// 折叠事件
const handleToggleCollapse = (id: string) => {
  const index = collapseList.value.findIndex(item => item === id);
  if (index > -1) {
    collapseList.value.splice(index, 1);
  } else {
    collapseList.value.push(id);
  }
};

// 选择模板文件
const handleFileChange = () => {
  emits('input', fileList.value);
};

defineExpose({
  fileList,
});

onBeforeMount(async () => {
  isLoading.value = true;
  await listTemplateSpace();
  await Promise.all(spaceList.value.map(item => getTemplateMetadata(item.id)));
  isLoading.value = false;
});
</script>
