<template>
  <bcs-select
    class="cluster-select"
    :value="localValue"
    :size="size"
    :clearable="false"
    :popover-min-width="320"
    :search-placeholder="$t('cluster.placeholder.searchCluster')"
    :loading="clusterGroupLoading">
    <bcs-option-group
      v-for="item, index in clusterGroups"
      :key="item.type"
      :name="item.title"
      :is-collapse="collapseList.includes(item.type)"
      :class="[
        'mt-[8px]',
        index === (clusterGroups.length - 1) ? 'mb-[4px]' : ''
      ]">
      <template #group-name>
        <CollapseTitle
          :title="`${item.title} (${item.list.length})`"
          :collapse="collapseList.includes(item.type)"
          @click="handleToggleCollapse(item.type)" />
      </template>
      <bcs-option
        v-for="cluster in item.list"
        :key="cluster.storage_cluster_id"
        :id="cluster.storage_cluster_id"
        :name="cluster.storage_cluster_name">
        <div
          class="flex items-center justify-between"
          @click.stop="handleClusterChange(cluster.storage_cluster_id)">
          <span class="bcs-ellipsis" v-bk-overflow-tips>{{ cluster.storage_cluster_name }}</span>
          <span class="text-[#979BA5] bcs-ellipsis">
            {{
              $t(
                'metrics.usage',
                [`${cluster.storage_usage}% (${
                  formatBytes((cluster.storage_usage / 100) * cluster.storage_total, 0)}/${
                  formatBytes(cluster.storage_total, 0)})`]
              )
            }}
          </span>
        </div>
      </bcs-option>
    </bcs-option-group>
    <template #extension>
      <SelectExtension
        :link-text="$t('logCollector.button.goBKLog')"
        :link="link"
        @refresh="handleGetLogCollectorClusterGroups" />
    </template>
  </bcs-select>
</template>
<script setup lang="ts">
import { computed, ref, watch } from 'vue';

import useLog, { IClusterGroup } from './use-log';

import $bkMessage from '@/common/bkmagic';
import { formatBytes } from '@/common/util';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import CollapseTitle from '@/components/cluster-selector/collapse-title.vue';
import SelectExtension from '@/components/select-extension.vue';
import { useProject } from '@/composables/use-app';
import $i18n from '@/i18n/i18n-setup';

type ClusterType = 'platform' | 'user';

interface IGroupData {
  list: IClusterGroup[]
  title: string
  type: ClusterType// 公共集群或者业务独享集群
}

const props = defineProps({
  clusterId: {
    type: String,
    default: '',
  },
  size: {
    type: String,
    default: '',
  },
});
const emits = defineEmits(['change']);

// es集群管理链接
const { projectCode } = useProject();
const link = computed(() => `${window.BK_LOG_HOST}/#/manage/es-cluster-manage?spaceUid=bkci__${projectCode.value}`);

const { getLogCollectorClusterGroups, switchStorageCluster } = useLog();

const localValue = ref<number|string>('');
// 折叠分组
const collapseList = ref<Array<ClusterType>>([]);
const handleToggleCollapse = (type: ClusterType) => {
  const index = collapseList.value.findIndex(item => item === type);
  if (index > -1) {
    collapseList.value.splice(index, 1);
  } else {
    collapseList.value.push(type);
  }
};
// ES存储集群列表
const clusterList = ref<IClusterGroup[]>([]);
const clusterGroups = computed<IGroupData[]>(() => clusterList.value.reduce<IGroupData[]>((data, item) => {
  // ES集群分组
  if (item.is_platform) {
    const group = data.find(item => item.type === 'platform');
    if (group) {
      group.list.push(item);
    } else {
      data.push({
        title: $i18n.t('bcs.cluster.share'),
        type: 'platform',
        list: [item],
      });
    }
  } else {
    const group = data.find(item => item.type === 'user');
    if (group) {
      group.list.push(item);
    } else {
      data.push({
        title: $i18n.t('logCollector.tips.bizPrivateCluster'),
        type: 'user',
        list: [item],
      });
    }
  }
  return data;
}, []));
const clusterGroupLoading = ref(false);
const handleGetLogCollectorClusterGroups = async () => {
  if (!props.clusterId) {
    localValue.value = '';
    clusterList.value = [];
    return;
  };

  clusterGroupLoading.value = true;
  clusterList.value = await getLogCollectorClusterGroups(props.clusterId);
  clusterGroupLoading.value = false;
  // 设置当前es集群信息
  const esCluster = clusterList.value.find(item => item.is_selected);
  localValue.value = esCluster?.storage_cluster_id || '';
  emits('change', esCluster);
};

// 切换ES集群
const handleClusterChange = (storage_cluster_id: string|number) => {
  if (storage_cluster_id === localValue.value) return;

  const data = clusterList.value.find(item => item.storage_cluster_id === storage_cluster_id);
  $bkInfo({
    type: 'warning',
    clsName: 'custom-info-confirm',
    title: $i18n.t('logCollector.title.confirmSwitchStorageCluster'),
    subTitle: data?.storage_cluster_name,
    defaultInfo: true,
    confirmFn: async () => {
      const result = await switchStorageCluster(props.clusterId, storage_cluster_id);
      if (result) {
        localValue.value = data?.storage_cluster_id || '';
        $bkMessage({
          theme: 'success',
          message: $i18n.t('generic.msg.success.ok'),
        });
      }
    },
  });
  emits('change', data);
};

watch(() => props.clusterId, () => {
  handleGetLogCollectorClusterGroups();
}, { immediate: true });
</script>

<style lang="postcss" scoped>
.cluster-select {
    width: 254px;
    &:not(.is-disabled) {
      background-color: #fff;
    }
    >>> .bk-select-name {
      text-align: left;
    }
    >>> .bk-select-loading {
      top: 5px;
    }
}
/deep/ .bk-option-group-name {
  border-bottom: 0 !important;
}
.bk-options .bk-option:first-child {
  margin-top: 0;
}
.bk-options .bk-option:last-child {
  margin-bottom: 0;
}
</style>
