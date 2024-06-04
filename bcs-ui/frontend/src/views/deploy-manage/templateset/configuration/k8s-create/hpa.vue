<!-- eslint-disable max-len -->
<template>
  <BcsContent>
    <template #header>
      <biz-header
        ref="commonHeader"
        @exception="exceptionHandler"
        @saveHPASuccess="saveHPASuccess"
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
              <div>
                {{$t('deploy.templateset.HPA')}}，
                <a v-if="$INTERNAL" class="bk-text-button" :href="PROJECT_CONFIG.k8sHpa" target="_blank">{{$t('plugin.tools.docs')}}</a>
              </div>
            </div>
          </bk-alert>
          <template v-if="!HPAs.length">
            <div class="biz-guide-box mt0">
              <bk-button icon="plus" type="primary" @click.stop.prevent="addLocalHPA">
                <span style="margin-left: 0;">{{$t('generic.button.add')}}HPA</span>
              </bk-button>
            </div>
          </template>
          <template v-else>
            <div class="biz-configuration-topbar">
              <div class="biz-list-operation">
                <div class="item" v-for="(hpa, index) in HPAs" :key="hpa.id">
                  <bk-button :class="['bk-button', { 'bk-primary': curHPA.id === hpa.id }]" @click.stop="setCurHPA(hpa, index)">
                    {{(hpa && hpa.config.metadata.name) || $t('deploy.templateset.unnamed')}}
                    <span class="biz-update-dot" v-show="hpa.isEdited"></span>
                  </bk-button>
                  <span class="bcs-icon bcs-icon-close" @click.stop="removeHPA(hpa, index)"></span>
                </div>

                <bcs-popover ref="hpaTooltip" :content="$t('deploy.templateset.addHPA')" placement="top">
                  <bk-button class="bk-button bk-default is-outline is-icon" @click.stop="addLocalHPA">
                    <i class="bcs-icon bcs-icon-plus"></i>
                  </bk-button>
                </bcs-popover>
              </div>
            </div>

            <div class="biz-configuration-content" style="position: relative; margin-bottom: 130px;">
              <div class="bk-form biz-configuration-form">
                <a href="javascript:void(0);" class="bk-text-button from-json-btn" @click.stop.prevent="showJsonPanel">{{$t('deploy.templateset.importYAML')}}</a>

                <bk-sideslider
                  :is-show.sync="toJsonDialogConf.isShow"
                  :title="toJsonDialogConf.title"
                  :width="toJsonDialogConf.width"
                  :quick-close="false"
                  class="biz-app-container-tojson-sideslider"
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
                    <input type="text" :class="['bk-form-input',{ 'is-danger': errors.has('hpaName') }]" :placeholder="$t('deploy.templateset.enterCharacterLimit')" style="width: 310px;" maxlength="64" v-model="curHPA.config.metadata.name" name="hpaName" v-validate="{ required: true, regex: /^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$/ }">
                    <div class="bk-form-tip" v-if="errors.has('hpaName')">
                      <p class="bk-tip-text">{{$t('deploy.templateset.nameIsRequired')}}</p>
                    </div>
                  </div>
                </div>

                <div class="bk-form-item is-required">
                  <label class="bk-label" style="width: 140px;">{{$t('deploy.templateset.associateApplication')}}：</label>
                  <div class="bk-form-content" style="margin-left: 140px;">
                    <div class="bk-dropdown-box" style="width: 310px;">
                      <bk-selector
                        :placeholder="$t('deploy.templateset.selectAssociatedApplication')"
                        :setting-key="'deploy_name'"
                        :display-key="'deploy_name'"
                        :selected.sync="curHPA.config.spec.scaleTargetRef.name"
                        :list="applicationList"
                        :prevent-init-trigger="'true'"
                        :is-loading="isLoadingApps">
                      </bk-selector>
                    </div>
                    <span class="biz-tip ml10" v-if="!isDataLoading && !applicationList.length">{{$t('deploy.templateset.configureDeploymentFirst')}}</span>
                  </div>
                </div>

                <div class="bk-form-item is-required">
                  <label class="bk-label" style="width: 140px;">{{$t('deploy.templateset.instanceRange')}}：</label>
                  <div class="bk-form-content" style="margin-left: 140px;">
                    <div class="bk-form-input-group is-addon-left mr10">
                      <span class="input-group-addon prefix" style="display: inline-block;">
                        {{$t('deploy.templateset.min')}}
                      </span>
                      <bkbcs-input
                        type="number"
                        :placeholder="$t('generic.placeholder.input')"
                        style="width: 105px;"
                        :min="1"
                        :value.sync="curHPA.config.spec.minReplicas"
                        :list="varList">
                      </bkbcs-input>
                    </div>

                    <div class="bk-form-input-group is-addon-left" style="display: inline-block;">
                      <span class="input-group-addon prefix">
                        {{$t('deploy.templateset.max')}}
                      </span>
                      <bkbcs-input
                        type="number"
                        :placeholder="$t('generic.placeholder.input')"
                        style="width: 103px;"
                        :min="curHPA.config.spec.minReplicas"
                        :value.sync="curHPA.config.spec.maxReplicas"
                        :list="varList">
                      </bkbcs-input>
                    </div>
                  </div>
                </div>
                <div class="biz-span">
                  <span data-v-d78ff3e4="" class="title">{{$t('deploy.templateset.scalingTriggerCondition')}}</span>
                </div>

                <div class="bk-form-item">
                  <label class="bk-label" style="width: 140px;"></label>
                  <div class="bk-form-content" style="margin-left: 140px;">
                    <table class="biz-simple-table">
                      <thead>
                        <tr>
                          <th style="width: 313px;">{{$t('k8s.kind')}}</th>
                          <th style="width: 225px;">{{$t('deploy.templateset.resourceTarget')}}</th>
                          <th></th>
                        </tr>
                      </thead>
                      <tbody>
                        <tr v-for="(metric, index) in curHPA.config.spec.metrics" :key="index">
                          <td>
                            <bk-selector
                              :placeholder="$t('deploy.templateset.selectResourceType')"
                              :setting-key="'name'"
                              :display-key="'description'"
                              :selected.sync="metric.resource.name"
                              :filter-list="metricFilterList"
                              :list="resourceList"
                              @item-selected="handlerMetricSelect(metric, index, ...arguments)">
                            </bk-selector>
                          </td>
                          <td>
                            <div class="bk-form-input-group">
                              <bkbcs-input
                                type="number"
                                style="width: 180px;"
                                :min="0"
                                :placeholder="$t('generic.placeholder.input')"
                                :value.sync="metric.resource.target.averageUtilization">
                              </bkbcs-input>
                              <span class="input-group-addon">
                                %
                              </span>
                            </div>
                          </td>
                          <td>
                            <bk-button class="action-btn" @click="addResource" v-if="curHPA.config.spec.metrics.length < resourceList.length">
                              <i class="bcs-icon bcs-icon-plus"></i>
                            </bk-button>
                            <bk-button class="action-btn" v-if="curHPA.config.spec.metrics.length > 1" @click="removeResource(metric, index)">
                              <i class="bcs-icon bcs-icon-minus"></i>
                            </bk-button>
                          </td>
                        </tr>
                      </tbody>
                    </table>
                  </div>
                </div>
              </div>
            </div>
          </template>
        </div>
      </div>
    </div>
  </BcsContent>
