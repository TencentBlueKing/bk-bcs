<template>
  <div class="biz-content">
    <div class="biz-top-bar">
      <div class="biz-cluster-node-overview-title">
        <i class="bcs-icon bcs-icon-arrows-left back" @click="goNode"></i>
        <span>{{nodeId}}</span>
      </div>
      <bk-guide></bk-guide>
    </div>
    <div class="biz-content-wrapper biz-cluster-node-overview">
      <div class="biz-cluster-node-overview-wrapper">
        <div class="biz-cluster-node-overview-header">
          <div class="header-item">
            <div class="key-label">IP：</div>
            <bcs-popover :content="nodeId" placement="bottom">
              <div class="value-label">{{nodeId || '--'}}</div>
            </bcs-popover>
          </div>
          <div class="header-item">
            <div class="key-label">CPU：</div>
            <bcs-popover :content="nodeInfo.cpu_count" placement="bottom">
              <div class="value-label">{{nodeInfo.cpu_count || '--'}}</div>
            </bcs-popover>
          </div>
          <div class="header-item">
            <div class="key-label">{{$t('内存：')}}</div>
            <bcs-popover :content="formatBytes(nodeInfo.memory, 0)" placement="bottom">
              <div class="value-label">{{formatBytes(nodeInfo.memory, 0) || '--'}}</div>
            </bcs-popover>
          </div>
          <div class="header-item">
            <div class="key-label">{{$t('存储：')}}</div>
            <bcs-popover :content="formatBytes(nodeInfo.disk, 0)" placement="bottom">
              <div class="value-label">{{formatBytes(nodeInfo.disk, 0) || '--'}}</div>
            </bcs-popover>
          </div>
          <div class="header-item">
            <div class="key-label">{{$t('内核：')}}</div>
            <bcs-popover :content="nodeInfo.release" placement="bottom">
              <div class="value-label">{{nodeInfo.release || '--'}}</div>
            </bcs-popover>
          </div>
          <div class="header-item">
            <div class="key-label">Docker：</div>
            <bcs-popover :content="nodeInfo.dockerVersion" placement="bottom">
              <div class="value-label">{{nodeInfo.dockerVersion || '--'}}</div>
            </bcs-popover>
          </div>
          <div class="header-item">
            <div class="key-label">{{$t('操作系统：')}}</div>
            <bcs-popover :content="nodeInfo.sysname" placement="bottom">
              <div class="value-label">{{nodeInfo.sysname || '--'}}</div>
            </bcs-popover>
          </div>
          <!-- <template v-if="isTkeCluster">
            <div class="header-item">
              <div class="key-label">{{$t('节点模板：')}}</div>
              <bcs-popover
                :content="nodeTemplateInfo.name"
                placement="bottom">
                <div class="value-label">
                  {{nodeTemplateInfo.name || '--' }}
                </div>
              </bcs-popover>
            </div>
            <div class="header-item">
              <div class="key-label">{{$t('Kubelet组件参数：')}}</div>
              <div
                class="value-label"
                v-bk-tooltips.bottom="nodeTemplateInfo.extraArgs.kubelet.split(';').join('<br>')">
                {{nodeTemplateInfo.extraArgs.kubelet}}
              </div>
            </div>
          </template> -->
        </div>
        <div class="biz-cluster-node-overview-chart-wrapper">
          <div class="biz-cluster-node-overview-chart">
            <div class="part top-left">
              <div class="info">
                <div class="left">{{$t('CPU使用率')}}</div>
                <div class="right">
                  <bk-dropdown-menu :align="'right'" ref="cpuDropdown">
                    <div style="cursor: pointer;" slot="dropdown-trigger">
                      <span>{{cpuToggleRangeStr}}</span>
                      <button class="biz-dropdown-button">
                        <i class="bcs-icon bcs-icon-angle-down" style="margin-top: 1px;"></i>
                      </button>
                    </div>
                    <ul class="bk-dropdown-list" slot="dropdown-content">
                      <li>
                        <a
                          href="javascript:;"
                          @click.stop="toggleRange('cpuDropdown', 'cpuToggleRangeStr', 'cpu_summary', '1')">
                          {{$t('1小时')}}</a>
                      </li>
                      <li>
                        <a
                          href="javascript:;"
                          @click.stop="toggleRange('cpuDropdown', 'cpuToggleRangeStr', 'cpu_summary', '2')">
                          {{$t('24小时')}}</a>
                      </li>
                      <li>
                        <a
                          href="javascript:;"
                          @click.stop="toggleRange('cpuDropdown', 'cpuToggleRangeStr', 'cpu_summary', '3')">
                          {{$t('近7天')}}</a>
                      </li>
                    </ul>
                  </bk-dropdown-menu>
                </div>
              </div>
              <ECharts :options="cpuChartOptsK8S" ref="cpuLine1" auto-resize></ECharts>
            </div>
            <div class="part top-right">
              <div class="info">
                <div class="left">{{$t('内存使用率')}}</div>
                <div class="right">
                  <bk-dropdown-menu :align="'right'" ref="memoryDropdown">
                    <div style="cursor: pointer;" slot="dropdown-trigger">
                      <span>{{memToggleRangeStr}}</span>
                      <button class="biz-dropdown-button">
                        <i class="bcs-icon bcs-icon-angle-down" style="margin-top: 1px;"></i>
                      </button>
                    </div>
                    <ul class="bk-dropdown-list" slot="dropdown-content">
                      <li>
                        <a
                          href="javascript:;"
                          @click.stop="toggleRange('memoryDropdown', 'memToggleRangeStr', 'mem', '1')">
                          {{$t('1小时')}}</a>
                      </li>
                      <li>
                        <a
                          href="javascript:;"
                          @click.stop="toggleRange('memoryDropdown', 'memToggleRangeStr', 'mem', '2')">
                          {{$t('24小时')}}</a>
                      </li>
                      <li>
                        <a
                          href="javascript:;"
                          @click.stop="toggleRange('memoryDropdown', 'memToggleRangeStr', 'mem', '3')">
                          {{$t('近7天')}}</a>
                      </li>
                    </ul>
                  </bk-dropdown-menu>
                </div>
              </div>
              <ECharts :options="memChartOptsK8S" ref="memoryLine1" auto-resize></ECharts>
            </div>
          </div>
          <div class="biz-cluster-node-overview-chart">
            <div class="part bottom-left">
              <div class="info">
                <div class="left">{{$t('网络')}}</div>
                <div class="right">
                  <bk-dropdown-menu :align="'right'" ref="networkDropdown">
                    <div style="cursor: pointer;" slot="dropdown-trigger">
                      <span>{{networkToggleRangeStr}}</span>
                      <button class="biz-dropdown-button" style="vertical-align: middle;">
                        <i class="bcs-icon bcs-icon-angle-down" style="margin-top: 1px;"></i>
                      </button>
                    </div>
                    <ul class="bk-dropdown-list" slot="dropdown-content">
                      <li>
                        <a
                          href="javascript:;"
                          @click.stop="toggleRange('networkDropdown', 'networkToggleRangeStr', 'net', '1')">
                          {{$t('1小时')}}</a>
                      </li>
                      <li>
                        <a
                          href="javascript:;"
                          @click.stop="toggleRange('networkDropdown', 'networkToggleRangeStr', 'net', '2')">
                          {{$t('24小时')}}</a>
                      </li>
                      <li>
                        <a
                          href="javascript:;"
                          @click.stop="toggleRange('networkDropdown', 'networkToggleRangeStr', 'net', '3')">
                          {{$t('近7天')}}</a>
                      </li>
                    </ul>
                  </bk-dropdown-menu>
                </div>
              </div>
              <ECharts :options="networkChartOptsK8S" ref="networkLine1" auto-resize></ECharts>
            </div>
            <div class="part">
              <div class="info">
                <div class="left">{{$t('IO使用率')}}</div>
                <div class="right">
                  <bk-dropdown-menu :align="'right'" ref="storageDropdown">
                    <div style="cursor: pointer;" slot="dropdown-trigger">
                      <span>{{storageToggleRangeStr}}</span>
                      <button class="biz-dropdown-button">
                        <i class="bcs-icon bcs-icon-angle-down" style="margin-top: 1px;"></i>
                      </button>
                    </div>
                    <ul class="bk-dropdown-list" slot="dropdown-content">
                      <li>
                        <a
                          href="javascript:;"
                          @click.stop="toggleRange('storageDropdown', 'storageToggleRangeStr', 'io', '1')">
                          {{$t('1小时')}}</a>
                      </li>
                      <li>
                        <a
                          href="javascript:;"
                          @click.stop="toggleRange('storageDropdown', 'storageToggleRangeStr', 'io', '2')">
                          {{$t('24小时')}}</a>
                      </li>
                      <li>
                        <a
                          href="javascript:;"
                          @click.stop="toggleRange('storageDropdown', 'storageToggleRangeStr', 'io', '3')">
                          {{$t('近7天')}}</a>
                      </li>
                    </ul>
                  </bk-dropdown-menu>
                </div>
              </div>
              <ECharts :options="diskioChartOptsK8S" ref="storageLine1" auto-resize></ECharts>
            </div>
          </div>
        </div>
        <bcs-tab class="mt20" type="card" :label-height="42">
          <bcs-tab-panel name="pod" label="Pod">
            <div class="layout-header">
              <div></div>
              <div class="select-wrapper">
                <span class="select-prefix">{{$t('命名空间')}}</span>
                <bcs-select
                  class="namespaces-select"
                  v-model="namespaceValue"
                  :loading="namespaceLoading"
                  searchable
                  :clearable="false"
                  :placeholder="$t('请选择命名空间')"
                  @selected="handleNamespaceSelected"
                >
                  <bcs-option
                    v-for="option in curNamespceList"
                    :key="option.value"
                    :id="option.value"
                    :name="option.label"></bcs-option>
                </bcs-select>
                <bk-input
                  class="search-input ml5"
                  clearable
                  v-model="searchValue"
                  right-icon="bk-icon icon-search"
                  :placeholder="$t('输入名称搜索')">
                </bk-input>
              </div>
            </div>
            <bk-table
              :data="curPodsData"
              :pagination="podsDataPagination"
              v-bkloading="{ isLoading: podLoading }"
              @page-change="handlePageChange"
              @page-limit-change="handlePageLimitChange"
            >
              <bk-table-column :label="$t('名称')" min-width="130" sortable :resizable="false">
                <template #default="{ row }">
                  <bk-button
                    class="bcs-button-ellipsis"
                    text
                    v-authority="{
                      clickable: podsWebAnnotations.perms.items[row.uid].detailBtn.clickable,
                      actionId: 'namespace_scoped_view',
                      resourceName: row.namespace,
                      disablePerms: true,
                      permCtx: {
                        project_id: projectId,
                        cluster_id: clusterId,
                        name: row.namespace
                      }
                    }"
                    @click="gotoPodDetail(row)"
                  >
                    {{ row.name }}
                  </bk-button>
                </template>
              </bk-table-column>
              <bk-table-column :label="$t('命名空间')" min-width="100" sortable :resizable="false">
                <template #default="{ row }">
                  <span>{{ row.namespace }}</span>
                </template>
              </bk-table-column>
              <bk-table-column :label="$t('镜像')" min-width="200" :resizable="false" :show-overflow-tooltip="false">
                <template slot-scope="{ row }">
                  <span v-bk-tooltips.top="(row.images || []).join('<br />')">
                    {{ (row.images || []).join(', ') }}
                  </span>
                </template>
              </bk-table-column>
              <bk-table-column label="Status" width="140" :resizable="false">
                <template slot-scope="{ row }">
                  <StatusIcon :status="row.status"></StatusIcon>
                </template>
              </bk-table-column>
              <bk-table-column label="Ready" width="100" :resizable="false">
                <template slot-scope="{ row }">
                  {{row.readyCnt}}/{{row.totalCnt}}
                </template>
              </bk-table-column>
              <bk-table-column label="Restarts" width="100" :resizable="false">
                <template slot-scope="{ row }">{{row.restartCnt}}</template>
              </bk-table-column>
              <bk-table-column label="Host IP" width="140" :resizable="false">
                <template slot-scope="{ row }">{{row.hostIP || '--'}}</template>
              </bk-table-column>
              <bk-table-column label="Pod IPv4" width="140" :resizable="false">
                <template slot-scope="{ row }">{{row.podIP || '--'}}</template>
              </bk-table-column>
              <bk-table-column label="Pod IPv6" min-width="140" :resizable="false">
                <template slot-scope="{ row }">{{row.podIPv6 || '--'}}</template>
              </bk-table-column>
              <bk-table-column label="Node" :resizable="false">
                <template slot-scope="{ row }">{{row.node || '--'}}</template>
              </bk-table-column>
              <bk-table-column label="Age" :resizable="false">
                <template #default="{ row }">
                  <span>{{row.age || '--'}}</span>
                </template>
              </bk-table-column>
              <bk-table-column :label="$t('编辑模式')" :resizable="false" width="100">
                <template #default="{ row }">
                  <span>{{row.editModel === 'form' ? $t('表单') : 'YAML'}}</span>
                </template>
              </bk-table-column>
              <bk-table-column :label="$t('操作')" :resizable="false" width="180">
                <template #default="{ row }">
                  <bk-button
                    text
                    v-authority="{
                      clickable: podsWebAnnotations.perms.items[row.uid].detailBtn.clickable,
                      actionId: 'namespace_scoped_view',
                      resourceName: row.namespace,
                      disablePerms: true,
                      permCtx: {
                        project_id: projectId,
                        cluster_id: clusterId,
                        name: row.namespace
                      }
                    }"
                    @click="handleShowLog(row, clusterId)"
                  >
                    {{ $t('日志') }}
                  </bk-button>
                  <bk-button
                    class="ml10"
                    text
                    v-authority="{
                      clickable: podsWebAnnotations.perms.items[row.uid].updateBtn.clickable,
                      actionId: 'namespace_scoped_update',
                      resourceName: row.namespace,
                      disablePerms: true,
                      permCtx: {
                        project_id: projectId,
                        cluster_id: clusterId,
                        name: row.namespace
                      }
                    }"
                    @click="handleUpdateResource(row)"
                  >
                    {{ $t('更新') }}
                  </bk-button>
                  <bk-button
                    class="ml10"
                    text
                    v-authority="{
                      clickable: podsWebAnnotations.perms.items[row.uid].deleteBtn.clickable,
                      actionId: 'namespace_scoped_delete',
                      resourceName: row.namespace,
                      disablePerms: true,
                      permCtx: {
                        project_id: projectId,
                        cluster_id: clusterId,
                        name: row.namespace
                      }
                    }"
                    @click="handleDeleteResource(row)"
                  >
                    {{ $t('删除') }}
                  </bk-button>
                </template>
              </bk-table-column>
            </bk-table>
          </bcs-tab-panel>
        </bcs-tab>
      </div>
    </div>
    <bcs-dialog class="log-dialog" v-model="logShow" width="80%" :show-footer="false" render-directive="if">
      <BcsLog
        :project-id="projectId"
        :cluster-id="clusterId"
        :namespace-id="curNamespace"
        :pod-id="curPodId"
        :default-container="defaultContainer"
        :global-loading="logLoading"
        :container-list="containerList">
      </BcsLog>
    </bcs-dialog>
  </div>
