<!-- eslint-disable max-len -->
<template>
  <ContentHeader :style="{ marginBottom: (isNewTemplate) ? '0px' : '55px' }">
    <i class="bcs-icon bcs-icon-arrows-left" @click="beforeLeave"></i>
    <div class="biz-templateset-title">
      <span>{{$t('deploy.templateset.returnToTemplateSet')}}</span>
    </div>
    <div class="biz-templateset-title">
      <span v-show="!isEditName">{{curTemplate.name}}</span>
      <input
        type="text"
        :placeholder="$t('deploy.templateset.enterMaxLength30')"
        maxlength="30"
        class="bk-form-input"
        v-model="editTemplate.name"
        v-bk-focus v-if="isEditName"
        @blur="saveTemplate"
        @keyup.enter="saveTemplate" />
      <a href="javascript:void(0)" class="bk-text-button bk-default" v-show="!isEditName" @click="editTemplateName">
        <i class="bcs-icon bcs-icon-edit"></i>
      </a>
    </div>
    <div class="biz-templateset-desc">
      <span v-show="!isEditDesc">{{curTemplate.desc}}</span>
      <input
        type="text"
        :placeholder="$t('deploy.templateset.enterMaxLength50')"
        maxlength="50"
        class="bk-form-input"
        v-model="editTemplate.desc"
        v-bk-focus
        v-if="isEditDesc"
        @blur="saveTemplate"
        @keyup.enter="saveTemplate" />
      <a href="javascript:void(0)" class="bk-text-button bk-default" v-show="!isEditDesc" @click="editTemplateDesc" @keyup.enter="saveTemplate">
        <i class="bcs-icon bcs-icon-edit"></i>
      </a>
    </div>
    <template #right>
      <div v-if="!exceptionCode && !isTemplateLoading">
        <!-- 如果不是新增状态的模板集并且有权限编辑才可查看加锁状态 -->
        <template v-if="String(curTemplateId) !== '0'">
          <template v-if="templateLockStatus.isLocked">
            <template v-if="templateLockStatus.isCurLocker">
              <div class="biz-lock-box">
                <div class="lock-wrapper warning">
                  <i class="bcs-icon bcs-icon-info-circle-shape"></i>
                  <strong class="desc">
                    {{$t('deploy.templateset.templateSetLocked')}}
                    <span v-if="lateShowVersionName">
                      （{{$t('deploy.templateset.currentVersionNumber')}}：{{lateShowVersionName}}
                      <bcs-popover
                        :delay="300"
                        :content="displayVersionNotes || '--'"
                        style="padding-left: 6px;"
                        placement="bottom">
                        <span style="color: #3a84ff;">{{$t('deploy.templateset.versionDescription')}}</span>
                      </bcs-popover>）
                    </span>
                  </strong>
                  <div class="action">
                    <bk-switcher
                      :selected="templateLockStatus.isLocked"
                      size="small"
                      @change="updateTemplateLockStatus">
                    </bk-switcher>
                  </div>
                </div>
              </div>
            </template>
            <template v-else>
              <div class="biz-lock-box">
                <div class="lock-wrapper warning">
                  <i class="bcs-icon bcs-icon-info-circle-shape"></i>
                  <strong class="desc">
                    {{$t('deploy.templateset.msg.multiOperator', templateLockStatus)}}
                    <span v-if="lateShowVersionName">
                      （{{$t('deploy.templateset.currentVersionNumber')}}：{{lateShowVersionName}}
                      <bcs-popover
                        :delay="300"
                        :content="displayVersionNotes || '--'"
                        style="padding-left: 6px;"
                        placement="bottom">
                        <span style="color: #3a84ff;">{{$t('deploy.templateset.versionDescription')}}</span>
                      </bcs-popover>）</span>
                  </strong>
                  <div class="action">
                    <a href="javascript: void(0);" class="bk-text-button" @click="reloadTemplateset">{{$t('deploy.templateset.clickRefresh')}}</a>
                  </div>
                </div>
              </div>
            </template>
          </template>
          <template v-else>
            <div class="biz-lock-box">
              <div class="lock-wrapper">
                <i class="bcs-icon bcs-icon-info-circle-shape"></i>
                <strong class="desc">
                  {{$t('deploy.templateset.protectionWhileEditing')}}
                  <span v-if="lateShowVersionName">
                    （{{$t('deploy.templateset.currentVersionNumber')}}：{{lateShowVersionName}}
                    <bcs-popover
                      :delay="300"
                      :content="displayVersionNotes || '--'"
                      style="padding-left: 6px;"
                      placement="bottom">
                      <span style="color: #3a84ff;">{{$t('deploy.templateset.versionDescription')}}</span>
                    </bcs-popover>）</span>
                </strong>
                <div class="action">
                  <bk-switcher
                    :selected="templateLockStatus.isLocked"
                    size="small"
                    @change="updateTemplateLockStatus">
                  </bk-switcher>
                </div>
              </div>
            </div>
          </template>
        </template>

        <!-- 如果模板集没有加锁或者当前用户是加锁者才可以操作 -->
        <template>
          <template v-if="templateLockStatus.isLocked && !templateLockStatus.isCurLocker">
            <bk-button disabled>{{$t('deploy.templateset.saveDraft')}}</bk-button>
            <bk-button type="primary" style="width: 70px;" disabled>{{$t('generic.button.save')}}</bk-button>
          </template>
          <template v-else>
            <span
              v-authority="{
                actionId: isNewTemplate ? 'templateset_create' : 'templateset_update',
                resourceName: curTemplate.name,
                permCtx: {
                  resource_type: isNewTemplate ? 'project' : 'templateset',
                  project_id: projectId,
                  template_id: isNewTemplate ? undefined : Number(curTemplateId)
                }
              }">
              <bk-button @click.stop.prevent="saveTemplateDraft">{{$t('deploy.templateset.saveDraft')}}</bk-button>
              <bk-button type="primary" :loading="isDataSaveing" :disabled="!isTemplateCanSave" style="width: 70px;" @click.stop.prevent="saveTemplateData">{{$t('generic.button.save')}}</bk-button>
            </span>
          </template>
        </template>

        <template>
          <bk-button
            :disabled="!canCreateInstance"
            v-authority="{
              clickable: isNewTemplate || isFromNewTemplate ? true : getAuthority('templateset_instantiate', Number(curTemplateId)),
              actionId: 'templateset_instantiate',
              resourceName: curTemplate.name,
              disablePerms: true,
              permCtx: {
                project_id: projectId,
                template_id: Number(curTemplateId)
              }
            }"
            class="ml5"
            @click.stop.prevent="createInstance">
            {{$t('deploy.templateset.tempalte')}}
          </bk-button>
        </template>

        <template>
          <bk-button
            v-authority="{
              clickable: isNewTemplate || isFromNewTemplate ? true : getAuthority('templateset_view', Number(curTemplateId)),
              actionId: 'templateset_view',
              resourceName: curTemplate.name,
              disablePerms: true,
              permCtx: {
                project_id: projectId,
                template_id: Number(curTemplateId)
              }
            }"
            @click.stop.prevent="showVersionPanel"
          >{{$t('deploy.templateset.versionList')}}</bk-button>
        </template>
      </div>
    </template>

    <bk-dialog
      :is-show.sync="versionDialogConf.isShow"
      :width="versionDialogConf.width"
      :has-header="false"
      :quick-close="false"
      :ext-cls="'create-project-dialog'"
      :content="versionDialogConf.content"
      @cancel="hideVersionBox">
      <template slot="content">
        <div class="version-box">
          <p class="title">{{$t('deploy.templateset.saveChangesTo')}}：</p>
          <ul :class="['version-list', { 'is-en': isEn }]">
            <template v-if="!isNewVersion">
              <li class="item mb10">
                <label class="bk-form-radio label-item">
                  <input type="radio" name="save-version-way" :class="{ 'is-checked': saveVersionWay === 'cur' }" value="cur" v-model="saveVersionWay">
                  <i class="bk-radio-text" style="display: inline-block; min-width: 70px;">{{$t('deploy.templateset.currentVersionNumber')}}：{{lateShowVersionName}}</i>
                </label>
              </li>

              <li class="item mb10">
                <label class="bk-form-radio label-item" style="margin-right: 0;">
                  <input type="radio" name="save-version-way" :class="{ 'is-checked': saveVersionWay === 'new' }" value="new" v-model="saveVersionWay">
                  <i class="bk-radio-text" style="display: inline-block; min-width: 70px;">{{$t('deploy.templateset.newVersion')}}：</i>
                  <bkbcs-input :placeholder="$t('deploy.templateset.enterVersionNumber')" @focus="saveVersionWay = 'new'" style="width: 176px; flex: 1;" v-model="versionKeyword" />
                </label>
              </li>

              <li class="item" v-if="withoutCurVersionList.length">
                <label class="bk-form-radio label-item" style="margin-right: 0;">
                  <input type="radio" name="save-version-way" :class="{ 'is-checked': saveVersionWay === 'old' }" value="old" v-model="saveVersionWay">
                  <i class="bk-radio-text" style="display: inline-block; min-width: 70px; letter-spacing: 0;">{{$t('deploy.templateset.otherVersions')}}：</i>
                  <bk-selector
                    style="width: 176px;"
                    :placeholder="$t('deploy.templateset.selectVersionNumber')"
                    :setting-key="'show_version_id'"
                    :selected.sync="selectedVersion"
                    :list="withoutCurVersionList"
                    @item-selected="selectVersion">
                  </bk-selector>
                </label>
              </li>
            </template>
            <template v-else>
              <li class="item">
                <label class="bk-form-radio label-item" style="margin-right: 0;">
                  <i class="bk-radio-text" style="display: inline-block; width: 70px; letter-spacing: 0;">{{$t('deploy.templateset.newVersion')}}：</i>
                  <bkbcs-input :placeholder="$t('deploy.templateset.enterVersionNumber')" @focus="saveVersionWay = 'new'" style="width: 203px; flex: 1;" v-model="versionKeyword" />
                </label>
              </li>
            </template>
            <li class="item">
              <label :class="['notes', 'label-item', { 'new-item': isNewVersion }]" style="margin-right: 0;">
                <i :class="['notes-text', { 'is-en-text': isEn, 'is-new': isNewVersion }]" :style="{ 'letter-spacing': 0, 'padding-left': isNewVersion ? 0 : '26px' }">{{$t('deploy.templateset.versionDescription')}}：</i>
                <bk-input type="textarea" class="notes-input" :style="{ width: isNewVersion ? '203px' : '176px' }" :placeholder="$t('deploy.templateset.enterVersionDescription')" v-model="curVersionNotes" />
              </label>
            </li>
          </ul>
        </div>
      </template>
      <div slot="footer">
        <template v-if="!canVersionSave">
          <bk-button type="primary" disabled>
            {{$t('generic.button.confirm')}}
          </bk-button>
        </template>
        <template v-else>
          <bk-button type="primary" :loading="isCreating" @click="saveVersion">
            {{$t('generic.button.confirm')}}
          </bk-button>
        </template>
        <bk-button type="button" :disabled="isCreating" @click="hideVersionBox">
          {{$t('generic.button.cancel')}}
        </bk-button>
      </div>
    </bk-dialog>

    <svg style="display: none;">
      <title>{{$t('deploy.templateset.icon')}}</title>
      <symbol id="biz-set-icon" viewBox="0 0 32 32">
        <path d="M6 3v3h-3v23h23v-3h3v-23h-23zM24 24v3h-19v-19h19v16zM27 24h-1v-18h-18v-1h19v19z"></path>
        <path d="M13.688 18.313h-6v6h6v-6z"></path>
        <path d="M21.313 10.688h-6v13.625h6v-13.625z"></path>
        <path d="M13.688 10.688h-6v6h6v-6z"></path>
      </symbol>
    </svg>

    <bk-dialog
      :is-show.sync="selectorConfirmDialog.isShow"
      title=""
      :width="500"
      :has-footer="false"
      :close-icon="selectorConfirmDialog.closeIcon"
      :quick-close="false"
      @cancel="selectorConfirmDialog.isShow = false">
      <template slot="content">
        <div class="biz-danger-tip">
          <div class="tip-header">
            <svg class="biz-set-icon"><use xlink:href="#biz-set-icon"></use></svg>
            <p class="title">{{$t('deploy.templateset.operationRisk')}}</p>
          </div>
          <div class="tip-content">
            <strong>{{$t('deploy.templateset.operationRiskAware')}}</strong>
            <p class="tip">{{$t('deploy.templateset.tip3')}}</p>
          </div>
          <div class="tip-footer">
            <bk-button class="bk-button bk-primary mr10">{{$t('generic.button.confirm')}}</bk-button>
            <bk-button>{{$t('generic.button.cancel')}}</bk-button>
          </div>
        </div>
      </template>
    </bk-dialog>

    <bk-sideslider
      :quick-close="true"
      :is-show.sync="versionSidePanel.isShow"
      :title="versionSidePanel.title"
      :width="900">
      <div class="p30" slot="content" v-bkloading="{ isLoading: isVersionListLoading }">
        <bk-table
          :data="versionList"
          :size="'medium'">
          <bk-table-column :label="$t('deploy.templateset.versionNumber')" :show-overflow-tooltip="false" min-width="200">
            <template slot-scope="props">
              <p>
                <span>{{props.row.name}}</span>
                <span v-if="props.row.show_version_id === curShowVersionId">{{$t('deploy.templateset.current')}}</span>
              </p>

              <bcs-popover
                v-if="props.row.comment"
                :delay="300"
                :content="props.row.comment"
                placement="right">
                <span style="color: #3a84ff; font-size: 12px;">{{$t('deploy.templateset.versionDescription')}}</span>
              </bcs-popover>
            </template>
          </bk-table-column>
          <bk-table-column :label="$t('cluster.labels.updatedAt')" :show-overflow-tooltip="true" min-width="150">
            <template slot-scope="props">
              {{props.row.updated}}
            </template>
          </bk-table-column>
          <bk-table-column :label="$t('deploy.templateset.lastUpdatedBy')" :show-overflow-tooltip="true" min-width="150">
            <template slot-scope="props">
              <bk-user-display-name v-if="props.row?.updator" :user-id="props.row?.updator">
              </bk-user-display-name>
              <span v-else>--</span>
            </template>
          </bk-table-column>
          <bk-table-column :label="$t('generic.label.action')" :show-overflow-tooltip="true" min-width="200">
            <template slot-scope="props">
              <a href="javascript:void(0);" class="bk-text-button" @click.stop.prevent="getTemplateByVersion(props.row.show_version_id)">{{$t('deploy.templateset.load')}}</a>
              <template v-if="props.row.show_version_id === -1">
                <a href="javascript:void(0);" class="bk-text-button is-disabled" v-bk-tooltips="$t('deploy.templateset.draftNotSupportYAML')">{{$t('deploy.templateset.convertToYAML')}}</a>
              </template>
              <template v-else>
                <a href="javascript:void(0);" class="bk-text-button" @click.stop.prevent="handleExportToYaml(props.row)">{{$t('deploy.templateset.convertToYAML')}}</a>
              </template>
              <!-- 只有一个版本时不能删除 -->
              <template v-if="versionList.length <= 1">
                <bcs-popover :delay="300" placement="right">
                  <a href="javascript:void(0);" class="bk-text-button is-disabled ml5" disabled>{{$t('generic.button.delete')}}</a>
                  <template slot="content">
                    <p class="biz-permission-tip">
                      {{$t('deploy.templateset.retainAtLeastOneVersion')}}
                    </p>
                  </template>
                </bcs-popover>
              </template>
              <template v-else>
                <!-- 有编辑权限 -->
                <template>
                  <!-- 已经加锁，且是当前加锁人 -->
                  <template v-if="templateLockStatus.isLocked && templateLockStatus.isCurLocker">
                    <a href="javascript:void(0);" class="bk-text-button" @click.stop.prevent="removeVersion(props.row)">{{$t('generic.button.delete')}}</a>
                  </template>
                  <!-- 已经加锁，但不是当前加锁人 -->
                  <template v-else-if="templateLockStatus.isLocked && !templateLockStatus.isCurLocker">
                    <bcs-popover :delay="300" placement="right">
                      <a href="javascript:void(0);" class="bk-text-button is-disabled ml5" disabled>{{$t('generic.button.delete')}}</a>
                      <template slot="content">
                        <p class="biz-permission-tip">
                          {{$t('deploy.templateset.msg.multiOperator', templateLockStatus)}}
                        </p>
                      </template>
                    </bcs-popover>
                  </template>
                  <!-- 没有加锁 -->
                  <template v-else>
                    <a href="javascript:void(0);" class="bk-text-button" @click.stop.prevent="removeVersion(props.row)">{{$t('generic.button.delete')}}</a>
                  </template>
                </template>
              </template>
            </template>
          </bk-table-column>
        </bk-table>
      </div>
    </bk-sideslider>
  </ContentHeader>
