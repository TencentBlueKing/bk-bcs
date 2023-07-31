<!-- eslint-disable max-len -->
<template>
  <div class="biz-content">
    <Header title="Polaris" :desc="`(${$t('plugin.tools.cluster', { name: clusterName })})`"></Header>
    <div class="biz-content-wrapper" style="padding: 0;" v-bkloading="{ isLoading: isInitLoading, opacity: 0.1 }">
      <template v-if="!isInitLoading">
        <div class="biz-panel-header">
          <div class="left">
            <bk-button icon="plus" type="primary" @click.stop.prevent="createPolarisRules">
              <span>{{$t('plugin.tools.create')}}</span>
            </bk-button>
          </div>
          <div class="right search-wrapper">
            <div class="left">
              <bk-selector
                style="width: 180px;"
                :searchable="true"
                :placeholder="$t('plugin.tools.polarisNS')"
                :selected.sync="searchParams.polaris_ns"
                :list="polarisNameSpaceList"
                :setting-key="'name'"
                :display-key="'name'"
                :allow-clear="true">
              </bk-selector>
              <bkbcs-input
                style="width: 180px;"
                class="ml-[5px]"
                :placeholder="$t('plugin.tools.polarisSvcName')"
                :value.sync="searchParams.polaris_name">
              </bkbcs-input>
            </div>
            <div class="left">
              <bk-button type="primary" :title="$t('generic.button.query')" icon="search" @click="handleSearch">
                {{$t('generic.button.query')}}
              </bk-button>
            </div>
          </div>
        </div>
        <div class="biz-crd-instance">
          <div class="biz-table-wrapper">
            <bk-table
              class="biz-namespace-table"
              v-bkloading="{ isLoading: isPageLoading && !isInitLoading }"
              :size="'medium'"
              :data="curPageData"
              :pagination="pageConf"
              @page-change="handlePageChange"
              @page-limit-change="handlePageSizeChange">
              <bk-table-column :label="$t('plugin.tools._ruleName')" prop="name" :show-overflow-tooltip="true" min-width="120">
                <template slot-scope="{ row }">
                  <p class="polaris-cell-item">{{ row.name }}</p>
                </template>
              </bk-table-column>
              <bk-table-column :label="$t('plugin.tools.cluster_ns')" min-width="220">
                <template slot-scope="{ row }">
                  <div>
                    <p class="polaris-cell-item" style="padding-bottom: 5px;" :title="clusterName">{{ $t('generic.label.cluster1') }}：{{ clusterName }}</p>
                    <p class="polaris-cell-item" :title="row.namespace">{{ $t('k8s.namespace') }}：{{ row.namespace }}</p>
                  </div>
                </template>
              </bk-table-column>
              <bk-table-column :label="$t('plugin.tools.polarisInfo')" min-width="180" :show-overflow-tooltip="true">
                <template slot-scope="{ row }">
                  <div>
                    <p class="polaris-cell-item" style="padding-bottom: 5px;">
                      <span class="label">{{ $t('generic.label.name') }}：</span>
                      <span>{{ row.crd_data.polaris.name }}</span>
                    </p>
                    <p class="polaris-cell-item">
                      <span class="label">{{ $t('k8s.namespace') }}：</span>
                      <span>{{ row.crd_data.polaris.namespace }}</span>
                    </p>
                  </div>
                </template>
              </bk-table-column>
              <bk-table-column :label="$t('deploy.templateset.service')" min-width="160" :show-overflow-tooltip="true">
                <template slot-scope="{ row }">
                  <div v-for="(service, index) in row.crd_data.services" :key="index" style="padding: 5px 0;">
                    <span style="display: inline-block; position: relative; bottom: 37px;">-</span>
                    <span style="display: inline-block;">
                      <p class="polaris-cell-item">name: {{ service.name }}</p>
                      <p class="polaris-cell-item">port: <span style="padding-left: 8px;">{{ service.port }}</span></p>
                      <p class="polaris-cell-item">direct: {{ service.direct === 'true' ? `${$t('units.boolean.true')}` : `${$t('units.boolean.false')}` }}</p>
                    </span>
                  </div>
                </template>
              </bk-table-column>
              <bk-table-column label="ip: port weight" min-width="180" :show-overflow-tooltip="true">
                <template slot-scope="{ row }">
                  <div v-if="row.status && row.status.syncStatus && row.status.syncStatus.lastRemoteInstances">
                    <div v-for="(remote, remoteIndex) in row.status.syncStatus.lastRemoteInstances" :key="remoteIndex" style="padding: 5px 0;">
                      <p class="polaris-cell-item">{{ remote.ip || '--' }}: {{ remote.port || '--' }} {{ remote.weight || '--' }}</p>
                    </div>
                  </div>
                  <div v-else>--</div>
                </template>
              </bk-table-column>
              <bk-table-column :label="$t('generic.label.status')" width="270">
                <template slot-scope="{ row }">
                  <div v-if="row.status && row.status.syncStatus" style="padding: 5px 0;">
                    <p class="polaris-cell-item" style="padding-bottom: 5px;" :title="row.status.syncStatus.state">{{ $t('plugin.tools.syncStatus') }}：{{ row.status.syncStatus.state || '--' }}</p>
                    <p class="polaris-cell-item" style="padding-bottom: 5px;" :title="row.status.syncStatus.lastSyncLatencyomitempty">{{ $t('plugin.tools.syncTime') }}：{{ row.status.syncStatus.lastSyncLatencyomitempty || '--' }}</p>
                    <p class="polaris-cell-item" :title="row.status.syncStatus.lastSyncTime">{{ $t('plugin.tools.lastSyncTime') }}：{{ row.status.syncStatus.lastSyncTime || '--' }}</p>
                  </div>
                  <div v-else>--</div>
                </template>
              </bk-table-column>
              <bk-table-column :label="$t('projects.operateAudit.record')" min-width="240">
                <template slot-scope="{ row }">
                  <p class="polaris-cell-item" style="padding-bottom: 5px;">{{ $t('generic.label.updator') }}：<span style="padding-left: 14px;">{{ row.operator || '--' }}</span></p>
                  <p class="polaris-cell-item" :title="row.updated">{{ $t('cluster.labels.updatedAt') }}：{{row.updated || '--'}}</p>
                </template>
              </bk-table-column>
              <bk-table-column :label="$t('generic.label.action')" min-width="140">
                <template slot-scope="{ row }">
                  <a href="javascript:void(0);" class="bk-text-button" @click="editCrdInstance(row)">{{$t('generic.button.update')}}</a>
                  <a href="javascript:void(0);" class="bk-text-button" @click="removeCrdInstance(row)">{{$t('generic.button.delete')}}</a>
                </template>
              </bk-table-column>
            </bk-table>
          </div>
        </div>
      </template>
    </div>
    <!-- 新建/更新 -->
    <bk-sideslider
      quick-close
      :is-show.sync="crdInstanceSlider.isShow"
      :title="crdInstanceSlider.title"
      :width="800"
      :before-close="handleBeforeClose">
      <div class="p30" slot="content">
        <div class="bk-form bk-form-vertical">
          <div class="bk-form-item">
            <div class="bk-form-content">
              <div class="bk-form-inline-item is-required" style="width: 320px;">
                <label class="bk-label">{{$t('plugin.tools._ruleName')}}：</label>
                <div class="bk-form-content">
                  <bkbcs-input
                    :placeholder="$t('generic.placeholder.input')"
                    :disabled="isReadonly"
                    :value.sync="curCrdInstance.name">
                  </bkbcs-input>
                </div>
              </div>
            </div>
          </div>

          <!-- <div class="bk-form-item">
            <div class="bk-form-content">
              <div class="bk-form-inline-item" style="width: 320px;">
                <label class="bk-label">{{$t('cluster.create.label.desc')}}：</label>
                <div class="bk-form-content">
                  <bk-input
                    type="textarea"
                    :placeholder="$t('generic.placeholder.input')"
                    :disabled="isReadonly"
                    v-model="curCrdInstance.description">
                  </bk-input>
                </div>
              </div>
            </div>
          </div> -->

          <div class="bk-form-item">
            <div class="bk-form-content">
              <div class="bk-form-inline-item is-required" style="width: 320px;">
                <label class="bk-label">{{$t('generic.label.cluster1')}}：</label>
                <div class="bk-form-content">
                  <bkbcs-input
                    :value.sync="clusterName"
                    :disabled="true">
                  </bkbcs-input>
                </div>
              </div>

              <div class="bk-form-inline-item is-required" style="width: 320px; margin-left: 35px;">
                <label class="bk-label">{{$t('k8s.namespace')}}：</label>
                <div class="bk-form-content">
                  <bk-selector
                    :searchable="true"
                    :placeholder="$t('generic.placeholder.select')"
                    :disabled="isReadonly"
                    :selected.sync="curCrdInstance.namespace"
                    :list="nameSpaceList">
                  </bk-selector>
                </div>
              </div>
            </div>
          </div>

          <div class="bk-form-item">
            <label class="bk-label">{{$t('plugin.tools.polarisInfo')}}：</label>
          </div>
          <div class="bk-form-item">
            <div class="bk-form-content">
              <div class="polaris-wrapper polaris-info">
                <div class="bk-form-inline-item is-required" style="width: 299px;">
                  <label class="bk-label">{{$t('generic.label.name')}}：</label>
                  <div class="bk-form-content">
                    <bkbcs-input
                      :value.sync="curCrdInstance.polaris.name"
                      :disabled="isReadonly"
                      :placeholder="$t('plugin.tools.allowNumLettersSymbols')">
                    </bkbcs-input>
                  </div>
                </div>
                <div class="bk-form-inline-item is-required" style="width: 320px; margin-left: 35px;">
                  <label class="bk-label">{{$t('k8s.namespace')}}：</label>
                  <div class="bk-form-content">
                    <bk-selector
                      :searchable="true"
                      :placeholder="$t('generic.placeholder.select')"
                      :selected.sync="curCrdInstance.polaris.namespace"
                      :disabled="isReadonly"
                      :list="polarisNameSpaceList">
                    </bk-selector>
                  </div>
                </div>
                <div class="bk-form-inline-item" style="width: 299px; margin-top: 30px; margin-right: 35px;">
                  <bk-checkbox v-model="isTokenExist" :disabled="isReadonly" name="cluster-classify-checkbox">
                    {{$t('plugin.tools.polaris')}}
                  </bk-checkbox>
                </div>
                <div v-if="isTokenExist" class="bk-form-inline-item" style="width: 350px; margin-top: 10px; height: 64px;">
                  <label class="bk-label">token：</label>
                  <div class="bk-form-content token">
                    <bkbcs-input
                      class="basic-input"
                      :placeholder="$t('generic.placeholder.input')"
                      :disabled="isReadonly"
                      :value.sync="curCrdInstance.polaris.token">
                    </bkbcs-input>
                    <i class="bcs-icon bcs-icon-question-circle token-icon ml10" v-bk-tooltips.top="$t('plugin.tools.createSvc')" />
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div class="bk-form-item">
            <label class="bk-label">{{$t('deploy.templateset.service')}}：</label>
          </div>
          <div class="bk-form-item">
            <div class="bk-form-content">
              <div class="polaris-wrapper">
                <section class="polaris-inner-wrapper mb10" v-for="(service, index) in curCrdInstance.services" :key="index">
                  <div class="bk-form-inline-item is-required" style="width: 284px;">
                    <label class="bk-label">{{$t('plugin.tools._serviceName')}}：</label>
                    <div class="bk-form-content">
                      <bkbcs-input
                        :placeholder="$t('generic.placeholder.input')"
                        :value.sync="service.name">
                      </bkbcs-input>
                    </div>
                  </div>
                  <div class="bk-form-inline-item is-required" style="width: 319px; margin-left: 35px;">
                    <label class="bk-label">{{$t('deploy.helm.port')}}：</label>
                    <div class="bk-form-content">
                      <bkbcs-input
                        :placeholder="$t('generic.placeholder.input')"
                        :value.sync="service.port">
                      </bkbcs-input>
                    </div>
                  </div>
                  <div class="bk-form-inline-item is-required" style="width: 284px; margin-top: 30px;">
                    <div class="bk-form-content">
                      <bk-checkbox v-model="service.direct" name="cluster-classify-checkbox">
                        {{$t('plugin.tools.pod')}}
                      </bk-checkbox>
                      <i class="bcs-icon bcs-icon-question-circle" v-bk-tooltips.top="$t('plugin.tools.nodePort')" />
                    </div>
                  </div>
                  <div class="bk-form-inline-item is-required" style="width: 319px; margin-top: 10px; margin-left: 35px;">
                    <label class="bk-label">{{$t('plugin.tools.weight')}}：</label>
                    <div class="bk-form-content">
                      <bkbcs-input
                        :placeholder="$t('generic.placeholder.input')"
                        :value.sync="service.weight">
                      </bkbcs-input>
                    </div>
                  </div>

                  <i class="bcs-icon bcs-icon-close polaris-close" @click="removeServiceMap(index)" v-if="curCrdInstance.services.length > 1"></i>
                </section>

                <bk-button class="polaris-block-btn mt10" @click="addServiceMap">
                  <i class="bcs-icon bcs-icon-plus"></i>
                  {{$t('plugin.tools.clickToAdd')}}
                </bk-button>
              </div>
            </div>
          </div>

          <div class="bk-form-item mt25">
            <bk-button type="primary" :loading="isDataSaveing" @click.stop.prevent="saveCrdInstance">{{curCrdInstance.crd_id ? $t('generic.button.update') : $t('generic.button.create')}}</bk-button>
            <bk-button @click.stop.prevent="hideCrdInstanceSlider" :disabled="isDataSaveing">{{$t('generic.button.cancel')}}</bk-button>
          </div>
        </div>
      </div>
    </bk-sideslider>
  </div>