</template>

<script lang="ts">
import { defineComponent, computed, ref, onMounted } from '@vue/composition-api';
import moment from 'moment';
import ECharts from 'vue-echarts/components/ECharts.vue';
import 'echarts/lib/chart/line';
import 'echarts/lib/component/tooltip';
import 'echarts/lib/component/legend';

import { BCS_CLUSTER } from '@/common/constant';
import StatusIcon from '@/views/dashboard/common/status-icon';
import BcsLog from '@/components/bcs-log/index';
import useLog from '@/views/dashboard/workload/detail/use-log';
import { useSelectItemsNamespace } from '@/views/dashboard/namespace/use-namespace';

import { nodeOverview } from '@/common/chart-option';
import { catchErrorHandler, formatBytes } from '@/common/util';
import { createChartOption } from './node-overview-chart-opts';
// import { getNodeTemplateInfo } from '@/api/base';

export default defineComponent({
  components: {  StatusIcon, BcsLog, ECharts },
  setup(props, ctx) {
    const { $route, $router, $bkInfo, $bkMessage, $store, $i18n } = ctx.root;
    const cpuLine = ref(nodeOverview.cpu);
    const memoryLine = ref(nodeOverview.memory);
    const networkLine = ref(nodeOverview.network);
    const storageLine = ref(nodeOverview.storage);
    const cpuChartOptsK8S = ref<any>(createChartOption());
    const memChartOptsK8S = ref<any>(createChartOption());
    const networkChartOptsK8S = ref<any>(createChartOption());
    const diskioChartOptsK8S = ref<any>(createChartOption());
    const cpuToggleRangeStr = ref($i18n.t('1小时'));
    const memToggleRangeStr = ref($i18n.t('1小时'));
    const networkToggleRangeStr = ref($i18n.t('1小时'));
    const storageToggleRangeStr = ref($i18n.t('1小时'));
    const nodeInfo = ref<any>({});
    const podsData = ref<any[]>([]);
    const podsWebAnnotations = ref<any>({});
    const podsDataPagination = ref({
      current: 1,
      count: 0,
      limit: 10,
    });
    const podLoading = ref(false);
    const cpuLine1 = ref<any>(null);
    const memoryLine1 = ref<any>(null);
    const storageLine1 = ref<any>(null);
    const networkLine1 = ref<any>(null);
    const searchValue = ref('');
    // 获取命名空间
    const namespaceValue = ref('all');
    const { namespaceLoading, namespaceList, getNamespaceData } = useSelectItemsNamespace(ctx);

    const projectId = computed(() => $route.params.projectId);
    const clusterId = computed(() => $route.params.clusterId);
    const isTkeCluster = computed(() => ($store.state as any).cluster.clusterList
      ?.find(item => item.clusterID === clusterId.value)?.provider === 'tencentCloud');
    const projectCode = computed(() => $route.params.projectCode);
    const nodeId = computed(() => $route.params.nodeId);
    const nodeName = computed(() => $route.params.nodeName);
    const curPodsData = computed(() => {
      const { limit, current } = podsDataPagination.value;
      let curData: any[] = [];
      if (namespaceValue.value === 'all') {
        curData = podsData.value.filter(i => i.name.includes(searchValue.value));
      } else {
        curData = podsData.value
          .filter(i => i.namespace.includes(namespaceValue.value) && i.name.includes(searchValue.value));
      }
      podsDataPagination.value.count = curData.length;
      return curData.slice(limit * (current - 1), limit * current);
    });
    const clusterList = computed(() => $store.state.cluster.clusterList || []);
    const curCluster = computed(() => clusterList.value.find(item => item.clusterID === clusterId.value));
    const curNamespceList = computed(() => {
      const list = namespaceList.value;
      list.unshift({
        label: `${$i18n.t('全部命名空间')}`,
        value: 'all',
      });
      return list;
    });
    /**
             * 获取上方的信息
             */
    const fetchNodeInfo = async () => {
      try {
        nodeInfo.value = await $store.dispatch('metric/clusterNodeInfo', {
          $projectCode: projectCode.value,
          $clusterId: clusterId.value,
          $nodeIP: nodeId.value,
        }) || {};
      } catch (e) {
        catchErrorHandler(e, ctx);
      }
    };

    /**
             * 获取中间图表数据 k8s
             *
             * @param {string} idx 标识，cpu / memory / network / storage
             * @param {string} range 时间范围，1: 1 小时，2: 24 小时，3：近 7 天
             */
    const fetchDataK8S = async (idx, range) => {
      const params = {
        start_at: '',
        end_at: moment().utc()
          .format(),
        $projectCode: projectCode.value,
        $nodeIP: nodeId.value,
        $clusterId: clusterId.value,
      };
      if (!params.$nodeIP) return;

      // 1 小时
      if (range === '1') {
        params.start_at = moment().subtract(1, 'hours')
          .utc()
          .format();
      } else if (range === '2') { // 24 小时
        params.start_at = moment().subtract(1, 'days')
          .utc()
          .format();
      } else if (range === '3') { // 近 7 天
        params.start_at = moment().subtract(7, 'days')
          .utc()
          .format();
      }

      try {
        if (idx === 'net') {
          const res = await Promise.all([
            $store.dispatch('metric/clusterNodeNetworkReceive', Object.assign({}, params)),
            $store.dispatch('metric/clusterNodeNetworkTransmit', Object.assign({}, params)),
          ]);
          renderNetChart(res[0].result, res[1].result);
        } else {
          let url = '';
          let renderFn = '';
          if (idx === 'cpu_summary') {
            url = 'metric/clusterNodeCpuUsage';
            renderFn = 'renderCpuChart';
          }

          if (idx === 'mem') {
            url = 'metric/clusterNodeMemoryUsage';
            renderFn = 'renderMemChart';
          }

          if (idx === 'io') {
            url = 'metric/clusterNodeDiskIOUsage';
            renderFn = 'renderDiskioChart';
          }

          const res = await $store.dispatch(url, params);
          if (renderFn === 'renderCpuChart') renderCpuChart(res.result || []);
          if (renderFn === 'renderMemChart') renderMemChart(res.result || []);
          if (renderFn === 'renderDiskioChart') renderDiskioChart(res.result || []);
        }
      } catch (e) {
        catchErrorHandler(e, ctx);
      }
    };

    /**
             * 渲染 cpu 图表
             *
             * @param {Array} list 图表数据
             */
    const renderCpuChart = (list) => {
      if (!cpuLine1.value) {
        return;
      }

      const curCpuChartOptsK8S = Object.assign({}, cpuChartOptsK8S.value);
      curCpuChartOptsK8S.series.splice(0, curCpuChartOptsK8S.series.length, ...[]);

      const data = list.length ? list : [{ values: [[parseInt(String(+new Date()).slice(0, 10), 10), '0']] }];

      data.forEach((item) => {
        item.values.forEach((d) => {
          d[0] = parseInt(d[0], 10) * 1000;
        });
        curCpuChartOptsK8S.series.push({
          type: 'line',
          showSymbol: false,
          smooth: true,
          hoverAnimation: false,
          areaStyle: {
            normal: {
              opacity: 0.2,
            },
          },
          itemStyle: {
            normal: {
              color: '#30d878',
            },
          },
          data: item.values,
        });
      });

      const label = $i18n.t('CPU使用率');
      cpuLine1.value.mergeOptions({
        tooltip: {
          formatter(params) {
            let ret = '';

            if (params[0].value[1] === '-') {
              ret = '<div>No Data</div>';
            } else {
              ret = `
                                    <div>${moment(parseInt(params[0].value[0], 10)).format('YYYY-MM-DD HH:mm:ss')}</div>
                                    <div>${label}：${parseFloat(params[0].value[1]).toFixed(2)}%</div>
                                `;
            }

            return ret;
          },
        },
      });

      cpuLine1.value.hideLoading();
    };

    /**
             * 渲染 mem 图表
             *
             * @param {Array} list 图表数据
             */
    const renderMemChart = (list) => {
      if (!memoryLine1.value) {
        return;
      }

      const curMemChartOptsK8S = Object.assign({}, memChartOptsK8S.value);
      curMemChartOptsK8S.series.splice(0, curMemChartOptsK8S.series.length, ...[]);

      const data = list.length ? list : [{ values: [[parseInt(String(+new Date()).slice(0, 10), 10), '0']] }];

      data.forEach((item) => {
        item.values.forEach((d) => {
          d[0] = parseInt(d[0], 10) * 1000;
        });
        curMemChartOptsK8S.series.push({
          type: 'line',
          showSymbol: false,
          smooth: true,
          hoverAnimation: false,
          areaStyle: {
            normal: {
              opacity: 0.2,
            },
          },
          itemStyle: {
            normal: {
              color: '#3a84ff',
            },
          },
          data: item.values,
        });
      });

      const label = $i18n.t('内存使用率');
      memoryLine1.value.mergeOptions({
        tooltip: {
          formatter(params) {
            let ret = '';

            if (params[0].value[1] === '-') {
              ret = '<div>No Data</div>';
            } else {
              ret = `
                                    <div>${moment(parseInt(params[0].value[0], 10)).format('YYYY-MM-DD HH:mm:ss')}</div>
                                    <div>${label}：${parseFloat(params[0].value[1]).toFixed(2)}%</div>
                                `;
            }

            return ret;
          },
        },
      });

      memoryLine1.value.hideLoading();
    };

    /**
             * 渲染 diskio 图表
             *
             * @param {Array} list 图表数据
             */
    const renderDiskioChart = (list) => {
      if (!storageLine1.value) {
        return;
      }

      const curDiskioChartOptsK8S = Object.assign({}, diskioChartOptsK8S.value);
      curDiskioChartOptsK8S.series.splice(0, curDiskioChartOptsK8S.series.length, ...[]);

      const data = list.length ? list : [{ values: [[parseInt(String(+new Date()).slice(0, 10), 10), '0']] }];

      data.forEach((item) => {
        item.values.forEach((d) => {
          d[0] = parseInt(d[0], 10) * 1000;
        });
        curDiskioChartOptsK8S.series.push({
          type: 'line',
          showSymbol: false,
          smooth: true,
          hoverAnimation: false,
          areaStyle: {
            normal: {
              opacity: 0.2,
            },
          },
          itemStyle: {
            normal: {
              color: '#ffbe21',
            },
          },
          data: item.values,
        });
      });

      const label = $i18n.t('磁盘IO');
      storageLine1.value.mergeOptions({
        tooltip: {
          formatter(params) {
            let ret = '';

            if (params[0].value[1] === '-') {
              ret = '<div>No Data</div>';
            } else {
              ret = `
                                    <div>${moment(parseInt(params[0].value[0], 10)).format('YYYY-MM-DD HH:mm:ss')}</div>
                                    <div>${label}：${parseFloat(params[0].value[1]).toFixed(2)}%</div>
                                `;
            }

            return ret;
          },
        },
      });

      storageLine1.value.hideLoading();
    };

    /**
             * 渲染 net 图表
             *
             * @param {Array} listReceive net 入流量数据
             * @param {Array} listTransmit net 出流量数据
             */
    const renderNetChart = (listReceive, listTransmit) => {
      if (!networkLine1.value) {
        return;
      }

      const curNetworkChartOptsK8S = Object.assign({}, networkChartOptsK8S.value);
      curNetworkChartOptsK8S.series.splice(0, curNetworkChartOptsK8S.series.length, ...[]);

      curNetworkChartOptsK8S.yAxis.splice(0, curNetworkChartOptsK8S.yAxis.length, ...[
        {
          axisLabel: {
            formatter(value) {
              return `${formatBytes(value, 0)}`;
            },
          },
        },
      ]);

      const dataReceive = listReceive.length
        ? listReceive
        : [{ values: [[parseInt(String(+new Date()).slice(0, 10), 10), '0']] }];

      const dataTransmit = listTransmit.length
        ? listTransmit
        : [{ values: [[parseInt(String(+new Date()).slice(0, 10), 10), '0']] }];

      dataReceive.forEach((item) => {
        item.values.forEach((d) => {
          d[0] = parseInt(d[0], 10) * 1000;
          d.push('receive');
        });
        curNetworkChartOptsK8S.series.push({
          type: 'line',
          showSymbol: false,
          smooth: true,
          hoverAnimation: false,
          areaStyle: {
            normal: {
              opacity: 0.2,
            },
          },
          itemStyle: {
            normal: {
              color: '#853cff',
            },
          },
          data: item.values,
        });
      });

      dataTransmit.forEach((item) => {
        item.values.forEach((d) => {
          d[0] = parseInt(d[0], 10) * 1000;
          d.push('transmit');
        });
        curNetworkChartOptsK8S.series.push({
          type: 'line',
          showSymbol: false,
          smooth: true,
          hoverAnimation: false,
          areaStyle: {
            normal: {
              opacity: 0.2,
            },
          },
          itemStyle: {
            normal: {
              color: '#3dda80',
            },
          },
          data: item.values,
        });
      });

      const labelReceive = $i18n.t('入流量');
      const labelTransmit = $i18n.t('出流量');
      networkLine1.value.mergeOptions({
        tooltip: {
          formatter(params) {
            let ret = ''
                                + `<div>${moment(parseInt(params[0].value[0], 10)).format('YYYY-MM-DD HH:mm:ss')}</div>`;

            params.forEach((p) => {
              if (p.value[2] === 'receive') {
                ret += `<div>${labelReceive}：${formatBytes(p.value[1], 0)}</div>`;
              } else if (p.value[2] === 'transmit') {
                ret += `<div>${labelTransmit}：${formatBytes(p.value[1], 0)}</div>`;
              }
            });

            return ret;
          },
        },
      });

      networkLine1.value.hideLoading();
    };

    /**
             * 切换时间范围
             *
             * @param {Object} dropdownRef dropdown 标识
             * @param {string} toggleRangeStr 标识
             * @param {string} idx 标识，cpu / memory / network / storage
             * @param {string} range 时间范围，1: 1 小时，2: 24 小时，3：近 7 天
             */
    const toggleRange = (dropdownRef, toggleRangeStr, idx, range) => {
      if (range === '1') {
        if (toggleRangeStr === 'cpuToggleRangeStr') cpuToggleRangeStr.value = $i18n.t('1小时');
        if (toggleRangeStr === 'memToggleRangeStr') memToggleRangeStr.value = $i18n.t('1小时');
        if (toggleRangeStr === 'networkToggleRangeStr') networkToggleRangeStr.value = $i18n.t('1小时');
        if (toggleRangeStr === 'storageToggleRangeStr') storageToggleRangeStr.value = $i18n.t('1小时');
      } else if (range === '2') {
        if (toggleRangeStr === 'cpuToggleRangeStr') cpuToggleRangeStr.value = $i18n.t('24小时');
        if (toggleRangeStr === 'memToggleRangeStr') memToggleRangeStr.value = $i18n.t('24小时');
        if (toggleRangeStr === 'networkToggleRangeStr') networkToggleRangeStr.value = $i18n.t('24小时');
        if (toggleRangeStr === 'storageToggleRangeStr') storageToggleRangeStr.value = $i18n.t('24小时');
      } else if (range === '3') {
        if (toggleRangeStr === 'cpuToggleRangeStr') cpuToggleRangeStr.value = $i18n.t('近7天');
        if (toggleRangeStr === 'memToggleRangeStr') memToggleRangeStr.value = $i18n.t('近7天');
        if (toggleRangeStr === 'networkToggleRangeStr') networkToggleRangeStr.value = $i18n.t('近7天');
        if (toggleRangeStr === 'storageToggleRangeStr') storageToggleRangeStr.value = $i18n.t('近7天');
      }

      const curRef = ref<any>(null);
      if (idx === 'cpu_summary') {
        curRef.value = cpuLine1.value;
      }
      if (idx === 'mem') {
        curRef.value = memoryLine1.value;
      }
      if (idx === 'io') {
        curRef.value = storageLine1.value;
      }
      if (idx === 'net') {
        curRef.value = networkLine1.value;
      }
      curRef.value?.showLoading({
        text: $i18n.t('正在加载中...'),
        color: '#30d878',
        maskColor: 'rgba(255, 255, 255, 0.8)',
      });

      fetchDataK8S(idx, range);
    };

    /**
             * 返回节点管理
             */
    const goNode = () => {
      $router.back();
    };

    /**
             * 获取pod数据
             */
    const fetchPodData = async () => {
      podLoading.value = true;
      const res = await $store.dispatch('cluster/fetchPodsData', {
        $projectId: projectId.value,
        $clusterId: clusterId.value,
        $nodename: nodeName.value,
      }).catch(() => ({ data: [], webAnnotations: {} }));
      podLoading.value = false;
      podsData.value = res.data;
      podsWebAnnotations.value = res.webAnnotations;
    };

    const handlePageChange = (page) => {
      podsDataPagination.value.current = page;
    };

    const handlePageLimitChange = (limit) => {
      podsDataPagination.value.limit = limit;
    };

    const updateViewMode = () => {
      localStorage.setItem('FEATURE_CLUSTER', 'done');
      localStorage.setItem(BCS_CLUSTER, curCluster.value.cluster_id);
      sessionStorage.setItem(BCS_CLUSTER, curCluster.value.cluster_id);
      $store.commit('cluster/forceUpdateCurCluster', curCluster.value.cluster_id ? curCluster.value : {});
      $store.commit('updateCurClusterId', curCluster.value.cluster_id);
      $store.commit('updateViewMode', 'dashboard');
      $store.dispatch('getFeatureFlag');
    };

    const gotoPodDetail = (row) => {
      updateViewMode();
      $router.push({
        name: 'dashboardWorkloadDetail',
        params: {
          category: 'pods',
          name: row.name,
          namespace: row.namespace,
          clusterId: clusterId.value,
          nodeId: nodeId.value,
          nodeName: nodeName.value,
          from: 'nodePods',
        },
        query: {
          kind: 'Pod',
        },
      });
    };

    const handleDeleteResource = (row) => {
      const { name, namespace } = row || {};
      $bkInfo({
        type: 'warning',
        clsName: 'custom-info-confirm',
        title: $i18n.t('确认删除当前资源'),
        subTitle: `Pod ${name}`,
        defaultInfo: true,
        confirmFn: async () => {
          let result = false;
          result = await $store.dispatch('dashboard/resourceDelete', {
            $namespaceId: namespace,
            $type: 'workloads',
            $category: 'pods',
            $name: name,
          });
          result && $bkMessage({
            theme: 'success',
            message: $i18n.t('删除成功'),
          });
          fetchPodData();
        },
      });
    };

    const handleUpdateResource = (row) => {
      updateViewMode();
      const { name, namespace, editMode } = row || {};
      if (editMode === 'yaml') {
        $router.push({
          name: 'dashboardResourceUpdate',
          params: {
            namespace,
            name,
          },
          query: {
            type: 'workloads',
            category: 'pods',
            kind: 'Pod',
          },
        });
      } else {
        $router.push({
          name: 'dashboardFormResourceUpdate',
          params: {
            namespace,
            name,
          },
          query: {
            type: 'workloads',
            category: 'pods',
            kind: 'Pod',
          },
        });
      }
    };

    const handleNamespaceSelected = (val) => {
      namespaceValue.value = val;
    };
    // const nodeTemplateInfo = ref({
    //   name: '',
    //   extraArgs: { kubelet: '' },
    // });
    // const fetchNodeTemplateInfo = async () => {
    //   if (!isTkeCluster.value) return;

    //   const data = await getNodeTemplateInfo({
    //     $innerIP: nodeId.value,
    //   });
    //   nodeTemplateInfo.value = data?.nodeTemplate || {
    //     name: '',
    //     extraArgs: { kubelet: '' },
    //   };
    // };

    onMounted(async () => {
      // eslint-disable-next-line max-len, no-multi-assign
      cpuLine.value.series[0].data = memoryLine.value.series[0].data = memoryLine.value.series[1].data  = networkLine.value.series[0].data = networkLine.value.series[1].data = storageLine.value.series[0].data = storageLine.value.series[1].data = [0];
      nodeOverview.storage.series[0].data = [9, 0, 22, 40, 12, 31, 2, 12, 18, 27, 27];
      // eslint-disable-next-line @typescript-eslint/prefer-optional-chain
      cpuLine1.value && cpuLine1.value.showLoading({
        text: $i18n.t('正在加载中...'),
        color: '#30d878',
        maskColor: 'rgba(255, 255, 255, 0.8)',
      });
      // eslint-disable-next-line @typescript-eslint/prefer-optional-chain
      memoryLine1.value && memoryLine1.value.showLoading({
        text: $i18n.t('正在加载中...'),
        color: '#30d878',
        maskColor: 'rgba(255, 255, 255, 0.8)',
      });
      // eslint-disable-next-line @typescript-eslint/prefer-optional-chain
      storageLine1.value && storageLine1.value.showLoading({
        text: $i18n.t('正在加载中...'),
        color: '#30d878',
        maskColor: 'rgba(255, 255, 255, 0.8)',
      });
      // eslint-disable-next-line @typescript-eslint/prefer-optional-chain
      networkLine1.value && networkLine1.value.showLoading({
        text: $i18n.t('正在加载中...'),
        color: '#30d878',
        maskColor: 'rgba(255, 255, 255, 0.8)',
      });

      // fetchNodeTemplateInfo();
      fetchNodeInfo();
      fetchDataK8S('cpu_summary', '1');
      fetchDataK8S('mem', '1');
      fetchDataK8S('net', '1');
      fetchDataK8S('io', '1');
      getNamespaceData({
        clusterId: clusterId.value,
      });
      fetchPodData();
    });

    return {
      // nodeTemplateInfo,
      isTkeCluster,
      memoryLine,
      networkLine,
      storageLine,
      cpuChartOptsK8S,
      memChartOptsK8S,
      networkChartOptsK8S,
      diskioChartOptsK8S,
      cpuToggleRangeStr,
      memToggleRangeStr,
      networkToggleRangeStr,
      storageToggleRangeStr,
      nodeInfo,
      podsData,
      podsWebAnnotations,
      podsDataPagination,
      podLoading,
      projectId,
      clusterId,
      projectCode,
      nodeId,
      curPodsData,
      clusterList,
      curCluster,
      cpuLine1,
      memoryLine1,
      storageLine1,
      networkLine1,
      namespaceLoading,
      namespaceValue,
      namespaceList,
      curNamespceList,
      searchValue,
      ...useLog(),
      handlePageLimitChange,
      handlePageChange,
      toggleRange,
      goNode,
      gotoPodDetail,
      handleDeleteResource,
      handleUpdateResource,
      handleNamespaceSelected,
      formatBytes,
    };
  },
});
</script>

