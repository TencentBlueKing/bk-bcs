<!-- eslint-disable max-len -->
<template>
  <div>
    <div class="flex overflow-hidden border-solid border-0 border-b border-[#dfe0e5]">
      <!--CPU使用率-->
      <div class="flex-1 p-[20px] h-[360px] border-solid border-0 border-r border-[#dfe0e5]">
        <div class="flex justify-between">
          <span class="text-[14px] font-bold">{{ $t('CPU使用率') }}</span>
          <div>
            <div class="flex justify-end">
              <span class="text-[32px]">
                {{ conversionPercentUsed(overviewData.cpu_usage.used, overviewData.cpu_usage.total) }}
              </span>
              <sup class="text-[20px]">%</sup>
            </div>
            <div class="font-bold text-[#3ede78] text-[14px]">
              {{ parseFloat(overviewData.cpu_usage.used || 0).toFixed(2) }}
              of
              {{ parseFloat(overviewData.cpu_usage.total || 0).toFixed(2) }}
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
      <div class="flex-1 p-[20px] h-[360px] border-solid border-0 border-r border-[#dfe0e5]">
        <div class="flex justify-between">
          <span class="text-[14px] font-bold">{{ $t('内存使用率') }}</span>
          <div>
            <div class="flex justify-end">
              <span class="text-[32px]">
                {{ conversionPercentUsed(overviewData.memory_usage.used_bytes, overviewData.memory_usage.total_bytes) }}
              </span>
              <sup class="text-[20px]">%</sup>
            </div>
            <div class="font-bold text-[#3a84ff] text-[14px]">
              {{ formatBytes(overviewData.memory_usage.used_bytes || 0) }}
              of
              {{ formatBytes(overviewData.memory_usage.total_bytes || 0) }}
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
      <!-- 磁盘容量 -->
      <div class="flex-1 p-[20px] h-[360px]">
        <div class="flex justify-between">
          <span class="text-[14px] font-bold">{{ $t('磁盘容量') }}</span>
          <div>
            <div class="flex justify-end">
              <span class="text-[32px]">
                {{ conversionPercentUsed(overviewData.disk_usage.used_bytes, overviewData.disk_usage.total_bytes) }}
              </span>
              <sup class="text-[20px]">%</sup>
            </div>
            <div class="font-bold text-[#853cff] text-[14px]">
              {{ formatBytes(overviewData.disk_usage.used_bytes || 0) }}
              of
              {{ formatBytes(overviewData.disk_usage.total_bytes || 0) }}
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
    <div class="flex overflow-hidden">
      <!--CPU装箱率-->
      <div class="flex-1 p-[20px] h-[360px] border-solid border-0 border-r border-[#dfe0e5]">
        <div class="flex justify-between">
          <span class="text-[14px] font-bold">
            {{ $t('CPU装箱率') }}
            <bk-popover theme="light">
              <span class="text-[#C4C6CC] relative top-[-1px]">
                <i class="bcs-icon bcs-icon-info-circle-shape"></i>
              </span>
              <template #content>
                <i18n
                  path="集群CPU装箱率 = 集群所有Pod CPU Request之和 / 集群所有节点CPU总核数（不包含Master节点），集群CPU装箱率越接近集群CPU使用率，集群CPU资源利用率越高 {0}">
                  <bk-link
                    theme="primary"
                    target="_blank"
                    href="https://kubernetes.io/zh-cn/docs/concepts/configuration/manage-resources-containers/">
                    <span class="text-[12px]"> {{ $t('了解更多') }}</span>
                  </bk-link>
                </i18n>
              </template>
            </bk-popover>
          </span>
          <div>
            <div class="flex justify-end">
              <span class="text-[32px]">
                {{ conversionPercentUsed(overviewData.cpu_usage.request, overviewData.cpu_usage.total) }}
              </span>
              <sup class="text-[20px]">%</sup>
            </div>
            <div class="font-bold text-[#3ede78] text-[14px]">
              {{ parseFloat(overviewData.cpu_usage.request || 0).toFixed(2) }}
              of
              {{ parseFloat(overviewData.cpu_usage.total || 0).toFixed(2) }}
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
      <div class="flex-1 p-[20px] h-[360px] border-solid border-0 border-r border-[#dfe0e5]">
        <div class="flex justify-between">
          <span class="text-[14px] font-bold">
            {{ $t('内存装箱率') }}
            <bk-popover theme="light">
              <span class="text-[#C4C6CC] relative top-[-1px]">
                <i class="bcs-icon bcs-icon-info-circle-shape"></i>
              </span>
              <template #content>
                <i18n
                  path="集群内存装箱率 = 集群所有Pod 内存 Request之和 / 集群所有节点内存总大小（不包含Master节点），集群内存装箱率越接近集群内存使用率，集群内存资源利用率越高 {0}">
                  <bk-link
                    theme="primary"
                    target="_blank"
                    href="https://kubernetes.io/zh-cn/docs/concepts/configuration/manage-resources-containers/">
                    <span class="text-[12px]"> {{ $t('了解更多') }}</span>
                  </bk-link>
                </i18n>
              </template>
            </bk-popover>
          </span>
          <div>
            <div class="flex justify-end">
              <span class="text-[32px]">
                {{
                  conversionPercentUsed(
                    overviewData.memory_usage.request_bytes,
                    overviewData.memory_usage.total_bytes
                  )
                }}
              </span>
              <sup class="text-[20px]">%</sup>
            </div>
            <div class="font-bold text-[#3a84ff] text-[14px]">
              {{ formatBytes(overviewData.memory_usage.request_bytes || 0) }}
              of
              {{ formatBytes(overviewData.memory_usage.total_bytes || 0) }}
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
      <!-- 磁盘IO -->
      <div class="flex-1 p-[20px] h-[360px]">
        <div class="flex justify-between">
          <span class="text-[14px] font-bold">{{ $t('磁盘IO') }}</span>
          <div>
            <div class="flex justify-end">
              <span class="text-[32px]">
                {{ conversionPercentUsed(
                  overviewData.diskio_usage.used, overviewData.diskio_usage.total) }}
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
          class="!h-[250px]"
          :colors="['#853cff']"
          :metrics="['diskio_usage']"
        />
      </div>
    </div>
  </div>
</template>
<script lang="ts">
import { defineComponent, onMounted, ref, toRefs } from 'vue';
import ClusterOverviewChart from './cluster-overview-chart.vue';
import $store from '@/store/index';
import { useProject } from '@/composables/use-app';
import { formatBytes } from '@/common/util';
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
    const { projectCode } = useProject();
    const overviewData = ref<{
      cpu_usage: any
      disk_usage: any
      memory_usage: any
      diskio_usage: any
    }>({
      cpu_usage: {},
      disk_usage: {},
      memory_usage: {},
      diskio_usage: {},
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
      overviewData,
      conversionPercentUsed,
      formatBytes,
    };
  },
});
</script>
