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
        v-for="(item, index) in group.children"
        :key="`${index}-${item.imageID}`"
        :id="item.imageID"
        :name="item.alias">
        <div class="flex items-center">
          <span class="bcs-ellipsis" v-bk-overflow-tips="{ interactive: false }">
            {{ `${item.alias} | ${item.osName} | ${item.imageID}` }}
          </span>
          <bcs-tag
            v-if="item.clusters
              && item.clusters.find(v => v.clusterID === clusterId)
              && item.provider === 'CLUSTER_IMAGE'"
            theme="info"
            radius="45px"
            class="shrink-0">
            {{ $t('tke.tips.currentClusterUsed') }}
          </bcs-tag>
          <bcs-tag
            class="shrink-0"
            theme="info"
            radius="45px"
            v-else-if="item.alias === recDefaultImage && item.provider === 'BCS_IMAGE'">
            {{ $t('tke.label.recommended') }}
          </bcs-tag>
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
  // recDefaultImage: {
  //   type: String,
  //   default: 'TencentOS Server 3.1 (TK4)',
  // },
  // recProvider: {
  //   type: String,
  //   default: 'PUBLIC_IMAGE',
  // },
  providerOrder: {
    type: Array as () => Array<string>,
    default: () => ['CLUSTER_IMAGE', 'BCS_IMAGE', 'PUBLIC_IMAGE', 'PRIVATE_IMAGE'],
  },
  clusterId: {
    type: String,
    default: '',
  },
});
const emits = defineEmits(['input', 'change', 'os-change', 'init']);

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
  CLUSTER_IMAGE: $i18n.t('tke.label.clusterImage'),
  BCS_IMAGE: $i18n.t('tke.label.bcsImage'),
  PUBLIC_IMAGE: $i18n.t('tke.label.publicImage'),
  MARKET_IMAGE: $i18n.t('tke.label.marketImage'),
  PRIVATE_IMAGE: $i18n.t('tke.label.privateImage'),
};
const imageListByGroup = computed<Record<string, IImageGroup>>(() => {
  if (osLoading.value) return {};
  const group: Record<string, IImageGroup> = imageList.value
    .sort((pre, current) => {
      if (pre.alias === recDefaultImage.value && current.alias !== recDefaultImage.value) return -1;

      if (pre.alias !== recDefaultImage.value && current.alias === recDefaultImage.value) return 1;

      if (pre.alias === recDefaultImage.value && current.alias === recDefaultImage.value) return 0;

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
    }, {});
  if (group.BCS_IMAGE && group.PRIVATE_IMAGE) {
    // 过滤掉 PRIVATE_IMAGE 里 BCS_IMAGE 的镜像
    const bcsImage = group.BCS_IMAGE.children.reduce((pre, item) => {
      // eslint-disable-next-line no-param-reassign
      pre[item.imageID] = item;
      return pre;
    }, {});
    group.PRIVATE_IMAGE.children = group.PRIVATE_IMAGE.children.filter(item => !bcsImage[item.imageID]);
  }
  // 按照 providerOrder 排序
  const data = {};
  props.providerOrder.forEach((provider: string) => {
    group[provider] && (data[provider] = group[provider]);
  });
  return data;
});

// 镜像列表
const recDefaultImage = ref('');
const currentClusterUsedImage = ref();
const osLoading = ref(false);
const handleGetOsList = async () => {
  if (!props.region || !props.cloudID) return;
  osLoading.value = true;
  imageList.value = await cloudOsImage({
    $cloudId: props.cloudID,
    clusterID: props.clusterId,
    region: props.region,
    provider: 'ALL',
  }).catch(() => []);
  $store.commit('cloudMetadata/updateOsList', imageList.value);
  // 设置默认镜像
  if (!props.value) {
    // 当前集群镜像
    let defaultImage = imageList.value.find((item) => {
      const image = item.clusters.find(item => item.clusterID === props.clusterId);
      return !!image;
    });
    currentClusterUsedImage.value = defaultImage;
    emits('init', currentClusterUsedImage.value);
    if (!defaultImage) {
      // 平台镜像
      defaultImage = imageList.value.find(item => item.provider === 'BCS_IMAGE');
    }
    recDefaultImage.value = defaultImage?.alias || '';
    handleImageChange(defaultImage?.imageID || '');
  }
  osLoading.value = false;
};
const handleImageChange = (imageID: string) => {
  const imageItem = imageList.value.find(item => item.imageID === imageID);
  // 兼容平台镜像
  const index = imageListByGroup.value.BCS_IMAGE?.children?.findIndex?.(item => item.imageID === imageID) || -1;
  if (!imageItem) return;

  if (imageItem.provider === 'PRIVATE_IMAGE' && index < 0) {
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
], () => {
  handleGetOsList();
});

onBeforeMount(() => {
  props.initData && handleGetOsList();
});
</script>
