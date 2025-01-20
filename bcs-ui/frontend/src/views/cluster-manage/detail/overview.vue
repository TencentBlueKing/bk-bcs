<!-- eslint-disable max-len -->
<template>
  <div>
    <div class="flex overflow-hidden pt-[20px] pb-[20px] h-[360px]">
      <!--CPU使用率-->
      <div class="flex-1 w-0 mr-[24px]">
        <div class="flex justify-between">
          <span class="text-[14px] font-bold">{{ $t('metrics.cpuUsage') }}</span>
          <div>
            <div class="flex justify-end items-center">
              <span class="text-[32px]">
                {{ conversionPercentUsed(overviewData.cpu_usage?.used, overviewData.cpu_usage?.total) }}
              </span>
              <sup class="text-[20px]">%</sup>
            </div>
            <div class="font-bold text-[#3ede78] text-[14px]">
              {{ parseFloat(overviewData.cpu_usage?.used || '0').toFixed(2) }}
              of
              {{ parseFloat(overviewData.cpu_usage?.total || '0').toFixed(2) }}
            </div>
          </div>
        </div>
        <ClusterOverviewChart
          :cluster-id="clusterId"
          class="!h-[250px]"
          :color="['#3ede78']"
          :metrics="['cpu_usage']"
        />
      </div>
      <!-- 内存使用率 -->
      <div class="flex-1 w-0 mr-[24px]">
        <div class="flex justify-between">
          <span class="text-[14px] font-bold">{{ $t('metrics.memUsage') }}</span>
          <div>
            <div class="flex justify-end items-center">
              <span class="text-[32px]">
                {{ conversionPercentUsed(overviewData.memory_usage?.used_bytes, overviewData.memory_usage?.total_bytes) }}
              </span>
              <sup class="text-[20px]">%</sup>
            </div>
            <div class="font-bold text-[#3a84ff] text-[14px]">
              {{ formatBytes(overviewData.memory_usage?.used_bytes || 0) }}
              of
              {{ formatBytes(overviewData.memory_usage?.total_bytes || 0) }}
            </div>
          </div>
        </div>
        <ClusterOverviewChart
          :cluster-id="clusterId"
          class="!h-[250px]"
          :colors="['#3a84ff']"
          :metrics="['memory_usage']"
        />
      </div>
      <!-- 磁盘容量(虚拟集群不展示磁盘信息) -->
      <div class="flex-1 w-0" v-if="curCluster?.clusterType !== 'virtual'">
        <div class="flex justify-between">
          <span class="text-[14px] font-bold">{{ $t('metrics.diskUsage') }}</span>
          <div>
            <div class="flex justify-end items-center">
              <span class="text-[32px]">
                {{ conversionPercentUsed(overviewData.disk_usage?.used_bytes, overviewData.disk_usage?.total_bytes) }}
              </span>
              <sup class="text-[20px]">%</sup>
            </div>
            <div class="font-bold text-[#853cff] text-[14px]">
              {{ formatBytes(overviewData.disk_usage?.used_bytes || 0) }}
              of
              {{ formatBytes(overviewData.disk_usage?.total_bytes || 0) }}
            </div>
          </div>
        </div>
        <ClusterOverviewChart
          :cluster-id="clusterId"
          class="!h-[250px]"
          :colors="['#853cff']"
          :metrics="['disk_usage']"
        />
      </div>
    </div>
    <div class="flex overflow-hidden pb-[20px] h-[360px]">
      <!--CPU装箱率-->
      <div class="flex-1 w-0 mr-[24px]">
        <div class="flex justify-between">
          <span class="text-[14px] font-bold">
            {{ $t('metrics.cpuRequestUsage.text') }}
            <bk-popover theme="light">
              <span class="text-[#C4C6CC] relative top-[-1px]">
                <i class="bcs-icon bcs-icon-info-circle-shape"></i>
              </span>
              <template #content>
                <i18n
                  path="metrics.cpuRequestUsage.clusterDesc">
                  <bk-link
                    theme="primary"
                    target="_blank"
                    href="https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/">
                    <span class="text-[12px]"> {{ $t('generic.button.learnMore') }}</span>
                  </bk-link>
                </i18n>
              </template>
            </bk-popover>
          </span>
          <div>
            <div class="flex justify-end items-center">
              <span class="text-[32px]">
                {{ conversionPercentUsed(overviewData.cpu_usage?.request, overviewData.cpu_usage?.total) }}
              </span>
              <sup class="text-[20px]">%</sup>
            </div>
            <div class="font-bold text-[#3ede78] text-[14px]">
              {{ parseFloat(overviewData.cpu_usage?.request || '0').toFixed(2) }}
              of
              {{ parseFloat(overviewData.cpu_usage?.total || '0').toFixed(2) }}
            </div>
          </div>
        </div>
        <ClusterOverviewChart
          :cluster-id="clusterId"
          class="!h-[250px]"
          :color="['#3ede78']"
          :metrics="['cpu_request_usage']"
        />
      </div>
      <!-- 内存装箱率 -->
      <div class="flex-1 w-0 mr-[24px]">
        <div class="flex justify-between">
          <span class="text-[14px] font-bold">
            {{ $t('metrics.memRequestUsage.text') }}
            <bk-popover theme="light">
              <span class="text-[#C4C6CC] relative top-[-1px]">
                <i class="bcs-icon bcs-icon-info-circle-shape"></i>
              </span>
              <template #content>
                <i18n
                  path="metrics.memRequestUsage.clusterDesc">
                  <bk-link
                    theme="primary"
                    target="_blank"
                    href="https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/">
                    <span class="text-[12px]"> {{ $t('generic.button.learnMore') }}</span>
                  </bk-link>
                </i18n>
              </template>
            </bk-popover>
          </span>
          <div>
            <div class="flex justify-end items-center">
              <span class="text-[32px]">
                {{
                  conversionPercentUsed(
                    overviewData.memory_usage?.request_bytes,
                    overviewData.memory_usage?.total_bytes
                  )
                }}
              </span>
              <sup class="text-[20px]">%</sup>
            </div>
            <div class="font-bold text-[#3a84ff] text-[14px]">
              {{ formatBytes(overviewData.memory_usage?.request_bytes || 0) }}
              of
              {{ formatBytes(overviewData.memory_usage?.total_bytes || 0) }}
            </div>
          </div>
        </div>
        <ClusterOverviewChart
          :cluster-id="clusterId"
          class="!h-[250px]"
          :colors="['#3a84ff']"
          :metrics="['memory_request_usage']"
        />
      </div>
      <!-- 磁盘IO(虚拟集群不展示磁盘信息) -->
      <div class="flex-1 w-0" v-if="curCluster?.clusterType !== 'virtual'">
        <div class="flex justify-between">
          <span class="text-[14px] font-bold">{{ $t('metrics.diskIOUsage') }}</span>
          <div>
            <div class="flex justify-end items-center">
              <span class="text-[32px]">
                {{ conversionPercentUsed(
                  overviewData.diskio_usage?.used, overviewData.diskio_usage?.total) }}
              </span>
              <sup class="text-[20px]">%</sup>
            </div>
            <!-- <div class="font-bold text-[#853cff] text-[14px]">
              {{ formatBytes(overviewData.disk_usage.used || 0) }}
              of
              {{ formatBytes(overviewData.disk_usage.total || 0) }}
            </div> -->
          </div>
        </div>
        <ClusterOverviewChart
          :cluster-id="clusterId"
          class="!h-[250px] mt-[17px]"
          :colors="['#853cff']"
          :metrics="['diskio_usage']"
        />
      </div>
    </div>
    <div class="flex overflow-hidden pb-[20px] h-[360px]" v-if="curCluster?.clusterType !== 'virtual'">
      <!--POD使用率-->
      <div class="flex-1 w-0 mr-[24px]">
        <div class="flex justify-between">
          <span class="text-[14px] font-bold">
            {{ $t('metrics.podUsage') }}
          </span>
          <div>
            <div class="flex justify-end items-center">
              <span class="text-[32px]">
                {{ conversionPercentUsed(overviewData.pod_usage?.used, overviewData.pod_usage?.total) }}
              </span>
              <sup class="text-[20px]">%</sup>
            </div>
            <div class="font-bold text-[#3ede78] text-[14px]">
              {{ overviewData.pod_usage?.used || '--' }}
              of
              {{ overviewData.pod_usage?.total || '--' }}
            </div>
          </div>
        </div>
        <ClusterOverviewChart
          :cluster-id="clusterId"
          class="!h-[250px]"
          :color="['#3ede78']"
          :metrics="['pod_usage']"
        />
      </div>
      <div class="flex-1 w-0 mr-[24px]"></div>
      <div class="flex-1 w-0"></div>
    </div>
  </div>
