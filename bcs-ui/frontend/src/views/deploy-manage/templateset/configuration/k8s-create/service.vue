<!-- eslint-disable max-len -->
<template>
  <BcsContent>
    <template #header>
      <biz-header
        ref="commonHeader"
        @exception="exceptionHandler"
        @saveServiceSuccess="saveServiceSuccess"
        @switchVersion="initResource"
        @exmportToYaml="exportToYaml">
      </biz-header>
    </template>
    <div class="biz-confignation-wrapper" v-bkloading="{ isLoading: isTemplateSaving }">
      <div class="biz-tab-box" v-show="!isDataLoading">
        <biz-tabs @tab-change="tabResource" ref="commonTab"></biz-tabs>
        <div class="biz-tab-content" v-bkloading="{ isLoading: isTabChanging }">
          <bk-alert type="info" class="mb20">
            <div slot="title">
              {{$t('deploy.templateset.serviceDescription')}}，
              <a class="bk-text-button" :href="PROJECT_CONFIG.k8sService" target="_blank">{{$t('plugin.tools.docs')}}</a>
            </div>
          </bk-alert>
          <template v-if="!services.length">
            <div class="biz-guide-box mt0">
              <bk-button icon="plus" type="primary" @click.stop.prevent="addLocalService">
                <span style="margin-left: 0;">{{$t('generic.button.add')}}Service</span>
              </bk-button>
            </div>
          </template>

          <template v-else>
            <div class="biz-configuration-topbar">
              <div class="biz-list-operation">
                <div class="item" v-for="(service, index) in services" :key="index">
                  <bk-button :class="['bk-button', { 'bk-primary': curService.id === service.id }]" @click.stop="setCurService(service, index)">
                    {{(service && service.config.metadata.name) || $t('deploy.templateset.unnamed')}}
                    <span class="biz-update-dot" v-show="service.isEdited"></span>
                  </bk-button>
                  <span class="bcs-icon bcs-icon-close" @click.stop="removeService(service, index)"></span>
                </div>

                <bcs-popover ref="serviceTooltip" :content="$t('deploy.templateset.addService')" placement="top">
                  <bk-button class="bk-button bk-default is-outline is-icon" @click.stop="addLocalService">
                    <i class="bcs-icon bcs-icon-plus"></i>
                  </bk-button>
                </bcs-popover>
              </div>
            </div>

            <div class="biz-configuration-content" style="position: relative; margin-bottom: 105px;">
              <div class="bk-form biz-configuration-form">
                <a href="javascript:void(0);" class="bk-text-button from-json-btn" @click.stop.prevent="showJsonPanel">{{$t('deploy.templateset.importYAML')}}</a>

                <bk-sideslider
                  :is-show.sync="toJsonDialogConf.isShow"
                  :title="toJsonDialogConf.title"
                  :width="toJsonDialogConf.width"
                  class="biz-app-container-tojson-sideslider"
                  :quick-close="false"
                  @hidden="closeToJson">
                  <div slot="content" style="position: relative;">
                    <div class="biz-log-box" :style="{ height: `${winHeight - 60}px` }" v-bkloading="{ isLoading: toJsonDialogConf.loading }">
                      <bk-button class="bk-button bk-primary save-json-btn" @click.stop.prevent="saveApplicationJson">{{$t('generic.button.import')}}</bk-button>
                      <bk-button class="bk-button bk-default hide-json-btn" @click.stop.prevent="hideApplicationJson">{{$t('generic.button.cancel')}}</bk-button>
                      <ace
                        :value="editorConfig.value"
                        :width="editorConfig.width"
                        :height="editorConfig.height"
                        :lang="editorConfig.lang"
                        :read-only="editorConfig.readOnly"
                        :full-screen="editorConfig.fullScreen"
                        @init="editorInitAfter">
                      </ace>
                    </div>
                  </div>
                </bk-sideslider>

                <div class="bk-form-item is-required">
                  <label class="bk-label" style="width: 140px;">{{$t('generic.label.name')}}：</label>
                  <div class="bk-form-content" style="margin-left: 140px;">
                    <bkbcs-input
                      type="text"
                      :placeholder="$t('deploy.templateset.enterCharacterLimit')"
                      style="width: 310px;"
                      maxlength="64"
                      :value.sync="curService.config.metadata.name"
                      :list="varList"
                    >
                    </bkbcs-input>
                    <div class="bk-form-tip" v-if="errors.has('serviceName')">
                      <p class="bk-tip-text">{{$t('deploy.templateset.nameIsRequired')}}</p>
                    </div>
                  </div>
                </div>
                <div class="bk-form-item is-required">
                  <label class="bk-label" style="width: 140px;">{{$t('deploy.templateset.associateApplication')}}：</label>
                  <div class="bk-form-content" style="margin-left: 140px;">
                    <div class="bk-dropdown-box" style="width: 310px;" @click="reloadApplications">
                      <bk-selector
                        :placeholder="$t('deploy.templateset.selectAssociatedApplication')"
                        :setting-key="'deploy_tag'"
                        :multi-select="true"
                        :display-key="'deploy_name'"
                        :selected.sync="curService.deploy_tag_list"
                        :list="applicationList"
                        :prevent-init-trigger="'true'"
                        :is-loading="isLoadingApps"
                        @item-selected="selectApps">
                      </bk-selector>
                    </div>
                    <span class="biz-tip ml10" v-if="!isDataLoading && !applicationList.length">{{$t('deploy.templateset.configureResourceFirst')}}</span>
                  </div>
                </div>
                <div class="bk-form-item is-required">
                  <label class="bk-label" style="width: 140px;">{{$t('deploy.templateset.associateLabel')}}：</label>
                  <div class="bk-form-content key-tip-wrapper" style="margin-left: 140px;">
                    <template v-if="appLabels.length && !isLabelsLoading">
                      <ul class="key-list">
                        <li v-for="(label,index) in appLabels" @click="selectLabel(label)" :key="index">
                          <span class="key">
                            <bk-checkbox name="linkapp" :value="label.isSelected"></bk-checkbox>
                          </span>
                          <span class="value">{{label.key}}:{{label.value}}</span>
                        </li>
                      </ul>
                      <p class="biz-tip mt5 mb15">{{$t('deploy.templateset.trafficWarning')}}</p>
                    </template>
                    <div v-else-if="!isLabelsLoading" class="biz-tip biz-danger" style="margin-top: 7px;">
                      {{existLinkApp.length ? $t('deploy.templateset.noCommonLabel') : $t('deploy.templateset.associateApplicationFirst')}}
                    </div>
                  </div>
                </div>
                <div class="bk-form-item">
                  <label class="bk-label" style="width: 140px;">{{$t('deploy.templateset.serviceType')}}：</label>
                  <div class="bk-form-content" style="margin-left: 140px;">
                    <div class="bk-dropdown-box" style="width: 310px;">
                      <bk-selector
                        :placeholder="$t('generic.placeholder.select')"
                        :setting-key="'id'"
                        :display-key="'name'"
                        :selected.sync="curService.config.spec.type"
                        :list="serviceTypeList"
                        @item-selected="selectServiceType">
                      </bk-selector>
                    </div>
                  </div>
                </div>
                <div class="bk-form-item" v-show="curService.config.spec.type !== 'NodePort'">
                  <label class="bk-label" style="width: 140px;">ClusterIP：</label>
                  <div class="bk-form-content" style="margin-left: 140px;">
                    <bkbcs-input :placeholder="$t('deploy.templateset.enterClusterIP')" style="width: 310px;" v-model="curService.config.spec.clusterIP" />
                    <!-- <p class="biz-tip mt5">{{$t('deploy.templateset.noFillOrNone')}}</p> -->
                  </div>
                </div>
                <div class="bk-form-item">
                  <label class="bk-label" style="width: 140px;">{{$t('dashboard.network.portmapping')}}：</label>
                  <div class="bk-form-content" style="margin-left: 140px;">
                    <div class="biz-keys-list mb10">
                      <template v-if="curService.deploy_tag_list.length">
                        <template v-if="appPortList.length && curService.config.spec.ports.length">
                          <table class="biz-simple-table">
                            <thead>
                              <tr>
                                <th style="width: 100px;">{{$t('deploy.templateset.portName')}}</th>
                                <th style="width: 100px;">{{$t('deploy.helm.port')}}</th>
                                <th style="width: 120px;">{{$t('deploy.templateset.protocol')}}</th>
                                <th style="width: 120px;">{{$t('deploy.templateset.targetPort')}}</th>
                                <th style="width: 100px;" v-if="curService.config.spec.type === 'NodePort' || curService.config.spec.type === 'LoadBalancer'">NodePort</th>
                                <th></th>
                              </tr>
                            </thead>
                            <tbody>
                              <tr v-for="(port, index) in curService.config.spec.ports" :key="index">
                                <td>
                                  <bkbcs-input
                                    type="text"
                                    :placeholder="$t('generic.placeholder.input')"
                                    style="width: 100px;"
                                    :value.sync="port.name"
                                    :list="varList"
                                  >
                                  </bkbcs-input>
                                </td>
                                <td>
                                  <bkbcs-input
                                    type="number"
                                    :placeholder="$t('generic.placeholder.input')"
                                    style="width: 100px;"
                                    :min="1"
                                    :max="65535"
                                    :value.sync="port.port"
                                    :list="varList"
                                  >
                                  </bkbcs-input>
                                </td>
                                <td>
                                  <bk-selector
                                    :placeholder="$t('deploy.templateset.protocol')"
                                    :setting-key="'id'"
                                    :allow-clear="true"
                                    :selected.sync="port.protocol"
                                    :list="protocolList">
                                  </bk-selector>
                                </td>
                                <td>
                                  <bk-selector
                                    :placeholder="$t('generic.placeholder.select')"
                                    :setting-key="'id'"
                                    :display-key="'name'"
                                    :selected.sync="port.id"
                                    :allow-clear="true"
                                    :filter-list="curServicePortList"
                                    :is-link="true"
                                    :init-prevent-trigger="true"
                                    :list="appPortList"
                                    @clear="clearPort(port)"
                                    @item-selected="selectPort(port)">
                                  </bk-selector>
                                </td>
                                <td v-if="curService.config.spec.type === 'NodePort' || curService.config.spec.type === 'LoadBalancer'">
                                  <bkbcs-input
                                    type="number"
                                    :placeholder="$t('generic.placeholder.input')"
                                    style="width: 76px;"
                                    :min="0"
                                    :max="32767"
                                    :disabled="curService.config.spec.type !== 'NodePort' && curService.config.spec.type !== 'LoadBalancer'"
                                    :value.sync="port.nodePort"
                                    :list="varList"
                                  >
                                  </bkbcs-input>
                                  <bcs-popover placement="top">
                                    <i class="bcs-icon bcs-icon-question-circle" style="vertical-align: middle; cursor: pointer;"></i>
                                    <div slot="content">
                                      {{$t('deploy.templateset.enterNodePortValue')}}
                                    </div>
                                  </bcs-popover>
                                </td>
                                <td>
                                  <bk-button class="action-btn ml5" @click.stop.prevent="addPort" v-show="curService.config.spec.ports.length < appPortList.length">
                                    <i class="bcs-icon bcs-icon-plus"></i>
                                  </bk-button>
                                  <bk-button class="action-btn" @click.stop.prevent="removePort(port, index)" v-show="curService.config.spec.ports.length > 1">
                                    <i class="bcs-icon bcs-icon-minus"></i>
                                  </bk-button>
                                </td>
                              </tr>
                            </tbody>
                          </table>
                        </template>
                        <template v-else>
                          <p class="mt5 biz-tip biz-danger">{{$t('deploy.templateset.fillAssociatedContainerPort')}}</p>
                        </template>
                      </template>
                      <template v-else>
                        <p class="mt5 biz-tip biz-danger">{{$t('deploy.templateset.associateApplicationFirst')}}</p>
                      </template>
                      <p class="biz-tip">
                        {{$t('deploy.templateset.clusterIPNoneSkipPortMapping')}}
                        <a href="javascript:void(0);" class="bk-text-button" @click="showPortExampleDialg">{{$t('deploy.templateset.viewExample')}}</a>
                      </p>
                    </div>
                  </div>
                </div>
                <div class="bk-form-item">
                  <label class="bk-label" style="width: 140px;">{{$t('generic.label.labelManage')}}：</label>
                  <div class="bk-form-content" style="margin-left: 140px;">
                    <bk-keyer :key-list.sync="curLabelList" ref="labelKeyer" @change="updateLabelList" :var-list="varList"></bk-keyer>
                  </div>
                </div>
                <div class="bk-form-item">
                  <label class="bk-label" style="width: 140px;">{{$t('deploy.templateset.annotationManagement')}}：</label>
                  <div class="bk-form-content" style="margin-left: 140px;">
                    <bk-keyer :key-list.sync="curRemarkList" :var-list="varList" ref="remarkKeyer" @change="updateApplicationRemark"></bk-keyer>
                  </div>
                </div>
              </div>
            </div>
          </template>
        </div>
      </div>
    </div>
    <bk-dialog
      :is-show.sync="exampleDialogConf.isShow"
      :width="exampleDialogConf.width"
      :title="exampleDialogConf.title"
      :close-icon="exampleDialogConf.closeIcon"
      :has-footer="false"
      :ext-cls="'biz-example-dialog'"
      @cancel="exampleDialogConf.isShow = false">
      <template slot="content">
        <img src="@/images/service-example.png" style="width: 100%;">
      </template>
    </bk-dialog>
  </BcsContent>