</template>

<script lang="ts">
import { defineComponent, reactive, toRefs, computed, onMounted, watch, toRef, h } from 'vue';
import { useNamespace } from '@/views/resource-view/namespace/use-namespace';
import Header from '@/components/layout/Header.vue';
import useSideslider from '@/composables/use-sideslider';
import $store from '@/store';
import $router from '@/router';
import $i18n from '@/i18n/i18n-setup';
import $bkMessage from '@/common/bkmagic';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';

export default defineComponent({
  name: 'CrdcontrollerPolarisInstances',
  components: {
    Header,
  },
  setup() {
    const $route = computed(() => toRef(reactive($router), 'currentRoute').value);

    const crdInstanceList = computed(() => Object.assign([], $store.state.crdcontroller.crdInstanceList));
    const clusterList = computed(() => $store.state.cluster.clusterList);
    const curProject = computed(() => $store.state.curProject).value;
    const projectId = computed(() => $route.value.params.projectId).value;
    const clusterId = computed(() => $route.value.params.clusterId).value;
    const clusterName = computed(() => {
      const cluster = clusterList.value.find(item => item.cluster_id === clusterId);
      return cluster ? cluster.name : '';
    });

    const state = reactive({
      crdKind: 'PolarisConfig',
      isInitLoading: true,
      isDataSaveing: false,
      isPageLoading: false,
      isTokenExist: false,
      isReadonly: false,
      pageConf: {
        count: 0,
        totalPage: 1,
        limit: 5,
        current: 1,
        show: true,
      },
      nameSpaceList: [],
      polarisNameSpaceList: [
        {
          id: 'Production',
          name: 'Production',
        },
        {
          id: 'Pre-release',
          name: 'Pre-release',
        },
        {
          id: 'Test',
          name: 'Test',
        },
        {
          id: 'Development',
          name: 'Development',
        },
      ],
      curPageData: [],
      crdInstanceSlider: {
        title: $i18n.t('plugin.tools.add'),
        isShow: false,
      },
      curCrdInstance: {
        name: '',
        namespace: '',
        polaris: {
          name: '',
          namespace: '',
          token: '',
        },
        services: [
          {
            name: '',
            port: '',
            direct: true,
            weight: '',
          },
        ],
      },
      searchParams: {
        polaris_name: '',
        polaris_ns: '',
      },
      appTypes: [
        {
          id: 'polaris_name',
          name: $i18n.t('plugin.tools.polarisSvcName'),
        },
      ],
    });

    const { handleBeforeClose, reset } = useSideslider(toRef(state, 'curCrdInstance'));

    watch(crdInstanceList, async () => {
      state.curPageData = await getDataByPage(state.pageConf.current);
    });

    onMounted(() => {
      getCrdInstanceList();
      getNameSpaceList();
    });

    const goBack = () => {
      $router.push({
        name: 'dbCrdcontroller',
        params: {
          projectId,
        },
      });
    };

    /**
             * 新建规则
             */
    const createPolarisRules = () => {
      state.curCrdInstance = {
        name: '',
        namespace: '',
        polaris: {
          name: '',
          namespace: '',
          token: '',
        },
        services: [{
          name: '',
          port: '',
          direct: true,
          weight: '',
        }],
      };
      state.crdInstanceSlider.title = $i18n.t('plugin.tools.add');
      state.isTokenExist = false;
      state.isReadonly = false;
      state.crdInstanceSlider.isShow = true;
      reset();
    };

    /**
             * 搜索列表
             */
    const handleSearch = () => {
      state.pageConf.current = 1;
      state.isPageLoading = true;
      getCrdInstanceList();
    };

    /**
             * 加载数据
             */
    const getCrdInstanceList = async () => {
      const { crdKind } = state;
      const params = {};

      if (state.searchParams.polaris_name) {
        params.polaris_name = state.searchParams.polaris_name;
      }
      if (state.searchParams.polaris_ns) {
        params.polaris_ns = state.searchParams.polaris_ns;
      }

      const res = await $store.dispatch('crdcontroller/getCrdInstanceList', {
        projectId,
        clusterId,
        crdKind,
        params,
      });
      // 晚消失是为了防止整个页面loading和表格数据loading效果叠加产生闪动
      setTimeout(() => {
        state.isPageLoading = false;
        state.isInitLoading = false;
      }, 200);

      if (!res) return;
      initPageConf();
      state.curPageData = getDataByPage(state.pageConf.current);
    };

    /**
             * 初始化分页配置
             */
    const initPageConf = () => {
      const total = crdInstanceList.value.length;
      state.pageConf.count = total;
      state.pageConf.totalPage = Math.ceil(total / state.pageConf.limit);
      if (state.pageConf.current > state.pageConf.totalPage) {
        state.pageConf.current = state.pageConf.totalPage;
      }
    };

    /**
             * 获取页数据
             * @param  {number} page 页
             * @return {object} data lb
             */
    const getDataByPage = (page) => {
      // 如果没有page，重置
      if (!page) {
        // eslint-disable-next-line no-multi-assign
        state.pageConf.current = page = 1;
      }
      let startIndex = (page - 1) * state.pageConf.limit;
      let endIndex = page * state.pageConf.limit;
      // state.isPageLoading = true
      if (startIndex < 0) {
        startIndex = 0;
      }
      if (endIndex > crdInstanceList.value.length) {
        endIndex = crdInstanceList.value.length;
      }
      state.isPageLoading = false;
      return crdInstanceList.value.slice(startIndex, endIndex);
    };

    /**
             * 分页大小更改
             *
             * @param {number} pageSize pageSize
             */
    const handlePageSizeChange = (pageSize) => {
      state.pageConf.limit = pageSize;
      state.pageConf.current = 1;
      initPageConf();
      handlePageChange();
    };

    /**
             * 分页改变回调
             * @param  {number} page 页
             */
    const handlePageChange = (page = 1) => {
      state.isPageLoading = true;
      state.pageConf.current = page;
      const data = getDataByPage(page);
      state.curPageData = JSON.parse(JSON.stringify(data));
    };


    const { getNamespaceData } = useNamespace();

    /**
             * 获取命名空间列表
             */
    const getNameSpaceList = async () => {
      const res = await getNamespaceData({
        $clusterId: clusterId,
      });
      if (!res) return;
      const list = res;
      list.forEach((item) => {
        item.isSelected = false;
        item.id = item.name;
      });
      state.nameSpaceList.splice(0, state.nameSpaceList.length, ...list);
    };

    /**
             * 增加一个关联Service
             */
    const addServiceMap = () => {
      const params = {
        name: '',
        port: '',
        direct: true,
        weight: '',
      };
      state.curCrdInstance.services.push(params);
    };

    /**
             * 移除一个关联Service
             */
    const removeServiceMap = (index) => {
      state.curCrdInstance.services.splice(index, 1);
    };

    /**
             * 隐藏侧面板
             */
    const hideCrdInstanceSlider = () => {
      state.crdInstanceSlider.isShow = false;
    };

    /**
             * 保存新建/更新
             */
    const actionCrdInstance = async (params) => {
      let url = '';
      if (state.curCrdInstance.id > 0) {
        url = 'crdcontroller/updateCrdInstance';
        state.isReadonly = true;
      } else {
        url = 'crdcontroller/addCrdInstance';
        state.isReadonly = false;
      }

      const data = JSON.parse(JSON.stringify(params));
      data.services.forEach((item) => {
        item.direct = String(item.direct);
      });

      const { crdKind } = state;
      state.isDataSaveing = true;

      const result = await $store.dispatch(url, { projectId, clusterId, crdKind, data }).catch(() => false);
      state.isDataSaveing = false;

      if (!result) return;

      $bkMessage({
        theme: 'success',
        message: $i18n.t('generic.msg.success.save1'),
      });
      getCrdInstanceList();
      hideCrdInstanceSlider();
    };

    /**
             * 保存 / 更新
             */
    const saveCrdInstance = () => {
      const params = {
        ...state.curCrdInstance,
      };
      if (checkData() && !state.isDataSaveing) {
        actionCrdInstance(params);
      }
    };

    /**
             * 编辑
             * @param  {object} crdInstance crdInstance
             * @param  {number} index 索引
             */
    const editCrdInstance = async (crdInstance) => {
      const { crdKind } = state;
      const crdId = crdInstance.id;
      const res = await $store.dispatch('crdcontroller/getCrdInstanceDetail', {
        crdKind,
        projectId,
        clusterId,
        crdId,
      }).catch(() => false);
      state.crdInstanceSlider.isShow = true;
      state.isReadonly = true;
      if (!res) return;
      res.data.crd_data.services.forEach((item) => {
        item.direct === 'true' ? item.direct = true : item.direct = false;
      });
      state.curCrdInstance = res.data.crd_data;
      state.curCrdInstance.crd_id = crdId;
      state.crdInstanceSlider.title = $i18n.t('generic.button.edit');
      reset();
    };

    /**
             * 删除
             * @param  {object} crdInstance crdInstance
             * @param  {number} index 索引
             */
    const removeCrdInstance = async (crdInstance) => {
      const { crdKind } = state;
      const crdId = crdInstance.id;

      $bkInfo({
        title: $i18n.t('generic.title.confirmDelete'),
        clsName: 'biz-remove-dialog',
        content: `${$i18n.t('plugin.tools.confirmDelete')}【${crdInstance.name}】？`,
        async confirmFn() {
          state.isPageLoading = true;
          const res = await $store.dispatch('crdcontroller/deleteCrdInstance', { projectId, clusterId, crdKind, crdId }).catch(() => false);
          state.isPageLoading = false;

          if (!res) return;

          $bkMessage({
            theme: 'success',
            message: $i18n.t('generic.msg.success.delete'),
          });
          getCrdInstanceList();
        },
      });
    };

    /**
             * 检查提交的数据
             * @return {boolean} true/false 是否合法
             */
    const checkData = () => {
      if (!state.curCrdInstance.name) {
        $bkMessage({
          theme: 'error',
          message: $i18n.t('plugin.tools.rule'),
          delay: 5000,
        });
        return false;
      }

      if (state.curCrdInstance.name.length > 63) {
        $bkMessage({
          theme: 'error',
          message: $i18n.t('plugin.tools.ruleRegex'),
          delay: 5000,
        });
        return false;
      }

      if (!/^([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]$/.test(state.curCrdInstance.name)) {
        $bkMessage({
          theme: 'error',
          message: $i18n.t('plugin.tools.ruleCharacterCriteria'),
          delay: 5000,
        });
        return false;
      }

      if (state.curCrdInstance.namespace === '') {
        $bkMessage({
          theme: 'error',
          message: $i18n.t('dashboard.ns.validate.emptyNs'),
        });
        return false;
      }

      if (!state.curCrdInstance.polaris.name) {
        $bkMessage({
          theme: 'error',
          message: $i18n.t('plugin.tools.enterPolarisInfoName'),
          delay: 5000,
        });
        return false;
      }

      if (state.curCrdInstance.polaris.name && !/^[\w-.:]{1,128}$/.test(state.curCrdInstance.polaris.name)) {
        $bkMessage({
          theme: 'error',
          message: $i18n.t('plugin.tools.polarisInfoNameCriteria'),
          delay: 5000,
        });
        return false;
      }

      if (!state.curCrdInstance.polaris.namespace) {
        $bkMessage({
          theme: 'error',
          message: $i18n.t('plugin.tools.selectPolarisNamespace'),
          delay: 5000,
        });
        return false;
      }

      if (state.curCrdInstance.services.length) {
        let status = true;
        state.curCrdInstance.services.forEach((i) => {
          if (!i.name) {
            $bkMessage({
              theme: 'error',
              message: $i18n.t('plugin.tools.inputSvcName'),
              delay: 5000,
            });
            status = false;
          }
          if (!i.port) {
            $bkMessage({
              theme: 'error',
              message: $i18n.t('plugin.tools.portInt'),
              delay: 5000,
            });
            status = false;
          }
          if (i.port && i.port < 0) {
            $bkMessage({
              theme: 'error',
              message: $i18n.t('plugin.tools.portNoNegative'),
              delay: 5000,
            });
            status = false;
          }
          if (!i.weight) {
            $bkMessage({
              theme: 'error',
              message: $i18n.t('plugin.tools.weightInt'),
              delay: 5000,
            });
            status = false;
          }
          if (i.port && i.weight < 0) {
            $bkMessage({
              theme: 'error',
              message: $i18n.t('plugin.tools.weightNoNegative'),
              delay: 5000,
            });
            status = false;
          }
        });
        return status;
      }
      return true;
    };

    return {
      ...toRefs(state),
      projectId,
      crdInstanceList,
      curProject,
      clusterId,
      clusterName,
      goBack,
      handleSearch,
      handlePageChange,
      handlePageSizeChange,
      createPolarisRules,
      hideCrdInstanceSlider,
      editCrdInstance,
      removeCrdInstance,
      saveCrdInstance,
      addServiceMap,
      removeServiceMap,
      handleBeforeClose,
    };
  },
});
</script>

<style scoped>
    @import './polaris_list.css';
</style>
