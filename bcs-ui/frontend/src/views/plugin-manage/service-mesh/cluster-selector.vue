<template>
  <bcs-select
    v-model="clusterValue"
    :clearable="false"
    :searchable="searchable"
    :disabled="disabled"
    :popover-min-width="320"
    :remote-method="remoteMethod"
    :search-placeholder="$t('cluster.placeholder.searchCluster')"
    :size="size"
    :scroll-height="460"
    :loading="loading"
    multiple
    :ext-cls="extCls"
    ref="clusterSelectorRef"
    @change="handleChange">
    <template #trigger v-if="trigger">
      <slot name="trigger"></slot>
    </template>
    <bcs-option
      v-for="cluster in clusterListOfMesh"
      :key="cluster.clusterID"
      :id="cluster.clusterID"
      :name="cluster.clusterName"
      :disabled="cluster.disabled">
      <div
        class="flex items-center justify-between"
        @mouseenter="hoverClusterID = cluster.clusterID"
        @mouseleave="hoverClusterID = ''">
        <div class="flex flex-col justify-center h-[50px]">
          <span class="leading-[20px] bcs-ellipsis" v-bk-overflow-tips>{{ cluster.clusterName }}</span>
          <span
            :class="[
              'leading-[20px]',
              {
                'text-[#979BA5]': normalStatusList.includes(cluster.status || ''),
                '!text-[#699DF4]': clusterValue.includes(cluster.clusterID || '') ||
                  (normalStatusList.includes(cluster.status || '')
                    && (hoverClusterID === cluster.clusterID) && !cluster.disabled),
              }]"
            v-bk-tooltips="{
              content: cluster.disabledDesc,
              disabled: !cluster.disabled,
              interactive: false
            }">
            ({{ cluster.clusterID }})
          </span>
        </div>
        <bcs-tag
          theme="danger"
          v-if="!normalStatusList.includes(cluster.status || '')">
          {{ $t('generic.label.abnormal') }}
        </bcs-tag>
      </div>
    </bcs-option>
  </bcs-select>
</template>
<script lang="ts">
import {  computed, defineComponent, onBeforeMount, PropType, ref, toRefs, watch } from 'vue';

import { meshClusters } from '@/api/modules/mesh-manager';
import { satisfiesVersion } from '@/common/util';
import { useProject } from '@/composables/use-app';
import $i18n from '@/i18n/i18n-setup';

type MeshCluster = {
  clusterID: string,
  clusterName: string,
  clusterType: string,
  isShared: boolean,
  isInstalled: boolean
  status: string,
  version: string,
};

export default defineComponent({
  name: 'ClusterSelector',
  model: {
    prop: 'value',
    event: 'change',
  },
  props: {
    value: {
      type: [String, Array] as PropType<string|string[]>,
      default: '',
    },
    searchable: {
      type: Boolean,
      default: true,
    },
    disabled: {
      type: Boolean,
      default: false,
    },
    size: {
      type: String,
      default: '',
    },
    validateClusterId: {
      type: Boolean,
      default: true,
    },
    multiple: {
      type: Boolean,
      default: false,
    },
    extCls: {
      type: String,
      default: '',
    },
    trigger: {
      type: Boolean,
      default: false,
    },
    allowedVersionRange: {
      type: String,
      default: '',
    },
  },
  emits: ['change'],
  setup(props, { emit }) {
    const { value } = toRefs(props);

    const { projectCode } = useProject();

    const normalStatusList = ['RUNNING'];
    const hoverClusterID = ref<string>();

    const clusterSelectorRef = ref<any>(null);
    const clusterValue = ref<string[]>([]);
    const clusterList = ref<MeshCluster[]>([]);
    const loading = ref(false);
    async function getClusterList() {
      loading.value = true;
      const res = await meshClusters({
        projectCode: projectCode.value,
      }).catch(() => ({ clusters: [] }));
      clusterList.value = res.clusters;
      loading.value = false;
    }

    // 不支持部署的集群
    function isDisabledCluster(item: MeshCluster) {
      return item.clusterType === 'virtual' || item.isShared || item.clusterType === 'federation';
    }
    function getDesc(item) {
      let desc = $i18n.t('serviceMesh.tips.clusterEnabled');
      if (item.clusterType === 'virtual') {
        desc = $i18n.t('serviceMesh.tips.disabledVirtual');
      } else if (item.isShared) {
        desc = $i18n.t('serviceMesh.tips.disabledShare');
      } else if (item.clusterType === 'federation') {
        desc = $i18n.t('serviceMesh.tips.disabledFederation');
      } else if (item.isInstalled) {
        desc = $i18n.t('serviceMesh.tips.isInstalled');
      }
      return desc;
    }
    const clusterListOfMesh = computed(() => clusterList.value
      .filter((item) => {
        const clusterID = item?.clusterID?.toLocaleLowerCase();
        const clusterName = item?.clusterName?.toLocaleLowerCase();
        const searchKey = keyword.value?.toLocaleLowerCase();
        return (clusterID?.includes(searchKey) || clusterName?.includes(searchKey));
      })
      .map(item => ({
        ...item,
        disabled: isDisabledCluster(item)
          || !isVersionSupported(item.version)
          || item.isInstalled,
        disabledDesc: getDesc(item),
      })));

    watch(value, (v) => {
      if (Array.isArray(v)) {
        clusterValue.value = v;
      } else {
        clusterValue.value = [v];
      }
    }, { immediate: true });

    // 远程搜索
    const keyword = ref('');
    const remoteMethod = (searhcKey) => {
      keyword.value = searhcKey;
    };

    function handleChange(value) {
      if (!props.multiple) {
        clusterValue.value = value[value.length - 1] ? [value[value.length - 1]] : [];
        clusterSelectorRef.value?.close?.();
      }
      emit('change', clusterValue.value);
    }

    const versions = computed(() => props.allowedVersionRange.split(',')
      .filter(v => !!v)
      .map(version => version?.trim?.()));
    function isVersionSupported(version: string) {
      if (!props.allowedVersionRange) return true;
      return versions.value.every(v => satisfiesVersion(version, v));
    }

    onBeforeMount(() => {
      getClusterList();
    });

    return {
      normalStatusList,
      hoverClusterID,
      clusterValue,
      clusterListOfMesh,
      clusterSelectorRef,
      loading,
      remoteMethod,
      handleChange,
      isVersionSupported,
    };
  },
});
</script>
