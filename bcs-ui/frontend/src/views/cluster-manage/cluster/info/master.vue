<template>
  <bk-form class="bcs-small-form px-[60px] py-[24px]">
    <template v-if="curCluster.manageType === 'INDEPENDENT_CLUSTER'">
      <bk-form-item :label="$t('集群类型')">
        <span class="text-[#313238]">
          {{ $t('独立集群') }}
        </span>
      </bk-form-item>
      <bk-form-item :label="$t('Master信息')">
        <bk-table :data="masterData" v-bkloading="{ isLoading }" class="max-w-[800px]">
          <bk-table-column :label="$t('主机名称')">
            <template #default="{ row }">
              {{ row.nodeName || '--' }}
            </template>
          </bk-table-column>
          <bk-table-column label="IPv4">
            <template #default="{ row }">
              {{ row.innerIP || '--' }}
            </template>
          </bk-table-column>
          <bk-table-column label="IPv6">
            <template #default="{ row }">
              {{ row.innerIPv6 || '--' }}
            </template>
          </bk-table-column>
          <template v-if="$INTERNAL">
            <bk-table-column :label="$t('机房')" prop="idc"></bk-table-column>
            <bk-table-column :label="$t('机架')" prop="rack"></bk-table-column>
            <bk-table-column :label="$t('机型')" prop="deviceClass"></bk-table-column>
          </template>
        </bk-table>
      </bk-form-item>
    </template>
    <template v-else>
      <bk-form-item :label="$t('集群类型')">
        <span class="text-[#313238]">
          {{ $t('托管集群') }}
        </span>
        <span class="text-[#979BA5]">
          ({{ $t('Kubernetes 集群的 Master 和 Etcd 会由 TKE 团队集中管理和维护，不需要关心集群 Master 的管理和维护。') }})
        </span>
      </bk-form-item>
      <bk-form-item :label="$t('集群规格')">
        <span class="text-[#313238]">{{ clusterLevel }}</span>
        <span class="text-[#979BA5]">
          ({{
            $t('当前集群规格最多管理 {nodes} 个节点，{pods} 个 Pod，{service} 个 ConfigMap，{crd} 个 CRD', {
              nodes: curClusterScale.level.split('L')[1],
              pods: curClusterScale.scale.maxNodePodNum,
              service: curClusterScale.scale.maxServiceNum,
              crd: curClusterScale.scale.cidrStep
            })
          }})
        </span>
      </bk-form-item>
    </template>
  </bk-form>
</template>
<script lang="ts">
import { useCluster } from '@/composables/use-app';
import { defineComponent, ref, computed, onBeforeMount } from 'vue';
import { masterList } from '@/api/modules/cluster-manager';
import clusterScaleData from '../create/cluster-scale.json';

export default defineComponent({
  name: 'ClusterMaster',
  props: {
    clusterId: {
      type: String,
      default: '',
      required: true,
    },
  },
  setup(props) {
    const { clusterList } = useCluster();
    const curCluster = computed(() => clusterList.value.find(item => item.clusterID === props.clusterId) || {});

    // 托管集群集群规格信息
    const clusterLevel = computed(() => curCluster.value?.clusterBasicSettings?.clusterLevel || '--');
    const clusterScale = ref(clusterScaleData.data);
    const curClusterScale = computed(() => clusterScale.value
      .find(item => item.level === clusterLevel.value)
      || { level: '', scale: { maxNodePodNum: 0, maxServiceNum: 0, cidrStep: 0 } });

    // 独立集群Master信息
    const isLoading = ref(false);
    const masterData = ref([]);
    const handleGetMasterData = async () => {
      isLoading.value = true;
      masterData.value = await masterList({
        $clusterId: props.clusterId,
      }).catch(() => []);
      isLoading.value = false;
    };

    onBeforeMount(() => {
      if (Object.keys(curCluster.value.master || {}).length) {
        handleGetMasterData();
      }
    });

    return {
      clusterLevel,
      curCluster,
      curClusterScale,
      isLoading,
      masterData,
    };
  },
});
</script>
<style lang="postcss" scoped>
@import './form.css';
</style>
