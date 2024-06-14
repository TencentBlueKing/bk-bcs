<template>
  <BcsContent>
    <template #header>
      <biz-header
        ref="commonHeader"
        @exception="exceptionHandler"
        @saveConfigmapSuccess="saveConfigmapSuccess"
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
              {{$t('deploy.templateset.configMapExplanation')}}，
              <a
                class="bk-text-button"
                :href="PROJECT_CONFIG.k8sConfigmap" target="_blank">{{$t('plugin.tools.docs')}}</a>
            </div>
          </bk-alert>
          <template v-if="!configmaps.length">
            <div class="biz-guide-box mt0">
              <bk-button icon="plus" type="primary" @click.stop.prevent="addLocalConfigmap">
                <span style="margin-left: 0;">{{$t('generic.button.add')}}ConfigMap</span>
              </bk-button>
            </div>
          </template>
          <template v-else>
            <div class="biz-configuration-topbar">
              <div class="biz-list-operation">
                <div class="item" v-for="(configmap, index) in configmaps" :key="configmap.id">
                  <bk-button
                    :class="['bk-button', { 'bk-primary': curConfigmap.id === configmap.id }]"
                    @click.stop="setCurConfigmap(configmap, index)">
                    {{(configmap && configmap.config.metadata.name) || $t('deploy.templateset.unnamed')}}
                    <span class="biz-update-dot" v-show="configmap.isEdited"></span>
                  </bk-button>
                  <span class="bcs-icon bcs-icon-close" @click.stop="removeConfigmap(configmap, index)"></span>
                </div>

                <bcs-popover ref="configmapTooltip" :content="$t('deploy.templateset.addConfigMap')" placement="top">
                  <bk-button class="bk-button bk-default is-outline is-icon" @click.stop="addLocalConfigmap">
                    <i class="bcs-icon bcs-icon-plus"></i>
                  </bk-button>
                </bcs-popover>
              </div>
            </div>

            <div class="biz-configuration-content" style="position: relative; margin-bottom: 105px;">
              <div class="bk-form biz-configuration-form">
                <a
                  href="javascript:void(0);" class="bk-text-button from-json-btn"
                  @click.stop.prevent="showJsonPanel">{{$t('deploy.templateset.importYAML')}}</a>

                <bk-sideslider
                  :is-show.sync="toJsonDialogConf.isShow"
                  :title="toJsonDialogConf.title"
                  :width="toJsonDialogConf.width"
                  class="biz-app-container-tojson-sideslider"
                  :quick-close="false"
                  @hidden="closeToJson">
                  <div slot="content" style="position: relative;">
                    <div
                      class="biz-log-box"
                      :style="{ height: `${winHeight - 60}px` }"
                      v-bkloading="{ isLoading: toJsonDialogConf.loading }">
                      <bk-button
                        class="bk-button bk-primary save-json-btn"
                        @click.stop.prevent="saveApplicationJson">{{$t('generic.button.import')}}</bk-button>
                      <bk-button
                        class="bk-button bk-default hide-json-btn"
                        @click.stop.prevent="hideApplicationJson">{{$t('generic.button.cancel')}}</bk-button>
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
                  <label class="bk-label" style="width: 105px;">{{$t('generic.label.name')}}：</label>
                  <div class="bk-form-content" style="margin-left: 105px;">
                    <input
                      type="text"
                      :class="['bk-form-input',{ 'is-danger': errors.has('configmapName') }]"
                      :placeholder="$t('deploy.templateset.enterCharacterLimit')"
                      style="width: 310px;"
                      maxlength="64"
                      v-model="curConfigmap.config.metadata.name"
                      name="configmapName"
                      v-validate="{
                        required: true,
                        regex: /^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$/
                      }">
                    <div class="bk-form-tip" v-if="errors.has('configmapName')">
                      <p class="bk-tip-text">{{$t('deploy.templateset.nameIsRequired')}}</p>
                    </div>
                  </div>
                </div>
                <template>
                  <div class="bk-form-item is-required">
                    <label class="bk-label" style="width: 105px;">{{$t('generic.label.key')}}：</label>
                    <div class="bk-form-content" style="margin-left: 105px;">
                      <div class="biz-list-operation">
                        <div class="item" v-for="(data, index) in curConfigmap.configmapKeyList" :key="index">
                          <bk-button
                            :class="['bk-button', { 'bk-primary': curKeyIndex === index }]"
                            @click.stop.prevent="setCurKey(data, index)"
                            v-if="!data.isEdit">
                            {{data.key || $t('deploy.templateset.unnamed')}}
                          </bk-button>
                          <bkbcs-input
                            type="text"
                            placeholder=""
                            v-else
                            style="width: 150px;"
                            :auto-focus="true"
                            :value.sync="data.key"
                            :list="varList"
                            @blur="setKey(data, index)">
                          </bkbcs-input>
                          <span
                            class="bcs-icon bcs-icon-edit" v-show="!data.isEdit"
                            @click.stop.prevent="editKey(data, index)"></span>
                          <span
                            class="bcs-icon bcs-icon-close" v-show="!data.isEdit"
                            @click.stop.prevent="removeKey(data, index)"></span>
                        </div>
                        <bcs-popover ref="keyTooltip" :content="$t('deploy.templateset.addKey')" placement="top">
                          <bk-button class="bk-button bk-default is-outline is-icon" @click.stop.prevent="addKey">
                            <i class="bcs-icon bcs-icon-plus"></i>
                          </bk-button>
                        </bcs-popover>
                      </div>
                    </div>
                  </div>
                  <template v-if="curKeyParams">
                    <div class="bk-form-item">
                      <label class="bk-label" style="width: 105px;">{{$t('generic.label.value')}}：</label>
                      <div class="bk-form-content" style="margin-left: 105px;">
                        <bk-input
                          type="textarea"
                          class="biz-resize-textarea"
                          ext-style="height: 300px;"
                          :rows="15"
                          v-model="curKeyParams.content"
                          :placeholder="$t('deploy.templateset.enterKeyName') + curKeyParams.key + $t('deploy.templateset.keyContent')">
                        </bk-input>
                      </div>
                    </div>
                  </template>
                </template>
              </div>
            </div>
          </template>
        </div>
      </div>
    </div>
  </BcsContent>
