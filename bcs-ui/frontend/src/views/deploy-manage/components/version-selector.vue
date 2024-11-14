<template>
  <bcs-select
    class="bg-[#fff]"
    :loading="versionLoading"
    :clearable="false"
    :popover-min-width="360"
    searchable
    :value="value"
    @change="handleVersionChange">
    <bcs-option
      v-for="item in versionList"
      :key="item.id"
      :id="item.id"
      :name="item.draft ?
        (item.version ?
          `${$t('templateFile.tag.draft')}( ${$t('templateFile.tag.baseOn')} ${item.version} )`
          : $t('templateFile.tag.draft'))
        : item.version">
      <div>
        <div class="flex items-center whitespace-nowrap bcs-ellipsis justify-between">
          <span class="flex items-center mr-[10px]">
            <span v-bk-overflow-tips class="mr-[8px] bcs-ellipsis">{{ item.draft ?
              (item.version ?
                `${$t('templateFile.tag.draft')}( ${$t('templateFile.tag.baseOn')} ${item.version} )`
                : $t('templateFile.tag.draft'))
              : item.version }}</span>
            <bcs-tag
              class="px-[6px] h-[20px] leading-[20px] flex-shrink-0"
              theme="warning"
              v-if="item.latest">Latest</bcs-tag>
            <bcs-tag
              v-if="item.version && item.latestDeployVersion === item.version"
              class="px-[6px] h-[20px] leading-[20px] flex-shrink-0">LatestDeployed</bcs-tag>
          </span>
          <span class="text-[#a2a6b0]">{{ (item?.createAt && formatDate(item.createAt * 1000)) || '--' }}</span>
        </div>
        <span
          class="text-[#a2a6b0] bcs-ellipsis"
          v-bk-overflow-tips>{{ item?.description || '--' }}</span>
      </div>
    </bcs-option>
    <template #extension v-if="showExtension">
      <SelectExtension
        :link-text="$t('templateFile.button.versionManage')"
        @link="handleLink"
        @refresh="listTemplateVersion" />
    </template>
  </bcs-select>
</template>
<script setup lang="ts">
import { computed, onActivated, ref, watch } from 'vue';

import { ITemplateVersionItem } from '@/@types/cluster-resource-patch';
import { TemplateSetService  } from '@/api/modules/new-cluster-resource';
import { formatDate } from '@/common/util';
import SelectExtension from '@/components/select-extension.vue';

const props = defineProps({
  value: {
    type: [String, Array],
    default: '',
  },
  id: {
    type: String,
    default: '',
  },
  onDraft: {
    type: Boolean,
    default: false,
  },
  showExtension: {
    type: Boolean,
    default: false,
  },
});

const emits = defineEmits(['change', 'input', 'link']);

// 版本列表
const versionLoading = ref(false);
const curVersionData = computed(() => versionList.value.find(item => item.id === props.value));
const versionList = ref<ITemplateVersionItem[]>([]);
const listTemplateVersion = async () => {
  if (!props.id) return;
  versionLoading.value = true;
  versionList.value = await TemplateSetService.ListTemplateVersion({
    $templateID: props.id,
  }).catch(() => []);
  // 筛选出非草稿版本
  if (props.onDraft) {
    versionList.value = versionList.value.filter(item => !item.draft);
  }
  // 版本不对时重置为第一个
  if (!versionList.value.find(item => item.id === props.value)) {
    // 检查数组是否为空,防止导致 handleVersionChange 接收到 undefined 参数。
    versionList.value.length && handleVersionChange(versionList.value.at(0)?.id);
  }
  versionLoading.value = false;
};

function handleVersionChange(val) {
  emits('input', val);
}

function handleLink() {
  emits('link');
}

watch(curVersionData, () => {
  curVersionData.value && emits('change', curVersionData.value);
});

watch(() => props.id, () => {
  listTemplateVersion();
}, { immediate: true });

defineExpose({
  listTemplateVersion,
});

onActivated(() => {
  listTemplateVersion();
});
</script>
