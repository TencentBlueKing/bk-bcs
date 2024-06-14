<!-- eslint-disable vue/multi-word-component-names -->
<!-- eslint-disable max-len -->
<template>
  <BcsContent hide-back :title="crdKind === 'BcsLog' ? $t('nav.log') : $t('plugin.tools.title')">
    <div>
      <Row class="mb-[16px]">
        <div class="right" slot="right">
          <ClusterSelectComb
            :search.sync="searchKeyword"
            :cluster-id.sync="searchScope"
            :show-search="false"
            :cluster-type="['independent', 'managed', 'virtual']"
            @search-change="search"
            @cluster-change="search"
            @refresh="refresh" />
        </div>
      </Row>
      <div class="biz-crdcontroller" v-bkloading="{ isLoading: isPageLoading }" style="min-height: 180px;">
        <svg style="display: none;">
          <title>{{$t('deploy.templateset.icon')}}</title>
          <symbol id="biz-set-icon" viewBox="0 0 60 60">
            <g id="图层_6">
              <g id="图层_32_1_">
                <path
                  class="st0" d="M12,8v4H8c-1.1,0-2,0.9-2,2v42c0,1.1,0.9,2,2,2h42c1.1,0,2-0.9,2-2v-4h4c1.1,0,2-0.9,2-2V8c0-1.1-0.9-2-2-2
                                        H14C12.9,6,12,6.9,12,8z M48,48v4v2H10V16h2h4h32V48z M54,48h-2V14c0-1.1-0.9-2-2-2H16v-2h38V48z" />
              </g>
              <path
                class="st1" d="M45.7,33.7h-1.8l-3.4-8.3l1.3-1.3c0.5-0.5,0.5-1.3,0-1.8l0,0c-0.5-0.5-1.3-0.5-1.8,0l-1.3,1.3l-8.4-3.5v-1.8
                                    c0-0.7-0.6-1.3-1.3-1.3l0,0c-0.7,0-1.3,0.6-1.3,1.3V20l-8.4,3.5l-1.2-1.2c-0.5-0.5-1.3-0.5-1.8,0l0,0c-0.5,0.5-0.5,1.3,0,1.8
                                    l1.2,1.2L14,33.7h-1.8c-0.7,0-1.3,0.6-1.3,1.3l0,0c0,0.7,0.6,1.3,1.3,1.3H14l3.5,8.4L16.2,46c-0.5,0.5-0.5,1.3,0,1.8l0,0
                                    c0.5,0.5,1.3,0.5,1.8,0l1.3-1.3l8.3,3.4v1.8c0,0.7,0.6,1.3,1.3,1.3l0,0c0.7,0,1.3-0.6,1.3-1.3v-1.9l8.3-3.4l1.3,1.3
                                    c0.5,0.5,1.3,0.5,1.8,0l0,0c0.5-0.5,0.5-1.3,0-1.8l-1.3-1.3l3.4-8.3h1.9c0.7,0,1.3-0.6,1.3-1.3l0,0C47,34.3,46.4,33.7,45.7,33.7z
                                     M30.3,23.4l6,2.5l-4.6,4.6c-0.4-0.2-0.9-0.4-1.3-0.6v-6.5H30.3z M27.7,23.4V30c-0.5,0.1-0.9,0.3-1.4,0.6l-4.7-4.7L27.7,23.4z
                                     M19.9,27.7l4.7,4.7c-0.2,0.4-0.4,0.9-0.5,1.3h-6.6L19.9,27.7z M17.4,36.3H24c0.1,0.5,0.3,0.9,0.6,1.3l-4.7,4.7L17.4,36.3z
                                     M27.7,46.5l-6-2.5l4.7-4.7c0.4,0.2,0.8,0.4,1.3,0.5V46.5z M29,37.5c-1.4,0-2.6-1.2-2.6-2.6c0-1.4,1.2-2.6,2.6-2.6s2.6,1.2,2.6,2.6
                                    C31.6,36.4,30.4,37.5,29,37.5z M30.3,46.5v-6.6c0.5-0.1,0.9-0.3,1.3-0.5l4.6,4.6L30.3,46.5z M38,42.2l-4.6-4.6
                                    c0.2-0.4,0.4-0.8,0.6-1.3h6.5L38,42.2z M34,33.7c-0.1-0.5-0.3-0.9-0.5-1.3l4.6-4.6l2.5,6H34V33.7z" />
              <g class="st2">
                <path class="st3" d="M41,49H17c-1.1,0-2-0.9-2-2V23c0-1.1,0.9-2,2-2h24c1.1,0,2,0.9,2,2v24C43,48.1,42.1,49,41,49z" />
              </g>
              <g>
                <path
                  class="st0" d="M42.2,25c-1.9,0-2.9,0.5-2.9,1.5v17.1c0,1,1,1.5,2.9,1.5v1.8H31.4V45c2,0,3-0.5,3-1.5v-8H23.6v8
                                        c0,1,1,1.5,3,1.5v1.8H15.8V45c1.9,0,2.8-0.5,2.8-1.5V26.4c0-1-0.9-1.5-2.8-1.5V23h10.8v2c-2,0-3,0.5-3,1.5v6.8h10.8v-6.8
                                        c0-1-1-1.5-3-1.5v-1.9h10.8V25z" />
              </g>
            </g>
          </symbol>
        </svg>
        <table class="bk-table biz-templateset-table mb20" v-if="crdKind !== 'BcsLog'">
          <thead>
            <tr>
              <th style="width: 120px; padding-left: 0;" class="center">{{$t('plugin.tools.icon')}}</th>
              <th style="width: 250px; padding-left: 20px;">{{$t('plugin.tools.toolName')}}</th>
              <th style="width: 120px; padding-left: 20px">{{$t('generic.label.version')}}</th>
              <th style="width: 150px; padding-left: 20px;">{{$t('generic.label.status')}}</th>
              <th style="padding-left: 0;">{{$t('cluster.create.label.desc')}}</th>
              <th style="width: 170px; padding-left: 0;">{{$t('generic.label.action')}}</th>
            </tr>
          </thead>
          <tbody>
            <template v-if="crdControllerList.length">
              <tr
                v-for="crdcontroller of crdControllerList"
                :key="crdcontroller.id">
                <td colspan="6">
                  <table class="biz-inner-table">
                    <tr>
                      <td class="logo">
                        <div class="logo-wrapper" v-if="logMap[crdcontroller.chart_name]">
                          <i :class="logMap[crdcontroller.chart_name]"></i>
                        </div>
                        <svg class="biz-set-icon" v-else>
                          <use xlink:href="#biz-set-icon"></use>
                        </svg>
                      </td>
                      <td class="name" style="width: 250px;">
                        <p class="text">{{crdcontroller.name || '--'}}</p>
                      </td>
                      <td class="version" style="width: 120px;padding: 0 10px 0 20px;">
                        <p class="text">{{crdcontroller.currentVersion || crdcontroller.version || '--'}}</p>
                      </td>
                      <td class="status">
                        <span class="biz-mark" v-if="crdcontroller.status === 'deployed'">
                          <bk-tag type="filled" theme="success">{{statusTextMap[crdcontroller.status] || $t('plugin.tools.deployed')}}</bk-tag>
                        </span>
                        <span class="biz-mark" v-else-if="!crdcontroller.status">
                          <bk-tag type="filled">{{statusTextMap[crdcontroller.status] || $t('generic.status.notEnable')}}</bk-tag>
                        </span>
                        <span class="biz-mark" v-else-if="crdcontroller.status === 'unknown'">
                          <bcs-popover :content="$t('plugin.tools.contact')" placement="top">
                            <bk-tag type="filled" theme="warning">{{statusTextMap[crdcontroller.status] || $t('generic.status.unknown')}}</bk-tag>
                          </bcs-popover>
                        </span>
                        <template v-else-if="pendingStatus.includes(crdcontroller.status)">
                          <div class="bk-spin-loading bk-spin-loading-mini bk-spin-loading-primary vm" style="margin-right: 3px;">
                            <div class="rotate rotate1"></div>
                            <div class="rotate rotate2"></div>
                            <div class="rotate rotate3"></div>
                            <div class="rotate rotate4"></div>
                            <div class="rotate rotate5"></div>
                            <div class="rotate rotate6"></div>
                            <div class="rotate rotate7"></div>
                            <div class="rotate rotate8"></div>
                          </div>
                          <span class="vm">{{statusTextMap[crdcontroller.status] || $t('plugin.tools.doing')}}</span>
                        </template>
                        <span class="biz-mark" v-else>
                          <bcs-popover :width="500" :content="crdcontroller.message" placement="top">
                            <bk-tag type="filled" theme="danger">{{statusTextMap[crdcontroller.status] || $t('generic.status.error')}}</bk-tag>
                          </bcs-popover>
                        </span>
                      </td>
                      <td class="description">
                        <p class="text">
                          {{crdcontroller.description || '--'}}
                          <a :href="crdcontroller.help_link" class="bk-text-button f12" target="_blank" v-if="crdcontroller.help_link">{{$t('plugin.tools.docs')}}</a>
                        </p>
                      </td>
                      <td class="action">
                        <template v-if="crdcontroller.status === 'deployed'">
                          <bk-dropdown-menu
                            class="dropdown-menu"
                            :align="'left'"
                            ref="dropdown">
                            <bk-button :class="['bk-button bk-default btn']" slot="dropdown-trigger" style="position: relative; width: 88px;">
                              <span>{{$t('generic.label.action')}}</span>
                              <i class="bcs-icon bcs-icon-angle-down dropdown-menu-angle-down ml5" style="font-size: 10px;"></i>
                            </bk-button>

                            <ul class="bk-dropdown-list" slot="dropdown-content">
                              <li v-if="crdcontroller.supported_actions.includes('config')">
                                <a href="javascript:void(0)" @click="goControllerInstances(crdcontroller)">{{$t('plugin.tools.config')}}</a>
                              </li>
                              <li v-if="crdcontroller.supported_actions.includes('upgrade')">
                                <a href="javascript:void(0)" @click="showInstanceDetail(crdcontroller)">{{$t('plugin.tools.upgrade')}}</a>
                              </li>
                              <li v-if="crdcontroller.supported_actions.includes('uninstall')">
                                <a href="javascript:void(0)" @click="handleUninstall(crdcontroller)">{{$t('plugin.tools.uninstall')}}</a>
                              </li>
                            </ul>
                          </bk-dropdown-menu>
                        </template>
                        <template v-else-if="!crdcontroller.status">
                          <bk-button type="primary" @click="haneldEnableCrdController(crdcontroller)">{{$t('logCollector.action.enable')}}</bk-button>
                        </template>
                        <template v-else-if="failedStatus.includes(crdcontroller.status)">
                          <template
                            v-if="!crdcontroller.supported_actions.includes('upgrade')
                              && !crdcontroller.supported_actions.includes('uninstall')">
                            <bk-button type="primary" @click="haneldEnableCrdController(crdcontroller)">{{$t('plugin.tools.restart')}}</bk-button>
                          </template>
                          <template v-else>
                            <bk-dropdown-menu
                              class="dropdown-menu"
                              :align="'left'"
                              ref="dropdown">
                              <bk-button :class="['bk-button bk-default btn']" slot="dropdown-trigger" style="position: relative; width: 88px;">
                                <span>{{$t('generic.label.action')}}</span>
                                <i class="bcs-icon bcs-icon-angle-down dropdown-menu-angle-down ml5" style="font-size: 10px;"></i>
                              </bk-button>
                              <ul class="bk-dropdown-list" slot="dropdown-content">
                                <li v-if="crdcontroller.supported_actions.includes('upgrade')">
                                  <a href="javascript:void(0)" @click="showInstanceDetail(crdcontroller)">{{$t('plugin.tools.upgrade')}}</a>
                                </li>
                                <li v-if="crdcontroller.supported_actions.includes('uninstall')">
                                  <a href="javascript:void(0)" @click="handleUninstall(crdcontroller)">{{$t('plugin.tools.uninstall')}}</a>
                                </li>
                              </ul>
                            </bk-dropdown-menu>
                          </template>
                        </template>
                        <template v-else-if="crdcontroller.status === 'unknown'">
                          <span v-bk-tooltips="$t('plugin.tools.contact')">
                            <bk-button :disabled="true">{{$t('logCollector.action.enable')}}</bk-button>
                          </span>
                        </template>
                        <template v-else-if="pendingStatus.includes(crdcontroller.status)">
                          <bk-button :disabled="true">{{statusTextMap[crdcontroller.status] || $t('plugin.tools.doing')}}</bk-button>
                        </template>
                      </td>
                    </tr>
                  </table>
                </td>
              </tr>
            </template>
            <template v-if="!crdControllerList.length">
              <tr>
                <td colspan="5">
                  <bcs-exception type="empty" scene="part"></bcs-exception>
                </td>
              </tr>
            </template>
          </tbody>
        </table>


      </div>
    </div>

    <bk-sideslider
      class="editor-slider"
      :quick-close="false"
      :is-show.sync="valueSlider.isShow"
      :title="valueSlider.title"
      :width="900">
      <template #header>
        <div class="flex place-content-between">
          <div>{{ valueSlider.title }}</div>
          <div>
            <bk-button class="bk-button bk-primary save-crd-btn" @click.stop.prevent="enableCrdController">{{$t('logCollector.action.enable')}}</bk-button>
            <bk-button class="bk-button bk-default hide-crd-btn" @click.stop.prevent="hideApplicationJson">{{$t('generic.button.cancel')}}</bk-button>
          </div>
        </div>
      </template>
      <div class="p0" slot="content">
        <div :class="['diff-editor-box', { 'editor-fullscreen': editorOptions.fullScreen }]" style="position: relative;">
          <monaco-editor
            ref="yamlEditor"
            class="editor"
            theme="monokai"
            language="yaml"
            :style="{ height: `${editorHeight}px`, width: '100%' }"
            v-model="editorOptions.content"
            :diff-editor="editorOptions.isDiff"
            :key="renderEditorKey"
            :options="editorOptions"
            :original="editorOptions.originContent">
          </monaco-editor>
        </div>
      </div>
    </bk-sideslider>
  </BcsContent>