</template>

<script>
/* eslint-disable @typescript-eslint/prefer-optional-chain */
/* eslint-disable @typescript-eslint/no-unused-vars */
/* eslint-disable no-prototype-builtins */
/* eslint-disable no-multi-assign */
/* eslint-disable no-case-declarations */
/* eslint-disable @typescript-eslint/no-this-alias */
/* eslint-disable max-len */
/* eslint-disable no-useless-escape */
import base64 from 'base-64';
import yamljs from 'js-yaml';
import shlex from 'shlex';

import { isObject } from '@/common/util';
import ContentHeader from '@/components/layout/Header.vue';

export default {
  name: 'HeaderIndex',
  components: { ContentHeader },
  data() {
    return {
      isFromNewTemplate: false,
      saveVersionWay: 'cur',
      isEditName: false,
      isEditDesc: false,
      isCreating: false,
      selectorConfirmDialog: {
        isShow: false,
      },
      editTemplate: {
        name: '',
        desc: '',
      },
      isTemplateLocking: false,
      isVersionListLoading: true,
      exceptionCode: null,
      versionSidePanel: {
        isShow: false,
        title: this.$t('deploy.templateset.versionList'),
      },
      versionDialogConf: {
        isShow: false,
        width: 600,
        closeIcon: false,
      },
      versionMetadata: {
        show_version_id: -1,
        name: '',
        real_version_id: 0,
      },
      isTemplateLoading: true,
      isDataSaveing: false,
      newTemplateId: 0,
      versionKeyword: '',
      canCreateInstance: false,
      selectedVersion: '',
      curApplicationCache: null,
      curVersionNotes: '',
      displayVersionNotes: '--',
      webAnnotations: null,
    };
  },
  computed: {
    isEn() {
      return this.$store.state.isEn;
    },
    projectCode() {
      return this.$route.params.projectCode;
    },
    canVersionSave() {
      if (this.saveVersionWay === 'cur' && this.curVersion) {
        return true;
      } if (this.saveVersionWay === 'old' && this.selectedVersion) {
        return true;
      } if (this.saveVersionWay === 'new' && this.versionKeyword) {
        return true;
      }
      return false;
    },
    lateShowVersionName() {
      let name = '';
      this.allVersionList.forEach((item) => {
        if (item.show_version_id === this.curShowVersionId) {
          name = item.name;
        }
      });
      return name;
    },
    metricList() {
      const ids = [];
      const list = this.$store.state.k8sTemplate.metricList;
      list.forEach((item) => {
        ids.push(item.id);
      });
      return ids;
    },
    isTemplateCanSave() {
      // 如果没有创建模板，查看是否有资源已经编辑过
      if (!this.curShowVersionId) {
        const { deployments } = this;
        const { daemonsets } = this;
        const { jobs } = this;
        const { statefulsets } = this;
        const { services } = this;
        const { configmaps } = this;
        const { secrets } = this;
        const { ingresss } = this;

        for (const deployment of deployments) {
          if (deployment.isEdited) {
            return true;
          }
        }

        for (const service of services) {
          if (service.isEdited) {
            return true;
          }
        }

        for (const daemonset of daemonsets) {
          if (daemonset.isEdited) {
            return true;
          }
        }

        for (const job of jobs) {
          if (job.isEdited) {
            return true;
          }
        }

        for (const statefulset of statefulsets) {
          if (statefulset.isEdited) {
            return true;
          }
        }

        for (const configmap of configmaps) {
          if (configmap.isEdited) {
            return true;
          }
        }

        for (const secret of secrets) {
          if (secret.isEdited) {
            return true;
          }
        }

        for (const ingress of ingresss) {
          if (ingress.isEdited) {
            return true;
          }
        }

        if (this.$store.state.k8sTemplate.canTemplateBindVersion) {
          return true;
        }
        return false;
      }
      return true;
    },
    curTemplateId() {
      return this.$store.state.k8sTemplate.curTemplateId || this.newTemplateId || this.$route.params.templateId;
    },
    userInfo() {
      return this.$store.state.user;
    },
    templateLockStatus() {
      const status = {
        isLocked: false,
        isCurLocker: false,
        locker: '',
      };
      // 模块集已经加锁
      if (this.curTemplate && this.curTemplate.is_locked) {
        status.isLocked = true;
        status.locker = this.curTemplate.locker;
        // 如果是当前用户加锁
        if (this.curTemplate.locker && this.curTemplate.locker === this.userInfo.username) {
          status.isCurLocker = true;
        } else {
          status.isCurLocker = false;
        }
      }
      return status;
    },
    curVersion() {
      return this.$store.state.k8sTemplate.curVersion;
    },
    curShowVersionId() {
      return this.$store.state.k8sTemplate.curShowVersionId;
    },
    isNewTemplate() {
      const { templateId } = this.$route.params;
      if (String(templateId) === '0') {
        return true;
      }
      return false;
    },
    curProject() {
      return this.$store.state.curProject;
    },
    curTemplate() {
      return this.$store.state.k8sTemplate.curTemplate;
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
    ingresss() {
      return this.$store.state.k8sTemplate.ingresss;
    },
    HPAs() {
      return this.$store.state.k8sTemplate.HPAs;
    },
    projectId() {
      return this.$route.params.projectId;
    },
    versionList() {
      const list = this.$store.state.k8sTemplate.versionList;
      if (list.length === 0) {
        this.canCreateInstance = false;
      } else if (list.length === 1 && list[0].show_version_id === -1) {
        this.canCreateInstance = false;
      } else {
        this.canCreateInstance = true;
      }
      return list;
    },
    allVersionList() {
      return this.$store.state.k8sTemplate.versionList;
    },
    withoutCurVersionList() {
      // 去掉草稿和当前版本
      return this.$store.state.k8sTemplate.versionList
        .filter(item => item.show_version_id !== -1 && item.show_version_id !== this.curShowVersionId);
    },
    imageList() {
      return this.$store.state.k8sTemplate.imageList;
    },
    linkServices() {
      return this.$store.state.k8sTemplate.linkServices.map(item => item.service_name);
    },
    isNewVersion() {
      return !(this.allVersionList.length && this.curShowVersionId !== -1);
    },
  },
  watch: {
    'versionDialogConf.isShow'() {
      if (!this.versionDialogConf.isShow) {
        this.versionKeyword = '';
      }
    },
    '$route'() {
      this.getVersionList();
    },
    'curShowVersionId'() {
      this.allVersionList.forEach((item) => {
        if (item.show_version_id === this.curShowVersionId) {
          this.curVersionNotes = item.comment;
          this.displayVersionNotes = item.comment;
        }
      });
    },
    'saveVersionWay'(val, old) {
      if (val && val === old) return;
      if (val === 'new') {
        this.curVersionNotes = '';
        return;
      }
      let item = null;
      if (val === 'cur' && this.curVersion) {
        item = this.allVersionList.find(item => item.show_version_id === this.curShowVersionId);
      } else if (val === 'old' && this.selectedVersion) {
        item = this.allVersionList.find(item => item.show_version_id === this.selectedVersion);
      }
      item && (this.curVersionNotes = item.comment);
    },
  },
  mounted() {
    this.getVersionList();
    // this.initImageList()
  },
  methods: {
    getAuthority(actionId, templateId) {
      return !!this.webAnnotations?.perms[templateId]?.[actionId];
    },
    beforeLeave() {
      const self = this;
      let isEdited = false;

      this.deployments.forEach((item) => {
        if (item.isEdited) {
          isEdited = true;
        }
      });
      this.services.forEach((item) => {
        if (item.isEdited) {
          isEdited = true;
        }
      });
      this.configmaps.forEach((item) => {
        if (item.isEdited) {
          isEdited = true;
        }
      });
      this.secrets.forEach((item) => {
        if (item.isEdited) {
          isEdited = true;
        }
      });
      this.daemonsets.forEach((item) => {
        if (item.isEdited) {
          isEdited = true;
        }
      });
      this.jobs.forEach((item) => {
        if (item.isEdited) {
          isEdited = true;
        }
      });
      this.statefulsets.forEach((item) => {
        if (item.isEdited) {
          isEdited = true;
        }
      });
      this.ingresss.forEach((item) => {
        if (item.isEdited) {
          isEdited = true;
        }
      });
      this.HPAs.forEach((item) => {
        if (item.isEdited) {
          isEdited = true;
        }
      });
      if (isEdited || this.$store.state.k8sTemplate.canTemplateBindVersion) {
        this.$bkInfo({
          title: this.$t('deploy.templateset.confirmExit'),
          content: this.$createElement('p', {
            style: {
              textAlign: 'left',
            },
          }, this.$t('deploy.templateset.unsavedTemplateWarning')),
          confirmFn() {
            self.goTemplatePage();
          },
        });
      } else {
        this.goTemplatePage();
      }
    },
    // initImageList () {
    //     const projectId = this.projectId
    //     this.$store.dispatch('k8sTemplate/getImageList', { projectId }).then(res => {
    //         const data = res.data
    //         this.$store.commit('k8sTemplate/updateImageList', data)
    //     }, res => {
    //         this.$bkMessage({
    //             theme: 'error',
    //             message: res.message,
    //             delay: 10000
    //         })
    //     })
    // },
    removeVersion(data) {
      const self = this;
      this.$bkInfo({
        title: this.$t('generic.title.confirmDelete'),
        content: this.$createElement('p', { style: { 'text-align': 'left' } }, `${this.$t('deploy.helm.deleteVersion')}：“${data.name}”`),
        confirmFn() {
          const { projectId } = self;
          const templateId = self.curTemplateId;
          const versionId = data.show_version_id;
          self.$store.dispatch('k8sTemplate/removeVersion', { projectId, templateId, versionId }).then((res) => {
            self.$bkMessage({
              theme: 'success',
              message: self.$t('generic.msg.success.ok'),
            });

            self.getVersionList().then((versionList) => {
              // 如果是删除当前版本
              if (versionId === self.curShowVersionId || String(versionId) === self.curShowVersionId) {
                // 加载第一项，优先选择非草稿
                if (self.versionList.length) {
                  let versionData = self.versionList[0];
                  if (versionData.show_version_id === -1 && self.versionList.length > 1) {
                    versionData = self.versionList[1];
                  }
                  self.getTemplateByVersion(versionData.show_version_id, true);
                } else {
                  self.getTemplateByVersion(-1);
                }
              }
            });
          }, (res) => {
            this.$bkMessage({
              theme: 'error',
              message: res.message,
              delay: '3000',
            });
          });
        },
      });
    },

    goTemplatePage() {
      // 清空数据
      this.$store.commit('k8sTemplate/clearCurTemplateData');
      this.$router.push({
        name: 'templateset',
        params: {
          projectId: this.projectId,
        },
      });
    },
    removeTemplate() {
      const self = this;
      const { projectId } = this;
      const templateId = this.curTemplateId;
      this.$bkInfo({
        title: this.$t('generic.title.confirmDelete'),
        content: this.$t('deploy.templateset._confirmDeleteTemplateSet'),
        confirmFn() {
          self.$store.dispatch('k8sTemplate/removeTemplate', { templateId, projectId }).then((res) => {
            this.$bkMessage({
              theme: 'success',
              message: this.$t('generic.msg.success.delete'),
            });
            self.goTemplatePage();
          }, (res) => {
            this.$store.commit('k8sTemplate/updateIsTemplateSaving', false);
            this.$bkMessage({
              theme: 'error',
              message: res.message,
              hasCloseIcon: true,
              delay: '3000',
            });
          });
        },
      });
    },
    createInstance() {
      if (this.canCreateInstance) {
        this.$router.push({
          name: 'instantiation',
          params: {
            projectId: this.projectId,
            templateId: this.curTemplate.id,
            curTemplate: this.curTemplate,
            curShowVersionId: this.curShowVersionId,
          },
        });
      }
    },
    editTemplateName() {
      this.isEditName = true;
      this.editTemplate = Object.assign({}, this.curTemplate);
    },
    cancelEditName() {
      setTimeout(() => {
        this.isEditName = false;
      }, 200);
    },
    editTemplateDesc() {
      this.isEditDesc = true;
      this.editTemplate = Object.assign({}, this.curTemplate);
    },
    cancelEditDesc() {
      setTimeout(() => {
        this.isEditDesc = false;
      }, 200);
    },
    updateTemplateLockStatus() {
      // 判断curTemplate name为空防止返回时清空当前数据解发switcher change事件
      if (this.isTemplateLocking || this.curTemplate.name === '') {
        return false;
      }

      if (this.templateLockStatus.isLocked) {
        this.unlockTemplateset();
      } else {
        this.lockTemplateset();
      }
    },
    async lockTemplateset() {
      const { projectId } = this;
      const templateId = this.curTemplateId;
      this.isTemplateLocking = true;
      try {
        await this.$store.dispatch('k8sTemplate/lockTemplateset', { projectId, templateId });
        this.$bkMessage({
          theme: 'success',
          message: this.$t('deploy.templateset.lockSuccess'),
        });
        this.reloadTemplateLockStatus();
      } catch (res) {
        this.$bkMessage({
          theme: 'error',
          message: res.message,
          hasCloseIcon: true,
          delay: '3000',
        });
      } finally {
        setTimeout(() => {
          this.isTemplateLocking = false;
        }, 1000);
      }
    },
    async unlockTemplateset() {
      const { projectId } = this;
      const templateId = this.curTemplateId;
      // 不是当前加锁者不能解锁
      if (!this.templateLockStatus.isCurLocker) {
        this.$bkMessage({
          theme: 'error',
          message: this.$t('deploy.templateset.msg.multiOperator', this.templateLockStatus),
        });
        return false;
      }
      this.isTemplateLocking = true;
      try {
        await this.$store.dispatch('k8sTemplate/unlockTemplateset', { projectId, templateId });
        this.$bkMessage({
          theme: 'success',
          message: this.$t('deploy.templateset.unlockSuccess'),
        });
        this.reloadTemplateLockStatus();
      } catch (res) {
        this.$bkMessage({
          theme: 'error',
          message: res.message,
          hasCloseIcon: true,
          delay: '3000',
        });
      } finally {
        setTimeout(() => {
          this.isTemplateLocking = false;
        }, 1000);
      }
    },
    saveTemplateDraft() {
      const { projectId } = this;
      const templateId = this.curTemplateId;

      const deployments = [];
      const services = [];
      const configmaps = [];
      const secrets = [];
      const daemonsets = [];
      const jobs = [];
      const statefulsets = [];
      const ingresss = [];
      const HPAs = [];

      this.deployments.forEach(async (deployment) => {
        const result = await this.formatDeploymentData(deployment);
        deployments.push(result);
      });

      this.services.forEach(async (service) => {
        const result = await this.formatServiceData(service);
        services.push(result);
      });

      this.configmaps.forEach(async (configmap) => {
        const result = await this.formatConfigmapData(configmap);
        configmaps.push(result);
      });

      this.secrets.forEach(async (secret) => {
        const result = await this.formatSecretData(secret);
        secrets.push(result);
      });

      this.jobs.forEach(async (job) => {
        const result = await this.formatJobData(job);
        jobs.push(result);
      });

      this.daemonsets.forEach(async (daemonset) => {
        const result = await this.formatDaemonsetData(daemonset);
        daemonsets.push(result);
      });

      this.statefulsets.forEach(async (statefulset) => {
        const result = await this.formatStatefulsetData(statefulset);
        statefulsets.push(result);
      });

      this.ingresss.forEach(async (ingress) => {
        const result = await this.formatIngressData(ingress);
        ingresss.push(result);
      });

      this.HPAs.forEach(async (HPA) => {
        const result = await this.formatHPAData(HPA);
        HPAs.push(result);
      });

      const data = {
        draft: {
          K8sDeployment: deployments,
          K8sService: services,
          K8sConfigMap: configmaps,
          K8sSecret: secrets,
          K8sDaemonSet: daemonsets,
          K8sJob: jobs,
          K8sStatefulSet: statefulsets,
          K8sIngress: ingresss,
          K8sHPA: HPAs,
        },
      };

      // 如果没有模板（template_id）
      if (this.isNewTemplate) {
        data.template = {
          name: this.curTemplate.name,
          desc: this.curTemplate.desc,
        };
        data.real_version_id = 0;
      } else {
        data.real_version_id = this.curVersion;
      }

      if (projectId) {
        this.$store.dispatch('k8sTemplate/updateTemplateDraft', { projectId, templateId, data }).then((res) => {
          this.$bkMessage({
            theme: 'success',
            message: this.$t('generic.msg.success.save'),
          });
          // 新创建则跳转
          if (this.isNewTemplate) {
            this.$router.push({
              name: this.$route.name,
              params: {
                projectId,
                templateId: res.data.template_id,
              },
            });
          }
        }, (res) => {
          this.$store.commit('k8sTemplate/updateIsTemplateSaving', false);
          this.$bkMessage({
            theme: 'error',
            message: res.message,
            hasCloseIcon: true,
            delay: '3000',
          });
        });
      }
    },
    showVersionPanel() {
      this.versionSidePanel.isShow = true;
      this.getVersionList();
    },
    async getVersionList() {
      const { projectId } = this;
      const templateId = this.curTemplateId;
      this.isVersionListLoading = true;
      if (templateId !== '0' && templateId !== 0) {
        await this.$store.dispatch('k8sTemplate/getVersionList', { projectId, templateId }).then((res) => {
          let versionList = [];
          if (res && res.data) {
            versionList = res.data;
            if (versionList) {
              versionList.forEach((item) => {
                if (item.show_version_id === Number(this.curShowVersionId) || item.show_version_id === this.curShowVersionId) {
                  this.versionMetadata = {
                    show_version_id: item.show_version_id,
                    name: item.name,
                    real_version_id: item.real_version_id,
                  };
                  this.curVersionNotes = item.comment;
                  this.displayVersionNotes = item.comment;
                }
              });
            }
          }

          this.isVersionListLoading = false;
          return versionList;
        }, (res) => {
          this.$bkMessage({
            theme: 'error',
            message: res.message,
            delay: '3000',
          });
        });
      } else {
        this.$store.commit('k8sTemplate/updateVersionList', []);
        this.isVersionListLoading = false;
        return [];
      }
    },
    getTemplateByVersion(versionId, isVersionRemove) {
      const { projectId } = this;
      const templateId = this.curTemplateId;
      this.$store.dispatch('k8sTemplate/getTemplateByVersion', { projectId, templateId, versionId }).then((res) => {
        const data = {
          deployments: res.data.K8sDeployment,
          services: res.data.K8sService,
          configmaps: res.data.K8sConfigmap,
          secrets: res.data.K8sSecret,
          daemonsets: res.data.K8sDaemonsets,
          jobs: res.data.K8sJobs,
          statefulsets: res.data.K8sStatefulsets,
          ingresss: res.data.K8sIngress,
          HPAs: res.data.K8sHPA,
        };
        this.$emit('switchVersion', data);
        // 如果不是操作删除版本，则可隐藏
        if (!isVersionRemove) {
          this.versionSidePanel.isShow = false;
        }
      });
    },

    // 生成最终实例化的配置文件
    filterForInstanceConf(appConf) {
      this.removeCustomKey(appConf); // 删除前端自定义字段
      this.removeEmptyValue(appConf); // 删除空数据
      this.formatKey(appConf); // 处理数据类型转换
    },

    // 将部分key类型转换
    formatKey(appConf, parent, parentKey) {
      const booleanKeys = [
        'hostNetwork',
      ];
      const arrayKeys = [
        'command',
        'args',
      ];
      for (const key in appConf) {
        if (key === 'image') {
          // 镜像替换变量
          appConf[key] = appConf[key].replace(DEVOPS_ARTIFACTORY_HOST, '{{SYS_JFROG_DOMAIN}}');
        } else if (arrayKeys.includes(key)) {
          // 转成数组
          appConf[key] = shlex.split(appConf[key]);
          console.log(appConf[key]);
        } else if (booleanKeys.includes(key)) {
          // 转成boolean
          appConf[key] = Boolean(appConf[key]);
        } else if (appConf.kind === 'Secret' && key === 'data') {
          this.baseEncode(appConf[key]);
        } else if (isObject(appConf[key])) {
          this.formatKey(appConf[key], appConf, key);
        }
      }
    },

    baseEncode(data) {
      for (const key in data) {
        data[key] = base64.encode(data[key]);
      }
    },

    // 删除key
    removeCustomKey(appConf) {
      const customKeys = [
        'webCache',
        'monitorLevel', // 重要级别
        'imageName', // 镜像名称
        'imageVersion', // 镜像版本
        'customLogLabel', // 日志采集
        'volumeClaimTemplates',
        'id',
      ];
      for (const key in appConf) {
        if (customKeys.includes(key)) {
          delete appConf[key];
        } else if (isObject(appConf[key])) {
          if (key === 'livenessProbe' || key === 'readinessProbe') {
            if ((!appConf[key].httpGet || !appConf[key].httpGet.path)
                                && (!appConf[key].tcpSocket || !appConf[key].tcpSocket.port)
                                && (!appConf[key].exec || !appConf[key].exec.command)) {
              delete appConf[key];
            }
          } else {
            this.removeCustomKey(appConf[key]);
          }
        }
      }
    },

    // 删除空值
    removeEmptyValue(appConf, parent, parentKey) {
      for (const key in appConf) {
        if (Array.isArray(appConf[key]) && !appConf[key].length) {
          delete appConf[key];
        } else if (isObject(appConf[key])) {
          if (JSON.stringify(appConf[key]) === '{}') {
            delete appConf[key];
          } else {
            this.removeEmptyValue(appConf[key], appConf, key);
          }
        } else if (appConf[key] === '') {
          delete appConf[key];
        }
      }
      if (parent && JSON.stringify(appConf) === '{}') {
        delete parent[parentKey];
      }
    },
    handleExportToYaml(versionData) {
      const versionId = versionData.show_version_id;
      const { name } = versionData;
      const realVersionId = versionData.real_version_id;
      const me = this;
      let msg = `转换YAML功能会将模板集版本【${name}】复制为一个新的YAML模板集，由平台完成模板配置的转换。使用中如有问题，请联系蓝鲸容器助手`;
      if (this.isEn) {
        msg = `The conversion YAML function will copy the template set version [${name}] into a new YAML template set, and the platform will complete the conversion of the template configuration. If you have any questions during use, please contact BCS`;
      }
      me.$bkInfo({
        title: me.$t('projects.operateAudit.note'),
        content: msg,
        confirmFn() {
          me.exportYamlTemplate(versionId, realVersionId);
        },
      });
    },
    async exportYamlTemplate(versionId, realVersionId) {
      try {
        const { projectId } = this;
        const templateId = this.curTemplateId;
        const res = await this.$store.dispatch('k8sTemplate/getTemplateSetDetail', { projectId, templateId, versionId });
        const serviceRes = await this.$store.dispatch('k8sTemplate/getServicesByVersion', { projectId, version: realVersionId });
        const resources = res.data;
        const curLinkServices = serviceRes.data;
        const data = {};
        if (resources.K8sDeployment && resources.K8sDeployment.length) {
          data.Deployment = resources.K8sDeployment.map((app) => {
            const appConfig = app.config;
            this.filterForInstanceConf(appConfig);
            return yamljs.dump(appConfig, { indent: 2 });
          });
        }

        if (resources.K8sService && resources.K8sService.length) {
          data.Service = resources.K8sService.map((app) => {
            const appConfig = app.config;
            this.filterForInstanceConf(appConfig);
            return yamljs.dump(appConfig, { indent: 2 });
          });
        }

        if (resources.K8sConfigMap && resources.K8sConfigMap.length) {
          data.ConfigMap = resources.K8sConfigMap.map((app) => {
            const appConfig = app.config;
            this.filterForInstanceConf(appConfig);
            return yamljs.dump(appConfig, { indent: 2 });
          });
        }

        if (resources.K8sSecret && resources.K8sSecret.length) {
          data.Secret = resources.K8sSecret.map((app) => {
            const appConfig = app.config;
            this.filterForInstanceConf(appConfig);
            return yamljs.dump(appConfig, { indent: 2 });
          });
        }

        if (resources.K8sDaemonSet && resources.K8sDaemonSet.length) {
          data.DaemonSet = resources.K8sDaemonSet.map((app) => {
            const appConfig = app.config;
            this.filterForInstanceConf(appConfig);
            return yamljs.dump(appConfig, { indent: 2 });
          });
        }

        if (resources.K8sJob && resources.K8sJob.length) {
          data.Job = resources.K8sJob.map((app) => {
            const appConfig = app.config;
            appConfig.apiVersion = 'batch/v1';
            this.filterForInstanceConf(appConfig);
            return yamljs.dump(appConfig, { indent: 2 });
          });
        }
        if (resources.K8sStatefulSet && resources.K8sStatefulSet.length) {
          data.StatefulSet = resources.K8sStatefulSet.map((app) => {
            const appConfig = app.config;
            const matchService = curLinkServices.find(item => item.service_tag === app.service_tag);
            if (matchService) {
              appConfig.spec.serviceName = matchService.service_name;
            }
            delete appConfig.spec.strategy;
            this.filterForInstanceConf(appConfig);
            return yamljs.dump(appConfig, { indent: 2 });
          });
        }

        if (resources.K8sIngress && resources.K8sIngress.length) {
          data.Ingress = resources.K8sIngress.map((app) => {
            const appConfig = app.config;
            this.filterForInstanceConf(appConfig);
            return yamljs.dump(appConfig, { indent: 2 });
          });
        }

        if (resources.K8sHPA && resources.K8sHPA.length) {
          data.HPA = resources.K8sHPA.map((app) => {
            const appConfig = app.config;
            this.filterForInstanceConf(appConfig);
            return yamljs.dump(appConfig, { indent: 2 });
          });
        }

        window.localStorage.cloneTemplateSet = JSON.stringify(data);
        this.$emit('exmportToYaml', data);
      } catch (res) {
        this.$bkMessage({
          theme: 'error',
          message: res.message,
          hasCloseIcon: true,
          delay: '3000',
        });
      }
    },
    async autoSaveResource(type) {
      switch (type) {
        case 'k8sTemplatesetDeployment':
          const { deployments } = this;
          // 对application资源数据检测
          for (const deployment of deployments) {
            if (deployment.isEdited) {
              const isValid = await this.checkDeploymentData(deployment);
              if (isValid) {
                const result = await this.saveApplication(deployment, 'deployment');
                if (result) {
                  this.$store.commit('k8sTemplate/updateBindVersion', true);
                  if (result.template_id) {
                    this.newTemplateId = result.template_id;
                    this.$store.commit('k8sTemplate/updateCurTemplateId', result.template_id);
                  }
                }
                return true;
              }
            }
          }
          break;

        case 'k8sTemplatesetDaemonset':
          const { daemonsets } = this;
          // 对application资源数据检测
          for (const daemonset of daemonsets) {
            if (daemonset.isEdited) {
              const isValid = await this.checkDaemonsetData(daemonset);
              if (isValid) {
                const result = await this.saveApplication(daemonset, 'daemonset');
                if (result) {
                  this.$store.commit('k8sTemplate/updateBindVersion', true);
                  if (result.template_id) {
                    this.newTemplateId = result.template_id;
                    this.$store.commit('k8sTemplate/updateCurTemplateId', result.template_id);
                  }
                }
                return true;
              }
            }
          }
          break;

        case 'k8sTemplatesetJob':
          const { jobs } = this;
          // 对application资源数据检测
          for (const job of jobs) {
            if (job.isEdited) {
              const isValid = await this.checkJobData(job);
              if (isValid) {
                const result = await this.saveApplication(job, 'job');
                if (result) {
                  this.$store.commit('k8sTemplate/updateBindVersion', true);
                  if (result.template_id) {
                    this.newTemplateId = result.template_id;
                    this.$store.commit('k8sTemplate/updateCurTemplateId', result.template_id);
                  }
                }
                return true;
              }
            }
          }
          break;

        case 'k8sTemplatesetStatefulset':
          const { statefulsets } = this;
          // 对application资源数据检测
          for (const statefulset of statefulsets) {
            if (statefulset.isEdited) {
              const isValid = await this.checkStatefulsetData(statefulset);
              if (isValid) {
                const result = await this.saveApplication(statefulset, 'statefulset');
                if (result) {
                  this.$store.commit('k8sTemplate/updateBindVersion', true);
                  if (result.template_id) {
                    this.newTemplateId = result.template_id;
                    this.$store.commit('k8sTemplate/updateCurTemplateId', result.template_id);
                  }
                }
                return true;
              }
            }
          }
          break;

        case 'k8sTemplatesetService':
          const { services } = this;
          // 对service资源数据检测
          for (const service of services) {
            const isValid = await this.checkServiceData(service);
            if (isValid) {
              const result = await this.saveService(service);
              this.$store.state.k8sTemplate.services.forEach((service) => {
                if (service.id === result.id) {
                  service.cache.deploy_tag_list = service.deploy_tag_list;
                }
              });
              if (result) {
                this.$store.commit('k8sTemplate/updateBindVersion', true);
                if (result.template_id) {
                  this.newTemplateId = result.template_id;
                  this.$store.commit('k8sTemplate/updateCurTemplateId', result.template_id);
                }
              }
              return true;
            }
          }
          break;

        case 'k8sTemplatesetConfigmap':
          const { configmaps } = this;
          // 对configmap资源数据检测
          for (const configmap of configmaps) {
            if (configmap.isEdited) {
              const isValid = await this.checkConfigmapData(configmap);
              if (isValid) {
                const result = await this.saveConfigmap(configmap);
                if (result) {
                  this.$store.commit('k8sTemplate/updateBindVersion', true);
                  if (result.template_id) {
                    this.newTemplateId = result.template_id;
                    this.$store.commit('k8sTemplate/updateCurTemplateId', result.template_id);
                  }
                }
                return true;
              }
            }
          }
          break;

        case 'k8sTemplatesetSecret':
          const { secrets } = this;
          // 对secret资源数据检测
          for (const secret of secrets) {
            if (secret.isEdited) {
              const isValid = await this.checkSecretData(secret);
              if (isValid) {
                const result = await this.saveSecret(secret);
                if (result) {
                  this.$store.commit('k8sTemplate/updateBindVersion', true);
                  if (result.template_id) {
                    this.newTemplateId = result.template_id;
                    this.$store.commit('k8sTemplate/updateCurTemplateId', result.template_id);
                  }
                }
                return true;
              }
            }
          }
          break;

        case 'k8sTemplatesetIngress':
          const { ingresss } = this;
          // 对ingress资源数据检测
          for (const ingress of ingresss) {
            if (ingress.isEdited) {
              const isValid = await this.checkIngressData(ingress);
              if (isValid) {
                const result = await this.saveIngress(ingress);
                if (result) {
                  this.$store.commit('k8sTemplate/updateBindVersion', true);
                  if (result.template_id) {
                    this.newTemplateId = result.template_id;
                    this.$store.commit('k8sTemplate/updateCurTemplateId', result.template_id);
                  }
                }
                return true;
              }
            }
          }
          break;

        case 'k8sTemplatesetHPA':
          const { HPAs } = this;
          // 对HPA资源数据检测
          for (const HPA of HPAs) {
            if (HPA.isEdited) {
              const isValid = await this.checkHPAData(HPA);
              if (isValid) {
                const result = await this.saveHPA(HPA);
                if (result) {
                  this.$store.commit('k8sTemplate/updateBindVersion', true);
                  if (result.template_id) {
                    this.newTemplateId = result.template_id;
                    this.$store.commit('k8sTemplate/updateCurTemplateId', result.template_id);
                  }
                }
                return true;
              }
            }
          }
          break;
      }
      return true;
    },
    async saveTemplate(event) {
      const { projectId } = this;
      const templateId = this.curTemplateId;
      const data = this.editTemplate;

      // 用户填空数据，用原数据
      if (!data.name) {
        data.name = this.curTemplate.name;
      }
      if (!data.desc) {
        data.desc = this.curTemplate.desc;
      }
      // 没有修改，不处理
      if (data.name === this.curTemplate.name && data.desc === this.curTemplate.desc) {
        this.isEditName = false;
        this.isEditDesc = false;
        return true;
      }

      if (templateId && String(templateId) !== '0') {
        try {
          await this.$store.dispatch('k8sTemplate/updateTemplate', { projectId, templateId, data });
          // const params = res.data
          // this.curTemplate = params
          this.$store.commit('k8sTemplate/updateCurTemplate', data);
          this.isEditName = false;
          this.isEditDesc = false;
          if (event) {
            this.$bkMessage({
              theme: 'success',
              message: this.$t('deploy.templateset.templateSetInfoSaved'),
            });
          }
        } catch (res) {
          this.$store.commit('k8sTemplate/updateIsTemplateSaving', false);
          this.$bkMessage({
            theme: 'error',
            message: res.message,
            hasCloseIcon: true,
            delay: '3000',
          });
        }
      } else {
        this.curTemplate = data;
        this.$store.commit('k8sTemplate/updateCurTemplate', data);
        this.isEditName = false;
        this.isEditDesc = false;
      }
      return true;
    },
    reloadTemplateLockStatus() {
      const templateId = this.curTemplateId;
      const { projectId } = this;
      this.isTemplateLocking = true;
      this.$store.dispatch('k8sTemplate/getTemplateById', { projectId, templateId }).then((res) => {
        const { data } = res;
        this.$store.commit('k8sTemplate/updateCurTemplate', data);
        this.webAnnotations = res.web_annotations;
      })
        .finally((res) => {
          setTimeout(() => {
            this.isTemplateLocking = false;
          }, 1000);
        });
    },
    reloadTemplateset() {
      this.$store.commit('k8sTemplate/clearCurTemplateData');
      this.$parent.$parent.reloadTemplateset();
    },
    initTemplate(callback) {
      if (this.curTemplate.id) {
        const data = {
          latest_version_id: this.curTemplate.latest_version_id,
          deployments: this.deployments,
          services: this.services,
          configmaps: this.configmaps,
          secrets: this.secrets,
          daemonsets: this.daemonsets,
          jobs: this.jobs,
          statefulsets: this.statefulsets,
          ingress: this.ingresss,
          HPAs: this.HPAs,
        };
        this.isTemplateLoading = false;
        this.isFromNewTemplate = true;
        callback(data);
      } else if (this.curTemplateId === 0 || this.curTemplateId === '0') {
        if (!this.curTemplate.name) {
          const templateParams = {
            id: 0,
            name: this.$t('deploy.templateset.templateSetPrefix') + (+new Date()),
            desc: this.$t('deploy.templateset.templateSetDescription'),
            is_locked: false,
            locker: '',
          };
          this.$store.commit('k8sTemplate/updateCurTemplate', templateParams);
        }

        this.isTemplateLoading = false;
        this.initResources(callback);
      } else {
        const templateId = this.curTemplateId;
        const { projectId } = this;
        this.isTemplateLoading = true;
        this.$store.dispatch('k8sTemplate/getTemplateById', { projectId, templateId }).then((res) => {
          const { data } = res;
          this.$store.commit('k8sTemplate/updateCurTemplate', data);
          this.webAnnotations = res.web_annotations;
          this.initResources(callback);
        }, (res) => {
          const { data } = res;
          if (data) {
            if (!data.code || data.code === 400 || data.code === 404) {
              this.exceptionCode = {
                code: '404',
                msg: this.$t('deploy.templateset.templateSetNotFound'),
              };
            } else if (data.code === 403) {
              this.exceptionCode = {
                code: '403',
                msg: this.$t('generic.msg.warning.403'),
              };
            } else {
              this.exceptionCode = {
                code: '403',
                msg: this.$t('generic.status.error'),
              };
            }
          } else {
            this.exceptionCode = {
              code: '403',
              msg: this.$t('generic.status.error'),
            };
          }
          this.$emit('exception', this.exceptionCode);
        })
          .finally((res) => {
            this.isTemplateLoading = false;
          });
      }
    },
    selectVersion(index, versionData) {
      this.versionMetadata.show_version_id = versionData.show_version_id;
      this.versionMetadata.name = versionData.name;
      this.versionMetadata.real_version_id = this.curVersion;
      this.curVersionNotes = versionData.comment;
    },
    initResources(callback) {
      const templateId = this.curTemplateId;
      const { projectId } = this;
      const version = this.curTemplate.latest_show_version_id;
      if (version) {
        this.$store.dispatch('k8sTemplate/getTemplateResource', { projectId, templateId, version }).then((res) => {
          const { data } = res;
          if (data.version) {
            this.$store.commit('k8sTemplate/updateCurVersion', data.version);
          }
          this.$store.commit('k8sTemplate/updateResources', data);
          const resources = {
            latest_version_id: this.curTemplate.latest_version_id,
            deployments: data.K8sDeployment,
            services: data.K8sService,
            configmaps: data.K8sConfigMap,
            secrets: data.K8sSecret,
            daemonsets: data.K8sDaemonSet,
            jobs: data.K8sJob,
            statefulsets: data.K8sStatefulSet,
            ingresss: data.K8sIngress,
            HPAs: data.K8sHPA,
          };
          callback(resources);
        }, (res) => {
          this.$store.commit('k8sTemplate/updateIsTemplateSaving', false);
          this.$bkMessage({
            theme: 'error',
            message: res.message,
            hasCloseIcon: true,
            delay: '10000',
          });
        });
      } else {
        const resources = {
          latest_version_id: null,
          deployments: this.deployments,
          services: this.services,
          configmaps: this.configmaps,
          secrets: this.secrets,
          daemonsets: this.daemonsets,
          jobs: this.jobs,
          statefulsets: this.statefulsets,
          ingresss: this.ingresss,
          HPAs: this.HPAs,
        };
        callback(resources);
      }
    },
    updateLocalApplicationData(application, data) {
      application.isEdited = false;
      const appId = application.id;
      if (data.id) {
        application.id = data.id;
      }
      if (data.version) {
        this.$store.commit('k8sTemplate/updateCurVersion', data.version);
      }
      this.$store.commit('k8sTemplate/updateApplicationById', { application, appId });
    },
    async checkApplicationData(application, type, callback) {
      const appName = application.config.metadata.name;
      const instance = application.config.spec.replicas;
      const nameReg1 = /^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$/;
      // const nameReg2 = /^[a-zA-Z]{1}[a-zA-Z0-9-_]{0,29}$/
      const pathReg = /\/((?!\.)[\w\d\-./~]+)*/;
      const portNameReg = /^[a-z]{1}[a-z0-9-]{0,255}$/;
      const volumeNameReg = /^[a-zA-Z]{1}[a-zA-Z0-9-]{0,253}$/;
      const chineseReg = /[\u4e00-\u9fa5]+/;
      const labelKeyReg = /^([A-Za-z0-9][-A-Za-z0-9_./]*)?[A-Za-z0-9]$/;
      const envKeyReg = /^([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]$/;
      const varReg = /\{\{([^\{\}]+)?\}\}/g;
      const ipReg = /^((2(5[0-5]|[0-4]\d))|[0-1]?\d{1,2})(\.((2(5[0-5]|[0-4]\d))|[0-1]?\d{1,2})){3}$/;
      let megPrefix = `"${appName}"${this.$t('deploy.templateset.in')}`;

      if (appName === '') {
        megPrefix += `${this.$t('generic.label.name')}：`;
        this.$bkMessage({
          theme: 'error',
          message: megPrefix + this.$t('deploy.templateset.validate.name'),
        });
        return false;
      }
      if (!nameReg1.test(appName)) {
        megPrefix += `${this.$t('generic.label.name')}：`;
        this.$bkMessage({
          theme: 'error',
          message: megPrefix + this.$t('deploy.templateset.applicationNameError'),
          delay: 8000,
        });
        return false;
      }

      if (instance === '') {
        megPrefix += `${this.$t('dashboard.workload.label.scaleNum')}：`;
        this.$bkMessage({
          theme: 'error',
          message: megPrefix + this.$t('deploy.templateset.enterInstanceQuantity'),
        });
        return false;
      }

      // 标签
      if (application.config.webCache.labelListCache) {
        const cacheLabels = application.config.webCache.labelListCache;
        for (const label of cacheLabels) {
          // const key11 = label.key.replace(varReg, 'key')
          if (label.key && !labelKeyReg.test(label.key.replace(varReg, 'key'))) {
            megPrefix += `${this.$t('k8s.label')}：`;
            this.$bkMessage({
              theme: 'error',
              delay: 8000,
              message: megPrefix + this.$t('deploy.templateset.startEndAlphaNumeric'),
            });
            return false;
          }
        }
      }
      const selector = application.config.webCache.labelListCache.filter(item => item.isSelector && item.key);

      // 除去job，其它要至少一个选择器
      if (!selector.length && type !== 'job') {
        megPrefix += `${this.$t('k8s.label')}：`;
        this.$bkMessage({
          theme: 'error',
          message: megPrefix + this.$t('deploy.templateset.mustAddSelector'),
        });
        return false;
      }

      // 注解
      if (application.config.webCache.remarkListCache) {
        const cacheLabels = application.config.webCache.remarkListCache;
        for (const label of cacheLabels) {
          if (label.key && !labelKeyReg.test(label.key.replace(varReg, 'key'))) {
            megPrefix += `${this.$t('k8s.annotation')}：`;
            this.$bkMessage({
              theme: 'error',
              delay: 8000,
              message: megPrefix + this.$t('deploy.templateset.startEndAlphaNumeric'),
            });
            return false;
          }
        }
      }

      // statefulset 关联service
      if (application.hasOwnProperty('service_tag') && !application.service_tag) {
        // application.service_tag = 'a'
        // megPrefix += `关联Service：`
        // this.$bkMessage({
        //     theme: 'error',
        //     message: megPrefix + '请选择要关联的service!'
        // })
        // return false
      }

      if (application.config.spec.hasOwnProperty('volumeClaimTemplates')) {
        for (const item of application.config.spec.volumeClaimTemplates) {
          if (item.metadata.name || item.spec.storageClassName || item.spec.accessModes.length) {
            if (!item.metadata.name) {
              megPrefix += `${this.$t('deploy.templateset.volumeTemplate')}：`;
              this.$bkMessage({
                theme: 'error',
                message: megPrefix + this.$t('deploy.templateset.enterMountName'),
              });
              return false;
            }
            if (!/^([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]$/.test(item.metadata.name)) {
              megPrefix += `${this.$t('deploy.templateset.volumeTemplate')}：`;
              this.$bkMessage({
                theme: 'error',
                message: megPrefix + this.$t('deploy.templateset.startEndAlphaNumericMountName'),
                delay: 8000,
              });
              return false;
            }
            if (!item.spec.storageClassName) {
              megPrefix += `${this.$t('deploy.templateset.volumeTemplate')}：`;
              this.$bkMessage({
                theme: 'error',
                message: megPrefix + this.$t('deploy.templateset.pleaseSelectStorageClassName'),
              });
              return false;
            }
            if (!item.spec.resources.requests.storage) {
              megPrefix += `${this.$t('deploy.templateset.volumeTemplate')}：`;
              this.$bkMessage({
                theme: 'error',
                message: megPrefix + this.$t('deploy.templateset.setSize'),
              });
              return false;
            }
            if (!item.spec.accessModes.length) {
              megPrefix += `${this.$t('deploy.templateset.volumeTemplate')}：`;
              this.$bkMessage({
                theme: 'error',
                message: megPrefix + this.$t('deploy.templateset.selectAccessMode'),
              });
              return false;
            }
          }
        }
      }

      if (application.config.webCache.volumes.length) {
        const { volumes } = application.config.webCache;
        for (const volume of volumes) {
          if (volume.type === 'emptyDir' && volume.name) {
            volume.source = '{}';
          }
          if (volume.name || volume.source) {
            if (!volume.name) {
              megPrefix += `${this.$t('deploy.templateset.volume')}：`;
              this.$bkMessage({
                theme: 'error',
                message: megPrefix + this.$t('deploy.templateset.mountNameNotEmpty'),
                delay: 8000,
              });
              return false;
            }
            if (!volume.source) {
              megPrefix += `${this.$t('deploy.templateset.volume')}：`;
              this.$bkMessage({
                theme: 'error',
                message: megPrefix + this.$t('deploy.templateset.mountSourceNotEmpty'),
                delay: 8000,
              });
              return false;
            }
            if (volume.name && !volumeNameReg.test(volume.name.replace(varReg, 'name'))) {
              megPrefix += `${this.$t('deploy.templateset.volume')}：`;
              this.$bkMessage({
                theme: 'error',
                message: megPrefix + this.$t('deploy.templateset.mountNameCriteria'),
                delay: 8000,
              });
              return false;
            }
          }
        }
      }

      if (application.config.webCache && application.config.webCache.metricIdList) {
        const result = application.config.webCache.metricIdList.filter(item => this.metricList.includes(item));
        application.config.webCache.metricIdList = result;
      }
      if (application.config.webCache.hostAliasesCache && application.config.webCache.hostAliasesCache.length) {
        const hostAliases = application.config.webCache.hostAliasesCache;
        for (const hostAlias of hostAliases) {
          if (hostAlias.ip || hostAlias.hostnames) {
            if (!hostAlias.ip) {
              megPrefix += 'hostAliases：';
              this.$bkMessage({
                theme: 'error',
                message: megPrefix + this.$t('deploy.templateset.ipNotEmpty'),
                delay: 8000,
              });
              return false;
            }
            if (hostAlias.ip && !ipReg.test(hostAlias.ip.replace(varReg, '255.255.255.255'))) {
              megPrefix += 'hostAliases：';
              this.$bkMessage({
                theme: 'error',
                message: megPrefix + this.$t('deploy.templateset.enterValidIPAddress'),
                delay: 8000,
              });
              return false;
            }
            if (!hostAlias.hostnames) {
              megPrefix += 'hostAliases：';
              this.$bkMessage({
                theme: 'error',
                message: megPrefix + this.$t('deploy.templateset.hostNamesNotEmpty'),
                delay: 8000,
              });
              return false;
            }
          }
        }
      }

      const containers = application.config.spec.template.spec.allContainers;
      let hasContainer = false;

      if (application.config.webCache.isUserConstraint && application.config.webCache.affinityYamlCache) {
        try {
          const yamlCode = application.config.webCache.affinityYamlCache;
          const json = yamljs.load(yamlCode);
          if (!json || Object.prototype.toString.call(json) === '[object String]') {
            megPrefix += `${this.$t('deploy.templateset.schedulingConstraints')}：`;
            this.$bkMessage({
              theme: 'error',
              message: megPrefix + this.$t('deploy.templateset.affinityConstraintInvalidFormat'),
              delay: 5000,
            });
            return false;
          }
        } catch (err) {
          megPrefix += `${this.$t('deploy.templateset.schedulingConstraints')}：`;
          this.$bkMessage({
            theme: 'error',
            message: megPrefix + this.$t('deploy.templateset.affinityConstraintInvalidFormat'),
            delay: 5000,
          });
          return false;
        }
      }

      for (const container of containers) {
        // 检查container name
        if (!container.name) {
          megPrefix += `${this.$t('dashboard.workload.container.name')}：`;
          this.$bkMessage({
            theme: 'error',
            message: megPrefix + this.$t('deploy.templateset.nameNotEmpty'),
          });
          return false;
        }

        if (!nameReg1.test(container.name)) {
          megPrefix += `${this.$t('dashboard.workload.container.name')}：`;
          this.$bkMessage({
            theme: 'error',
            message: megPrefix + this.$t('deploy.templateset.nameError'),
            delay: 8000,
          });
          return false;
        }

        if (container.webCache.containerType === 'container') {
          hasContainer = true;
        }
        // 检查container镜像设置
        if (!container.image) {
          this.$bkMessage({
            theme: 'error',
            delay: 5000,
            message: megPrefix + this.$t('deploy.templateset.configImageAndVersion', container),
          });
          return false;
        }

        // 端口映射检查
        const portNameCache = {};
        for (const item of container.ports) {
          if (item.name || item.containerPort) {
            if (!item.name) {
              this.$bkMessage({
                theme: 'error',
                delay: 5000,
                message: megPrefix + this.$t('deploy.templateset.portMappingNameNotEmpty', container),
              });
              return false;
            }
            if (!portNameReg.test(item.name.replace(varReg, 'name'))) {
              this.$bkMessage({
                theme: 'error',
                message: megPrefix + this.$t('deploy.templateset.portMappingNameError', container),
                delay: 8000,
              });
              return false;
            }
            if (portNameCache[item.name]) {
              this.$bkMessage({
                theme: 'error',
                message: megPrefix + this.$t('deploy.templateset.portMappingNameNotDuplicate', container),
                delay: 8000,
              });
              return false;
            }
            portNameCache[item.name] = true;


            if (!item.containerPort) {
              this.$bkMessage({
                theme: 'error',
                delay: 5000,
                message: megPrefix + this.$t('deploy.templateset.containerPortNotEmpty', container),
              });
              return false;
            }
            if (parseInt(item.containerPort) < 1 || parseInt(item.containerPort) > 65535) {
              this.$bkMessage({
                theme: 'error',
                delay: 5000,
                message: megPrefix + this.$t('deploy.templateset.containerPortRange', container),
              });
              return false;
            }
          }
        }

        // 命令
        if (container.command && chineseReg.test(container.command)) {
          this.$bkMessage({
            theme: 'error',
            delay: 5000,
            message: megPrefix + this.$t('deploy.templateset.commandNoChinese', container),
          });
          return false;
        }

        if (container.args && chineseReg.test(container.args)) {
          this.$bkMessage({
            theme: 'error',
            delay: 5000,
            message: megPrefix + this.$t('deploy.templateset.commandParamsNoChinese', container),
          });
          return false;
        }

        // 检查container volumes
        if (container.volumeMounts.length) {
          for (const item of container.volumeMounts) {
            if (item.name || item.mountPath) {
              if (!item.name) {
                this.$bkMessage({
                  theme: 'error',
                  delay: 5000,
                  message: megPrefix + this.$t('deploy.templateset.mountVolumeConfigNotEmpty', container),
                });
                return false;
              }
              if (!item.mountPath) {
                this.$bkMessage({
                  theme: 'error',
                  message: megPrefix + this.$t('deploy.templateset.containerDirectoryNotEmpty', container),
                  delay: 5000,
                });
                return false;
              }
              // const mountePathReg = /^[\/\-a-z0-9]*$/

              // if (!mountePathReg.test(item.mountPath.replace(varReg, 'path'))) {
              //     this.$bkMessage({
              //         theme: 'error',
              //         delay: 5000,
              //         message: megPrefix + `容器"${container.name}"的挂载卷配置：容器目录不正确，只能包含小写字母、数字、斜线（/）、横线（-）！`
              //     })
              //     return false
              // }

              // if (!mountePathReg.test(item.subPath.replace(varReg, 'path'))) {
              //     this.$bkMessage({
              //         theme: 'error',
              //         delay: 5000,
              //         message: megPrefix + `容器"${container.name}"的挂载卷配置：子目录不正确，只能包含小写字母、数字、斜线（/）、横线（-）！`
              //     })
              //     return false
              // }
            }
          }
        }

        // 环境变量检查
        const envList = container.webCache.env_list;
        for (const env of envList) {
          if (env.key || env.value) {
            if (['valueFrom', 'custom', 'configmapKey', 'secretKey'].includes(env.type)) {
              if (!env.key) {
                this.$bkMessage({
                  theme: 'error',
                  message: megPrefix + this.$t('deploy.templateset.envVarKeyNotEmpty', container),
                  delay: 5000,
                });
                return false;
              }
              if (!envKeyReg.test(env.key.replace(varReg, 'key'))) {
                this.$bkMessage({
                  theme: 'error',
                  message: megPrefix + this.$t('deploy.templateset.startEndAlphaNumericEnvVar', container),
                  delay: 8000,
                });
                return false;
              }
            } else {
              env.key = '';
            }

            if (env.type !== 'custom' && !env.value) {
              this.$bkMessage({
                theme: 'error',
                message: megPrefix + this.$t('deploy.templateset.valueNotEmpty', container),
                delay: 5000,
              });
              return false;
            }
          }
        }

        // 资源限制
        const { resources } = container;
        if (resources.limits.cpu && resources.requests.cpu && !varReg.test(resources.limits.cpu) && !varReg.test(resources.requests.cpu) && (resources.requests.cpu > resources.limits.cpu)) {
          this.$bkMessage({
            theme: 'error',
            message: megPrefix + this.$t('deploy.templateset.cpuLimitMinLessThanOrEqualMax', container),
            delay: 7000,
          });
          return false;
        }
        if (resources.limits.memory && resources.requests.memory && !varReg.test(resources.limits.memory) && !varReg.test(resources.requests.memory) && (resources.requests.memory > resources.limits.memory)) {
          this.$bkMessage({
            theme: 'error',
            message: megPrefix + this.$t('deploy.templateset.memoryLimitMinLessThanOrEqualMax', container),
            delay: 7000,
          });
          return false;
        }

        if (container.webCache.logListCache.length) {
          for (const log of container.webCache.logListCache) {
            log.value = log.value.trim();
            if (log.value && !pathReg.test(log.value.replace(varReg, '/path'))) {
              this.$bkMessage({
                theme: 'error',
                delay: 5000,
                message: megPrefix + this.$t('deploy.templateset.nonStandardLogInvalidPath', container),
              });
              return false;
            }
          }
        }

        // 生命周期
        if (container.lifecycle.preStop.exec.command && chineseReg.test(container.lifecycle.preStop.exec.command)) {
          this.$bkMessage({
            theme: 'error',
            delay: 5000,
            message: megPrefix + this.$t('deploy.templateset.preStopNoChinese', container),
          });
          return false;
        }

        if (container.lifecycle.postStart.exec.command && chineseReg.test(container.lifecycle.postStart.exec.command)) {
          this.$bkMessage({
            theme: 'error',
            delay: 5000,
            message: megPrefix + this.$t('deploy.templateset.postStartNoChinese', container),
          });
          return false;
        }
      }
      if (!hasContainer) {
        megPrefix += this.$t('dashboard.workload.container.title');
        this.$bkMessage({
          theme: 'error',
          message: megPrefix + this.$t('deploy.templateset.mustAddAtLeastOneContainer'),
          delay: 5000,
        });
        return false;
      }
      return true;
    },
    async formatApplicationData(application) {
      const params = JSON.parse(JSON.stringify(application));
      params.template = {
        name: this.curTemplate.name,
        desc: this.curTemplate.desc,
      };
      delete params.isEdited;
      // 键值转换
      const { webCache } = application.config;
      if (webCache && webCache.remarkListCache) {
        const remarkKeyList = this.tranListToObject(webCache.remarkListCache);
        params.config.spec.template.metadata.annotations = remarkKeyList;
      }

      if (webCache && webCache.logLabelListCache) {
        const logLabelKeyList = this.tranListToObject(webCache.logLabelListCache);
        params.config.customLogLabel = logLabelKeyList;
      }

      if (webCache && webCache.nodeSelectorList) {
        const selectorKeyList = this.tranListToObject(webCache.nodeSelectorList);
        params.config.spec.template.spec.nodeSelector = selectorKeyList;
      }

      if (webCache && webCache.labelListCache) {
        const labelKeyList = this.tranListToObject(webCache.labelListCache);
        params.config.spec.template.metadata.labels = labelKeyList;
        params.config.spec.selector.matchLabels = {};
        webCache.labelListCache.forEach((item) => {
          if (item.isSelector && item.key && item.value) {
            params.config.spec.selector.matchLabels[item.key] = item.value;
          }
        });
      }
      if (webCache && webCache.hostAliasesCache && webCache.hostAliasesCache.length) {
        params.config.spec.template.spec.hostAliases = [];
        webCache.hostAliasesCache.forEach((item) => {
          params.config.spec.template.spec.hostAliases.push({
            ip: item.ip,
            hostnames: item.hostnames.replace(/ /g, '').split(';'),
          });
        });
      } else {
        delete params.config.spec.template.spec.hostAliases;
      }

      if (webCache && webCache.volumes) {
        const cacheColumes = webCache.volumes;
        const volumes = [];
        cacheColumes.forEach((volume) => {
          if ((volume.name && volume.source) || (volume.name && volume.type === 'emptyDir')) {
            switch (volume.type) {
              case 'emptyDir':
                volumes.push({
                  name: volume.name,
                  emptyDir: {},
                });
                break;

              case 'persistentVolumeClaim':
                volumes.push({
                  name: volume.name,
                  persistentVolumeClaim: {
                    claimName: volume.source,
                  },
                });
                break;

              case 'hostPath':
                const item = {
                  name: volume.name,
                  hostPath: {
                    path: volume.source,
                  },
                };
                if (volume.hostType) {
                  item.hostPath.type = volume.hostType;
                }
                volumes.push(item);
                break;

              case 'configMap':
                // 针对已经存的configmap处理
                let volumeSource = volume.source;
                if (volume.is_exist) {
                  volumeSource = volume.source.split(':')[0];
                }
                volumes.push({
                  name: volume.name,
                  configMap: {
                    name: volumeSource,
                  },
                });
                break;

              case 'secret':
                volumes.push({
                  name: volume.name,
                  secret: {
                    secretName: volume.source,
                  },
                });
                break;
              case 'emptyDir(Memory)':
                volumes.push({
                  name: volume.name,
                  emptyDir: {
                    medium: 'Memory',
                    sizeLimit: `${volume.source}Gi`,
                  },
                });
                break;
            }
          }
        });

        params.config.spec.template.spec.volumes = volumes;
      }

      if (params.config.webCache.isUserConstraint) {
        try {
          const yamlCode = params.config.webCache.affinityYamlCache || '';
          params.config.webCache.affinityYaml = yamlCode;
          const json = yamljs.load(yamlCode);
          if (json) {
            params.config.spec.template.spec.affinity = json;
          } else {
            params.config.spec.template.spec.affinity = {};
          }
        } catch (err) {
          // error
        }
      } else {
        params.config.spec.template.spec.affinity = {};
      }

      // 镜像凭证处理
      if (params.config.spec.template.spec.imagePullSecrets) {
        const imageSecrets = params.config.spec.template.spec.imagePullSecrets;
        const validImageSecrets = imageSecrets.filter(item => item.name);
        if (validImageSecrets.length) {
          params.config.spec.template.spec.imagePullSecrets = validImageSecrets;
        } else {
          delete params.config.spec.template.spec.imagePullSecrets;
        }
      }

      // 服务帐户处理
      if (!params.config.spec.template.spec.serviceAccountName) {
        delete params.config.spec.template.spec.serviceAccountName;
      }

      // 转换命令参数和环境变量
      const { allContainers } = params.config.spec.template.spec;
      params.config.spec.template.spec.containers = [];
      params.config.spec.template.spec.initContainers = [];

      allContainers.forEach((container) => {
        // 端口
        const { ports } = container;
        const validatePorts = [];
        ports.forEach((item) => {
          if (item.containerPort) {
            validatePorts.push({
              id: item.id,
              containerPort: item.containerPort,
              name: item.name,
              protocol: item.protocol,
            });
          }
        });
        container.ports = validatePorts;

        // volumes
        const volumes = container.volumeMounts;
        let validateVolumes = [];
        validateVolumes = volumes.filter(item => item.mountPath && item.name);

        // 用户不填subPath，删除此key
        validateVolumes.forEach((item) => {
          if (!item.subPath) {
            delete item.subPath;
          }
        });
        container.volumeMounts = validateVolumes;

        // logpath
        const paths = [];
        const logList = container.webCache.logListCache;
        logList.forEach((item) => {
          if (item.value) {
            paths.push(item.value);
          }
        });
        container.logPathList = paths;

        if (container.webCache.containerType === 'initContainer') {
          // initContainer类型去除健康检查、就绪检查、生命周期
          delete container.livenessProbe;
          delete container.readinessProbe;
          delete container.lifecycle;
          params.config.spec.template.spec.initContainers.push(container);
        } else {
          params.config.spec.template.spec.containers.push(container);
        }
      });

      delete params.config.spec.template.spec.allContainers;
      delete params.cache;
      return params;
    },
    checkLinkData(application) {
      // const appName = application.config.metadata.name
      // let megPrefix = `"${appName}"中`
      // const containers = application.config.spec.template.spec.allContainers
      return true;
    },
    async checkDeploymentData(application) {
      const result = await this.checkApplicationData(application, 'deployment');
      return result;
    },
    async checkDaemonsetData(application) {
      const result = await this.checkApplicationData(application, 'daemonset');
      return result;
    },
    async checkJobData(application) {
      const result = await this.checkApplicationData(application, 'job');
      return result;
    },
    async checkStatefulsetData(application) {
      const result = await this.checkApplicationData(application, 'statefulset');
      return result;
    },
    async formatDeploymentData(application) {
      const result = await this.formatApplicationData(application);
      return result;
    },
    async formatDaemonsetData(application) {
      const result = await this.formatApplicationData(application);
      return result;
    },
    async formatJobData(application) {
      const result = await this.formatApplicationData(application);
      return result;
    },
    async formatStatefulsetData(application) {
      const result = await this.formatApplicationData(application);
      return result;
    },
    saveTemplateData() {
      // 利用setTimeout事件来先让当前容器的blur事件执行完才切换
      setTimeout(() => {
        this.updateTemplateData();
      }, 500);
    },
    async updateTemplateData() {
      if (!this.isTemplateCanSave) {
        return false;
      }
      const { deployments } = this;
      const { daemonsets } = this;
      const { jobs } = this;
      const { statefulsets } = this;
      const { services } = this;
      const { configmaps } = this;
      const { secrets } = this;
      const { ingresss } = this;
      const { HPAs } = this;

      // 对deployment资源数据检测
      for (const deployment of deployments) {
        if (deployment.isEdited) {
          const isValid = await this.checkDeploymentData(deployment);
          if (!isValid) {
            return false;
          }
        } else {
          const isValid = this.checkLinkData(deployment);
          if (!isValid) {
            return false;
          }
        }
      }

      // 对service资源数据检测
      for (const service of services) {
        if (service.isEdited) {
          const isValid = await this.checkServiceData(service);
          if (!isValid) {
            return false;
          }
        }
      }

      // 对configmap资源数据检测
      for (const configmap of configmaps) {
        if (configmap.isEdited) {
          const isValid = await this.checkConfigmapData(configmap);
          if (!isValid) {
            return false;
          }
        }
      }

      // 对secret资源数据检测
      for (const secret of secrets) {
        if (secret.isEdited) {
          const isValid = await this.checkSecretData(secret);
          if (!isValid) {
            return false;
          }
        }
      }

      // 对daemonset资源数据检测
      for (const daemonset of daemonsets) {
        if (daemonset.isEdited) {
          const isValid = await this.checkDaemonsetData(daemonset);
          if (!isValid) {
            return false;
          }
        }
      }

      // 对job资源数据检测
      for (const job of jobs) {
        if (job.isEdited) {
          const isValid = await this.checkJobData(job);
          if (!isValid) {
            return false;
          }
        }
      }

      // 对statefulset资源数据检测
      for (const statefulset of statefulsets) {
        if (statefulset.isEdited) {
          const isValid = await this.checkStatefulsetData(statefulset);
          if (!isValid) {
            return false;
          }
        }
      }

      // 对ingress资源数据检测
      for (const ingress of ingresss) {
        if (ingress.isEdited) {
          const isValid = await this.checkIngressData(ingress);
          if (!isValid) {
            return false;
          }
        }
      }

      // 对HPA资源数据检测
      for (const HPA of HPAs) {
        if (HPA.isEdited) {
          const isValid = await this.checkHPAData(HPA);
          if (!isValid) {
            return false;
          }
        }
      }

      if (this.isDataSaveing) return;
      this.$store.commit('k8sTemplate/updateIsTemplateSaving', true);
      this.isDataSaveing = true;

      // 保存deployments
      for (const deployment of deployments) {
        if (!deployment.isEdited) {
          continue;
        }
        const preId = deployment.id;
        const result = await this.saveApplication(deployment, 'deployment');
        if (!result) {
          return false;
        }
        this.$emit('saveDeploymentSuccess', {
          responseData: result,
          resource: deployment,
          preId,
        });
      }

      // 保存services
      for (const service of services) {
        if (!service.isEdited) {
          continue;
        }
        const preId = service.id;
        const result = await this.saveService(service);
        if (!result) {
          return false;
        }
        this.$emit('saveServiceSuccess', {
          responseData: result,
          resource: service,
          preId,
        });
      }

      // 保存configmaps
      for (const configmap of configmaps) {
        if (!configmap.isEdited) {
          continue;
        }
        const preId = configmap.id;
        const result = await this.saveConfigmap(configmap);
        if (!result) {
          return false;
        }
        this.$emit('saveConfigmapSuccess', {
          responseData: result,
          resource: configmap,
          preId,
        });
      }

      // 保存secrets
      for (const secret of secrets) {
        if (!secret.isEdited) {
          continue;
        }
        const preId = secret.id;
        const result = await this.saveSecret(secret);
        if (!result) {
          return false;
        }
        this.$emit('saveSecretSuccess', {
          responseData: result,
          resource: secret,
          preId,
        });
      }

      // 保存daemonsets
      for (const daemonset of daemonsets) {
        if (!daemonset.isEdited) {
          continue;
        }
        const preId = daemonset.id;
        const result = await this.saveApplication(daemonset, 'daemonset');
        if (!result) {
          return false;
        }
        this.$emit('saveDaemonsetSuccess', {
          responseData: result,
          resource: daemonset,
          preId,
        });
      }

      // 保存jobs
      for (const job of jobs) {
        if (!job.isEdited) {
          continue;
        }
        const preId = job.id;
        const result = await this.saveApplication(job, 'job');
        if (!result) {
          return false;
        }
        this.$emit('saveJobSuccess', {
          responseData: result,
          resource: job,
          preId,
        });
      }

      // 保存statefulsets
      for (const statefulset of statefulsets) {
        if (!statefulset.isEdited) {
          continue;
        }
        const preId = statefulset.id;
        const result = await this.saveApplication(statefulset, 'statefulset');
        if (!result) {
          return false;
        }
        this.$emit('saveStatefulsetSuccess', {
          responseData: result,
          resource: statefulset,
          preId,
        });
      }

      // 保存ingresss
      for (const ingress of ingresss) {
        if (!ingress.isEdited) {
          continue;
        }
        const preId = ingress.id;
        const result = await this.saveIngress(ingress);
        if (!result) {
          return false;
        }
        this.$emit('saveIngressSuccess', {
          responseData: result,
          resource: ingress,
          preId,
        });
      }

      // 保存HPAs
      for (const HPA of HPAs) {
        if (!HPA.isEdited) {
          continue;
        }
        const preId = HPA.id;
        const result = await this.saveHPA(HPA);
        if (!result) {
          return false;
        }
        this.$emit('saveHPASuccess', {
          responseData: result,
          resource: HPA,
          preId,
        });
      }

      this.$store.commit('k8sTemplate/updateIsTemplateSaving', false);
      await this.getVersionList();
      this.versionSidePanel.isShow = false;
      this.versionDialogConf.isShow = true;
      this.isDataSaveing = false;
    },
    hideVersionBox() {
      if (this.isNewTemplate) {
        if (this.newTemplateId) {
          this.$router.push({
            name: 'k8sTemplatesetDeployment',
            params: {
              projectId: this.projectId,
              projectCode: this.projectCode,
              templateId: this.newTemplateId,
            },
          });
        }
      }
      this.versionDialogConf.isShow = false;
      this.selectedVersion = '';
    },
    checkVersionData() {
      const nameReg = /^[a-zA-Z0-9-_.]{1,45}$/;

      if (!this.versionKeyword) {
        this.$bkMessage({
          theme: 'error',
          message: this.$t('deploy.templateset.enterVersionNumber'),
        });
        return false;
      }

      if (!nameReg.test(this.versionKeyword)) {
        this.$bkMessage({
          theme: 'error',
          message: this.$t('deploy.templateset.compositionAlphaNumSymbols'),
        });
        return false;
      }

      for (const item of this.versionList) {
        if (item.name === this.versionKeyword) {
          this.$bkMessage({
            theme: 'error',
            message: this.$t('deploy.templateset.ersionExists', { versionKeyword: this.versionKeyword }),
          });
          return false;
        }
      }
      return true;
    },
    async saveVersion() {
      const { projectId } = this;
      const templateId = this.curTemplateId;
      // 根据不同方式组装数据
      this.versionMetadata.real_version_id = this.curVersion;
      if (this.saveVersionWay === 'cur') {
        this.versionMetadata.show_version_id = this.curShowVersionId;
      } else if (this.saveVersionWay === 'old' && this.selectedVersion) {
        this.versionMetadata.show_version_id = this.selectedVersion;
      } else if (this.saveVersionWay === 'new') {
        if (this.checkVersionData()) {
          this.versionMetadata.name = this.versionKeyword;
          this.versionMetadata.show_version_id = 0;
        } else {
          return false;
        }
      }

      // 匹配name
      if (this.versionList) {
        // 如果有版本，自动默认选中原来版本号
        this.versionList.forEach((item) => {
          if (String(item.show_version_id) === String(this.versionMetadata.show_version_id)) {
            this.versionMetadata.name = item.name;
          }
        });
      }

      const params = this.versionMetadata;
      params.comment = this.curVersionNotes;

      this.isCreating = true;
      await this.$store.dispatch('k8sTemplate/saveVersion', { projectId, templateId, params }).then((res) => {
        this.$bkMessage({
          theme: 'success',
          message: this.$t('generic.msg.success.save'),
          delay: 3000,
        });
        this.$store.commit('k8sTemplate/updateBindVersion', false);

        if (res.data.show_version_id) {
          this.$store.commit('k8sTemplate/updateCurShowVersionId', res.data.show_version_id);
        }
        if (this.isNewTemplate && this.curTemplateId) {
          this.$router.push({
            name: this.$route.name,
            params: {
              projectId: this.projectId,
              templateId: this.curTemplateId,
            },
          });
        }

        this.curTemplate.latest_show_version = this.versionMetadata.name;
        this.curTemplate.latest_show_version_id = res.data.show_version_id;
        this.curTemplate.latest_version_id = res.data.real_version_id;
        this.saveVersionWay = 'cur';
        this.versionKeyword = '';
        this.selectedVersion = '';
        this.versionDialogConf.isShow = false;
        this.getVersionList();
      }, (res) => {
        this.$store.commit('k8sTemplate/updateIsTemplateSaving', false);
        this.$bkMessage({
          theme: 'error',
          message: res.message,
        });
      })
        .finally(() => {
          this.isCreating = false;
        });
    },
    getKeyList(list) {
      let results = [];
      results = list.filter(item => item.key && item.value);
      return results;
    },
    tranListToObject(list) {
      const results = this.getKeyList(list);
      if (results.length === 0) {
        return {};
      }
      const obj = {};
      results.forEach((item) => {
        if (item.key) {
          obj[item.key] = item.value;
        }
      });
      return obj;
    },
    async saveApplication(application, type) {
      let result;
      let data;
      switch (type) {
        case 'deployment':
          data = await this.formatDeploymentData(application);
          break;

        case 'daemonset':
          data = await this.formatDaemonsetData(application);
          break;

        case 'job':
          data = await this.formatJobData(application);
          break;

        case 'statefulset':
          data = await this.formatStatefulsetData(application);
          break;
      }

      if (this.curVersion) {
        if (application.id.indexOf && (application.id.indexOf('local') > -1)) {
          result = await this.createApplication(data, application, type);
        } else {
          result = await this.updateApplication(data, application, type);
        }
      } else {
        result = await this.createFirstApplication(data, application, type);
      }
      return result;
    },
    async saveDaemonset(daemonset) {
      let result;
      const data = await this.formatDaemonsetData(daemonset);
      if (this.curVersion) {
        if (daemonset.id.indexOf && (daemonset.id.indexOf('local') > -1)) {
          result = await this.createDaemonset(data, daemonset);
        } else {
          result = await this.updateDaemonset(data, daemonset);
        }
      } else {
        result = await this.createFirstDaemonset(data, daemonset);
      }
      return result;
    },
    updateLocalData(responseData, targetData, resourceType) {
      targetData.isEdited = false;
      if (targetData.id) {
        const preId = targetData.id;
        switch (resourceType) {
          case 'deployment':
            this.$store.commit('k8sTemplate/updateDeploymentById', {
              deployment: responseData,
              targetData,
              preId,
            });
            break;
          case 'daemonset':
            this.$store.commit('k8sTemplate/updateDaemonsetById', {
              daemonset: responseData,
              targetData,
              preId,
            });
            break;
          case 'job':
            this.$store.commit('k8sTemplate/updateJobById', {
              job: responseData,
              targetData,
              preId,
            });
            break;
          case 'statefulset':
            this.$store.commit('k8sTemplate/updateStatefulsetById', {
              statefulset: responseData,
              targetData,
              preId,
            });
            break;
          case 'service':
            this.$store.commit('k8sTemplate/updateServiceById', {
              service: responseData,
              preId,
            });
            break;
          case 'configmap':
            this.$store.commit('k8sTemplate/updateConfigmapById', {
              configmap: responseData,
              targetData,
              preId,
            });
            break;
          case 'secret':
            this.$store.commit('k8sTemplate/updateSecretById', {
              secret: responseData,
              targetData,
              preId,
            });
            break;
          case 'ingress':
            this.$store.commit('k8sTemplate/updateIngressById', {
              ingress: responseData,
              targetData,
              preId,
            });
            break;
          case 'HPA':
            this.$store.commit('k8sTemplate/updateHPAById', {
              HPA: responseData,
              targetData,
              preId,
            });
            break;
        }
      }
      if (responseData.template_id) {
        this.newTemplateId = responseData.template_id;
        this.$store.commit('k8sTemplate/updateCurTemplateId', responseData.template_id);
      }
      if (responseData.version) {
        this.$store.commit('k8sTemplate/updateCurVersion', responseData.version);
      }
    },
    async createApplication(data, resource, resourceType) {
      const version = this.curVersion;
      const { projectId } = this;
      const operationMap = {
        deployment: 'k8sTemplate/addDeployment',
        daemonset: 'k8sTemplate/addDaemonset',
        job: 'k8sTemplate/addJob',
        statefulset: 'k8sTemplate/addStatefulset',
      };
      const result = await this.$store.dispatch(operationMap[resourceType], { projectId, version, data }).then((res) => {
        const responseData = res.data;
        resource.config.spec.template.metadata.labels = data.config.spec.template.metadata.labels;
        resource.config.spec.selector.matchLabels = data.config.spec.selector.matchLabels;
        this.updateLocalData(responseData, resource, resourceType);
        return responseData;
      }, (res) => {
        this.$store.commit('k8sTemplate/updateIsTemplateSaving', false);
        this.$bkMessage({
          theme: 'error',
          message: res.message,
          hasCloseIcon: true,
          delay: '10000',
        });
        this.isDataSaveing = false;
      });
      return result;
    },
    async createFirstApplication(data, resource, resourceType) {
      const templateId = this.curTemplateId;
      const { projectId } = this;
      const operationMap = {
        deployment: 'k8sTemplate/addFirstDeployment',
        daemonset: 'k8sTemplate/addFirstDaemonset',
        job: 'k8sTemplate/addFirstJob',
        statefulset: 'k8sTemplate/addFirstStatefulset',
      };

      const result = await this.$store.dispatch(operationMap[resourceType], { projectId, templateId, data }).then((res) => {
        const responseData = res.data;
        resource.config.spec.template.metadata.labels = data.config.spec.template.metadata.labels;
        resource.config.spec.selector.matchLabels = data.config.spec.selector.matchLabels;
        this.updateLocalData(responseData, resource, resourceType);
        return responseData;
      }, (res) => {
        this.$store.commit('k8sTemplate/updateIsTemplateSaving', false);
        this.$bkMessage({
          theme: 'error',
          message: res.message,
          hasCloseIcon: true,
          delay: '10000',
        });
        this.isDataSaveing = false;
      });
      return result;
    },
    async updateApplication(data, resource, resourceType) {
      const version = this.curVersion;
      const { projectId } = this;
      const { id } = data;
      const operationMap = {
        deployment: 'k8sTemplate/updateDeployment',
        daemonset: 'k8sTemplate/updateDaemonset',
        job: 'k8sTemplate/updateJob',
        statefulset: 'k8sTemplate/updateStatefulset',
      };

      const result = await this.$store.dispatch(operationMap[resourceType], { projectId, version, data, id }).then((res) => {
        const responseData = res.data;
        resource.config.spec.template.metadata.labels = data.config.spec.template.metadata.labels;
        resource.config.spec.selector.matchLabels = data.config.spec.selector.matchLabels;
        this.updateLocalData(responseData, resource, resourceType);
        return responseData;
      }, (res) => {
        this.$store.commit('k8sTemplate/updateIsTemplateSaving', false);
        this.$bkMessage({
          theme: 'error',
          message: res.message,
          hasCloseIcon: true,
          delay: '10000',
        });
        this.isDataSaveing = false;
      });

      return result;
    },
    async saveService(service) {
      let result;
      const data = await this.formatServiceData(service);

      if (this.curVersion) {
        if (service.id.indexOf && (service.id.indexOf('local') > -1)) {
          result = await this.createService(data, service);
        } else {
          result = await this.updateService(data, service);
        }
      } else {
        result = await this.createFirstService(data, service);
      }
      // 如果成功，且绑定的是statefulset则自动同步到相应的statefulset资源
      if (result && result.version) {
        const statefulsetItem = service.deploy_tag_list.find(item => item.indexOf('K8sStatefulSet') > -1);
        if (statefulsetItem) {
          const statefulsetId = statefulsetItem.split('|')[0];
          try {
            // 绑定
            this.statefulsets.forEach((statefulset) => {
              // 把其它已经绑定的statefulset进行解绑
              if (statefulset.deploy_tag !== statefulsetId && statefulset.service_tag === service.service_tag) {
                statefulset.service_tag = '';
                this.$store.dispatch('k8sTemplate/bindServiceForStatefulset', {
                  projectId: this.projectId,
                  versionId: result.version,
                  statefulsetId: statefulset.deploy_tag,
                  data: {
                    service_tag: '',
                  },
                });
              }
              // 给绑定的statefulset同步本地数据
              if (String(statefulset.deploy_tag) === statefulsetId) {
                statefulset.service_tag = service.service_tag;
              }
            });
            // 同步到接口
            await this.$store.dispatch('k8sTemplate/bindServiceForStatefulset', {
              projectId: this.projectId,
              versionId: result.version,
              statefulsetId,
              data: {
                service_tag: service.service_tag,
              },
            });
          } catch (res) {
            this.$bkMessage({
              theme: 'error',
              message: res.message,
              hasCloseIcon: true,
              delay: '3000',
            });
          }
        } else {
          // 如果原来已经存在statefulset，现在取消那需要解绑
          if (service.cache && service.cache.deploy_tag_list) {
            const statefulsetItem = service.cache.deploy_tag_list.find(item => item.indexOf('K8sStatefulSet') > -1);
            if (statefulsetItem) {
              const statefulsetId = statefulsetItem.split('|')[0];
              // 绑定
              this.statefulsets.forEach((statefulset) => {
                // 把其它已经绑定的statefulset进行解绑
                if (String(statefulset.deploy_tag) === statefulsetId) {
                  statefulset.service_tag = '';
                  this.$store.dispatch('k8sTemplate/bindServiceForStatefulset', {
                    projectId: this.projectId,
                    versionId: result.version,
                    statefulsetId: statefulset.deploy_tag,
                    data: {
                      service_tag: '',
                    },
                  });
                }
              });
            }
          }
        }
      }
      return result;
    },

    async saveIngress(ingress) {
      let result;
      const data = await this.formatIngressData(ingress);

      if (this.curVersion) {
        if (ingress.id.indexOf && (ingress.id.indexOf('local') > -1)) {
          result = await this.createIngress(data, ingress);
        } else {
          result = await this.updateIngress(data, ingress);
        }
      } else {
        result = await this.createFirstIngress(data, ingress);
      }
      return result;
    },
    async checkIngressData(ingress) {
      const ingressName = ingress.config.metadata.name;
      const nameReg = /^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$/;
      const labelKeyReg = /^([A-Za-z0-9][-A-Za-z0-9_./]*)?[A-Za-z0-9]$/;
      const varReg = /\{\{([^\{\}]+)?\}\}/g;
      const pathReg = /\/((?!\.)[\w\d\-./~]+)*/;
      let megPrefix = `"${ingressName}"${this.$t('deploy.templateset.in')}`;

      if (ingressName === '') {
        megPrefix += this.$t('deploy.templateset.nameLabel');
        this.$bkMessage({
          theme: 'error',
          message: this.$t('deploy.templateset.validate.name1'),
        });
        return false;
      }
      if (!nameReg.test(ingressName)) {
        megPrefix += this.$t('deploy.templateset.nameLabel');
        this.$bkMessage({
          theme: 'error',
          message: megPrefix + this.$t('deploy.templateset._nameError'),
          delay: 8000,
        });
        return false;
      }

      // 标签
      if (ingress.config.webCache.labelListCache) {
        const cacheLabels = ingress.config.webCache.labelListCache;
        for (const label of cacheLabels) {
          if (label.key && !labelKeyReg.test(label.key.replace(varReg, 'key'))) {
            megPrefix += `${this.$t('k8s.label')}：`;
            this.$bkMessage({
              theme: 'error',
              delay: 8000,
              message: megPrefix + this.$t('deploy.templateset.startEndAlphaNumeric'),
            });
            return false;
          }
        }
      }

      // 注解
      if (ingress.config.webCache.remarkListCache) {
        const cacheLabels = ingress.config.webCache.remarkListCache;
        for (const label of cacheLabels) {
          if (label.key && !labelKeyReg.test(label.key.replace(varReg, 'key'))) {
            megPrefix += `${this.$t('k8s.annotation')}：`;
            this.$bkMessage({
              theme: 'error',
              delay: 8000,
              message: megPrefix + this.$t('deploy.templateset.startEndAlphaNumeric'),
            });
            return false;
          }
        }
      }

      for (const rule of ingress.config.spec.rules) {
        // 检查rule
        if (!rule.host) {
          megPrefix += `${this.$t('generic.label.rule')}：`;
          this.$bkMessage({
            theme: 'error',
            message: megPrefix + this.$t('deploy.templateset.hostnameNotEmpty'),
          });
          return false;
        }

        // if (!nameReg.test(rule.host)) {
        //     megPrefix += this.$t('deploy.templateset._ruleHostname', { host: rule.host })
        //     this.$bkMessage({
        //         theme: 'error',
        //         message: megPrefix + this.$t('deploy.templateset.nameErrorWithLength'),
        //         delay: 8000
        //     })
        //     return false
        // }

        const { paths } = rule.http;

        for (const path of paths) {
          if (path.path && !pathReg.test(path.path)) {
            megPrefix += this.$t('deploy.templateset._pathGroup', { host: rule.host });
            this.$bkMessage({
              theme: 'error',
              message: megPrefix + this.$t('deploy.templateset.pathIncorrect'),
              delay: 8000,
            });
            return false;
          }

          if (path.backend.serviceName && !path.backend.servicePort) {
            megPrefix += this.$t('deploy.templateset._pathGroup', { host: rule.host });
            this.$bkMessage({
              theme: 'error',
              message: megPrefix + this.$t('deploy.templateset.associateServicePort'),
              delay: 8000,
            });
            return false;
          }

          if (path.backend.serviceName && !this.linkServices.includes(path.backend.serviceName)) {
            megPrefix += this.$t('deploy.templateset._pathGroup', { host: rule.host });
            this.$bkMessage({
              theme: 'error',
              message: megPrefix + this.$t('deploy.templateset.associatedServiceNotExist', {
                serviceName: path.backend.serviceName,
              }),
              delay: 8000,
            });
            return false;
          }
        }
      }
      return true;
    },
    async saveHPA(HPA) {
      let result;
      const data = await this.formatHPAData(HPA);

      if (this.curVersion) {
        if (HPA.id.indexOf && (HPA.id.indexOf('local') > -1)) {
          result = await this.createHPA(data, HPA);
        } else {
          result = await this.updateHPA(data, HPA);
        }
      } else {
        result = await this.createFirstHPA(data, HPA);
      }
      return result;
    },
    async checkHPAData(HPA) {
      const HPAName = HPA.config.metadata.name;
      const nameReg = /^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$/;
      let megPrefix = `"${HPAName}"${this.$t('deploy.templateset.in')}`;

      if (HPAName === '') {
        megPrefix += this.$t('deploy.templateset.nameLabel');
        this.$bkMessage({
          theme: 'error',
          message: this.$t('deploy.templateset.validate.name1'),
        });
        return false;
      }

      if (!nameReg.test(HPAName)) {
        megPrefix += this.$t('deploy.templateset.nameLabel');
        this.$bkMessage({
          theme: 'error',
          message: megPrefix + this.$t('deploy.templateset._nameError'),
          delay: 5000,
        });
        return false;
      }

      if (!HPA.config.spec.scaleTargetRef.name) {
        megPrefix += this.$t('deploy.templateset.associatedApplication');
        this.$bkMessage({
          theme: 'error',
          message: megPrefix + this.$t('deploy.templateset.associateApplicationFirst'),
          delay: 5000,
        });
        return false;
      }

      if (HPA.config.spec.minReplicas === '') {
        megPrefix += this.$t('deploy.templateset._instanceRange');
        this.$bkMessage({
          theme: 'error',
          message: megPrefix + this.$t('deploy.templateset.minInstanceNotEmpty'),
          delay: 5000,
        });
        return false;
      }

      if (HPA.config.spec.maxReplicas === '') {
        megPrefix += this.$t('deploy.templateset._instanceRange');
        this.$bkMessage({
          theme: 'error',
          message: megPrefix + this.$t('deploy.templateset.maxInstanceNotEmpty'),
          delay: 5000,
        });
        return false;
      }

      if (HPA.config.spec.maxReplicas < HPA.config.spec.minReplicas) {
        megPrefix += this.$t('deploy.templateset._instanceRange');
        this.$bkMessage({
          theme: 'error',
          message: megPrefix + this.$t('deploy.templateset.maxInstanceNotLessThanMin'),
          delay: 5000,
        });
        return false;
      }

      if (HPA.config.spec.metrics.length) {
        for (const metric of HPA.config.spec.metrics) {
          if (!metric.type) {
            megPrefix += this.$t('deploy.templateset._scalingTriggerCondition');
            this.$bkMessage({
              theme: 'error',
              message: megPrefix + this.$t('deploy.templateset.selectResourceType'),
              delay: 5000,
            });
            return false;
          }

          if (metric.type === 'Resource' && !metric.resource.target.averageUtilization) {
            megPrefix += this.$t('deploy.templateset._scalingTriggerCondition');
            this.$bkMessage({
              theme: 'error',
              message: megPrefix + this.$t('deploy.templateset.resourceTargetNotEmpty'),
              delay: 5000,
            });
            return false;
          }
        }
      }

      return true;
    },
    async formatHPAData(HPA) {
      const params = JSON.parse(JSON.stringify(HPA));
      params.template = {
        name: this.curTemplate.name,
        desc: this.curTemplate.desc,
      };
      delete params.isEdited;
      delete params.cache;
      return params;
    },
    async createHPA(data, HPA) {
      const version = this.curVersion;
      const { projectId } = this;
      const result = this.$store.dispatch('k8sTemplate/addHPA', { projectId, version, data }).then((res) => {
        const responseData = res.data;
        this.updateLocalData(responseData, HPA, 'HPA');
        return responseData;
      }, (res) => {
        this.$store.commit('k8sTemplate/updateIsTemplateSaving', false);
        this.$bkMessage({
          theme: 'error',
          message: res.message,
          hasCloseIcon: true,
          delay: '10000',
        });
        this.isDataSaveing = false;
      });
      return result;
    },
    async updateHPA(data, HPA) {
      const version = this.curVersion;
      const { projectId } = this;
      const HPAId = data.id;
      const result = this.$store.dispatch('k8sTemplate/updateHPA', { projectId, version, data, HPAId }).then((res) => {
        const { data } = res;
        this.updateLocalData(data, HPA, 'HPA');
        this.isDataSaveing = false;
        return data;
      }, (res) => {
        this.$store.commit('k8sTemplate/updateIsTemplateSaving', false);
        this.$bkMessage({
          theme: 'error',
          message: res.message,
          hasCloseIcon: true,
          delay: '10000',
        });
        this.isDataSaveing = false;
      });
      return result;
    },
    async createFirstHPA(data, HPA) {
      const templateId = this.curTemplateId;
      const { projectId } = this;
      const result = await this.$store.dispatch('k8sTemplate/addFirstHPA', { projectId, templateId, data }).then((res) => {
        const responseData = res.data;
        this.updateLocalData(responseData, HPA, 'HPA');
        this.isDataSaveing = false;
        return data;
      }, (res) => {
        this.$store.commit('k8sTemplate/updateIsTemplateSaving', false);
        this.$bkMessage({
          theme: 'error',
          message: res.message,
          hasCloseIcon: true,
          delay: '10000',
        });
        this.isDataSaveing = false;
      });
      return result;
    },
    async checkServiceData(service) {
      const serviceName = service.config.metadata.name;
      const serviceNameReg = /^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$/;
      const serviceIPReg = /^((25[0-5]|2[0-4]\d|[01]?\d\d?)($|(?!\.$)\.)){4}$/;
      const varReg = /\{\{([^\{\}]+)?\}\}/g;

      let megPrefix = `"${serviceName}"${this.$t('deploy.templateset.in')}`;

      if (serviceName === '') {
        megPrefix += this.$t('deploy.templateset.nameLabel');
        this.$bkMessage({
          theme: 'error',
          message: this.$t('deploy.templateset.validate.name1'),
        });
        return false;
      }
      if (!serviceNameReg.test(serviceName.replace(varReg, 'service'))) {
        megPrefix += this.$t('deploy.templateset.nameLabel');
        this.$bkMessage({
          theme: 'error',
          message: megPrefix + this.$t('deploy.templateset._nameError'),
          delay: 8000,
        });
        return false;
      }

      if (!service.deploy_tag_list.length) {
        megPrefix += this.$t('deploy.templateset.associatedApplication');
        this.$bkMessage({
          theme: 'error',
          message: megPrefix + this.$t('deploy.templateset.selectAssociatedApplication'),
          delay: 3000,
        });
        return false;
      }

      const statefulsetList = service.deploy_tag_list.filter(item => item.indexOf('K8sStatefulSet') > -1);
      if (statefulsetList.length >= 2) {
        this.$bkMessage({
          theme: 'error',
          message: this.$t('deploy.templateset.multipleStatefulSetsError'),
          hasCloseIcon: true,
          delay: '3000',
        });
        return false;
      }

      if (!service.config.webCache.link_labels.length) {
        megPrefix += this.$t('deploy.templateset.associatedLabel');
        this.$bkMessage({
          theme: 'error',
          message: megPrefix + this.$t('deploy.templateset.selectAssociatedLabel'),
          delay: 3000,
        });
        return false;
      }

      const serviceIp = service.config.spec.clusterIP;
      if (serviceIp && serviceIp !== 'None' && !serviceIPReg.test(serviceIp)) {
        megPrefix += 'IP：';
        this.$bkMessage({
          theme: 'error',
          message: megPrefix + this.$t('deploy.templateset.enterCorrectIpOrNone'),
          delay: 3000,
        });
        return false;
      }

      // 端口映射检查
      const { ports } = service.config.spec;
      if (serviceIp !== 'None') {
        let hasPort = false;
        for (const item of ports) {
          if (item.name && item.port && item.targetPort) {
            hasPort = true;
          }
        }
        // 没有端口
        if (!hasPort) {
          // 如果已经关联应用
          if (service.deploy_tag_list.length) {
            megPrefix += this.$t('dashboard.network.portmapping');
            this.$bkMessage({
              theme: 'error',
              delay: 8000,
              message: megPrefix + this.$t('deploy.templateset.associatedApplicationFillPort'),
            });
            return false;
          }
          megPrefix += this.$t('dashboard.network.portmapping');
          this.$bkMessage({
            theme: 'error',
            delay: 8000,
            message: megPrefix + this.$t('deploy.templateset.clusterIPNoneOrAssociateApplication'),
          });
          return false;
        }
      }
      for (const item of ports) {
        if (item.name || item.port || item.targetPort) {
          if (item.name && !/^[a-z]{1}[a-z0-9-]{0,29}$/.test(item.name.replace(varReg, 'name'))) {
            megPrefix += this.$t('dashboard.network.portmapping');
            this.$bkMessage({
              theme: 'error',
              delay: 8000,
              message: megPrefix + this.$t('deploy.templateset.portNameCriteria'),
            });
            return false;
          }
          if (!item.port) {
            megPrefix += this.$t('dashboard.network.portmapping');
            this.$bkMessage({
              theme: 'error',
              delay: 5000,
              message: megPrefix + this.$t('deploy.templateset.portNotEmpty'),
            });
            return false;
          }
          if (!item.protocol) {
            megPrefix += this.$t('dashboard.network.portmapping');
            this.$bkMessage({
              theme: 'error',
              delay: 5000,
              message: megPrefix + this.$t('deploy.templateset.selectProtocol'),
            });
            return false;
          }
          if (item.nodePort || item.nodePort === 0) {
            if (!varReg.test(item.nodePort) && (item.nodePort < 30000 || item.nodePort > 32767)) {
              megPrefix += this.$t('dashboard.network.portmapping');
              this.$bkMessage({
                theme: 'error',
                delay: 5000,
                message: megPrefix + this.$t('deploy.templateset.nodePortRange'),
              });
              return false;
            }
          }
        }
      }

      return true;
    },
    async formatServiceData(service) {
      const params = JSON.parse(JSON.stringify(service));
      const { webCache } = params.config;
      if (webCache && webCache.labelListCache) {
        const labelKeyList = this.tranListToObject(webCache.labelListCache);
        params.config.metadata.labels = labelKeyList;
      }

      if (webCache && webCache.remarkListCache) {
        const remarkKeyList = this.tranListToObject(webCache.remarkListCache);
        params.config.metadata.annotations = remarkKeyList;
      }
      if (webCache && webCache.link_labels) {
        const selector = {};
        webCache.link_labels.forEach((item) => {
          const values = item.split(':');
          selector[values[0]] = values[1];
        });
        params.config.spec.selector = selector;
      }

      const { ports } = params.config.spec;
      const validPorts = ports.filter(port => port.name && port.port && port.targetPort);
      params.config.spec.ports = validPorts;

      params.template = {
        name: this.curTemplate.name,
        desc: this.curTemplate.desc,
      };
      return params;
    },
    async formatIngressData(ingress) {
      const params = JSON.parse(JSON.stringify(ingress));
      params.template = {
        name: this.curTemplate.name,
        desc: this.curTemplate.desc,
      };
      delete params.isEdited;
      // 键值转换
      const { webCache } = ingress.config;
      if (webCache && webCache.remarkListCache) {
        const remarkKeyList = this.tranListToObject(webCache.remarkListCache);
        params.config.metadata.annotations = remarkKeyList;
      }

      if (webCache && webCache.labelListCache) {
        const labelKeyList = this.tranListToObject(webCache.labelListCache);
        params.config.metadata.labels = labelKeyList;
      }
      // 如果不是变量，转为数组形式
      const varReg = /\{\{([^\{\}]+)?\}\}/g;
      const tls = [];
      params.config.spec.tls.forEach((item) => {
        if (item.hosts || item.secretName) {
          if (!varReg.test(item.hosts)) {
            item.hosts = item.hosts.split(',');
          }
          tls.push(item);
        }
      });

      // 如果用户有填写，把tls传过去，否则删除
      if (tls.length) {
        params.config.spec.tls = tls;
      } else {
        delete params.config.spec.tls;
      }

      delete params.cache;
      return params;
    },
    async createIngress(data, ingress) {
      const version = this.curVersion;
      const { projectId } = this;
      const result = this.$store.dispatch('k8sTemplate/addIngress', { projectId, version, data }).then((res) => {
        const responseData = res.data;
        this.updateLocalData(responseData, ingress, 'ingress');
        return responseData;
      }, (res) => {
        this.$store.commit('k8sTemplate/updateIsTemplateSaving', false);
        this.$bkMessage({
          theme: 'error',
          message: res.message,
          hasCloseIcon: true,
          delay: '10000',
        });
        this.isDataSaveing = false;
      });
      return result;
    },
    async updateIngress(data, ingress) {
      const version = this.curVersion;
      const { projectId } = this;
      const ingressId = data.id;
      const result = this.$store.dispatch('k8sTemplate/updateIngress', { projectId, version, data, ingressId }).then((res) => {
        const { data } = res;
        this.updateLocalData(data, ingress, 'ingress');
        this.isDataSaveing = false;
        return data;
      }, (res) => {
        this.$store.commit('k8sTemplate/updateIsTemplateSaving', false);
        this.$bkMessage({
          theme: 'error',
          message: res.message,
          hasCloseIcon: true,
          delay: '10000',
        });
        this.isDataSaveing = false;
      });
      return result;
    },
    async createFirstIngress(data, ingress) {
      const templateId = this.curTemplateId;
      const { projectId } = this;
      const result = await this.$store.dispatch('k8sTemplate/addFirstIngress', { projectId, templateId, data }).then((res) => {
        const responseData = res.data;
        this.updateLocalData(responseData, ingress, 'ingress');
        this.isDataSaveing = false;
        return data;
      }, (res) => {
        this.$store.commit('k8sTemplate/updateIsTemplateSaving', false);
        this.$bkMessage({
          theme: 'error',
          message: res.message,
          hasCloseIcon: true,
          delay: '10000',
        });
        this.isDataSaveing = false;
      });
      return result;
    },
    async createService(data, service) {
      const version = this.curVersion;
      const { projectId } = this;
      const result = await this.$store.dispatch('k8sTemplate/addService', { projectId, version, data }).then((res) => {
        const responseData = res.data;
        this.updateLocalData(responseData, service, 'service');
        return responseData;
      }, (res) => {
        this.$store.commit('k8sTemplate/updateIsTemplateSaving', false);
        this.$bkMessage({
          theme: 'error',
          message: res.message,
          hasCloseIcon: true,
          delay: '10000',
        });
        this.isDataSaveing = false;
      });
      return result;
    },
    async createFirstService(data, service) {
      const templateId = this.curTemplateId;
      const { projectId } = this;
      const result = await this.$store.dispatch('k8sTemplate/addFirstService', { projectId, templateId, data }).then((res) => {
        const responseData = res.data;
        this.updateLocalData(responseData, service, 'service');
        this.isDataSaveing = false;
        return responseData;
      }, (res) => {
        this.$store.commit('k8sTemplate/updateIsTemplateSaving', false);
        this.$bkMessage({
          theme: 'error',
          message: res.message,
          hasCloseIcon: true,
          delay: '10000',
        });
        this.isDataSaveing = false;
      });
      return result;
    },
    async updateService(data, service) {
      const version = this.curVersion;
      const { projectId } = this;
      const serviceId = data.id;
      const result = this.$store.dispatch('k8sTemplate/updateService', { projectId, version, data, serviceId }).then((res) => {
        const { data } = res;
        this.updateLocalData(data, service, 'service');
        this.isDataSaveing = false;
        return data;
      }, (res) => {
        this.$store.commit('k8sTemplate/updateIsTemplateSaving', false);
        this.$bkMessage({
          theme: 'error',
          message: res.message,
          hasCloseIcon: true,
          delay: '10000',
        });
        this.isDataSaveing = false;
      });
      return result;
    },
    async createConfigmap(data, configmap) {
      const version = this.curVersion;
      const { projectId } = this;
      const result = await this.$store.dispatch('k8sTemplate/addConfigmap', { projectId, version, data }).then((res) => {
        const responseData = res.data;
        this.updateLocalData(responseData, configmap, 'configmap');
        return responseData;
      }, (res) => {
        this.$store.commit('k8sTemplate/updateIsTemplateSaving', false);
        this.$bkMessage({
          theme: 'error',
          message: res.message,
          hasCloseIcon: true,
          delay: '10000',
        });
        this.isDataSaveing = false;
      });
      return result;
    },
    async createFirstConfigmap(data, configmap) {
      const templateId = this.curTemplateId;
      const { projectId } = this;
      const result = await this.$store.dispatch('k8sTemplate/addFirstConfigmap', { projectId, templateId, data }).then((res) => {
        const responseData = res.data;
        this.updateLocalData(responseData, configmap, 'configmap');
        this.isDataSaveing = false;
        return responseData;
      }, (res) => {
        this.$store.commit('k8sTemplate/updateIsTemplateSaving', false);
        this.$bkMessage({
          theme: 'error',
          message: res.message,
          hasCloseIcon: true,
          delay: '10000',
        });
        this.isDataSaveing = false;
      });
      return result;
    },
    async updateConfigmap(data, configmap) {
      const version = this.curVersion;
      const { projectId } = this;
      const configmapId = data.id;
      const result = await this.$store.dispatch('k8sTemplate/updateConfigmap', { projectId, version, data, configmapId }).then((res) => {
        const responseData = res.data;
        this.updateLocalData(responseData, configmap, 'configmap');
        return responseData;
      }, (res) => {
        this.$store.commit('k8sTemplate/updateIsTemplateSaving', false);
        this.$bkMessage({
          theme: 'error',
          message: res.message,
          hasCloseIcon: true,
          delay: '10000',
        });
        this.isDataSaveing = false;
      });
      return result;
    },
    async checkConfigmapData(configmap) {
      const configmapName = configmap.config.metadata.name;
      const nameReg1 = /^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$/;
      const nameReg2 = /^[a-zA-Z{]{1}[a-zA-Z0-9-_.{}]{0,254}$/;
      const keys = configmap.configmapKeyList;
      let megPrefix = `"${configmapName}"${this.$t('deploy.templateset.in')}`;

      if (configmapName === '') {
        megPrefix += this.$t('deploy.templateset.nameLabel');
        this.$bkMessage({
          theme: 'error',
          message: this.$t('deploy.templateset.validate.name1'),
        });
        return false;
      }
      if (!nameReg1.test(configmapName)) {
        megPrefix += this.$t('deploy.templateset.nameLabel');
        this.$bkMessage({
          theme: 'error',
          message: megPrefix + this.$t('deploy.templateset._nameError'),
          delay: 8000,
        });
        return false;
      }
      if (keys && keys.length) {
        for (const item of keys) {
          if (!nameReg2.test(item.key)) {
            megPrefix += this.$t('deploy.templateset.keyLabel');
            this.$bkMessage({
              theme: 'error',
              message: megPrefix + this.$t('deploy.templateset.msg.labelKey'),
              delay: 8000,
            });
            return false;
          }
        }
      } else {
        megPrefix += this.$t('deploy.templateset.keyLabel');
        this.$bkMessage({
          theme: 'error',
          message: megPrefix + this.$t('deploy.templateset.addKeyFirst'),
        });
        return false;
      }
      return true;
    },
    async formatConfigmapData(configmap) {
      const params = JSON.parse(JSON.stringify(configmap));

      const keyObj = {};
      const keys = params.configmapKeyList;
      if (keys && keys.length) {
        keys.forEach((item) => {
          keyObj[item.key] = item.content;
        });
        params.config.data = keyObj;
        configmap.config.data = keyObj;
      }

      params.template = {
        name: this.curTemplate.name,
        desc: this.curTemplate.desc,
      };
      return params;
    },
    async saveConfigmap(configmap) {
      const data = await this.formatConfigmapData(configmap);
      let result;
      if (this.curVersion) {
        if (configmap.id.indexOf && (configmap.id.indexOf('local') > -1)) {
          result = this.createConfigmap(data, configmap);
        } else {
          result = this.updateConfigmap(data, configmap);
        }
      } else {
        result = this.createFirstConfigmap(data, configmap);
      }
      return result;
    },
    async createSecret(data, secret) {
      const version = this.curVersion;
      const { projectId } = this;
      const result = await this.$store.dispatch('k8sTemplate/addSecret', { projectId, version, data }).then((res) => {
        const { data } = res;
        this.updateLocalData(data, secret, 'secret');
        return data;
      }, (res) => {
        this.$store.commit('k8sTemplate/updateIsTemplateSaving', false);
        this.$bkMessage({
          theme: 'error',
          message: res.message,
          hasCloseIcon: true,
          delay: '10000',
        });
        this.isDataSaveing = false;
      });
      return result;
    },
    async createFirstSecret(data, secret) {
      const templateId = this.curTemplateId;
      const { projectId } = this;
      const result = await this.$store.dispatch('k8sTemplate/addFirstSecret', { projectId, templateId, data }).then((res) => {
        const { data } = res;
        this.updateLocalData(data, secret, 'secret');
        return data;
      }, (res) => {
        this.$store.commit('k8sTemplate/updateIsTemplateSaving', false);
        this.$bkMessage({
          theme: 'error',
          message: res.message,
          hasCloseIcon: true,
          delay: '10000',
        });
        this.isDataSaveing = false;
      });
      return result;
    },
    async updateSecret(data, secret) {
      const version = this.curVersion;
      const { projectId } = this;
      const secretId = secret.id;
      const result = await this.$store.dispatch('k8sTemplate/updateSecret', { projectId, version, data, secretId }).then((res) => {
        const { data } = res;
        this.updateLocalData(data, secret, 'secret');
        return data;
      }, (res) => {
        this.$store.commit('k8sTemplate/updateIsTemplateSaving', false);
        this.$bkMessage({
          theme: 'error',
          message: res.message,
          hasCloseIcon: true,
          delay: '10000',
        });
        this.isDataSaveing = false;
      });
      return result;
    },
    async checkSecretData(secret) {
      const secretName = secret.config.metadata.name;
      const nameReg1 = /^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$/;
      const nameReg2 = /^[a-zA-Z0-9-_.{}]{0,255}$/;
      const keys = secret.secretKeyList;
      let megPrefix = `"${secretName}"${this.$t('deploy.templateset.in')}`;

      if (secretName === '') {
        megPrefix += this.$t('deploy.templateset.nameLabel');
        this.$bkMessage({
          theme: 'error',
          message: this.$t('deploy.templateset.validate.name1'),
        });
        return false;
      }
      if (!nameReg1.test(secretName)) {
        megPrefix += this.$t('deploy.templateset.nameLabel');
        this.$bkMessage({
          theme: 'error',
          message: megPrefix + this.$t('deploy.templateset._nameError'),
          delay: 8000,
        });
        return false;
      }

      if (keys && keys.length) {
        for (const item of keys) {
          if (!nameReg2.test(item.key)) {
            megPrefix += this.$t('deploy.templateset.keyLabel');
            this.$bkMessage({
              theme: 'error',
              message: megPrefix + this.$t('deploy.templateset.domainRegex'),
              delay: 8000,
            });
            return false;
          }
        }
      } else {
        megPrefix += this.$t('deploy.templateset.keyLabel');
        this.$bkMessage({
          theme: 'error',
          message: megPrefix + this.$t('deploy.templateset.addKeyFirst'),
        });
        return false;
      }

      return true;
    },
    async formatSecretData(secret) {
      const params = JSON.parse(JSON.stringify(secret));
      const keyObj = {};
      const keys = params.secretKeyList;
      if (keys && keys.length) {
        keys.forEach((item) => {
          keyObj[item.key] = item.content;
        });
        params.config.data = keyObj;
        secret.config.data = keyObj;
      }

      params.template = {
        name: this.curTemplate.name,
        desc: this.curTemplate.desc,
      };
      return params;
    },
    async saveSecret(secret) {
      const data = await this.formatSecretData(secret);
      let result;
      if (this.curVersion) {
        if (secret.id.indexOf && (secret.id.indexOf('local') > -1)) {
          result = this.createSecret(data, secret);
        } else {
          result = this.updateSecret(data, secret);
        }
      } else {
        result = this.createFirstSecret(data, secret);
      }
      return result;
    },
  },
};
</script>

<style scoped lang="postcss">
    @import './header.css';
    .biz-templateset-title {
        display: inline-block;
        height: 52px;
        line-height: 52px;
        font-size: 16px;
        vertical-align: middle;
        .bk-form-input {
            margin-top: -2px;
            width: 220px;
        }
        .bcs-icon {
            font-size: 12px;
            visibility: hidden;
        }
        &:hover {
            .bcs-icon {
                visibility: visible;
            }
        }
    }
</style>
