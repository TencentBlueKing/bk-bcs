<template>
  <div class="autoscaler-management">
    <!-- 自动扩缩容配置 -->
    <section class="autoscaler">
      <div class="group-header">
        <div>
          <span class="group-header-title">{{$t('Cluster Autoscaler配置')}}</span>
          <span class="switch-autoscaler">
            {{$t('Cluster Autoscaler')}}
            <bcs-switcher
              size="small"
              v-model="autoscalerData.enableAutoscale"
              :disabled="autoscalerData.status === 'UPDATING'"
              :pre-check="handleToggleAutoScaler"
            ></bcs-switcher>
          </span>
        </div>
        <bcs-button
          theme="primary"
          :disabled="autoscalerData.status === 'UPDATING'"
          @click="handleEditAutoScaler">{{$t('编辑配置')}}</bcs-button>
      </div>
      <div v-bkloading="{ isLoading: configLoading }">
        <LayoutGroup :title="$t('基本配置')" class="mb10">
          <AutoScalerFormItem
            :list="basicScalerConfig"
            :autoscaler-data="autoscalerData">
          </AutoScalerFormItem>
        </LayoutGroup>
        <LayoutGroup :title="$t('自动扩容配置')" class="mb10">
          <AutoScalerFormItem
            :list="autoScalerConfig"
            :autoscaler-data="autoscalerData">
          </AutoScalerFormItem>
        </LayoutGroup>
        <LayoutGroup collapsible class="mb15" :expanded="!!autoscalerData.isScaleDownEnable">
          <template #title>
            <span>{{$t('自动缩容配置')}}</span>
            <span class="switch-autoscaler">
              {{$t('允许缩容节点')}}
              <bcs-switcher
                size="small"
                :disabled="autoscalerData.status === 'UPDATING'"
                v-model="autoscalerData.isScaleDownEnable"
                :pre-check="handleChangeScalerDown">
              </bcs-switcher>
            </span>
          </template>
          <AutoScalerFormItem
            :list="autoScalerDownConfig"
            :autoscaler-data="autoscalerData">
          </AutoScalerFormItem>
        </LayoutGroup>
      </div>
    </section>
    <!-- 节点池配置 -->
    <section class="nodepool">
      <div class="group-header">
        <div class="group-header-title">{{$t('节点池管理')}}</div>
        <bcs-button theme="primary" icon="plus" @click="handleCreatePool">{{$t('新建节点池')}}</bcs-button>
      </div>
      <bcs-table
        :data="curPageData"
        :pagination="pagination"
        v-bkloading="{ isLoading: nodepoolLoading }"
        @page-change="pageChange"
        @page-limit-change="pageSizeChange">
        <bcs-table-column :label="$t('节点池 ID (名称)')" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            <bcs-button text @click="handleGotoDetail(row)">
              <span class="bcs-ellipsis">{{`${row.nodeGroupID}（${row.name}）`}}</span>
            </bcs-button>
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('节点数量范围')" width="120">
          <template #default="{ row }">
            <span>
              {{ `${row.autoScaling.minSize} ~ ${row.autoScaling.maxSize}` }}
            </span>
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('节点数量')" align="right" width="120">
          <template #default="{ row }">
            <bcs-button
              text
              :disabled="row.autoScaling.desiredSize === 0"
              @click="handleShowNodeManage(row)">
              {{row.autoScaling.desiredSize}}
            </bcs-button>
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('机型')">
          <template #default="{ row }">
            {{ row.launchTemplate.instanceType }}
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('操作系统')" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.launchTemplate.imageInfo.imageName || '--' }}
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('节点池状态')">
          <template #default="{ row }">
            <LoadingIcon v-if="['CREATING', 'DELETING', 'UPDATING'].includes(row.status)">
              {{ statusTextMap[row.status] }}
            </LoadingIcon>
            <StatusIcon status="unknown" v-else-if="!row.enableAutoscale && row.status === 'RUNNING'">
              {{$t('已关闭')}}
            </StatusIcon>
            <StatusIcon
              :status="row.status"
              :status-color-map="statusColorMap"
              v-else>
              {{ statusTextMap[row.status] }}
            </StatusIcon>
          </template>
        </bcs-table-column>
        <!-- <bcs-table-column :label="$t('自动扩缩容')" :render-header="renderHeader">
                    <template #default="{ row }">
                        <span v-if="row.enableAutoscale">{{$t('已启用')}}</span>
                        <span v-else class="disabled">{{$t('已关闭')}}</span>
                    </template>
                </bcs-table-column> -->
        <bcs-table-column :label="$t('操作')" width="170">
          <template #default="{ row }">
            <div class="operate">
              <bcs-button text @click="handleShowRecord(row)">{{$t('扩缩容记录')}}</bcs-button>
              <bcs-popover
                placement="bottom"
                theme="light dropdown"
                :arrow="false"
                class="ml15"
                :disabled="row.status === 'DELETING'"
                :ref="row.nodeGroupID">
                <span class="more-icon"><i class="bcs-icon bcs-icon-more"></i></span>
                <div slot="content">
                  <ul>
                    <li
                      :class="['dropdown-item', {
                        disabled: (row.enableAutoscale && disabledAutoscaler)
                          || ['CREATING', 'DELETING', 'UPDATING'].includes(row.status)
                      }]"
                      v-bk-tooltips="{
                        content: $t('Cluster Autoscaler 需要至少一个节点池开启，请停用 Cluster Autoscaler 后再关闭'),
                        disabled: !(row.enableAutoscale && disabledAutoscaler)
                      }"
                      @click="handleToggleNodeScaler(row)">
                      {{row.enableAutoscale ? $t('关闭节点池') : $t('启用节点池')}}
                    </li>
                    <li class="dropdown-item" @click="handleEditPool(row)">{{$t('编辑节点池')}}</li>
                    <li
                      :class="['dropdown-item', { disabled: disabledDelete || !!row.autoScaling.desiredSize }]"
                      v-bk-tooltips="{
                        content: !!row.autoScaling.desiredSize
                          ? $t('请删除节点后再删除节点池')
                          : $t('Cluster Autoscaler 需要至少一个节点池，请停用 Cluster Autoscaler 后再删除'),
                        disabled: !(disabledDelete || !!row.autoScaling.desiredSize),
                        placements: 'left'
                      }"
                      @click="handleDeletePool(row)">{{$t('删除节点池')}}</li>
                  </ul>
                </div>
              </bcs-popover>
            </div>
          </template>
        </bcs-table-column>
      </bcs-table>
    </section>
    <!-- 节点数量 -->
    <bcs-dialog
      theme="primary"
      header-position="left"
      :title="$t('管理节点数量')"
      :width="700"
      v-model="showNodeManage"
      @cancel="handleNodeManageCancel">
      <bcs-alert type="info" :title="$t('注意：若节点池已开启自动伸缩， 则数量将会随集群负载自动调整')"></bcs-alert>
      <bcs-form class="form-content mt15" :label-width="100">
        <bcs-form-item class="form-content-item" :label="$t('节点池名称')">
          <span>{{ currentOperateRow.name }}</span>
        </bcs-form-item>
        <bcs-form-item class="form-content-item" :label="$t('节点数量范围')">
          <span>
            {{
              currentOperateRow.autoScaling
                ? `${currentOperateRow.autoScaling.minSize} ~ ${currentOperateRow.autoScaling.maxSize}`
                : '--'
            }}
          </span>
        </bcs-form-item>
        <bcs-form-item class="form-content-item" :label="$t('节点数量')">
          <span>
            {{
              currentOperateRow.autoScaling
                ? currentOperateRow.autoScaling.desiredSize
                : '--'
            }}
          </span>
        </bcs-form-item>
      </bcs-form>
      <bcs-table
        class="mt15"
        v-bkloading="{ isLoading: nodeListLoading }"
        :data="nodeCurPageData"
        :pagination="nodePagination"
        @page-change="nodePageChange"
        @page-limit-change="nodePageSizeChange">
        <bcs-table-column :label="$t('节点名称')" prop="innerIP"></bcs-table-column>
        <bcs-table-column :label="$t('状态')">
          <template #default="{ row }">
            <LoadingIcon v-if="['DELETING', 'INITIALIZATION'].includes(row.status)">
              {{ nodeStatusMap[row.status] }}
            </LoadingIcon>
            <StatusIcon
              :status="row.status"
              :status-color-map="nodeColorMap"
              v-else>
              {{ nodeStatusMap[row.status] }}
            </StatusIcon>
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('操作')" width="120">
          <template #default="{ row }">
            <div class="operate">
              <bcs-button text @click="handleToggleCordon(row)">
                {{row.unSchedulable ? $t('允许调度') : $t('停止调度')}}
              </bcs-button>
              <bcs-popover
                placement="bottom"
                theme="light dropdown"
                :arrow="false"
                class="ml15">
                <span class="more-icon"><i class="bcs-icon bcs-icon-more"></i></span>
                <div slot="content">
                  <ul>
                    <li class="dropdown-item" @click="handleNodeDrain(row)">{{$t('pod迁移')}}</li>
                    <li
                      :class="['dropdown-item', { disabled: !row.unSchedulable }]"
                      v-bk-tooltips="{
                        content: $t('请先停止调度'),
                        disabled: row.unSchedulable
                      }"
                      @click="handleDeleteNodeGroupNode(row)"
                    >{{$t('删除节点')}}</li>
                  </ul>
                </div>
              </bcs-popover>
            </div>
          </template>
        </bcs-table-column>
      </bcs-table>
    </bcs-dialog>
    <!-- 扩缩容记录 -->
    <bcs-dialog
      theme="primary"
      header-position="left"
      :title="$t('扩缩容记录')"
      :width="960"
      v-model="showRecord"
      @cancel="handleRecordCancel">
      <div class="mb15 flex-between">
        <bcs-date-picker
          :shortcuts="shortcuts"
          type="datetimerange"
          shortcut-close
          :use-shortcut-text="false"
          v-model="timeRange"
          @change="handleTimeRangeChange">
        </bcs-date-picker>
      </div>
      <bcs-table
        v-bkloading="{ isLoading: recordLoading }"
        :data="recordList"
        :pagination="recordPagination"
        @page-change="recordPageChange"
        @page-limit-change="recordPageSizeChange">
        <bcs-table-column type="expand" width="30">
          <template #default="{ row }">
            <bcs-table
              :data="row.task ? row.task.stepSequence : []"
              :outer-border="false"
              :header-cell-style="{ background: '#fff', borderRight: 'none' }">
              <bcs-table-column :label="$t('步骤名称')" show-overflow-tooltip>
                <template #default="{ row: key }">
                  {{ row.task.steps[key].name }}
                </template>
              </bcs-table-column>
              <bcs-table-column :label="$t('步骤信息')" show-overflow-tooltip>
                <template #default="{ row: key }">
                  {{ row.task.steps[key].message }}
                </template>
              </bcs-table-column>
              <bcs-table-column :label="$t('开始时间')" show-overflow-tooltip>
                <template #default="{ row: key }">
                  {{ row.task.steps[key].start }}
                </template>
              </bcs-table-column>
              <bcs-table-column :label="$t('结束时间')" show-overflow-tooltip>
                <template #default="{ row: key }">
                  {{ row.task.steps[key].end }}
                </template>
              </bcs-table-column>
              <bcs-table-column :label="$t('状态')">
                <template #default="{ row: key }">
                  <StatusIcon
                    :status-color-map="taskStatusColorMap"
                    :status="row.task.steps[key].status">
                    {{ taskStatusMap[row.task.steps[key].status] }}
                  </StatusIcon>
                </template>
              </bcs-table-column>
            </bcs-table>
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('事件类型')" show-overflow-tooltip>
          <template #default="{ row }">
            {{row.task ? row.task.taskName : '--'}}
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('事件信息')" prop="message" min-width="200" show-overflow-tooltip></bcs-table-column>
        <bcs-table-column :label="$t('开始时间')" prop="createTime" show-overflow-tooltip></bcs-table-column>
        <bcs-table-column :label="$t('结束时间')" show-overflow-tooltip>
          <template #default="{ row }">
            {{row.task ? row.task.end : '--'}}
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('状态')">
          <template #default="{ row }">
            <StatusIcon
              :status-color-map="taskStatusColorMap"
              :status="row.task.status"
              v-if="row.task">
              {{ taskStatusMap[row.task.status] }}
            </StatusIcon>
            <span v-else>--</span>
          </template>
        </bcs-table-column>
      </bcs-table>
    </bcs-dialog>
  </div>