</template>

<script>
import { addonsDetail, addonsInstall, addonsList, addonsUninstall } from '@/api/modules/helm';
import { catchErrorHandler } from '@/common/util';
import ClusterSelectComb from '@/components/cluster-selector/cluster-select-comb.vue';
import BcsContent from '@/components/layout/Content.vue';
import Row from '@/components/layout/Row.vue';
import MonacoEditor from '@/components/monaco-editor/editor.vue';

export default {
  components: {
    MonacoEditor,
    BcsContent,
    ClusterSelectComb,
    Row,
  },
  data() {
    return {
      pendingStatus: ['uninstalling', 'pending-install', 'pending-upgrade', 'pending-rollback'],
      failedStatus: ['failed', 'failed-install', 'failed-upgrade', 'failed-rollback', 'failed-uninstall'],
      statusTextMap: {
        unknown: this.$t('generic.status.error'),
        deployed: this.$t('generic.status.ready'),
        uninstalled: this.$t('generic.status.deleted'),
        superseded: this.$t('deploy.helm.invalidate'),
        failed: this.$t('generic.status.failed'),
        uninstalling: this.$t('generic.status.deleting'),
        'pending-install': this.$t('deploy.helm.pending'),
        'pending-upgrade': this.$t('generic.status.updating'),
        'pending-rollback': this.$t('deploy.helm.pendingRollback'),
        'failed-install': this.$t('deploy.helm.failed'),
        'failed-upgrade': this.$t('generic.status.updateFailed'),
        'failed-rollback': this.$t('deploy.helm.rollbackFailed'),
        'failed-uninstall': this.$t('generic.status.deleteFailed'),
      },
      isInitLoading: true,
      isPageLoading: false,
      crdControllerList: [],
      crdControllerListCache: [],
      curCrdcontroller: null,
      searchKeyword: '',
      searchScope: '',
      statusTimer: {},
      valueSlider: {
        isShow: false,
        fullScreen: false,
        title: '',
      },
      renderEditorKey: 0,
      editorOptions: {
        readOnly: false,
        fontSize: 14,
        fullScreen: false,
        content: '',
        originContent: '',
        isDiff: false,
      },
      dataSource: {
        std_data_name: '',
        file_data_name: '',
        sys_data_name: '',
      },
      logMap: {
        'db-privilege': 'bcs-icon bcs-icon-db-auth',
        'bk-log-collector': 'bcs-icon bcs-icon-log',
        'bcs-gamestatefulset-operator': 'bcs-icon bcs-icon-gss',
        'bcs-gamedeployment-operator': 'bcs-icon bcs-icon-gd',
        'prometheus-adapter': 'bcs-icon bcs-icon-prom',
        'bcs-ingress-controller': 'bcs-icon bcs-icon-bi-2',
        'bcs-hook-operator': 'bcs-icon bcs-icon-bh',
        'bcs-polaris-operator': 'bcs-icon bcs-icon-pol',
      },
    };
  },
  computed: {
    isEn() {
      return this.$store.state.isEn;
    },
    projectId() {
      return this.$route.params.projectId;
    },
    curProject() {
      return this.$store.state.curProject;
    },
    crdKind() {
      return this.$route.meta.crdKind;
    },
    clusterList() {
      return this.$store.state.cluster.clusterList;
    },
    searchScopeList() {
      const { clusterList } = this;
      let results = [];
      if (clusterList.length) {
        results = [];
        clusterList.forEach((item) => {
          results.push({
            id: item.cluster_id,
            name: item.name,
          });
        });
      }

      return results;
    },
    editorHeight() {
      const height = window.innerHeight;
      return this.editorOptions.fullScreen ? height : height - 80;
    },
    curClusterId() {
      return this.$store.getters.curClusterId;
    },
  },
  // watch: {
  //   curClusterId() {
  //     this.searchScope = this.curClusterId;
  //     this.search();
  //   },
  // },
  mounted() {
    this.init();
  },
  beforeRouteLeave(to, from, next) {
    this.clearAllInterval();
    next();
  },
  methods: {
    async init() {
      try {
        if (this.clusterList.length) {
          // if (this.curClusterId) {
          //     this.searchScope = this.curClusterId
          // } else {
          //     this.searchScope = this.clusterList[0].cluster_id
          // }
          this.getCrdControllersByCluster();
        } else {
          this.isInitLoading = false;
          this.isPageLoading = false;
        }
      } catch (e) {
        catchErrorHandler(e, this);
        this.isInitLoading = false;
        this.isPageLoading = false;
      }
    },

    async haneldEnableCrdController(crdcontroller) {
      // 清空数据
      this.editorOptions.content = '';
      this.editorOptions.originContent = '';

      this.curCrdcontroller = crdcontroller;
      if (crdcontroller.default_values) {
        this.valueSlider.title = `${this.$t('plugin.tools.enable')}${crdcontroller.name}`;
        this.editorOptions.content = crdcontroller.default_values;
        this.editorOptions.originContent = crdcontroller.default_values;
        // eslint-disable-next-line no-plusplus
        this.renderEditorKey++;
        this.valueSlider.isShow = true;
      } else {
        this.enableCrdController();
      }
    },

    async enableCrdController() {
      try {
        const crdcontroller = this.curCrdcontroller;
        const clusterId = this.searchScope;
        this.valueSlider.isShow = false;
        await addonsInstall({
          $clusterId: clusterId,
          name: crdcontroller.name,
          version: crdcontroller.version,
          values: this.editorOptions.content,
        }).catch(() => false);
        this.getCrdControllersByCluster();
        this.refreshList();
        // this.getCrdcontrollerStatus(crdcontroller);
      } catch (e) {
        catchErrorHandler(e, this);
      } finally {
        this.editorOptions.readOnly = false;
      }
    },

    hideApplicationJson() {
      this.valueSlider.isShow = false;
      // 清空数据
      this.editorOptions.content = '';
      this.editorOptions.originContent = '';
    },

    async refreshList() {
      const clusterId = this.searchScope;
      const list = await addonsList({
        $clusterId: clusterId,
      }).catch(() => []);
      const data = list.map(item => ({
        ...item,
        // 兼容旧数据
        chart_name: item.chartName,
        default_values: item.defaultValues,
        description: item.description,
        help_link: item.docsLink,
        id: item.name,
        cluster_id: clusterId,
        message: item.message,
        status: item.status,
        logo: item.logo,
        name: item.name,
        supported_actions: item.supportedActions,
        version: item.version,
      }));
        // 搜索
      let results = data.filter((item) => {
        if (this.crdKind === 'BcsLog') {
          return item.chart_name === 'bk-log-collector';
        }
        return item.chart_name !== 'bk-log-collector';
      });
      if (this.searchKeyword.trim()) {
        results = [];
        const keyword = this.searchKeyword.trim();
        const keyList = ['name'];
        const list = data;

        list.forEach((item) => {
          item.isChecked = false;
          for (const key of keyList) {
            if (item[key].indexOf(keyword) > -1) {
              results.push(item);
              return true;
            }
          }
        });
      }
      // results[0].status = 'pending'
      this.crdControllerList = results;
      const isPending = this.crdControllerList.some(item => this.pendingStatus.includes(item.status));
      if (isPending) {
        setTimeout(() => {
          this.refreshList();
        }, 5000);
      }
    },
    async getCrdControllersByCluster() {
      if (this.isPageLoading) {
        return false;
      }
      if (!this.searchScope) {
        return false;
      }

      this.isPageLoading = true;
      try {
        await this.refreshList();
        this.clearAllInterval();
        // this.crdControllerList.forEach((item) => {
        //   if (this.pendingStatus.includes(item.status)) {
        //     this.getCrdcontrollerStatus(item);
        //   }
        // });
      } catch (e) {
        catchErrorHandler(e, this);
      } finally {
        setTimeout(() => {
          this.isInitLoading = false;
          this.isPageLoading = false;
        }, 200);
      }
    },

    clearAllInterval() {
      for (const key in this.statusTimer) {
        clearInterval(this.statusTimer[key]);
      }
      this.statusTimer = {};
    },

    async enableLogPlans() {
      const { projectId } = this;
      try {
        await this.$store.dispatch('enableLogPlans', projectId);
      } catch (e) {
        catchErrorHandler(e, this);
      }
    },

    showInstanceDetail(crdcontroller) {
      this.$router.push({
        name: 'crdcontrollerInstanceDetail',
        params: {
          clusterId: this.searchScope,
          id: crdcontroller.id,
          chartName: crdcontroller.chart_name,
        },
      });
    },

    async goControllerInstances(crdcontroller) {
      if (this.crdKind === 'BcsLog') {
        try {
          if (this.$INTERNAL) {
            const { projectId } = this;
            await this.$store.dispatch('enableLogPlans', projectId);
          }

          this.$router.push({
            name: 'crdcontrollerLogInstances',
            params: {
              clusterId: this.searchScope,
            },
          });
        } catch (e) {
          if (e.code !== 404) {
            catchErrorHandler(e, this);
          }
        }
      } else {
        if (crdcontroller.chart_name === 'db-privilege') {
          this.$router.push({
            name: 'crdcontrollerDBInstances',
            params: {
              clusterId: this.searchScope,
            },
          });
        } else if (crdcontroller.chart_name === 'bcs-polaris-operator') {
          this.$router.push({
            name: 'crdcontrollerPolarisInstances',
            params: {
              clusterId: this.searchScope,
            },
          });
        }
      }
    },

    search() {
      this.getCrdControllersByCluster();
    },

    refresh() {
      this.searchKeyword = '';
      this.getCrdControllersByCluster();
    },

    /**
           * 简单判断是否为图片
           * @param  {string} img 图片url
           * @return {Boolean} true/false
           */
    isImage(img) {
      if (!img) {
        return false;
      }
      if (img.startsWith('http://') || img.startsWith('https://') || img.startsWith('data:image/')) {
        return true;
      }
      return false;
    },

    /**
           * 获取crdcontroller状态
           * @param  {object} crdcontroller crdcontroller
           * @param  {number} index 索引
           */
    getCrdcontrollerStatus(crdcontroller) {
      if (crdcontroller.id === undefined) {
        return false;
      }
      const crdcontrollerId = crdcontroller.id;
      const clusterId = crdcontroller.cluster_id || this.searchScope;
      // eslint-disable-next-line @typescript-eslint/no-this-alias
      const self = this;

      clearInterval(this.statusTimer[crdcontroller.id]);
      // eslint-disable-next-line @typescript-eslint/no-misused-promises
      this.statusTimer[crdcontroller.id] = setInterval(async () => {
        try {
          const item = await addonsDetail({
            $clusterId: clusterId,
            $name: crdcontroller.name,
          }).catch(() => ({}));
          const data = {
            ...item,
            // 兼容旧数据
            chart_name: item.chartName,
            default_values: item.defaultValues,
            description: item.description,
            help_link: item.docsLink,
            id: item.name,
            cluster_id: clusterId,
            message: item.message,
            status: item.status,
            logo: item.logo,
            name: item.name,
            supported_actions: item.supportedActions,
            version: item.version,
          };
          if (!this.pendingStatus.includes(data.status) || !data.status) {
            clearInterval(self.statusTimer[crdcontroller.id]);
            this.crdControllerList.forEach((item) => {
              if (item.id === crdcontrollerId) {
                item.status = data.status;
                item.message = data.message;
              }
            });
          }
        } catch (e) {
          catchErrorHandler(e, this);
        }
      }, 2000);
    },
    // 卸载组件
    handleUninstall(crdcontroller) {
      const clusterId = crdcontroller.cluster_id || this.searchScope;
      this.$bkInfo({
        type: 'warning',
        clsName: 'custom-info-confirm',
        subTitle: crdcontroller.name,
        title: this.$t('plugin.tools.confirmUninstall'),
        defaultInfo: true,
        confirmFn: async () => {
          const result = await addonsUninstall({
            $clusterId: clusterId,
            $name: crdcontroller.name,
          }).then(() => true)
            .catch(() => false);
          if (result) {
            this.getCrdControllersByCluster();
          }
        },
      });
    },
  },
};
</script>

<style lang="postcss" scoped>
    @import './index.css';
    /deep/ .bk-dropdown-menu .bk-dropdown-list {
        overflow: hidden;
    }
</style>
