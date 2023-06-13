<template>
  <LayoutContent :title="$t('日志采集')" hide-back>
    <LayoutRow class="mb20">
      <template #right>
        <ClusterSelect cluster-type="all" class="w-[300px]" v-model="clusterID"></ClusterSelect>
      </template>
    </LayoutRow>
    <bcs-table size="large" :data="list" v-bkloading="{ isLoading: loading }">
      <bcs-table-column :label="$t('图标')" width="80">
        <template #default>
          <span class="text-[#42a5d8] text-[48px]"><i class="bcs-icon bcs-icon-log"></i></span>
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('组件名称')" width="120" prop="name"></bcs-table-column>
      <bcs-table-column :label="$t('版本')" width="120" prop="installed_info.chart_version">
        <template #default="{ row }">
          <span>{{row.installed_info.chart_version || '--'}}</span>
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('状态')" width="120" prop="installed_info.status">
        <template #default="{ row }">
          <StatusIcon
            :status="row.installed_info.status"
            :status-color-map="statusColorMap">
            {{statusTextMap[row.installed_info.status] || $t('未启用')}}
          </StatusIcon>
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('日志查询入口')">
        <template #default>
          <div class="flex flex-col items-start" v-if="logLink">
            <bk-link theme="primary" :href="logLink.std_log_url" target="_blank">
              <span class="text-[12px]">{{$t('标准日志')}}</span>
            </bk-link>
            <bk-link theme="primary" :href="logLink.file_log_url" target="_blank">
              <span class="text-[12px]">{{$t('文件路径日志')}}</span>
            </bk-link>
          </div>
          <bk-link v-else>{{$t('未配置规则')}}</bk-link>
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('描述')" prop="description" min-width="400"></bcs-table-column>
      <bcs-table-column :label="$t('操作')">
        <template #default="{ row }">
          <template
            v-if="!row.installed_info.status
              || (row.installed_info.status === 'failed' && !row.supported_actions.length)">
            <bcs-button text class="mr5" @click="handleEnableTool(row)">
              {{ row.installed_info.status === 'failed' ? $t('重新启用') : $t('启用')}}
            </bcs-button>
          </template>
          <template v-else>
            <bcs-button
              text
              class="mr5"
              v-if="row.supported_actions.includes('config')"
              @click="handleConfigTool"
            >{{$t('前往配置')}}</bcs-button>
            <bcs-button
              text
              class="mr5"
              v-if="row.supported_actions.includes('upgrade')"
              @click="handleUpgradeTool(row)">
              {{$t('更新组件')}}
            </bcs-button>
            <bcs-button
              text
              class="mr5"
              v-if="row.supported_actions.includes('uninstall')"
              @click="handleUninstallTool(row)">
              {{$t('卸载组件')}}
            </bcs-button>
          </template>
        </template>
      </bcs-table-column>
    </bcs-table>
  </LayoutContent>
</template>
<script lang="ts">
import { defineComponent, watch, ref } from 'vue';
import LayoutContent from '@/components/layout/Content.vue';
import LayoutRow from '@/components/layout/Row.vue';
import ClusterSelect from '@/components/cluster-selector/cluster-select.vue';
import StatusIcon from '@/components/status-icon';
import useLog from './use-log';
import $store from '@/store';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';

export default defineComponent({
  name: 'LogCollector',
  components: { LayoutContent, LayoutRow, ClusterSelect, StatusIcon },
  setup() {
    const clusterID = ref('');
    const loading = ref(false);
    const list = ref<any[]>([]);
    const statusColorMap = ref({
      deployed: 'green',
      failed: 'red',
      unknown: 'gray',
      '': 'gray',
    });
    const statusTextMap = ref({
      deployed: $i18n.t('已启用'),
      failed: $i18n.t('启用失败'),
      unknown: $i18n.t('未知'),
    });
    const {
      handleGetEntrypoints,
    } = useLog();

    const logLink = ref<Record<string, string>>({});
    const handleGetLogLinks = async () => {
      logLink.value = await handleGetEntrypoints(clusterID.value);
    };
    // 列表
    const handleGetToolsList = async () => {
      loading.value = true;
      const data = await $store.dispatch('crdcontroller/clusterTools', { $clusterId: clusterID.value });
      // 组件库和日志采集公用一个接口，拆分成的两个界面
      list.value = (data || []).filter(item => item.chart_name === 'bk-log-collector');
      loading.value = false;
    };
    // 启用
    const handleEnableTool = async (row) => {
      loading.value = true;
      await $store.dispatch('crdcontroller/clusterToolsInstall', {
        $clusterId: clusterID.value,
        $toolId: row.id,
        values: '',
      });
      await handleGetToolsList();
      loading.value = false;
    };
    // 前往配置
    const handleConfigTool = async () => {
      $router.push({
        name: 'newLogList',
        params: {
          clusterId: clusterID.value,
        },
      });
    };
    // todo: 更新
    const handleUpgradeTool = async (row) => {
      $router.push({
        name: 'crdcontrollerInstanceDetail',
        params: {
          clusterId: clusterID.value,
          id: row.id,
          chartName: row.chart_name,
        },
      });
    };
    // 卸载
    const handleUninstallTool = async (row) => {
      $bkInfo({
        type: 'warning',
        clsName: 'custom-info-confirm',
        subTitle: row.name,
        title: $i18n.t('确定卸载'),
        defaultInfo: true,
        confirmFn: async () => {
          const result = await $store.dispatch('crdcontroller/clusterToolsUninstall', {
            $clusterId: clusterID.value,
            $toolId: row.id,
          });
          result && handleGetToolsList();
        },
      });
    };

    watch(clusterID, () => {
      handleGetToolsList();
      handleGetLogLinks();
    });

    return {
      logLink,
      clusterID,
      loading,
      list,
      statusColorMap,
      statusTextMap,
      handleEnableTool,
      handleConfigTool,
      handleUpgradeTool,
      handleUninstallTool,
    };
  },
});
</script>
