<!-- eslint-disable max-len -->
<template>
  <BcsContent
    title="Charts"
    :tabs="repos"
    :active-tab="activeRepo"
    hide-back
    v-bkloading="{ isLoading: loading }"
    @tab-change="handleTabChange">
    <template #header-right>
      <a
        class="bk-text-button"
        :href="PROJECT_CONFIG.helm"
        target="_blank">
        {{ $t('deploy.helm.pushRepo') }}
      </a>
    </template>
    <!-- 空仓库状态 -->
    <bcs-exception type="empty" v-if="!repos.length && !loading">
      <div>{{$t('generic.msg.empty.noData3')}}</div>
      <bcs-button
        theme="primary"
        class="mt-[10px]"
        :loading="initRepoLoading"
        v-authority="{
          actionId: 'project_edit',
          resourceName: curProject.project_name,
          permCtx: {
            resource_type: 'project',
            project_id: curProject.project_id
          }
        }"
        @click="handleInitRepo">
        {{$t('deploy.helm.createRepo')}}
      </bcs-button>
    </bcs-exception>
    <!-- 仓库列表 -->
    <template v-else>
      <Row class="mb-[16px]">
        <!-- 搜索功能 -->
        <template #left>
          <bcs-popover trigger="click" theme="light">
            <bcs-button>{{$t('deploy.helm.repoInfo')}}</bcs-button>
            <template #content>
              <div class="py-[10px]">
                <div class="flex leading-[20px]">
                  <span class="flex w-[80px]">{{$t('deploy.helm.repo')}}:</span>
                  {{curRepoItem.repoURL}}
                  <span class="bcs-icon-btn ml-[8px]" @click="handleCopyData(curRepoItem.repoURL)">
                    <i class="bcs-icon bcs-icon-copy"></i>
                  </span>
                </div>
                <template v-if="curRepoItem.username && curRepoItem.password">
                  <div class="flex leading-[20px]">
                    <span class="flex w-[80px]">{{$t('deploy.helm.username')}}:</span>
                    {{curRepoItem.username}}
                    <span class="bcs-icon-btn ml-[8px]" @click="handleCopyData(curRepoItem.username)">
                      <i class="bcs-icon bcs-icon-copy"></i>
                    </span>
                  </div>
                  <div class="flex leading-[20px]">
                    <span class="flex w-[80px]">{{$t('deploy.helm.password')}}:</span>
                    {{showPassword ? curRepoItem.password : new Array(10).fill('*').join('')}}
                    <span class="bcs-icon-btn ml-[8px]" @click="showPassword = !showPassword">
                      <i :class="['bcs-icon', showPassword ? 'bcs-icon-eye' : 'bcs-icon-eye-slash']"></i>
                    </span>
                    <span class="bcs-icon-btn ml-[8px]" @click="handleCopyData(curRepoItem.password)">
                      <i class="bcs-icon bcs-icon-copy"></i>
                    </span>
                  </div>
                  <div class="flex items-center leading-[20px]">
                    <span class="flex w-[100px]">{{$t('deploy.helm.addRepo')}}:</span>
                    <bcs-button
                      text
                      size="small"
                      class="!px-0"
                      @click="handleCopyData(`helm repo add ${projectCode} ${curRepoItem.repoURL} --username=${curRepoItem.username} --password=${curRepoItem.password}`)">
                      {{$t('deploy.helm.copy')}}
                    </bcs-button>
                  </div>
                </template>
              </div>
            </template>
          </bcs-popover>
        </template>
        <template #right>
          <bcs-input
            right-icon="bk-icon icon-search"
            class="min-w-[360px]"
            clearable
            :placeholder="$t('generic.placeholder.searchName')"
            v-model="searchName">
          </bcs-input>
          <!-- <bcs-button class="ml-[8px]" @click="handleGetChartsTableData">
            <i class="bcs-icon bcs-icon-reset"></i>
            {{$t('generic.log.button.refresh')}}
          </bcs-button> -->
        </template>
      </Row>
      <bcs-table
        :data="curTableConfig.data"
        :pagination="curTableConfig.pagination"
        v-bkloading="{ isLoading: chartsLoading }"
        @page-change="(page) => pageChange(page, curRepoItem.name)"
        @page-limit-change="(size) => pageSizeChange(size, curRepoItem.name)">
        <bcs-table-column :label="$t('generic.label.name')" prop="name" show-overflow-tooltip>
          <template #default="{ row }">
            <div class="flex items-center">
              <span class="flex items-center justify-center w-[24px] h-[24px] text-[24px] text-[#C4C6CC] mr-[8px]">
                <img :src="row.icon" v-if="row.icon" />
                <i v-else class="bcs-icon bcs-icon-helm-app"></i>
              </span>
              <bcs-button text @click="handleShowDetail(row)">
                <span class="bcs-ellipsis">{{row.name}}</span>
              </bcs-button>
            </div>
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('generic.label.version')" prop="latestVersion" width="160" show-overflow-tooltip></bcs-table-column>
        <bcs-table-column :label="$t('deploy.helm.latestUpdate')" prop="updateTime" width="200"></bcs-table-column>
        <bcs-table-column :label="$t('cluster.create.label.desc')" prop="latestDescription" show-overflow-tooltip></bcs-table-column>
        <bcs-table-column :label="$t('generic.label.action')" width="180">
          <template #default="{ row }">
            <bcs-button text @click="handleReleaseChart(row)">{{ $t('deploy.helm.install') }}</bcs-button>
            <bcs-button
              text
              class="ml-[10px]"
              @click="handleShowVersionDialog(row, 'download')">
              {{$t('deploy.helm.download')}}
            </bcs-button>
            <bk-popover
              placement="bottom"
              theme="light dropdown"
              :arrow="false"
              class="ml-[5px]"
              trigger="click"
              v-if="!curRepoItem.public">
              <span class="bcs-icon-more-btn"><i class="bcs-icon bcs-icon-more"></i></span>
              <template #content>
                <ul>
                  <li
                    class="bcs-dropdown-item"
                    v-authority="{
                      actionId: 'project_edit',
                      resourceName: curProject.project_name,
                      permCtx: {
                        resource_type: 'project',
                        project_id: curProject.project_id
                      }
                    }"
                    @click="handleDeleteChart(row)">{{$t('deploy.helm.deleteChart')}}</li>
                  <li
                    class="bcs-dropdown-item"
                    v-authority="{
                      actionId: 'project_edit',
                      resourceName: curProject.project_name,
                      permCtx: {
                        resource_type: 'project',
                        project_id: curProject.project_id
                      }
                    }"
                    @click="handleShowVersionDialog(row, 'delete')">{{$t('deploy.helm.deleteVersion')}}</li>
                </ul>
              </template>
            </bk-popover>
          </template>
        </bcs-table-column>
        <template #empty>
          <BcsEmptyTableStatus :type="searchName ? 'search-empty' : 'empty'" @clear="searchName = ''" />
        </template>
      </bcs-table>
      <!-- chart详情 -->
      <bcs-sideslider
        :is-show.sync="showDetail"
        quick-close
        :width="1200"
        :title="curRow.name">
        <template #header>
          <div class="flex justify-between items-center pr-[30px]">
            {{curRow.name}}
            <bcs-button theme="primary" @click="handleReleaseChart(curRow)">{{$t('deploy.helm.install')}}</bcs-button>
          </div>
        </template>
        <template #content>
          <ChartDetail :repo-name="activeRepo" :chart="curRow" />
        </template>
      </bcs-sideslider>
      <!-- 删除chart版本或者下载chart版本 -->
      <bcs-dialog
        :title="versionDialogType === 'delete'
          ? $t('deploy.helm.deleteChartVer', { name: curRow.name })
          : $t('deploy.helm.downloadChartVer', { name: curRow.name })"
        :width="600"
        v-model="showChartVersion">
        <bcs-select :loading="versionLoading" searchable :clearable="false" v-model="version">
          <bcs-option
            v-for="item in versionList"
            :key="item.version"
            :id="item.version"
            :name="item.version">
          </bcs-option>
        </bcs-select>
        <ChartReleasesTable
          :data="chartReleaseData"
          class="mt-[10px]"
          v-if="!!chartReleaseData.length && versionDialogType === 'delete'" />
        <template #footer>
          <bcs-button
            theme="primary"
            :loading="dialogLoading"
            @click="handleDeleteOrDownloadVersion">{{$t('generic.button.confirm')}}</bcs-button>
          <bcs-button @click="handleCancelVersionDialog">{{$t('generic.button.cancel')}}</bcs-button>
        </template>
      </bcs-dialog>
      <!-- 删除chart -->
      <bcs-dialog
        :title="$t('deploy.helm.confirmDeleteChart', { name: curRow.name })"
        width="600"
        v-model="showDeleteChartDialog">
        <ChartReleasesTable
          :data="chartReleaseData"
          v-if="!!chartReleaseData.length" />
        <template #footer>
          <bcs-button
            theme="primary"
            :loading="dialogLoading"
            @click="handleConfirmDelete">{{$t('generic.button.confirm')}}</bcs-button>
          <bcs-button @click="showDeleteChartDialog = false">{{$t('generic.button.cancel')}}</bcs-button>
        </template>
      </bcs-dialog>
    </template>
  </BcsContent>
