<template>
  <bk-form class="bcs-small-form px-[60px] py-[24px]">
    <template v-if="curCluster.manageType === 'INDEPENDENT_CLUSTER'">
      <bk-form-item :label="$t('cluster.labels.clusterType')">
        <span class="text-[#313238]">
          {{ $t('bcs.cluster.selfDeployed') }}
        </span>
      </bk-form-item>
      <bk-form-item :label="$t('cluster.labels.masterInfo')">
        <div class="flex max-w-[800px]">
          <bk-table :data="masterData" v-bkloading="{ isLoading }">
            <bk-table-column :label="$t('cluster.labels.hostName')">
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
              <bk-table-column :label="$t('generic.ipSelector.label.idc')" prop="idc"></bk-table-column>
              <bk-table-column :label="$t('cluster.labels.rack')" prop="rack"></bk-table-column>
              <bk-table-column :label="$t('generic.ipSelector.label.serverModel')" prop="deviceClass"></bk-table-column>
            </template>
          </bk-table>
        </div>
      </bk-form-item>
    </template>
    <template v-else>
      <bk-form-item :label="$t('cluster.labels.clusterType')">
        <span class="text-[#313238]">
          {{ $t('bcs.cluster.managed') }}
        </span>
        <span class="text-[#979BA5]">
          ({{ $t('cluster.create.label.manageType.managed.desc') }})
        </span>
      </bk-form-item>
      <bk-form-item :label="$t('cluster.create.label.manageType.managed.clusterLevel.text')">
        <span class="text-[#313238]">{{ clusterLevel }}</span>
        <span class="text-[#979BA5]">
          ({{
            $t('cluster.create.label.manageType.managed.clusterLevel.desc', {
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
import { copyText } from '@/common/util';
import $bkMessage from '@/common/bkmagic';
import $i18n from '@/i18n/i18n-setup';

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
    const masterData = ref<any[]>([]);
    const handleGetMasterData = async () => {
      isLoading.value = true;
      masterData.value = await masterList({
        $clusterId: props.clusterId,
      }).catch(() => []);
      isLoading.value = false;
    };

    // 复制IP
    const handleCopyIPv4 = () => {
      copyText(masterData.value.map(item => item.innerIP).join('\n'));
      $bkMessage({
        theme: 'success',
        message: $i18n.t('generic.msg.success.copy'),
      });
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
      handleCopyIPv4,
    };
  },
});
</script>
<style lang="postcss" scoped>
@import './form.css';
</style>