</template>

<script>
import yamljs from 'js-yaml';
import _ from 'lodash';

import header from './header.vue';
import tabs from './tabs.vue';

import { catchErrorHandler } from '@/common/util';
import ace from '@/components/ace-editor';
import BcsContent from '@/components/layout/Content.vue';
import configmapParams from '@/json/k8s-configmap.json';
import k8sBase from '@/mixins/configuration/k8s-base';
import mixinBase from '@/mixins/configuration/mixin-base';

export default {
  name: 'ConfigMap',
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
      winHeight: 0,
      exceptionCode: null,
      isDataLoading: true,
      isDataSaveing: false,
      curConfigmapCache: Object.assign({}, configmapParams),
      curConfigmap: configmapParams,
      configmapKeyList: [],
      curKeyIndex: 0,
      curKeyParams: null,
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
  },
  async beforeRouteLeave(to, form, next) {
    // 修改模板集信息
    // await this.$refs.commonHeader.saveTemplate()
    this.updateConfigmapDatas();
    next(true);
  },
  mounted() {
    this.$refs.commonHeader.initTemplate((data) => {
      this.initResource(data);
      this.isDataLoading = false;
    });
    this.winHeight = window.innerHeight;
  },
  methods: {
    exceptionHandler(exceptionCode) {
      this.isDataLoading = false;
      this.exceptionCode = exceptionCode;
    },
    setCurKey(data, index) {
      this.curKeyParams = data;
      this.curKeyIndex = index;
    },
    editKey(data) {
      data.isEdit = true;
    },
    removeKey(data, index) {
      if (this.curKeyIndex > index) {
        this.curKeyIndex = this.curKeyIndex - 1;
      } else if (this.curKeyIndex === index) {
        this.curKeyIndex = 0;
      }
      this.curConfigmap.configmapKeyList.splice(index, 1);
      this.curKeyParams = this.curConfigmap.configmapKeyList[this.curKeyIndex];
    },
    setKey(data, index) {
      if (data.key === '') {
        data.key = `key-${this.curConfigmap.configmapKeyList.length}`;
      } else {
        const nameReg = /^[a-zA-Z]{1}[a-zA-Z0-9-_.]{0,254}$/;
        // eslint-disable-next-line no-useless-escape
        const varReg = /\{\{([^\{\}]+)?\}\}/g;

        if (!nameReg.test(data.key.replace(varReg, 'key'))) {
          this.$bkMessage({
            theme: 'error',
            message: this.$t('deploy.templateset.msg.labelKey'),
            delay: 5000,
          });
          return false;
        }

        const keyObj = {};
        for (const item of this.curConfigmap.configmapKeyList) {
          if (!keyObj[item.key]) {
            keyObj[item.key] = true;
          } else {
            this.$bkMessage({
              theme: 'error',
              message: this.$t('deploy.templateset.keyNotDuplicate'),
              delay: 5000,
            });
            return false;
          }
        }
      }
      this.curKeyParams = this.curConfigmap.configmapKeyList[index];
      this.curKeyIndex = index;
      data.isEdit = false;
    },
    addKey() {
      const index = this.curConfigmap.configmapKeyList.length + 1;
      this.curConfigmap.configmapKeyList.push({
        key: `key-${index}`,
        isEdit: true,
        content: '',
      });
      this.curKeyParams = this.curConfigmap.configmapKeyList[index - 1];
      this.curKeyIndex = index - 1;
      this.$refs.keyTooltip.visible = false;
    },
    updateLocalData(data) {
      if (data.id) {
        this.curConfigmap.id = data.id;
      }
      if (data.version) {
        this.$store.commit('k8sTemplate/updateCurVersion', data.version);
      }
      this.$store.commit('k8sTemplate/updateConfigmaps', this.configmaps);
      setTimeout(() => {
        this.configmaps.forEach((item) => {
          if (item.id === data.id) {
            this.setCurConfigmap(item);
          }
        });
      }, 500);
    },
    setCurConfigmap(configmap) {
      // 同步上一个键值
      const params = {};
      const keys = this.curConfigmap.configmapKeyList;
      if (keys?.length) {
        keys.forEach((item) => {
          params[item.key] = item.content;
        });
        this.curConfigmap.config.data = params;
      }

      // 切换到当前项
      configmap.configmapKeyList = [];
      this.curConfigmap = configmap;
      this.initConfigmapKeyList(configmap);

      clearInterval(this.compareTimer);
      clearTimeout(this.setTimer);
      this.setTimer = setTimeout(() => {
        if (!this.curConfigmap.cache) {
          this.curConfigmap.cache = JSON.parse(JSON.stringify(configmap));
        }
        this.watchChange();
      }, 500);
    },
    initConfigmapKeyList(configmap) {
      const list = [];
      const keys = configmap.config.data;
      for (const [key, value] of Object.entries(keys)) {
        list.push({
          key,
          isEdit: false,
          content: value,
        });
      }
      this.curKeyIndex = 0;
      if (list.length) {
        this.curKeyParams = list[0];
      } else {
        this.curKeyParams = null;
      }

      configmap.configmapKeyList.splice(0, configmap.configmapKeyList.length, ...list);
    },
    watchChange() {
      this.compareTimer = setInterval(() => {
        const appCopy = JSON.parse(JSON.stringify(this.curConfigmap));
        const cacheCopy = JSON.parse(JSON.stringify(this.curConfigmap.cache));

        // 删除无用属性
        delete appCopy.isEdited;
        delete appCopy.cache;
        delete appCopy.id;
        if (appCopy.configmapKeyList) {
          appCopy.configmapKeyList.forEach((item) => {
            delete item.isEdit;
          });
        }

        delete cacheCopy.isEdited;
        delete cacheCopy.cache;
        delete cacheCopy.id;
        if (cacheCopy.configmapKeyList) {
          cacheCopy.configmapKeyList.forEach((item) => {
            delete item.isEdit;
          });
        }

        const appStr = JSON.stringify(appCopy);
        const cacheStr = JSON.stringify(cacheCopy);
        const keyObj = {};

        const keys = this.curConfigmap.configmapKeyList;
        keys.forEach((item) => {
          keyObj[item.key] = item.content;
        });

        if (String(this.curConfigmap.id).indexOf('local_') > -1) {
          this.curConfigmap.isEdited = true;
        } else if (appStr !== cacheStr) {
          this.curConfigmap.isEdited = true;
        } else {
          this.curConfigmap.isEdited = false;
        }
      }, 1000);
    },
    removeLocalConfigmap(configmap, index) {
      // 是否删除当前项
      if (this.curConfigmap.id === configmap.id) {
        if (index === 0 && this.configmaps[index + 1]) {
          this.setCurConfigmap(this.configmaps[index + 1]);
        } else if (this.configmaps[0]) {
          this.setCurConfigmap(this.configmaps[0]);
        }
      }
      this.configmaps.splice(index, 1);
    },
    removeConfigmap(configmap, index) {
      // eslint-disable-next-line @typescript-eslint/no-this-alias
      const self = this;
      const { projectId } = this;
      const version = this.curVersion;
      const configmapId = configmap.id;

      this.$bkInfo({
        title: this.$t('generic.title.confirmDelete'),
        content: this.$createElement('p', { style: { 'text-align': 'left' } }, `${this.$t('deploy.templateset.deleteConfigMap')}：${configmap.config.metadata.name || this.$t('deploy.templateset.unnamed')}`),
        async confirmFn() {
          if (configmapId.indexOf && configmapId.indexOf('local_') > -1) {
            self.removeLocalConfigmap(configmap, index);
          } else {
            try {
              const res = await self.$store.dispatch('k8sTemplate/removeConfigmap', {
                configmapId,
                version,
                projectId,
              });

              const { data } = res;
              self.removeLocalConfigmap(configmap, index);

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

    addLocalConfigmap() {
      const configmap = JSON.parse(JSON.stringify(configmapParams));
      const index = this.configmaps.length + 1;
      const now = +new Date();

      configmap.id = `local_${now}`;
      configmap.isEdited = true;
      configmap.config.metadata.name = `configmap-${index}`;
      this.configmaps.push(configmap);
      this.setCurConfigmap(configmap);
      this.$refs.configmapTooltip && (this.$refs.configmapTooltip.visible = false);
      this.$store.commit('k8sTemplate/updateConfigmaps', this.configmaps);
    },
    saveConfigmapSuccess(params) {
      this.configmaps.forEach((item) => {
        if (params.responseData.id === item.id || params.preId === item.id) {
          item.cache = JSON.parse(JSON.stringify(item));
        }
      });
      if (params.responseData.id === this.curConfigmap.id || params.preId === this.curConfigmap.id) {
        this.updateLocalData(params.resource);
      }
    },
    updateConfigmapDatas() {
      const keyObj = {};
      const keys = this.curConfigmap.configmapKeyList;
      if (keys) {
        keys.forEach((item) => {
          keyObj[item.key] = item.content;
        });
        this.curConfigmap.config.data = keyObj;
      }
    },
    initResource(data) {
      if (data.configmaps?.length) {
        this.setCurConfigmap(data.configmaps[0], 0);
      } else if (data.configmap?.length) {
        this.setCurConfigmap(data.configmap[0], 0);
      }
    },
    exportToYaml() {
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
      this.toJsonDialogConf.title = `${this.curConfigmap.config.metadata.name}.yaml`;
      const keyObj = {};
      const keys = this.curConfigmap.configmapKeyList;
      const appConfig = JSON.parse(JSON.stringify(this.curConfigmap.config));

      if (keys?.length) {
        keys.forEach((item) => {
          keyObj[item.key] = item.content;
        });
        appConfig.data = keyObj;
      }

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
      const appParams = configmapParams.config;
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
      if (annot?.length) {
        editor.gotoLine(annot[0].row, annot[0].column, true);
        return false;
      }
      const newConfObj = _.merge({}, configmapParams.config, appObj);
      const jsonFromat = this.formatJson(newConfObj);
      this.curConfigmap.config = jsonFromat;
      this.initConfigmapKeyList(this.curConfigmap);
      this.toJsonDialogConf.isShow = false;
    },
  },
};
</script>

<style scoped>
    @import './configmap.css';
</style>