<style scoped lang="postcss">
    @import '@/css/variable.css';
    @import './pod-log.css';

    .biz-cluster-node-overview {
        padding: 20px;
    }

    .biz-cluster-node-overview-title {
        display: inline-block;
        height: 60px;
        line-height: 60px;
        font-size: 16px;
        margin-left: 20px;
        cursor: pointer;

        .back {
            font-size: 16px;
            font-weight: 700;
            position: relative;
            top: 1px;
            color: $iconPrimaryColor;
        }
    }

    .biz-cluster-node-overview-wrapper {
        background-color: $bgHoverColor;
        display: inline-block;
        width: 100%;
    }

    .biz-cluster-node-overview-header {
        display: flex;
        border: 1px solid $borderWeightColor;
        border-radius: 2px;

        .header-item {
            font-size: 14px;
            flex: 1;
            height: 75px;
            border-right: 1px solid $borderWeightColor;
            padding: 0 20px;

            &:last-child {
                border-right: none;
            }

            .key-label {
                font-weight: 700;
                margin: 12px 0 5px 0;
                overflow: hidden;
                text-overflow: ellipsis;
                white-space: normal;
                word-break: break-all;
                display: -webkit-box;
                -webkit-line-clamp: 1;
                -webkit-box-orient: vertical;
            }

            .value-label {
                max-width: 130px;
                padding-top: 4px;
                overflow: hidden;
                text-overflow: ellipsis;
                white-space: nowrap;
            }
        }
    }

    .biz-cluster-node-overview-chart-wrapper {
        margin-top: 20px;
        background-color: #fff;
        box-shadow: 1px 0 2px rgba(0, 0, 0, 0.1);
        border: 1px solid $borderWeightColor;
        font-size: 0;
        border-radius: 2px;

        .biz-cluster-node-overview-chart {
            display: inline-block;
            width: 100%;

            .part {
                width: 50%;
                float: left;
                height: 250px;

                &.top-left {
                    border-right: 1px solid $borderWeightColor;
                    border-bottom: 1px solid $borderWeightColor;
                }

                &.top-right {
                    border-bottom: 1px solid $borderWeightColor;
                }

                &.bottom-left {
                    border-right: 1px solid $borderWeightColor;
                }

                .info {
                    font-size: 14px;
                    display: flex;
                    padding: 20px 30px;

                    .left,
                    .right {
                        flex: 1;
                    }

                    .left {
                        font-weight: 700;
                    }

                    .right {
                        text-align: right;
                    }
                }
            }
        }
    }

    .echarts {
        width: 100%;
        height: 160px;
        z-index: 100;
    }

    .biz-cluster-node-overview-table-wrapper {
        margin-top: 20px;
    }

    .biz-cluster-node-overview-table {
        border-bottom: none;

        .name {
            width: 400px;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        .mirror {
            width: 500px;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        i {
            top: 1px;
            position: relative;
            margin-right: 7px;

            &.running {
                color: $iconSuccessColor;
            }

            &.warning {
                color: $iconWarningColor;
            }

            &.danger {
                color: $failColor;
            }
        }
    }

    .biz-cluster-node-overview-page {
        border-top: 1px solid #e6e6e6;
        padding: 20px 40px 20px 0;
    }

    @media screen and (max-width: $mediaWidth) {
        .biz-cluster-node-overview-table {
            border-bottom: none;

            .name {
                width: 300px;
            }

            .mirror {
                width: 400px;
            }
        }
    }
    .layout-header {
        display: flex;
        justify-content: space-between;
        margin-bottom: 20px;
    }
    .select-wrapper {
        display: flex;
        .select-prefix {
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 0 10px;
            border: 1px solid #c4c6cc;
            margin-right: -1px;
            font-size: 12px;
        }
        .namespaces-select {
            width: 200px;
        }
        .search-input {
            width: 250px;
        }
    }
</style>