</template>
<script lang="ts">
import { computed, defineComponent, onMounted, ref, watch } from 'vue';

import ChartDetail from './chart-detail.vue';
import ChartReleasesTable from './chart-releases-table.vue';
import useHelm from './use-helm';

import $bkMessage from '@/common/bkmagic';
import { copyText } from '@/common/util';
import BcsContent from '@/components/layout/Content.vue';
import Row from '@/components/layout/Row.vue';
import { useProject } from '@/composables/use-app';
import useDebouncedRef from '@/composables/use-debounce';
import { IPagination } from '@/composables/use-page';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';

interface ITableConfig {
  name: string;
  data: any[]
  pagination: Partial<IPagination>
}
export default defineComponent({
  name: 'ChartRepo',
  components: { BcsContent, Row, ChartDetail, ChartReleasesTable },
  props: {
    name: {
      type: String,
      default: '',
    },
  },
  setup(props) {
    const {
      loading,
      repos,
      handleGetReposList,
      handleGetRepoCharts,
      handleCreateRepo,
      handleDeleteRepoChart,
      handleGetRepoChartVersions,
      handleDeleteRepoChartVersion,
      handleDownloadChart,
      handleGetChartReleases,
    } = useHelm();

    const searchName = useDebouncedRef<string>('', 300);
    const activeRepo = ref(props.name);
    const chartTableConfig = ref<ITableConfig[]>([]);
    const chartsLoading = ref(false);
    const showPassword = ref(false);
    const curTableConfig = computed(() => chartTableConfig.value.find(item => item.name === activeRepo.value) || {
      name: '',
      data: [],
      pagination: {},
    });
    const curRepoItem = computed(() => repos.value.find(item => item.name === activeRepo.value) || {});
    // 复制
    const handleCopyData = (value: string) => {
      copyText(value);
      $bkMessage({
        theme: 'success',
        message: $i18n.t('generic.msg.success.copy'),
      });
    };
    // 切换仓库
    const handleTabChange = (item) => {
      activeRepo.value = item.name;
      searchName.value = '';
      $router.replace({
        name: 'chartList',
        query: {
          name: item.name,
        },
      });
      handleGetChartsTableData();
    };
    // 搜索chart列表
    watch(searchName, async () => {
      handleGetChartsTableData();
    });
    // chart列表
    const handleGetChartsTableData = async () => {
      if (!activeRepo.value) return;
      chartsLoading.value = true;
      const data = await handleGetRepoCharts(activeRepo.value, 1, 10, searchName.value);
      const tableConfig: ITableConfig = {
        name: activeRepo.value,
        data: data.data,
        pagination: {
          count: data.total,
          current: 1,
          limit: 10,
          showTotalCount: true,
        },
      };
      const index = chartTableConfig.value.findIndex(item => item.name === activeRepo.value);
      chartTableConfig.value.splice(index, 1, tableConfig);
      chartsLoading.value = false;
    };
    const pageChange = async (page: number, repoName: string) => {
      const config = chartTableConfig.value?.find(item => item.name === repoName);
      if (config) {
        chartsLoading.value = true;
        const data = await handleGetRepoCharts(config.name, page, config.pagination.limit || 10, searchName.value);
        config.data = data.data;
        config.pagination.count = data.total;
        config.pagination.current = page;
        chartsLoading.value = false;
      }
    };
    const pageSizeChange = async (size: number, repoName: string) => {
      const config = chartTableConfig.value?.find(item => item.name === repoName);
      if (config) {
        chartsLoading.value = true;
        const data = await handleGetRepoCharts(config.name, 1, size, searchName.value);
        config.data = data.data;
        config.pagination.count = data.total;
        config.pagination.current = 1;
        config.pagination.limit = size;
        chartsLoading.value = false;
      }
    };
    const curRow = ref<Record<string, any>>({});
    // chart详情
    const showDetail = ref(false);
    const handleShowDetail = (row) => {
      showDetail.value = true;
      curRow.value = row;
    };
    // 部署chart
    const handleReleaseChart = (row) => {
      $router.push({
        name: 'releaseChart',
        params: {
          repoName: activeRepo.value,
          chartName: row.name,
        },
      });
    };
    // 创建仓库
    const { projectCode, curProject } = useProject();
    const initRepoLoading = ref(false);
    const handleInitRepo = async () => {
      initRepoLoading.value = true;
      const result = await handleCreateRepo({
        name: projectCode.value,
        type: 'HELM',
      });
      result && handleGetData();
      initRepoLoading.value = false;
    };
    // 获取当前chart已经release的数据
    const dialogLoading = ref(false);
    const chartReleaseData = ref<any[]>([]);
    const handleGetChartReleasesData = async (versions: string[] = []) => {
      dialogLoading.value = true;
      chartReleaseData.value = await handleGetChartReleases({
        $repoName: activeRepo.value,
        $chartName: curRow.value.name,
        versions,
      });
      dialogLoading.value = false;
    };

    // 删除chart
    const showDeleteChartDialog = ref(false);
    const handleDeleteChart = (row) => {
      curRow.value = row;
      chartReleaseData.value = [];
      showDeleteChartDialog.value = true;
      handleGetChartReleasesData();
    };
    const handleConfirmDelete = async () => {
      dialogLoading.value = true;
      const result = await handleDeleteRepoChart(activeRepo.value, curRow.value.name);
      if (result) {
        handleGetData();
        showDeleteChartDialog.value = false;
      }
      dialogLoading.value = false;
    };
    // 删除或者下载指定chart的版本
    const version = ref('');
    const versionList = ref<any[]>([]);
    const versionLoading = ref(false);
    const showChartVersion = ref(false);
    const versionDialogType = ref<'delete'|'download'>();
    const handleShowVersionDialog = async (row, type: 'delete' | 'download') => {
      curRow.value = row;
      versionDialogType.value = type;
      chartReleaseData.value = [];

      showChartVersion.value = true;
      versionLoading.value = true;
      const data = await handleGetRepoChartVersions(activeRepo.value, row.name);
      versionLoading.value = false;
      versionList.value = data?.data || [];
      version.value = versionList.value[0]?.version;
      type === 'delete' && handleGetChartReleasesData([version.value]);
    };
    watch(version, () => {
      handleGetChartReleasesData([version.value]);
    });
    const handleDeleteOrDownloadVersion = async () => {
      if (!version.value) return;

      if (versionDialogType.value === 'delete') {
        dialogLoading.value = true;
        const result = await handleDeleteRepoChartVersion(activeRepo.value, curRow.value.name, version.value);
        dialogLoading.value = false;
        result && handleGetData();
      } else {
        await handleDownloadChart(activeRepo.value, curRow.value.name, version.value);
      }
      showChartVersion.value = false;
    };
    const handleCancelVersionDialog = () => {
      version.value = '';
      versionList.value = [];
      showChartVersion.value = false;
    };

    const handleGetData = async () => {
      // 获取仓库列表
      await handleGetReposList();
      // 默认选择第一个仓库
      if (!activeRepo.value) {
        activeRepo.value = repos.value[0]?.name;
      }
      handleGetChartsTableData();
    };

    onMounted(() => {
      handleGetData();
    });

    return {
      curRepoItem,
      curProject,
      projectCode,
      showPassword,
      searchName,
      loading,
      activeRepo,
      repos,
      chartsLoading,
      curTableConfig,
      showDetail,
      curRow,
      showDeleteChartDialog,
      dialogLoading,
      chartReleaseData,
      initRepoLoading,
      showChartVersion,
      versionList,
      version,
      versionLoading,
      versionDialogType,
      pageChange,
      pageSizeChange,
      handleReleaseChart,
      handleDeleteChart,
      handleShowVersionDialog,
      handleShowDetail,
      handleInitRepo,
      handleDeleteOrDownloadVersion,
      handleCancelVersionDialog,
      handleTabChange,
      handleConfirmDelete,
      handleCopyData,
      handleGetChartsTableData,
    };
  },
});
</script>
