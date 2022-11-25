<!-- eslint-disable max-len -->
<template>
  <BcsContent :title="$t('Helm Chart仓库')" hide-back>
    <template #header-right>
      <a class="bk-text-button" :href="PROJECT_CONFIG.doc.helm" target="_blank">{{ $t('如何推送Helm Chart到项目仓库？') }}</a>
    </template>
    <!-- 空仓库状态 -->
    <bcs-exception type="empty" v-if="!repos.length && !loading">
      <div>{{$t('没有数据')}}</div>
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
        {{$t('创建仓库')}}
      </bcs-button>
    </bcs-exception>
    <!-- 仓库列表 -->
    <template v-else>
      <bcs-tab
        v-bkloading="{ isLoading: loading }"
        :active.sync="activeRepo"
        @tab-change="handleTabChange">
        <bcs-tab-panel
          v-for="item in repos"
          :name="item.name"
          :label="item.displayName"
          :key="item.name">
          <Row class="mb-[16px]">
            <!-- 搜索功能 -->
            <template #left>
              <bcs-input
                right-icon="bk-icon icon-search"
                class="min-w-[360px]"
                :placeholder="$t('输入名称搜索')"
                v-model="searchName">
              </bcs-input>
            </template>
            <template #right>
              <bcs-popover trigger="click" theme="light">
                <bcs-button>{{$t('查看仓库信息')}}</bcs-button>
                <template #content>
                  <div class="py-[10px]">
                    <div class="flex leading-[20px]">
                      <span class="flex w-[80px]">{{$t('仓库地址')}}:</span>
                      {{item.repoURL}}
                      <span class="bcs-icon-btn ml-[8px]" @click="handleCopyData(item.repoURL)">
                        <i class="bcs-icon bcs-icon-copy"></i>
                      </span>
                    </div>
                    <template v-if="item.username && item.password">
                      <div class="flex leading-[20px]">
                        <span class="flex w-[80px]">{{$t('用户名')}}:</span>
                        {{item.username}}
                        <span class="bcs-icon-btn ml-[8px]" @click="handleCopyData(item.username)">
                          <i class="bcs-icon bcs-icon-copy"></i>
                        </span>
                      </div>
                      <div class="flex leading-[20px]">
                        <span class="flex w-[80px]">{{$t('密码')}}:</span>
                        {{showPassword ? item.password : new Array(10).fill('*').join('')}}
                        <span class="bcs-icon-btn ml-[8px]" @click="showPassword = !showPassword">
                          <i :class="['bcs-icon', showPassword ? 'bcs-icon-eye' : 'bcs-icon-eye-slash']"></i>
                        </span>
                        <span class="bcs-icon-btn ml-[8px]" @click="handleCopyData(item.password)">
                          <i class="bcs-icon bcs-icon-copy"></i>
                        </span>
                      </div>
                      <div class="flex items-center leading-[20px]">
                        <span class="flex w-[100px]">{{$t('添加repo仓库')}}:</span>
                        <bcs-button
                          text
                          size="small"
                          class="!px-0"
                          @click="handleCopyData(`helm repo add ${projectCode} ${item.repoURL} --username=${item.username} --password=${item.password}`)">
                          {{$t('点击复制')}}
                        </bcs-button>
                      </div>
                    </template>
                  </div>
                </template>
              </bcs-popover>
              <bcs-button class="ml-[8px]" @click="handleGetChartsTableData">
                <i class="bcs-icon bcs-icon-reset"></i>
                {{$t('刷新')}}
              </bcs-button>
            </template>
          </Row>
          <bcs-table
            :data="curTableConfig.data"
            :pagination="curTableConfig.pagination"
            v-bkloading="{ isLoading: chartsLoading }"
            @page-change="(page) => pageChange(page, item.name)"
            @page-limit-change="(size) => pageSizeChange(size, item.name)">
            <bcs-table-column :label="$t('名称')" prop="name" show-overflow-tooltip>
              <template #default="{ row }">
                <bcs-button text @click="handleShowDetail(row)">
                  <span class="bcs-ellipsis">{{row.name}}</span>
                </bcs-button>
              </template>
            </bcs-table-column>
            <bcs-table-column :label="$t('版本')" prop="latestVersion" width="100"></bcs-table-column>
            <bcs-table-column :label="$t('最近更新')" prop="updateTime" width="200"></bcs-table-column>
            <bcs-table-column :label="$t('描述')" prop="latestDescription" show-overflow-tooltip></bcs-table-column>
            <bcs-table-column :label="$t('操作')" width="150">
              <template #default="{ row }">
                <bcs-button text @click="handleReleaseChart(row)">{{ $t('部署') }}</bcs-button>
                <bcs-button
                  text
                  class="ml-[10px]"
                  @click="handleShowVersionDialog(row, 'download')">
                  {{$t('下载版本')}}
                </bcs-button>
                <bk-popover
                  placement="bottom"
                  theme="light dropdown"
                  :arrow="false"
                  class="ml-[5px]"
                  v-if="!item.public">
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
                        @click="handleDeleteChart(row)">{{$t('删除Chart')}}</li>
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
                        @click="handleShowVersionDialog(row, 'delete')">{{$t('删除版本')}}</li>
                    </ul>
                  </template>
                </bk-popover>
              </template>
            </bcs-table-column>
          </bcs-table>
        </bcs-tab-panel>
      </bcs-tab>
      <!-- chart详情 -->
      <bcs-sideslider
        :is-show.sync="showDetail"
        quick-close
        :width="1000"
        :title="curRow.name">
        <template #header>
          <div class="flex justify-between items-center pr-[30px]">
            {{curRow.name}}
            <bcs-button theme="primary" @click="handleReleaseChart(curRow)">{{$t('部署')}}</bcs-button>
          </div>
        </template>
        <template #content>
          <ChartDetail :repo-name="activeRepo" :chart="curRow" />
        </template>
      </bcs-sideslider>
      <!-- 删除chart版本或者下载chart版本 -->
      <bcs-dialog
        :title="versionDialogType === 'delete'
          ? $t('删除 {name} Chart的版本', { name: curRow.name })
          : $t('下载 {name} Chart的版本', { name: curRow.name })"
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
            :disabled="!!chartReleaseData.length && versionDialogType === 'delete'"
            @click="handleDeleteOrDownloadVersion">{{$t('确定')}}</bcs-button>
          <bcs-button @click="handleCancelVersionDialog">{{$t('取消')}}</bcs-button>
        </template>
      </bcs-dialog>
      <!-- 删除chart -->
      <bcs-dialog
        :title="$t('确定删除 {name}', { name: curRow.name })"
        width="600"
        v-model="showDeleteChartDialog">
        <ChartReleasesTable
          :data="chartReleaseData"
          v-if="!!chartReleaseData.length" />
        <template #footer>
          <bcs-button
            theme="primary"
            :loading="dialogLoading"
            :disabled="!!chartReleaseData.length"
            @click="handleConfirmDelete">{{$t('确定')}}</bcs-button>
          <bcs-button @click="showDeleteChartDialog = false">{{$t('取消')}}</bcs-button>
        </template>
      </bcs-dialog>
    </template>
  </BcsContent>