</template>
<script lang="ts">
import { defineComponent, onMounted, ref, computed, onBeforeUnmount, CreateElement, getCurrentInstance } from 'vue';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $store from '@/store/index';
import StatusIcon from '@/components/status-icon';
import LoadingIcon from '@/components/loading-icon.vue';
import usePage from '@/composables/use-page';
import useInterval from '@/composables/use-interval';
import LayoutGroup from '@/views/cluster-manage/components/layout-group.vue';
import AutoScalerFormItem from './form-item.vue';
import $bkMessage from '@/common/bkmagic';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';

export default defineComponent({
  name: 'AutoScaler',
  components: { StatusIcon, LoadingIcon, LayoutGroup, AutoScalerFormItem },
  props: {
    clusterId: {
      type: String,
      default: '',
    },
  },
  setup(props) {
    const configLoading = ref(false);
    const autoscalerData = ref<Record<string, string>>({});
    const basicScalerConfig = ref([
      {
        prop: 'status',
        name: $i18n.t('状态'),
      },
      {
        prop: 'scanInterval',
        name: $i18n.t('扩缩容检测时间间隔'),
        unit: $i18n.t('秒'),
      },
      {
        prop: 'okTotalUnreadyCount',
        name: $i18n.t('允许unready节点'),
        unit: $i18n.t('个'),
        desc: $i18n.t('自动扩缩容保护机制，集群中unready节点大于允许unready节点数量，且unready节点的比例大于设置的比例，会停止Cluster Autoscaler功能，否则Cluster Autoscaler功能正常运行'),
      },
      {
        prop: 'maxTotalUnreadyPercentage',
        name: $i18n.t('unready节点超过集群总节点'),
        unit: '%',
        suffix: $i18n.t('时停止自动扩缩容'),
      },
    ]);
    const autoScalerConfig = ref([
      {
        prop: 'expander',
        name: $i18n.t('扩容算法'),
        isBasicProp: true,
        desc: $i18n.t('random：在有多个节点池时，随机选择节点池<br/>least-waste：在有多个节点池时，以最小浪费原则选择，选择有最少可用资源的节点池<br/>most-pods：在有多个节点池时，选择容量最大（可以创建最多Pod）的节点池'),
      },
      {
        prop: 'bufferResourceRatio',
        name: $i18n.t('触发扩容资源阈值'),
        isBasicProp: true,
        unit: '%',
      },
      {
        prop: 'maxNodeProvisionTime',
        name: $i18n.t('等待节点提供最长时间'),
        unit: $i18n.t('秒'),
        desc: $i18n.t('如果节点池在设置的时间范围内没有提供可用资源，会导致此次自动扩容失败'),
      },
      {
        prop: 'scaleUpFromZero',
        name: $i18n.t('(没有ready节点时) 允许自动扩容'),
      },
    ]);
    const autoScalerDownConfig = ref([
      {
        prop: 'scaleDownUtilizationThreahold',
        name: $i18n.t('触发缩容资源阈值 (CPU/内存)'),
        isBasicProp: true,
        unit: '%',
      },
      {
        prop: 'scaleDownUnneededTime',
        name: $i18n.t('执行缩容等待时间'),
        isBasicProp: true,
        unit: $i18n.t('秒'),
        desc: $i18n.t('Cluster Autocaler组件评估集群可以缩容多久后开始执行缩容，防止集群容量在短时间内或高或低于设置的缩容阈值造成频繁扩缩容操作'),
      },
      {
        prop: 'maxGracefulTerminationSec',
        name: $i18n.t('等待 Pod 退出最长时间'),
        isBasicProp: true,
        unit: $i18n.t('秒'),
        desc: $i18n.t('缩容节点时，等待 pod 停止的最长时间（不会遵守 terminationGracefulPeriodSecond，超时强杀）'),
      },
      {
        prop: 'scaleDownDelayAfterAdd',
        name: $i18n.t('扩容后判断缩容时间间隔'),
        unit: $i18n.t('秒'),
        desc: $i18n.t('扩容节点后多久才继续缩容判断，如果业务自定义初始化任务所需时间比较长，需要适当上调此值'),
      },
      {
        prop: 'scaleDownDelayAfterDelete',
        name: $i18n.t('连续两次缩容时间间隔'),
        unit: $i18n.t('秒'),
        desc: $i18n.t('缩容节点后多久再继续缩容节点，默认设置为0，代表与扩缩容检测时间间隔设置的值相同'),
      },
      {
        prop: 'scaleDownDelayAfterFailure',
        name: $i18n.t('缩容失败后重试时间间隔'),
        unit: $i18n.t('秒'),
      },
      {
        prop: 'scaleDownUnreadyTime',
        name: $i18n.t('unready节点缩容等待时间'),
        unit: $i18n.t('秒'),
      },
      {
        prop: 'maxEmptyBulkDelete',
        name: $i18n.t('单次缩容最大节点数'),
        unit: $i18n.t('个'),
      },
    ]);
    const getAutoScalerConfig = async () => {
      if (!props.clusterId) return;
      autoscalerData.value = await $store.dispatch('clustermanager/clusterAutoScaling', {
        $clusterId: props.clusterId,
      });
      if (autoscalerData.value.status !== 'UPDATING') {
        stop();
      }
    };
    const handleGetAutoScalerConfig = async () => {
      configLoading.value = true;
      await getAutoScalerConfig();
      if (autoscalerData.value.status === 'UPDATING') {
        start();
      }
      configLoading.value = false;
    };
    const { start, stop } = useInterval(getAutoScalerConfig, 5000); // 轮询
    // 自动扩容开启｜关闭
    const user = computed(() => $store.state.user);
    // eslint-disable-next-line @typescript-eslint/no-misused-promises
    const handleToggleAutoScaler = async value => new Promise(async (resolve, reject) => {
      if (!autoscalerData.value.enableAutoscale
                        && (!nodepoolList.value.length || nodepoolList.value.every(item => !item.enableAutoscale))) {
        // 开启时前置判断是否存在节点池 或 节点池都是未开启状态时，要提示至少开启一个
        $bkInfo({
          type: 'warning',
          clsName: 'custom-info-confirm',
          title: !nodepoolList.value.length
            ? $i18n.t('请创建节点池并启用节点池自动扩缩容功能')
            : $i18n.t('请至少启用 1 个节点池的自动扩缩容功能或创建新的节点池'),
          defaultInfo: true,
          okText: $i18n.t('立即新建'),
          confirmFn: () => {
            handleCreatePool();
          },
          cancelFn: () => {
            // eslint-disable-next-line prefer-promise-reject-errors
            reject(false);
          },
        });
      } else {
        // 开启或关闭扩缩容
        const result = await $store.dispatch('clustermanager/toggleClusterAutoScalingStatus', {
          enable: value,
          $clusterId: props.clusterId,
          updater: user.value.username,
        });
        if (result) {
          $bkMessage({
            theme: 'success',
            message: $i18n.t('操作成功'),
          });
          handleGetAutoScalerConfig();
          resolve(true);
        } else {
          // eslint-disable-next-line prefer-promise-reject-errors
          reject(false);
        }
      }
    });
    // eslint-disable-next-line @typescript-eslint/no-misused-promises
    const handleChangeScalerDown = async value => new Promise(async (resolve, reject) => {
      configLoading.value = true;
      const result = await $store.dispatch('clustermanager/updateClusterAutoScaling', {
        ...autoscalerData.value,
        isScaleDownEnable: value,
        updater: user.value.username,
        $clusterId: props.clusterId,
      });
      configLoading.value = false;
      if (result) {
        $bkMessage({
          theme: 'success',
          message: $i18n.t('更新成功'),
        });
        resolve(true);
      } else {
        // eslint-disable-next-line prefer-promise-reject-errors
        reject(false);
      }
    });
    // 获取节点池列表
    const renderHeader = (h: CreateElement, data) => h('span', {
      class: 'custom-header-cell',
      directives: [
        {
          name: 'bkTooltips',
          value: {
            content: $i18n.t('已启用时会作为Cluster Autocaler的资源池，已关闭时不会作为Cluster Autocaler的资源池'),
          },
        },
      ],
    }, [data.column.label]);
    const statusTextMap = { // 节点池状态
      RUNNING: $i18n.t('正常'),
      CREATING: $i18n.t('创建中'),
      DELETING: $i18n.t('删除中'),
      UPDATING: $i18n.t('更新中'),
      DELETED: $i18n.t('已删除'),
      'CREATE-FAILURE': $i18n.t('创建失败'),
      'UPDATE-FAILURE': $i18n.t('更新失败'),
    };
    const statusColorMap = {
      RUNNING: 'green',
      DELETED: 'gray',
      'CREATE-FAILURE': 'red',
      'UPDATE-FAILURE': 'red',
    };
    const nodepoolList = ref<any[]>([]);
    const nodepoolLoading = ref(false);
    const {
      pagination,
      curPageData,
      pageChange,
      pageSizeChange,
    } = usePage(nodepoolList);
    const getNodePoolList = async () => {
      nodepoolList.value = await $store.dispatch('clustermanager/nodeGroup', {
        clusterID: props.clusterId,
      });
      if (!nodepoolList.value.some(pool => [
        'CREATING',
        'DELETING',
        'UPDATING',
      ].includes(pool.status))) {
        stopPoolInterval();
      } else {
        startPoolInterval();
      }
    };
    const handleGetNodePoolList = async () => {
      nodepoolLoading.value = true;
      await getNodePoolList();
      if (nodepoolList.value.some(pool => [
        'CREATING',
        'DELETING',
        'UPDATING',
      ].includes(pool.status))) {
        startPoolInterval();
      }
      nodepoolLoading.value = false;
    };
    const { start: startPoolInterval, stop: stopPoolInterval } = useInterval(getNodePoolList, 5000); // 轮询
    // 节点池详情
    const handleGotoDetail = (row) => {
      $router.push({
        name: 'nodePoolDetail',
        params: {
          clusterId: props.clusterId,
          nodeGroupID: row.nodeGroupID,
        },
      });
    };
    // 至少保证一个节点池处于开启状态
    const disabledAutoscaler = computed(() => autoscalerData.value.enableAutoscale
                    && nodepoolList.value.filter(item => item.enableAutoscale).length <= 1);
    // 单节点开启和关闭弹性伸缩
    const { proxy } = getCurrentInstance() || { proxy: null };
    const handleToggleNodeScaler = async (row) => {
      if (nodepoolLoading.value || ['CREATING', 'DELETING', 'UPDATING'].includes(row.status)) return;

      const $refs = proxy?.$refs || {};
      $refs[row.nodeGroupID] && ($refs[row.nodeGroupID] as any).hideHandler();
      nodepoolLoading.value = true;
      let result = false;
      if (row.enableAutoscale) {
        // 关闭时校验是否时最后一个开启状态
        if (disabledAutoscaler.value) {
          nodepoolLoading.value = false;
          return;
        }
        // 关闭
        result = await $store.dispatch('clustermanager/disableNodeGroupAutoScale', {
          $nodeGroupID: row.nodeGroupID,
        });
      } else {
        // 启用
        result = await $store.dispatch('clustermanager/enableNodeGroupAutoScale', {
          $nodeGroupID: row.nodeGroupID,
        });
      }
      if (result) {
        $bkMessage({
          theme: 'success',
          message: $i18n.t('操作成功'),
        });
        await handleGetNodePoolList();
      }
      nodepoolLoading.value = false;
    };
    // 当前操作行
    const currentOperateRow = ref<Record<string, any>>({});
    // 删除node pool
    const disabledDelete = computed(() =>
    // 至少保证一个节点池
      autoscalerData.value.enableAutoscale
                    && nodepoolList.value.length <= 1);
    const handleDeletePool = (row) => {
      if (disabledDelete.value || !!row.autoScaling.desiredSize) return;

      $bkInfo({
        type: 'warning',
        clsName: 'custom-info-confirm',
        title: $i18n.t('确定删除节点池 {name} ', { name: `${row.nodeGroupID}（${row.name}）` }),
        defaultInfo: true,
        confirmFn: async () => {
          const result = await $store.dispatch('clustermanager/deleteNodeGroup', {
            $nodeGroupID: row.nodeGroupID,
            operator: user.value.username,
          });
          if (result) {
            $bkMessage({
              theme: 'success',
              message: $i18n.t('操作成功'),
            });
            handleGetNodePoolList();
          }
        },
      });
    };
    // 节点池节点数量管理
    const nodeStatusMap = {
      INITIALIZATION: $i18n.t('初始化中'),
      RUNNING: $i18n.t('正常'),
      DELETING: $i18n.t('删除中'),
      DELETED: $i18n.t('已删除'),
      'DELETE-FAILURE': $i18n.t('删除失败'),
      'ADD-FAILURE': $i18n.t('扩容节点失败'),
      'REMOVE-FAILURE': $i18n.t('缩容节点失败'),
      REMOVABLE: $i18n.t('不可调度'),
    };
    const nodeColorMap = {
      RUNNING: 'green',
      'DELETE-FAILURE': 'red',
      'ADD-FAILURE': 'red',
      'REMOVE-FAILURE': 'red',
    };
    const nodeListLoading = ref(false);
    const nodeList = ref<any[]>([]);
    const {
      pagination: nodePagination,
      curPageData: nodeCurPageData,
      pageChange: nodePageChange,
      pageSizeChange: nodePageSizeChange,
    } = usePage(nodeList);
    const showNodeManage = ref(false);
    const handleNodeManageCancel = () => {
      currentOperateRow.value = {};
      nodeList.value = [];
      stopNodeInterval();
    };
    const handleShowNodeManage = (row) => {
      currentOperateRow.value = row;
      showNodeManage.value = true;
      handleGetNodeList();
    };
    const getNodeList = async () => {
      nodeList.value = await $store.dispatch('clustermanager/nodeGroupNodeList', {
        $nodeGroupID: currentOperateRow.value.nodeGroupID,
        output: 'wide',
      });
      if (!nodeList.value.some(node => ['DELETING', 'INITIALIZATION'].includes(node.status))) {
        stopNodeInterval();
      } else {
        startNodeInterval();
      }
    };
    const handleGetNodeList = async () => {
      nodeListLoading.value = true;
      await getNodeList();
      if (nodeList.value.some(node => ['DELETING', 'INITIALIZATION'].includes(node.status))) {
        startNodeInterval();
      }
      nodeListLoading.value = false;
    };
    const { start: startNodeInterval, stop: stopNodeInterval } = useInterval(getNodeList, 5000); // 轮询
    const handleNodeDrain = async (row) => {
      if (nodeListLoading.value) return;
      // POD迁移
      nodeListLoading.value = true;
      const result = await $store.dispatch('clustermanager/clusterNodeDrain', {
        innerIPs: [row.innerIP],
        clusterID: props.clusterId,
        updater: user.value.username,
      });
      if (result) {
        $bkMessage({
          theme: 'success',
          message: $i18n.t('POD迁移成功'),
        });
        await getNodeList();
      }
      nodeListLoading.value = false;
    };
    const handleDeleteNodeGroupNode = async (row) => {
      if (nodeListLoading.value || !row.unSchedulable) return;
      // 删除节点组节点
      nodeListLoading.value = true;
      const result = await $store.dispatch('clustermanager/deleteNodeGroupNode', {
        $nodeGroupID: currentOperateRow.value.nodeGroupID,
        nodes: row.innerIP,
        clusterID: props.clusterId,
        operator: user.value.username,
      });
      if (result) {
        $bkMessage({
          theme: 'success',
          message: $i18n.t('操作成功'),
        });
        await Promise.all([
          handleGetNodePoolList(),
          getNodeList(),
        ]);
      }
      nodeListLoading.value = false;
    };
    const handleToggleCordon = async (row) => {
      // 停止调度和允许调度
      nodeListLoading.value = true;
      let result = false;
      if (row.unSchedulable) {
        // 允许调度
        result = await $store.dispatch('clustermanager/nodeUnCordon', {
          innerIPs: [row.innerIP],
          clusterID: props.clusterId,
          updater: user.value.username,
        });
      } else {
        // 停止调度
        result = await $store.dispatch('clustermanager/nodeCordon', {
          innerIPs: [row.innerIP],
          clusterID: props.clusterId,
          updater: user.value.username,
        });
      }
      if (result) {
        $bkMessage({
          theme: 'success',
          message: $i18n.t('调度成功'),
        });
        await getNodeList();
      }
      nodeListLoading.value = false;
    };

    // 扩缩容记录
    const taskStatusMap = {
      INITIALIZING: $i18n.t('初始化中'),
      RUNNING: $i18n.t('执行中'),
      SUCCESS: $i18n.t('执行成功'),
      FAILURE: $i18n.t('执行失败'),
      TIMEOUT: $i18n.t('执行超时'),
      FORCETERMINATE: $i18n.t('强制终止'),
      NOTSTARTED: $i18n.t('未启动'),
    };
    const taskStatusColorMap = {
      INITIALIZING: 'green',
      RUNNING: 'green',
      SUCCESS: 'green',
      FAILURE: 'red',
      TIMEOUT: 'red',
      FORCETERMINATE: 'red',
      NOTSTARTED: 'gray',
    };
    const shortcuts = ref([
      {
        text: $i18n.t('今天'),
        value() {
          const end = new Date();
          const start = new Date(end.getFullYear(), end.getMonth(), end.getDate());
          return [start, end];
        },
      },
      {
        text: $i18n.t('近7天'),
        value() {
          const end = new Date();
          const start = new Date();
          start.setTime(start.getTime() - 3600 * 1000 * 24 * 7);
          return [start, end];
        },
      },
      {
        text: $i18n.t('近15天'),
        value() {
          const end = new Date();
          const start = new Date();
          start.setTime(start.getTime() - 3600 * 1000 * 24 * 15);
          return [start, end];
        },
      },
      {
        text: $i18n.t('近30天'),
        value() {
          const end = new Date();
          const start = new Date();
          start.setTime(start.getTime() - 3600 * 1000 * 24 * 30);
          return [start, end];
        },
      },
    ]);
    const timeRange = ref<Date[]>([]);
    const showRecord = ref(false);
    const recordLoading = ref(false);
    const recordList = ref<any[]>([]);
    const recordPagination = ref({
      current: 1,
      limit: 10,
      count: 0,
    });
    const recordPageChange = (page) => {
      recordPagination.value.current = page;
      handleGetRecordList();
    };
    const recordPageSizeChange = (limit) => {
      recordPagination.value.current = 1;
      recordPagination.value.limit = limit;
      handleGetRecordList();
    };
    const handleTimeRangeChange = () => {
      handleGetRecordList();
    };
    const handleShowRecord = (row) => {
      const end = new Date();
      const start = new Date(end.getFullYear(), end.getMonth(), end.getDate());
      timeRange.value = [
        start,
        end,
      ];
      currentOperateRow.value = row;
      showRecord.value = true;
      handleGetRecordList();
    };
    const handleRecordCancel = () => {
      currentOperateRow.value = {};
      recordPagination.value = {
        current: 1,
        limit: 10,
        count: 0,
      };
    };
    const handleGetRecordList = async () => {
      recordLoading.value = true;
      const { results = [], count = 0 } = await $store.dispatch('clustermanager/clusterAutoScalingLogs', {
        resourceType: 'nodegroup',
        resourceID: currentOperateRow.value.nodeGroupID,
        startTime: Math.floor(new Date(timeRange.value[0]).getTime() / 1000),
        endTime: Math.floor(new Date(timeRange.value[1]).getTime() / 1000),
        limit: recordPagination.value.limit,
        page: recordPagination.value.current,
      });
      recordList.value = results;
      recordPagination.value.count = count;
      recordLoading.value = false;
    };

    // 编辑自动扩缩容
    const handleEditAutoScaler = () => {
      $router.push({
        name: 'autoScalerConfig',
        params: {
          clusterId: props.clusterId,
        },
      });
    };
    // 新建节点池
    const handleCreatePool = () => {
      $router.push({
        name: 'nodePool',
        params: {
          clusterId: props.clusterId,
        },
      });
    };
    // 编辑节点池
    const handleEditPool = (row) => {
      $router.push({
        name: 'editNodePool',
        params: {
          clusterId: props.clusterId,
          nodeGroupID: row.nodeGroupID,
        },
      });
    };
    onMounted(() => {
      handleGetAutoScalerConfig();
      handleGetNodePoolList();
    });
    onBeforeUnmount(() => {
      stop();
      stopPoolInterval();
      stopNodeInterval();
    });
    return {
      disabledAutoscaler,
      disabledDelete,
      currentOperateRow,
      nodeListLoading,
      showNodeManage,
      configLoading,
      nodepoolLoading,
      pagination,
      curPageData,
      pageChange,
      pageSizeChange,
      nodeStatusMap,
      nodeColorMap,
      nodePagination,
      nodeCurPageData,
      handleNodeManageCancel,
      nodePageChange,
      nodePageSizeChange,
      handleNodeDrain,
      handleDeleteNodeGroupNode,
      handleToggleCordon,
      autoscalerData,
      basicScalerConfig,
      autoScalerConfig,
      autoScalerDownConfig,
      statusColorMap,
      statusTextMap,
      showRecord,
      timeRange,
      recordLoading,
      handleTimeRangeChange,
      shortcuts,
      recordPagination,
      recordList,
      renderHeader,
      recordPageChange,
      recordPageSizeChange,
      handleRecordCancel,
      handleGotoDetail,
      handleShowNodeManage,
      handleToggleNodeScaler,
      handleDeletePool,
      handleToggleAutoScaler,
      handleGetNodeList,
      handleShowRecord,
      handleEditAutoScaler,
      handleCreatePool,
      handleChangeScalerDown,
      handleEditPool,
      taskStatusMap,
      taskStatusColorMap,
    };
  },
});
</script>
<style lang="postcss" scoped>
.autoscaler-management {
    padding: 0px 32px 20px 32px;
    font-size: 12px;
    .group-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 20px 0;
        &-title {
            font-size: 14px;
            font-weight: bold;
            color: #63656E;
        }
    }
    .autoscaler {
        .switch-autoscaler {
            margin-left: 16px;
            padding-left: 16px;
            border-left: 1px solid #DCDEE5;
            color: #63656E;
        }
    }
    .nodepool {
        border-top: 1px solid #DCDEE5;
    }
    .disabled {
        color: #C4C6CC;
    }
}
.flex-between {
    display: flex;
    align-items: center;
    justify-content: space-between;
}
.operate {
    display: flex;
    align-items: center;
    >>> .more-icon {
        display: flex;
        align-items: center;
        justify-content: center;
        color: #63656E;
        font-size: 18px;
        cursor: pointer;
        margin-top: 2px;
        color: #3A84FF;
        &:hover {
            background: #eaf2ff;
            border-radius: 50%;
        }
    }
}
>>> .form-content {
    display: flex;
    flex-wrap: wrap;
    &-item {
        height: 32px;
        margin-top: 0;
        font-size: 12px;
        width: 100%;
    }
    .bk-label {
        font-size: 12px;
        color: #979BA5;
        text-align: left;
    }
    .bk-form-content {
        font-size: 12px;
        color: #313238;
        display: flex;
        align-items: center;
    }
}
>>> .dropdown-item {
    height: 32px;
    line-height: 32px;
    padding: 0 16px;
    color: #63656e;
    font-size: 12px;
    text-decoration: none;
    white-space: nowrap;
    cursor: pointer;
    &:hover {
        background-color: #eaf3ff;
        color: #3a84ff;
    }
    &.disabled {
        color: #C4C6CC;
        cursor: not-allowed;
    }
}
.mw88 {
    min-width: 88px;
}
.delete-content {
    display: flex;
    flex-direction: column;
    align-items: center;
}
.delete-title {
    color: #313238;
    font-size: 20px;
}
>>> .custom-header-cell {
    text-decoration: underline;
    text-decoration-style: dashed;
    text-underline-position: under;
}
</style>