</template>
<script lang="ts">
import { computed, defineComponent, onMounted, ref, toRefs } from 'vue';

import ClusterOverviewChart from './components/cluster-overview-chart.vue';

import { formatBytes } from '@/common/util';
import { useCluster, useProject } from '@/composables/use-app';
import $store from '@/store/index';

interface IUsageData {
  request?: string
  total?: string
  used?: string
}

interface IByteData {
  request_bytes?: string
  total_bytes?: string
  used_bytes?: string
}

export default defineComponent({
  name: 'ClusterOverview',
  components: { ClusterOverviewChart },
  props: {
    clusterId: {
      type: String,
      default: '',
      required: true,
    },
  },
  setup(props) {
    const { clusterId } = toRefs(props);
    const { clusterList } = useCluster();
    const curCluster = computed(() => clusterList.value.find(item => item.clusterID === clusterId.value));
    const { projectCode } = useProject();
    const overviewData = ref<{
      cpu_usage: IUsageData
      disk_usage: IByteData
      memory_usage: IByteData
      diskio_usage: IUsageData
      pod_usage: IUsageData
    }>({
      cpu_usage: {},
      disk_usage: {},
      memory_usage: {},
      diskio_usage: {},
      pod_usage: {},
    });
    const getClusterOverview = async () => {
      overviewData.value = await $store.dispatch('metric/clusterOverview', {
        $projectCode: projectCode.value,
        $clusterId: clusterId.value,
      });
    };
    const conversionPercentUsed = (used, total) => {
      if (!total || parseFloat(total) === 0) {
        return 0;
      }
      let ret: any = parseFloat(used || 0) / parseFloat(total) * 100;
      if (ret !== 0 && ret !== 100) {
        ret = ret.toFixed(2);
      }
      return ret;
    };
    onMounted(() => {
      getClusterOverview();
    });
    return {
      curCluster,
      overviewData,
      conversionPercentUsed,
      formatBytes,
    };
  },
});
</script>