</template>
<script lang="ts">
import { computed, defineComponent, onMounted, ref, watch } from '@vue/composition-api';
import BcsContent from '@/components/layout/Content.vue';
import Row from '@/components/layout/Row.vue';
import ChartReleasesTable from './chart-releases-table.vue';
import ChartDetail from './chart-detail.vue';
import useHelm from './use-helm';
import { IPagination } from '../dashboard/common/use-page';
import { useProject } from '@/common/use-app';
import useDebouncedRef from '@/common/use-debounce';
import $router from '@/router';
import { copyText } from '@/common/util';
import $i18n from '@/i18n/i18n-setup';

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
  setup(props, ctx) {
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
    const curTableConfig = computed(() => chartTableConfig.value.find(item => item.name === activeRepo.value) || {
      name: '',
      data: [],
      pagination: {},
    });

    const showPassword = ref(false);
    // 复制
    const handleCopyData = (value: string) => {
      copyText(value);
      ctx.root.$bkMessage({
        theme: 'success',
        message: $i18n.t('复制成功'),
      });
    };
    // 切换仓库
    const handleTabChange = (name) => {
      searchName.value = '';
      $router.replace({
        query: {
          name,
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
      chartsLoading.value = true;
      const data = await handleGetRepoCharts(activeRepo.value, 1, 10, searchName.value);
      const tableConfig: ITableConfig = {
        name: activeRepo.value,
        data: data.data,
        pagination: {
          count: data.total,
          current: 1,
          limit: 10,
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
        const data = await handleGetRepoCharts(config.name, page, config.pagination.limit || 10);
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
        const data = await handleGetRepoCharts(config.name, 1, size);
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
      result && handleGetData();
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
