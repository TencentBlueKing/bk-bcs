<template>
  <bcs-select
    searchable
    :clearable="false"
    :loading="osLoading"
    :value="value"
    :disabled="disabled"
    @change="handleImageChange">
    <bcs-option-group
      v-for="group in imageListByGroup"
      :key="group.provider"
      :name="group.name"
      :is-collapse="collapseList.includes(group.provider)"
      :class="[
        'mt-[8px]'
      ]">
      <template #group-name>
        <CollapseTitle
          :title="`${group.name} (${group.children.length})`"
          :collapse="collapseList.includes(group.provider)"
          @click="handleToggleCollapse(group.provider)" />
      </template>
      <bcs-option
        v-for="item in group.children"
        :key="item.imageID"
        :id="item.imageID"
        :name="item.alias">
        <div
          class="flex items-center justify-between"
          v-bk-tooltips="{
            content: item.clusters ? item.clusters.map(v => v.clusterID).join(',') : '--',
            disabled: !item.clusters.length
          }">
          <span class="flex items-center">
            {{ `${item.alias}(${item.imageID})` }}
            <bcs-tag
              theme="info"
              radius="45px"
              v-if="item.alias === recDefaultImage && item.provider === 'PUBLIC_IMAGE'">
              {{ $t('tke.label.recommended') }}
            </bcs-tag>
          </span>
          <span v-if="item.clusters.length" class="text-[#979BA5]">
            {{ $t('tke.tips.imageUsedInCluster') }}
          </span>
        </div>
      </bcs-option>
    </bcs-option-group>
  </bcs-select>
</template>
<script setup lang="ts">
import { computed, onBeforeMount, ref, watch } from 'vue';

import { cloudOsImage } from '@/api/modules/cluster-manager';
import CollapseTitle from '@/components/cluster-selector/collapse-title.vue';
import $i18n from '@/i18n/i18n-setup';
import $store from '@/store';
import { IImageGroup, IImageItem } from '@/views/cluster-manage/types/types';

const props = defineProps({
  value: {
    type: String,
  },
  region: {
    type: String,
    default: '',
  },
  cloudAccountID: {
    type: String,
    default: '',
  },
  cloudID: {
    type: String,
    default: '',
  },
  disabled: {
    type: Boolean,
    default: false,
  },
  initData: {
    type: Boolean,
    default: false,
  },
});
const emits = defineEmits(['input', 'change', 'os-change']);

// 折叠组
const collapseList = ref<string[]>([]);
const handleToggleCollapse = (provider: string) => {
  const index = collapseList.value.findIndex(item => item === provider);
  if (index > -1) {
    collapseList.value.splice(index, 1);
  } else {
    collapseList.value.push(provider);
  }
};
const imageList = ref<Array<IImageItem>>($store.state.cloudMetadata.osList);
const providerMap = {
  PUBLIC_IMAGE: $i18n.t('tke.label.publicImage'),
  MARKET_IMAGE: $i18n.t('tke.label.marketImage'),
  PRIVATE_IMAGE: $i18n.t('tke.label.privateImage'),
};
const imageListByGroup = computed<Record<string, IImageGroup>>(() => imageList.value
  .sort((pre, current) => {
    if (pre.alias === recDefaultImage && current.alias !== recDefaultImage) return -1;

    if (pre.alias !== recDefaultImage && current.alias === recDefaultImage) return 1;

    if (pre.alias === recDefaultImage && current.alias === recDefaultImage) return 0;

    return pre.alias.localeCompare(current.alias);
  })
  .reduce((pre, item) => {
    if (!pre[item.provider]) {
      pre[item.provider] = {
        name: providerMap[item.provider],
        provider: item.provider,
        children: [item],
      };
    } else {
      pre[item.provider].children.push(item);
    }
    return pre;
  }, {}));

// 镜像列表
const recDefaultImage = 'TencentOS Server 3.1 (TK4)';
const osLoading = ref(false);
const handleGetOsList = async () => {
  if (!props.region || !props.cloudAccountID || !props.cloudID) return;
  osLoading.value = true;
  imageList.value = await cloudOsImage({
    $cloudId: props.cloudID,
    accountID: props.cloudAccountID,
    region: props.region,
    provider: 'ALL',
  }).catch(() => []);
  $store.commit('cloudMetadata/updateOsList', imageList.value);
  // 设置默认镜像
  const defaultImageID = imageList.value
    .find(item => item.alias === recDefaultImage && item.provider === 'PUBLIC_IMAGE')?.imageID || '';
  handleImageChange(defaultImageID);
  osLoading.value = false;
};
const handleImageChange = (imageID: string) => {
  const imageItem = imageList.value.find(item => item.imageID === imageID);
  if (!imageItem) return;

  if (imageItem.provider === 'PRIVATE_IMAGE') {
    // 私有镜像取ID
    emits('os-change', imageItem.imageID);
  } else {
    // 其他镜像取名称
    emits('os-change', imageItem.osName);
  }
  emits('change', imageID);
  emits('input', imageID);
  $store.commit('cloudMetadata/updateImageID', imageID);
};

watch([
  () => props.region,
  () => props.cloudAccountID,
], () => {
  handleGetOsList();
});

onBeforeMount(() => {
  props.initData && handleGetOsList();
});
</script>