</template>

<script>
/* eslint-disable @typescript-eslint/prefer-optional-chain */
/* eslint-disable @typescript-eslint/no-unused-vars */
/* eslint-disable no-prototype-builtins */
/* eslint-disable no-multi-assign */
/* eslint-disable no-case-declarations */
/* eslint-disable @typescript-eslint/no-this-alias */
import yamljs from 'js-yaml';
import _ from 'lodash';

import header from './header.vue';
import tabs from './tabs.vue';

import { catchErrorHandler } from '@/common/util';
import ace from '@/components/ace-editor';
import BcsContent from '@/components/layout/Content.vue';
import hpaParams from '@/json/k8s-hpa.json';
import k8sBase from '@/mixins/configuration/k8s-base';
import mixinBase from '@/mixins/configuration/mixin-base';

export default {
  name: 'K8SHPA',
  components: {
    'biz-header': header,
    'biz-tabs': tabs,
    ace,
    BcsContent,
  },
  mixins: [mixinBase, k8sBase],
  data() {
    return {
      isTabChanging: false,
      applicationList: [],
      winHeight: 0,
      exceptionCode: null,
      isDataLoading: true,
      isDataSaveing: false,
      curHPACache: Object.assign({}, hpaParams),
      curHPA: hpaParams,
      isLoadingApps: false,
      compareTimer: 0,
      setTimer: 0,
      resourceList: [],
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
    isTemplateSaving() {
      return this.$store.state.k8sTemplate.isTemplateSaving;
    },
    curTemplate() {
      return this.$store.state.k8sTemplate.curTemplate;
    },
    curVersion() {
      return this.$store.state.k8sTemplate.curVersion;
    },
    deployments() {
      return this.$store.state.k8sTemplate.deployments;
    },
    HPAs() {
      return this.$store.state.k8sTemplate.HPAs;
    },
    projectId() {
      return this.$route.params.projectId;
    },
    templateId() {
      return this.$route.params.templateId;
    },
    metricFilterList() {
      return this.curHPA.config.spec.metrics.map(item => item.resource.name);
    },
  },
  watch: {
    deployments() {
      if (this.curVersion) {
        this.initApplications(this.curVersion);
      }
    },
  },
  mounted() {
    this.$refs.commonHeader.initTemplate((data) => {
      this.initResource(data);
      this.getHPAMetric();
      this.isDataLoading = false;
    });
    this.winHeight = window.innerHeight;
  },
  methods: {
    exceptionHandler(exceptionCode) {
      this.isDataLoading = false;
      this.exceptionCode = exceptionCode;
    },
    updateLocalData(data) {
      if (data.id) {
        this.curHPA.id = data.id;
      }
      if (data.version) {
        this.$store.commit('k8sTemplate/updateCurVersion', data.version);
      }
      this.$store.commit('k8sTemplate/updateHPAs', this.HPAs);
      setTimeout(() => {
        this.HPAs.forEach((item) => {
          if (item.id === data.id) {
            this.setCurHPA(item);
          }
        });
      }, 500);
    },
    setCurHPA(hpa) {
      // 切换到当前项
      this.curHPA = hpa;
      clearInterval(this.compareTimer);
      clearTimeout(this.setTimer);
      this.setTimer = setTimeout(() => {
        if (!this.curHPA.cache) {
          this.curHPA.cache = JSON.parse(JSON.stringify(hpa));
        }
        this.watchChange();
      }, 500);
    },
    watchChange() {
      this.compareTimer = setInterval(() => {
        const appCopy = JSON.parse(JSON.stringify(this.curHPA));
        const cacheCopy = JSON.parse(JSON.stringify(this.curHPA.cache));

        // 删除无用属性
        delete appCopy.isEdited;
        delete appCopy.cache;
        delete appCopy.id;

        delete cacheCopy.isEdited;
        delete cacheCopy.cache;
        delete cacheCopy.id;

        const appStr = JSON.stringify(appCopy);
        const cacheStr = JSON.stringify(cacheCopy);

        if (String(this.curHPA.id).indexOf('local_') > -1) {
          this.curHPA.isEdited = true;
        } else if (appStr !== cacheStr) {
          this.curHPA.isEdited = true;
        } else {
          this.curHPA.isEdited = false;
        }
      }, 1000);
    },
    removeLocalHPA(hpa, index) {
      // 是否删除当前项
      if (this.curHPA.id === hpa.id) {
        if (index === 0 && this.HPAs[index + 1]) {
          this.setCurHPA(this.HPAs[index + 1]);
        } else if (this.HPAs[0]) {
          this.setCurHPA(this.HPAs[0]);
        }
      }
      this.HPAs.splice(index, 1);
    },
    removeHPA(hpa, index) {
      const self = this;
      const { projectId } = this;
      const version = this.curVersion;
      const HPAId = hpa.id;

      this.$bkInfo({
        title: this.$t('generic.title.confirmDelete'),
        content: this.$createElement('p', { style: { 'text-align': 'left' } }, `${this.$t('deploy.templateset.deleteHPA')}：${hpa.config.metadata.name || this.$t('deploy.templateset.unnamed')}`),
        async confirmFn() {
          if (HPAId.indexOf && HPAId.indexOf('local_') > -1) {
            self.removeLocalHPA(hpa, index);
          } else {
            try {
              const res = await self.$store.dispatch('k8sTemplate/removeHPA', { HPAId, version, projectId });
              const { data } = res;
              self.removeLocalHPA(hpa, index);

              if (data.version) {
                self.$store.commit('k8sTemplate/updateCurVersion', data.version);
                self.$store.commit('k8sTemplate/updateBindVersion', true);
              }
            } catch (e) {
              catchErrorHandler(e, this);
            }
          }
        },
      });
    },
    addLocalHPA() {
      const hpa = JSON.parse(JSON.stringify(hpaParams));
      const index = this.HPAs.length + 1;
      const now = +new Date();

      hpa.id = `local_${now}`;
      hpa.isEdited = true;
      hpa.config.metadata.name = `hpa-${index}`;
      this.HPAs.push(hpa);
      this.setCurHPA(hpa);
      this.$refs.hpaTooltip && (this.$refs.hpaTooltip.visible = false);
      this.$store.commit('k8sTemplate/updateHPAs', this.HPAs);
    },
    saveHPASuccess(params) {
      this.HPAs.forEach((item) => {
        if (params.responseData.id === item.id || params.preId === item.id) {
          item.cache = JSON.parse(JSON.stringify(item));
        }
      });
      if (params.responseData.id === this.curHPA.id || params.preId === this.curHPA.id) {
        this.updateLocalData(params.resource);
      }
    },
    async initResource(data) {
      const version = data.latest_version_id || data.version;

      if (version) {
        await this.initApplications(version);
      }
      if (data.HPAs && data.HPAs.length) {
        this.setCurHPA(data.HPAs[0], 0);
        this.checkApplication();
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
    async getHPAMetric() {
      try {
        const res = await this.$store.dispatch('k8sTemplate/getHPAMetric', this.projectId);
        this.resourceList = res.data;
      } catch (e) {
        catchErrorHandler(e, this);
      }
    },
    async tabResource(type, target) {
      this.isTabChanging = true;
      await this.$refs.commonHeader.saveTemplate();
      await this.$refs.commonHeader.autoSaveResource(type);
      this.$refs.commonTab.goResource(target);
    },
    reloadApplications() {
      if (this.curVersion) {
        this.isLoadingApps = true;
        this.initApplications(this.curVersion);
      }
    },
    async initApplications(version) {
      const { projectId } = this;
      this.linkAppVersion = version;

      try {
        const res = await this.$store.dispatch('k8sTemplate/getAppsByVersion', { projectId, version });
        this.applicationList = res.data.Deployment ? res.data.Deployment : [];
      } catch (e) {
        catchErrorHandler(e, this);
      } finally {
        this.isLoadingApps = false;
      }
    },
    showJsonPanel() {
      this.toJsonDialogConf.title = `${this.curHPA.config.metadata.name}.yaml`;
      const appConfig = JSON.parse(JSON.stringify(this.curHPA.config));

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
      const appParams = hpaParams.config;
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

      const newConfObj = _.merge({}, hpaParams.config, appObj);
      const jsonFromat = this.formatJson(newConfObj);
      this.curHPA.config = jsonFromat;
      this.toJsonDialogConf.isShow = false;
    },
    addResource() {
      this.curHPA.config.spec.metrics.push({
        type: '',
        resource: {
          name: '',
          target: {
            type: 'Utilization',
            averageUtilization: '',
          },
        },
      });
    },
    removeResource(metric, index) {
      this.curHPA.config.spec.metrics.splice(index, 1);
    },
    handlerMetricSelect(metric, metricIndex, resourceName, resource) {
      const trigger = this.curHPA.config.spec.metrics[metricIndex];
      trigger.type = resource.type;

      if (resource.type === 'Resource') {
        delete metric.pods;
        trigger.resource = {
          name: resource.name,
          target: {
            type: 'Utilization',
            averageUtilization: '',
          },
        };
      }
    },
    checkApplication() {
      const appNames = this.applicationList.map(app => app.deploy_name);
      const deploymentName = this.curHPA.config.spec.scaleTargetRef.name;
      if (deploymentName && !appNames.includes(deploymentName)) {
        this.$bkMessage({
          theme: 'error',
          message: this.$t('deploy.templateset.rebindWarning', {
            curHPA: this.curHPA.config.metadata.name,
            deploymentName,
          }),
          delay: 5000,
        });
        return false;
      }
    },
  },
};
</script>

<style scoped>
    @import './hpa.css'
</style>