</template>

<script>
/* eslint-disable @typescript-eslint/prefer-optional-chain */
/* eslint-disable @typescript-eslint/no-unused-vars */
/* eslint-disable no-prototype-builtins */
/* eslint-disable no-multi-assign */
/* eslint-disable no-case-declarations */
/* eslint-disable @typescript-eslint/no-this-alias */
/* eslint-disable @typescript-eslint/no-require-imports */
import yamljs from 'js-yaml';
import _ from 'lodash';

import header from './header.vue';
import tabs from './tabs.vue';

import ace from '@/components/ace-editor';
import bkKeyer from '@/components/keyer';
import BcsContent from '@/components/layout/Content.vue';
import serviceParams from '@/json/k8s-service.json';
import k8sBase from '@/mixins/configuration/k8s-base';
import mixinBase from '@/mixins/configuration/mixin-base';

export default {
  name: 'K8SService',
  components: {
    'bk-keyer': bkKeyer,
    'biz-header': header,
    'biz-tabs': tabs,
    ace,
    BcsContent,
  },
  mixins: [mixinBase, k8sBase],
  data() {
    return {
      isTabChanging: false,
      exceptionCode: null,
      isDataLoading: true,
      isDataSaveing: false,
      isWeightError: false,
      isLoadingApps: false,
      portTimer: 0,
      existLinkApp: [],
      algorithmList: [
        {
          id: 'roundrobin',
          name: 'roundrobin',
        },
        {
          id: 'source',
          name: 'source',
        },
        {
          id: 'leastconn',
          name: 'leastconn',
        },
      ],
      exampleDialogConf: {
        isShow: false,
        title: this.$t('deploy.templateset.portMappingExample'),
        width: 800,
        closeIcon: true,
      },
      isLabelsLoading: true,
      serviceTypeList: [
        {
          id: 'ClusterIP',
          name: 'ClusterIP',
        },
        {
          id: 'NodePort',
          name: 'NodePort',
        },
        {
          id: 'LoadBalancer',
          name: 'LoadBalancer',
        },
      ],
      weight: 10,
      curServiceIPs: '',
      linkAppVersion: 0,
      protocolIndex: -1,
      protocolList: [
        {
          id: 'TCP',
          name: 'TCP',
        },
        {
          id: 'UDP',
          name: 'UDP',
        },
      ],
      appPortList: [],
      appLabels: [],
      curServiceCache: Object.assign({}, serviceParams),
      curService: serviceParams,
      winHeight: 0,
      toJsonDialogConf: {
        isShow: false,
        title: '',
        timer: null,
        width: 800,
        loading: false,
      },
      editorConfig: {
        width: '100%',
        height: '100%',
        lang: 'yaml',
        readOnly: false,
        fullScreen: false,
        value: '',
        editor: null,
      },
      yamlEditorConfig: {
        width: '100%',
        height: '100%',
        lang: 'yaml',
        readOnly: false,
        fullScreen: false,
        value: '',
        editor: null,
      },
    };
  },
  computed: {
    varList() {
      return this.$store.state.variable.varList;
    },
    curTemplate() {
      return this.$store.state.k8sTemplate.curTemplate;
    },
    applicationList() {
      return this.$store.state.k8sTemplate.linkApplications;
    },
    isTemplateSaving() {
      return this.$store.state.k8sTemplate.isTemplateSaving;
    },
    curVersion() {
      return this.$store.state.k8sTemplate.curVersion;
    },
    deployments() {
      return this.$store.state.k8sTemplate.deployments;
    },
    services() {
      return this.$store.state.k8sTemplate.services;
    },
    configmaps() {
      return this.$store.state.k8sTemplate.configmaps;
    },
    secrets() {
      return this.$store.state.k8sTemplate.secrets;
    },
    daemonsets() {
      return this.$store.state.k8sTemplate.daemonsets;
    },
    jobs() {
      return this.$store.state.k8sTemplate.jobs;
    },
    statefulsets() {
      return this.$store.state.k8sTemplate.statefulsets;
    },
    projectId() {
      return this.$route.params.projectId;
    },
    templateId() {
      return this.$route.params.templateId;
    },
    curServicePortList() {
      const results = [];
      const { ports } = this.curService.config.spec;
      ports.forEach((item) => {
        results.push(item.targetPort);
      });
      return results;
    },
    curRemarkList() {
      const list = [];
      const { annotations } = this.curService.config.metadata;
      // 如果有缓存直接使用
      if (this.curService.config.webCache && this.curService.config.webCache.remarkListCache) {
        return this.curService.config.webCache.remarkListCache;
      }
      if (annotations) {
        for (const [key, value] of Object.entries(annotations)) {
          list.push({
            key,
            value,
          });
        }
      }

      if (!list.length) {
        list.push({
          key: '',
          value: '',
        });
      }
      return list;
    },
    curLabelList() {
      const list = [];
      const { labels } = this.curService.config.metadata;
      // 如果有缓存直接使用
      if (this.curService.config.webCache && this.curService.config.webCache.labelListCache) {
        return this.curService.config.webCache.labelListCache;
      }
      for (const [key, value] of Object.entries(labels)) {
        list.push({
          key,
          value,
        });
      }
      if (!list.length) {
        list.push({
          key: '',
          value: '',
        });
      }
      return list;
    },
  },
  watch: {
    'deployments'() {
      if (this.curVersion) {
        this.initApplications(this.curVersion);
      }
    },
    'daemonsets'() {
      if (this.curVersion) {
        this.initApplications(this.curVersion);
      }
    },
    'jobs'() {
      if (this.curVersion) {
        this.initApplications(this.curVersion);
      }
    },
    'statefulsets'() {
      if (this.curVersion) {
        this.initApplications(this.curVersion);
      }
    },
  },
  mounted() {
    this.$refs.commonHeader.initTemplate((data) => {
      this.initResource(data);
      this.isDataLoading = false;
    });
    this.winHeight = window.innerHeight;
  },
  methods: {
    showPortExampleDialg() {
      this.exampleDialogConf.isShow = true;
    },
    selectLabel(labels) {
      labels.isSelected = !labels.isSelected;
      this.curService.config.webCache.link_labels = [];
      this.curService.config.spec.selector = {};
      this.appLabels.forEach((label) => {
        if (label.isSelected) {
          this.curService.config.webCache.link_labels.push(label.id);
          this.curService.config.spec.selector[label.key] = label.value;
        }
      });
    },
    selectServiceType(index, item) {
      if (index !== 'NodePort' && index !== 'LoadBalancer') {
        this.curService.config.spec.ports.forEach((port) => {
          port.nodePort = '';
        });
      }
      if (index === 'NodePort') {
        delete this.curService.config.spec.clusterIP;
      } else {
        this.$set(this.curService.config.spec, 'clusterIP', '');
      }
    },
    async initResource(data) {
      const self = this;
      const version = data.latest_version_id || data.version;
      if (version) {
        this.initApplications(version, () => {
          if (data.services && data.services.length) {
            self.$nextTick(() => {
              self.setCurService(data.services[0], 0);
            });
          }
        });
      } else {
        if (data.services && data.services.length) {
          this.$nextTick(() => {
            this.setCurService(data.services[0], 0);
          });
        }
      }
    },
    exportToYaml(data) {
      this.$router.push({
        name: 'K8sYamlTemplateset',
        params: {
          projectId: this.projectId,
          projectCode: this.projectCode,
          templateId: 0,
        },
        query: {
          action: 'export',
        },
      });
    },
    async tabResource(type, target) {
      this.isTabChanging = true;
      await this.$refs.commonHeader.saveTemplate();
      await this.$refs.commonHeader.autoSaveResource(type);
      this.$refs.commonTab.goResource(target);
    },
    showJsonPanel() {
      this.toJsonDialogConf.title = `${this.curService.config.metadata.name}.yaml`;
      const appConfig = JSON.parse(JSON.stringify(this.curService.config));
      const { webCache } = appConfig;
      // 在处理yaml导入时，保存一份原数据，方便对导入的数据进行合并处理
      this.curServiceCache = JSON.parse(JSON.stringify(this.curService.config));

      // 标签
      if (webCache && webCache.labelListCache) {
        const labelKeyList = this.tranListToObject(webCache.labelListCache);
        appConfig.metadata.labels = labelKeyList;
      }

      // 注解
      if (webCache && webCache.remarkListCache) {
        const remarkKeyList = this.tranListToObject(webCache.remarkListCache);
        appConfig.metadata.annotations = remarkKeyList;
      }

      delete appConfig.webCache;
      const yamlStr = yamljs.dump(appConfig);
      this.editorConfig.value = yamlStr;
      this.toJsonDialogConf.isShow = true;
    },
    hideApplicationJson() {
      this.toJsonDialogConf.isShow = false;
    },
    closeToJson() {
      this.toJsonDialogConf.isShow = false;
      this.toJsonDialogConf.title = '';
      this.editorConfig.value = '';
    },
    editorInitAfter(editor) {
      this.editorConfig.editor = editor;
      this.editorConfig.editor.setStyle('biz-app-container-tojson-ace');
    },
    getAppParamsKeys(obj, result) {
      for (const key in obj) {
        if (key === 'data') continue;
        if (Object.prototype.toString.call(obj) === '[object Array]') {
          this.getAppParamsKeys(obj[key], result);
        } else if (Object.prototype.toString.call(obj) === '[object Object]') {
          if (!result.includes(key)) {
            result.push(key);
          }
          this.getAppParamsKeys(obj[key], result);
        }
      }
    },
    checkJson(jsonObj) {
      const { editor } = this.editorConfig;
      const appParams = serviceParams.config;
      const appParamKeys = [
        'id',
        'creationTimestamp',
      ];
      const jsonParamKeys = [];

      this.getAppParamsKeys(appParams, appParamKeys);
      this.getAppParamsKeys(jsonObj, jsonParamKeys);

      // application查看无效字段
      for (const key of jsonParamKeys) {
        if (!appParamKeys.includes(key)) {
          this.$bkMessage({
            theme: 'error',
            message: `${key}${this.$t('deploy.templateset.invalidField')}`,
          });
          const match = editor.find(`${key}`);
          if (match) {
            editor.moveCursorTo(match.end.row, match.end.column);
          }
          return false;
        }
      }
      return true;
    },
    formatJson(jsonObj) {
      // 标签
      const keyList = [];
      const { labels } = jsonObj.metadata;

      for (const [key, value] of Object.entries(labels)) {
        const params = {
          key,
          value,
        };
        keyList.push(params);
      }
      if (!keyList.length) {
        keyList.push({
          key: '',
          value: '',
        });
      }
      jsonObj.webCache.labelListCache = keyList;

      // 注解
      const remarkKeyList = [];
      const { annotations } = jsonObj.metadata;

      for (const [key, value] of Object.entries(annotations)) {
        const params = {
          key,
          value,
        };
        remarkKeyList.push(params);
      }
      if (!remarkKeyList.length) {
        remarkKeyList.push({
          key: '',
          value: '',
        });
      }
      jsonObj.webCache.remarkListCache = remarkKeyList;

      // 关联标签
      const { selector } = jsonObj.spec;
      if (selector) {
        for (const [key, value] of Object.entries(selector)) {
          const params = `${key}:${value}`;
          jsonObj.webCache.link_labels.push(params);
        }
      }

      return jsonObj;
    },
    saveApplicationJson() {
      const { editor } = this.editorConfig;
      const yaml = editor.getValue();
      let appObj = null;
      if (!yaml) {
        this.$bkMessage({
          theme: 'error',
          message: this.$t('deploy.templateset.enterYAML'),
        });
        return false;
      }

      try {
        appObj = yamljs.load(yaml);
      } catch (err) {
        this.$bkMessage({
          theme: 'error',
          message: this.$t('deploy.templateset.enterValidYAML'),
        });
        return false;
      }

      const annot = editor.getSession().getAnnotations();
      if (annot && annot.length) {
        editor.gotoLine(annot[0].row, annot[0].column, true);
        return false;
      }
      const newConfObj = _.merge({}, serviceParams.config, appObj);
      const jsonFromat = this.formatJson(newConfObj);
      this.curService.config = jsonFromat;
      this.toJsonDialogConf.isShow = false;

      this.appLabels.forEach((label) => {
        // eslint-disable-next-line max-len
        if (this.curService.config.webCache.link_labels && this.curService.config.webCache.link_labels.indexOf(label.id) > -1) {
          label.isSelected = true;
        } else {
          label.isSelected = false;
        }
      });
    },
    exceptionHandler(exceptionCode) {
      this.isDataLoading = false;
      this.exceptionCode = exceptionCode;
    },
    reloadApplications() {
      if (this.curVersion) {
        this.isLoadingApps = true;
        this.initApplications(this.curVersion);
      }
    },
    selectPort(port) {
      const { id } = port;
      this.appPortList.forEach((item) => {
        if (item.id === id) {
          port.targetPort = item.name;
        }
      });
    },
    clearPort(port) {
      port.targetPort = '';
    },
    toggleRouter(target) {
      this.$router.push({
        name: target,
        params: {
          projectId: this.projectId,
          templateId: this.templateId,
        },
      });
    },
    addPort() {
      const { ports } = this.curService.config.spec;
      const port = {
        id: '',
        name: '',
        port: '',
        protocol: 'TCP',
        targetPort: '',
        nodePort: '',
      };
      ports.push(port);
    },
    removePort(port, index) {
      const { ports } = this.curService.config.spec;
      ports.splice(index, 1);
    },
    initApplications(version, callback) {
      const { projectId } = this;
      this.linkAppVersion = version;
      this.$store.dispatch('k8sTemplate/getAppsByVersion', { projectId, version }).then((res) => {
        this.isLoadingApps = false;

        setTimeout(() => {
          callback && callback();
        }, 10);
      }, (res) => {
        const { message } = res;
        this.$bkMessage({
          theme: 'error',
          message,
          hasCloseIcon: true,
          delay: '10000',
        });
      });
    },
    updateLocalData(data) {
      if (data.id) {
        this.curService.id = data.id;
      }
      if (data.version) {
        this.$store.commit('k8sTemplate/updateCurVersion', data.version);
      }
      this.$store.commit('k8sTemplate/updateServices', this.services);
      setTimeout(() => {
        this.services.forEach((item) => {
          if (item.id === data.id) {
            this.setCurService(item);
          }
        });
      }, 500);
    },
    setCurService(service, index) {
      this.isLabelsLoading = true;
      this.curService = service;
      this.curServiceIPs = this.curService.config.spec.clusterIP;
      if (!this.curService.config.spec.ports.length) {
        this.addPort();
      }
      clearInterval(this.compareTimer);
      clearTimeout(this.setTimer);
      this.setTimer = setTimeout(() => {
        this.appPortList = [];
        this.initLinkResource();

        if (!this.curService.cache) {
          this.curService.cache = JSON.parse(JSON.stringify(service));
        }
        this.watchChange();
      }, 500);
    },
    watchChange() {
      this.compareTimer = setInterval(() => {
        const appCopy = JSON.parse(JSON.stringify(this.curService));
        const cacheCopy = JSON.parse(JSON.stringify(this.curService.cache));

        // 删除无用属性
        delete appCopy.isEdited;
        delete appCopy.cache;
        delete appCopy.id;

        delete cacheCopy.isEdited;
        delete cacheCopy.cache;
        delete cacheCopy.id;

        const appStr = JSON.stringify(appCopy);
        const cacheStr = JSON.stringify(cacheCopy);
        if (String(this.curService.id).indexOf('local_') > -1) {
          this.curService.isEdited = true;
        } else if (appStr !== cacheStr) {
          this.curService.isEdited = true;
        } else {
          this.curService.isEdited = false;
        }
      }, 1000);
    },
    getProtocalById(id) {
      let result = null;
      this.appPortList.forEach((item) => {
        if (item.id === id) {
          result = item;
        }
      });
      if (result) {
        return result.protocol;
      }
      return '';
    },
    getTargetPortById(id) {
      let result = null;
      this.appPortList.forEach((item) => {
        if (item.id === id) {
          result = item;
        }
      });
      if (result) {
        return result.target_port;
      }
      return '';
    },
    selectApps(appIds, data) {
      this.curService.config.webCache.link_labels = [];
      this.curService.config.spec.selector = {};

      this.existLinkApp = appIds;
      this.getPorts(appIds, this.curService.config.metadata.name);
      this.getLabels(appIds, this.curService.config.metadata.name);
      // 如果关联应用, 且clusterIp为None
      if (appIds && appIds.length) {
        if (this.curService.config.spec.clusterIP === 'None') {
          this.curService.config.spec.clusterIP = '';
        }
      } else {
        if (!this.curService.config.spec.clusterIP) {
          this.curService.config.spec.clusterIP = 'None';
        }
      }
    },
    initLinkResource() {
      const appIds = [];
      const appKeys = [];

      // 过滤已经删除的app
      this.applicationList.forEach((item) => {
        item.children.forEach((child) => {
          appKeys.push(child.deploy_tag);
        });
      });
      this.curService.deploy_tag_list.forEach((item) => {
        if (appKeys.includes(item)) {
          appIds.push(item);
        } else {
          const type = item.split('|')[1];
          this.$bkMessage({
            theme: 'error',
            message: this.$t('deploy.templateset.reselectAppWarning', {
              name: this.curService.config.metadata.name,
              type,
            }),
          });
        }
      });

      this.existLinkApp = appIds;
      this.getPorts(appIds, this.curService.config.metadata.name);
      this.getLabels(appIds, this.curService.config.metadata.name);
    },
    getLabels(apps, serviceName) {
      this.isLabelsLoading = true;
      const { projectId } = this;
      const version = this.curVersion;

      this.$store.dispatch('k8sTemplate/getLabelsByDeployments', { projectId, version, apps }).then((res) => {
        if (!res.data) {
          return false;
        }
        // 防止不断点tab发起请求导致数据冲突
        if (serviceName !== this.curService.config.metadata.name) {
          return false;
        }
        const labels = [];
        // eslint-disable-next-line no-restricted-syntax
        for (const key in res.data) {
          const params = {
            id: `${key}:${res.data[key]}`,
            key,
            value: res.data[key],
            isSelected: false,
          };
          // eslint-disable-next-line max-len
          if (this.curService.config.webCache.link_labels && this.curService.config.webCache.link_labels.indexOf(params.id) > -1) {
            params.isSelected = true;
          }
          labels.push(params);
        }
        this.appLabels.splice(0, this.appLabels.length, ...labels);
      }, (res) => {
        this.appLabels.splice(0, this.appLabels.length);
      })
        .finally(() => {
          this.isLabelsLoading = false;
        });
    },
    getPorts(apps, serviceName) {
      const { projectId } = this;
      const version = this.curVersion;
      this.$store.dispatch('k8sTemplate/getPortsByDeployments', { projectId, version, apps }).then((res) => {
        if (!res.data) {
          return false;
        }
        // 防止不断点tab发起请求导致数据冲突
        if (serviceName !== this.curService.config.metadata.name) {
          return false;
        }
        const ports = res.data.filter(item => item.name);
        const keys = [];
        let results = [];
        ports.forEach((port) => {
          keys.push(port.id);
        });
        this.appPortList.splice(0, this.appPortList.length, ...ports);
        results = this.curService.config.spec.ports.filter((item) => {
          if (!item.id) {
            return true;
          } if (keys.includes(item.id)) {
            return true;
          }
          return false;
        });

        this.curService.config.spec.ports.splice(0, this.curService.config.spec.ports.length, ...results);
        if (!this.curService.config.spec.ports.length) {
          this.addPort();
        }
      }, (res) => {
        this.curService.config.spec.ports.splice(0, this.appPortList.length);
      });
    },
    removeLocalService(service, index) {
      // 是否删除当前项
      if (this.curService.id === service.id) {
        if (index === 0 && this.services[index + 1]) {
          this.setCurService(this.services[index + 1]);
        } else if (this.services[0]) {
          this.setCurService(this.services[0]);
        }
      }
      this.services.splice(index, 1);
    },
    removeService(service, index) {
      const self = this;
      const serviceId = service.id;

      this.$bkInfo({
        title: this.$t('generic.title.confirmDelete'),
        content: this.$createElement('p', { style: { 'text-align': 'left' } }, `${this.$t('deploy.templateset.deleteService')}：${service.config.metadata.name || this.$t('deploy.templateset.unnamed')}`),
        confirmFn() {
          if (serviceId.indexOf && serviceId.indexOf('local_') > -1) {
            self.removeLocalService(service, index);
          } else {
            self.deleteService(service, index);
          }
        },
      });
    },
    async deleteService(service, index) {
      const { projectId } = this;
      const version = this.curVersion;
      const serviceId = service.id;

      try {
        const res = await this.$store.dispatch('k8sTemplate/removeService', { serviceId, version, projectId });
        const { data } = res;
        this.removeLocalService(service, index);

        if (data.version) {
          this.$store.commit('k8sTemplate/updateCurVersion', data.version);
          this.$store.commit('k8sTemplate/updateBindVersion', true);
        }
        this.unBindStatefulset(service, data.version);
      } catch (res) {
        const { message } = res;
        this.$bkMessage({
          theme: 'error',
          message,
        });
      }
    },
    async unBindStatefulset(service, version) {
      const statefulsetItem = service.deploy_tag_list.find(item => item.indexOf('K8sStatefulSet') > -1);

      if (statefulsetItem) {
        const statefulsetId = statefulsetItem.split('|')[0];
        try {
          // 绑定
          this.statefulsets.forEach((statefulset) => {
            // 把其它已经绑定的statefulset进行解绑
            if (statefulset.deploy_tag === statefulsetId) {
              statefulset.service_tag = '';
              this.$store.dispatch('k8sTemplate/bindServiceForStatefulset', {
                projectId: this.projectId,
                versionId: version,
                statefulsetId: statefulset.deploy_tag,
                data: {
                  service_tag: '',
                },
              });
            }
          });
        } catch (res) {
          this.$bkMessage({
            theme: 'error',
            message: res.message,
            hasCloseIcon: true,
          });
        }
      }
    },
    saveServiceSuccess(params) {
      this.services.forEach((item) => {
        if (params.responseData.id === item.id || params.preId === item.id) {
          item.cache = JSON.parse(JSON.stringify(item));
        }
      });
      if (params.responseData.id === this.curService.id || params.preId === this.curService.id) {
        this.updateLocalData(params.resource);
      }
    },
    addLocalService() {
      const service = JSON.parse(JSON.stringify(serviceParams));
      const index = this.services.length;
      const now = +new Date();

      service.id = `local_${now}`;
      service.isEdited = true;
      service.config.metadata.name = `service-${index + 1}`;
      this.services.push(service);

      this.setCurService(service, index);
      this.$refs.serviceTooltip && (this.$refs.serviceTooltip.visible = false);
      this.$store.commit('k8sTemplate/updateServices', this.services);
    },
    updateLabelList(list, data) {
      if (!this.curService.config.webCache) {
        this.curService.config.webCache = {};
      }
      this.curService.config.webCache.labelListCache = list;
    },
    updateApplicationRemark(list, data) {
      if (!this.curService.config.webCache) {
        this.curService.config.webCache = {};
      }
      this.curService.config.webCache.remarkListCache = list;
    },
  },
};
</script>

<style scoped>
    @import './service.css';
</style>
