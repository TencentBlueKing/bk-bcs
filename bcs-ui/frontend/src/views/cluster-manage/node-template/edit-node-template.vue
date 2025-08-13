<template>
  <div class="add-node-template" v-bkloading="{ isLoading }">
    <bcs-resize-layout
      placement="right"
      ext-cls="template-resize"
      collapsible
      :initial-divide="400"
      :border="false"
      :min="3"
      disabled>
      <template #aside>
        <ActionDoc class="node-template-aside" :title="$t('cluster.nodeTemplate.title.initConfig')" />
      </template>
      <template #main>
        <bk-form :model="formData" :rules="rules" ref="formRef">
          <FormGroup :title="$t('generic.title.basicInfo1')" :allow-toggle="false">
            <bk-form-item
              :label="$t('cluster.nodeTemplate.label.templateName')"
              required
              property="name"
              error-display-type="normal">
              <bk-input class="mw524" :disabled="isEdit" v-model="formData.name"></bk-input>
            </bk-form-item>
            <bk-form-item
              :label="$t('cluster.create.label.desc')"
              property="desc"
              error-display-type="normal">
              <bk-input
                v-model="formData.desc"
                type="textarea"
                :rows="3"
                :maxlength="100"
                class="mw524"
              ></bk-input>
            </bk-form-item>
          </FormGroup>
          <FormGroup class="mt15" :title="$t('cluster.nodeTemplate.title.labelsAndTaints')" :allow-toggle="false">
            <bk-form-item :label="$t('k8s.label')" property="labels" error-display-type="normal">
              <KeyValue
                class="max-w-[420px]"
                :disable-delete-item="false"
                :min-item="0"
                v-model="formData.labels">
              </KeyValue>
            </bk-form-item>
            <bk-form-item
              :label="$t('k8s.taint')"
              property="taints"
              error-display-type="normal">
              <span class="add-btn" v-if="!formData.taints.length" @click="handleAddTaints">
                <i class="bk-icon icon-plus-circle-shape mr5"></i>
                {{$t('generic.button.add')}}
              </span>
              <Taints
                style="max-width: 600px;"
                v-else
                v-model="formData.taints"
                @add="handleValidateForm"
                @delete="handleValidateForm"
              />
            </bk-form-item>
            <bk-form-item :label="$t('k8s.annotation')" property="annotations" error-display-type="normal">
              <KeyValue
                class="max-w-[420px]"
                :disable-delete-item="false"
                :min-item="0"
                :value-rules="[]"
                v-model="formData.annotations.values">
              </KeyValue>
            </bk-form-item>
          </FormGroup>
          <FormGroup class="mt15" :title="$t('cluster.nodeTemplate.kubelet.title.argsConfig1')" :allow-toggle="false">
            <div style="padding: 0 24px">
              <div class="bcs-flex-between kubelet mb15">
                <span class="left"></span>
                <div class="right">
                  <bcs-input
                    v-model="searchValue"
                    :placeholder="$t('generic.placeholder.params')"
                    right-icon="bk-icon icon-search"
                    clearable>
                  </bcs-input>
                  <i
                    class="bcs-icon bcs-icon-zhongzhishuju ml15"
                    v-bk-tooltips.top="$t('cluster.nodeTemplate.kubelet.button.resetArgs')"
                    @click="handleReset"></i>
                  <i
                    class="bcs-icon bcs-icon-yulan ml15"
                    v-bk-tooltips.top="$t('cluster.nodeTemplate.kubelet.button.preview')"
                    @click="handlePreview"></i>
                </div>
              </div>
              <bcs-table
                :data="curPageData"
                :pagination="pagination"
                v-bkloading="{ isLoading: loading }"
                @row-mouse-enter="handlekubeletMouseEnter"
                @page-change="pageChange"
                @page-limit-change="pageSizeChange">
                <bcs-table-column
                  :label="$t('cluster.nodeTemplate.kubelet.label.argsName')"
                  prop="flagName"></bcs-table-column>
                <bcs-table-column
                  :label="$t('cluster.nodeTemplate.kubelet.label.argsDesc')"
                  prop="flagDesc"
                  show-overflow-tooltip>
                </bcs-table-column>
                <bcs-table-column
                  :label="$t('cluster.nodeTemplate.kubelet.label.defaultValue')"
                  prop="defaultValue"></bcs-table-column>
                <bcs-table-column :label="$t('cluster.nodeTemplate.kubelet.label.curValue')">
                  <template #default="{ row }">
                    <div class="kubelet-value">
                      <InputType
                        v-if="editKey === row.flagName"
                        :type="row.flagType"
                        :options="row.flagValueList"
                        :range="row.range"
                        ref="editInputRef"
                        v-model="kubeletParams[row.flagName]"
                        @blur="handleEditBlur"
                        @enter="handleEditBlur"
                      ></InputType>
                      <template v-else>
                        <span>{{kubeletParams[row.flagName] || '--'}}</span>
                        <i
                          class="bcs-icon bcs-icon-edit2 ml5"
                          v-show="activeKubeletFlagName === row.flagName"
                          @click="handleEditkubelet(row)"></i>
                      </template>
                      <span
                        class="error-tips" v-if="row.regex
                          && kubeletParams[row.flagName]
                          && !new RegExp(row.regex.validator).test(kubeletParams[row.flagName])">
                        <i
                          v-bk-tooltips="row.regex ? row.regex.message : ''"
                          class="bk-icon icon-exclamation-circle-shape"></i>
                      </span>
                    </div>
                  </template>
                </bcs-table-column>
                <template #empty>
                  <BcsEmptyTableStatus :type="searchValue ? 'search-empty' : 'empty'" @clear="searchValue = ''" />
                </template>
              </bcs-table>
            </div>
          </FormGroup>
          <FormGroup class="mt15" :title="$t('cluster.nodeTemplate.title.initConfig')" :allow-toggle="false">
            <bk-form-item
              :label="$t('cluster.nodeTemplate.label.preInstall.title')"
              :desc="$t('cluster.nodeTemplate.label.preInstall.desc')">
              <bcs-input
                type="textarea"
                class="mt10 mw524"
                :rows="6"
                placeholder="#!/bin/bash"
                v-model="formData.preStartUserScript"></bcs-input>
            </bk-form-item>
            <bk-form-item
              :label="$t('cluster.nodeTemplate.label.postInstall.title')"
              :desc="$t('cluster.nodeTemplate.label.postInstall.desc')">
              <bcs-select class="mw524" :clearable="false" v-model="postActionType">
                <bcs-option id="simple" :name="$t('cluster.nodeTemplate.label.postInstall.type.scripts')"></bcs-option>
                <bcs-option id="complex" :name="$t('cluster.nodeTemplate.label.postInstall.type.sops')"></bcs-option>
              </bcs-select>
              <bcs-input
                type="textarea"
                class="mt10 mw524"
                :rows="6"
                placeholder="#!/bin/bash"
                v-model="formData.userScript"
                v-if="postActionType === 'simple'">
              </bcs-input>
            </bk-form-item>
            <bk-form-item :label="$t('cluster.nodeTemplate.sops.label.sops')" v-if="postActionType === 'complex'">
              <div class="sops-wrapper">
                <bcs-select
                  :loading="bkSopsLoading"
                  :clearable="false"
                  class="mw524"
                  searchable
                  style="flex: 1"
                  v-model="bkSopsTemplateID">
                  <bcs-option
                    v-for="item in bkSopsList"
                    :key="item.templateID"
                    :id="item.templateID"
                    :name="item.templateName">
                  </bcs-option>
                </bcs-select>
                <span
                  class="ml10"
                  v-if="templateUrl"
                  v-bk-tooltips.top="$t('cluster.nodeTemplate.sops.tips.gotoSops')"
                  @click="handleGotoSops">
                  <i class="bcs-icon bcs-icon-fenxiang"></i>
                </span>
                <span
                  class="ml10"
                  v-bk-tooltips.top="$t('generic.button.refresh')"
                  @click="handleRefreshList">
                  <i class="bcs-icon bcs-icon-reset"></i>
                </span>
              </div>
              <div class="bk-sops-params mw524" v-bkloading="{ isLoading: sopsParamsLoading }">
                <div class="title">
                  <span
                    v-bk-tooltips.top="{ content: $t('cluster.nodeTemplate.sops.title.taskArgs.tips') }"
                    class="name">
                    {{$t('cluster.nodeTemplate.sops.title.taskArgs.text')}}
                  </span>
                </div>
                <div class="content">
                  <div class="content-item mb15" v-for="item in sopsParamsList" :key="item.key">
                    <div class="content-item-label">
                      <span
                        :class="{ desc: !!item.desc }"
                        v-bk-tooltips.top="{
                          content: item.desc,
                          disabled: !item.desc
                        }"
                      >{{item.name}}</span>
                    </div>
                    <bcs-input
                      behavior="simplicity"
                      :placeholder="$t('cluster.nodeTemplate.sops.placeholder.input')"
                      v-model="sopsParams[item.key]">
                    </bcs-input>
                  </div>
                  <span
                    v-bk-tooltips="{
                      disabled: !isSopsParamsExitVar,
                      content: $t('cluster.nodeTemplate.sops.button.debug.tips')
                    }">
                    <bcs-button
                      theme="primary"
                      outline
                      :disabled="isSopsParamsExitVar"
                      @click="handleDebug">
                      {{$t('cluster.nodeTemplate.sops.button.debug.text')}}
                    </bcs-button>
                  </span>
                </div>
              </div>
            </bk-form-item>
          </FormGroup>
        </bk-form>
        <bcs-dialog
          :title="$t('cluster.nodeTemplate.kubelet.button.preview')"
          :show-footer="false"
          header-position="left"
          width="640"
          v-model="showPreview">
          <bcs-table :data="kubeletDiffData" :key="JSON.stringify(kubeletDiffData)">
            <bcs-table-column :label="$t('plugin.tools.toolName')" prop="moduleID"></bcs-table-column>
            <bcs-table-column
              :label="$t('cluster.nodeTemplate.kubelet.label.flagName')"
              prop="flagName"></bcs-table-column>
            <bcs-table-column :label="$t('cluster.nodeTemplate.kubelet.label.beforeEdit')" prop="origin">
              <template #default="{ row }">
                {{row.origin || '--'}}
              </template>
            </bcs-table-column>
            <bcs-table-column
              :label="$t('cluster.nodeTemplate.kubelet.label.afterEdit')"
              prop="value"></bcs-table-column>
          </bcs-table>
        </bcs-dialog>
        <!-- 任务调试状态 -->
        <bcs-dialog
          :show-footer="false"
          :mask-close="false"
          width="400"
          v-model="showDebugStatus"
          :on-close="handleDebugDialogClose">
          <div class="task-status">
            <div
              class="loading-icon"
              v-show="['INITIALIZING', 'RUNNING'].includes(taskData.status)"
              v-bkloading="{
                isLoading: ['INITIALIZING', 'RUNNING'].includes(taskData.status),
                opacity: 1,
                theme: 'primary',
                mode: 'spin'
              }"></div>
            <template v-if="['INITIALIZING', 'RUNNING'].includes(taskData.status)">
              <div class="title mt15">{{$t('cluster.nodeTemplate.sops.status.running.text')}}...</div>
              <div class="operator mt15">
                <bcs-button
                  text
                  size="small"
                  :disabled="!taskUrl"
                  @click="handleGotoTaskDetail">
                  {{$t('cluster.nodeTemplate.sops.status.running.detailBtn')}}
                </bcs-button>
              </div>
            </template>
            <template v-else-if="taskData.status === 'SUCCESS'">
              <div class="bcs-flex-center">
                <span class="status-icon success"><i class="bcs-icon bcs-icon-check-1"></i></span>
              </div>
              <div class="title mt20">{{$t('cluster.nodeTemplate.sops.status.success')}}</div>
              <div class="operator mt20">
                <bcs-button
                  class="mw88"
                  theme="primary"
                  :disabled="!taskUrl"
                  @click="handleGotoTaskDetail"
                >{{$t('cluster.nodeTemplate.sops.status.running.detailBtn')}}</bcs-button>
                <bcs-button
                  class="ml5"
                  style="min-width: 88px;"
                  @click="showDebugStatus = false">{{$t('generic.status.done')}}</bcs-button>
              </div>
            </template>
            <template v-else-if="taskData.status === 'FAILURE'">
              <div class="bcs-flex-center">
                <span class="status-icon failure"><i class="bcs-icon bcs-icon-close"></i></span>
              </div>
              <div class="title mt20">{{$t('cluster.nodeTemplate.sops.status.failed.text')}}</div>
              <div class="operator mt20">
                <bcs-button
                  class="mw88"
                  theme="primary"
                  :disabled="!taskUrl"
                  @click="handleGotoTaskDetail"
                >{{$t('cluster.nodeTemplate.sops.status.running.detailBtn')}}</bcs-button>
                <bcs-button
                  class="mw88 ml5"
                  theme="primary"
                  @click="handleDebug">{{$t('cluster.nodeTemplate.sops.button.retry')}}</bcs-button>
              </div>
            </template>
          </div>
        </bcs-dialog>
      </template>
    </bcs-resize-layout>
    <div class="bcs-fixed-footer">
      <bcs-button class="mw88" theme="primary" :loading="btnLoading" @click="handleCreateOrUpdate">
        {{isEdit ? $t('generic.button.save') : $t('generic.button.create')}}
      </bcs-button>
      <bcs-button class="mw88 ml10" @click="handleCancel">{{$t('generic.button.cancel')}}</bcs-button>
    </div>
  </div>
</template>
<script lang="ts">
import { cloneDeep } from 'lodash';
import { computed, defineComponent, getCurrentInstance, onMounted, ref, watch } from 'vue';
import xss from 'xss';

import ActionDoc from '../components/action-doc.vue';

import $bkMessage from '@/common/bkmagic';
import FormGroup from '@/components/form-group.vue';
import useInterval from '@/composables/use-interval';
import usePage from '@/composables/use-page';
import useSearch from '@/composables/use-search';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $store from '@/store/index';
import InputType from '@/views/cluster-manage/components/input-type.vue';
import KeyValue from '@/views/cluster-manage/components/key-value.vue';
import Taints from '@/views/cluster-manage/components/new-taints.vue';

export default defineComponent({
  components: { FormGroup, KeyValue, Taints, InputType, ActionDoc },
  props: {
    nodeTemplateID: {
      type: [String, Number],
      default: '',
    },
  },
  setup(props) {
    const curProject = computed(() => $store.state.curProject);
    const user = computed(() => $store.state.user);
    const isEdit = computed(() => !!props.nodeTemplateID);

    const postActionType = ref<'complex' | 'simple'>('simple');
    watch(postActionType, () => {
      if (postActionType.value === 'complex' && !bkSopsList.value.length) {
        handleGetbkSopsList();
      }
    });
    const formRef = ref<any>(null);
    const formData = ref({
      projectID: curProject.value.project_id,
      name: '',
      desc: '',
      labels: {},
      taints: [],
      annotations: { values: {} },
      preStartUserScript: '',
      userScript: '',
      extraArgs: {
        kubelet: '',
      },
      scaleOutExtraAddons: {
        plugins: {},
      },
    });
    const rules = ref({
      name: [{
        required: true,
        message: $i18n.t('generic.validate.required'),
        trigger: 'blur',
      }],
      labels: [
        {
          message: $i18n.t('generic.validate.labelValueEmpty'),
          trigger: 'custom',
          // eslint-disable-next-line max-len
          validator: () => Object.keys(formData.value.labels).every(key => !!formData.value.labels[key]),
        },
        {
          message: $i18n.t('generic.validate.labelKey'),
          trigger: 'custom',
          validator: () => {
            const keys = Object.keys(formData.value.labels);
            const values = keys.map(key => formData.value.labels[key]);
            return keys.every(v => /^[A-Za-z0-9._/-]+$/.test(v)) && values.every(v => /^[A-Za-z0-9._/-]+$/.test(v));
          },
        },
      ],
      taints: [
        {
          validator: () => formData.value.taints.every((item: any) => item.key && item.value && item.effect),
          message: $i18n.t('cluster.nodeTemplate.validate.keyValue'),
          trigger: 'custom',
        },
        {
          message: $i18n.t('generic.validate.repeatKey'),
          trigger: 'custom',
          validator: () => {
            const data = (formData.value.taints as any[]).reduce((pre, item) => {
              if (item.key) {
                pre.push(item.key);
              }
              return pre;
            }, []);
            const removeDuplicateData = new Set(data);
            return data.length === removeDuplicateData.size;
          },
        },
        {
          message: $i18n.t('generic.validate.labelKey'),
          trigger: 'custom',
          validator: () => (formData.value.taints as any[])
            .every(item => /^[A-Za-z0-9._/-]+$/.test(item.key) && /^[A-Za-z0-9._/-]+$/.test(item.value)),
        },
      ],
      annotations: [
        {
          message: $i18n.t('generic.validate.annotationValueEmpty'),
          trigger: 'custom',
          // eslint-disable-next-line max-len
          validator: () => Object.keys(formData.value.annotations.values).every(key => !!formData.value.annotations.values[key]),
        },
        {
          message: $i18n.t('generic.validate.labelKey'),
          trigger: 'custom',
          validator: () => {
            const keys = Object.keys(formData.value.annotations.values);
            return keys.every(v => /^[A-Za-z0-9._/-]+$/.test(v));
          },
        },
      ],
    });
    const handleValidateForm = async () => {
      await formRef.value?.validate();
    };

    // kubelet 组件参数
    const loading = ref(false);
    const editKey = ref('');
    const showPreview = ref(false);
    const kubeletParams = ref({});
    const originKubeletParams = ref<any>({});
    const kubeletDiffData = computed(() => Object.keys(kubeletParams.value).reduce<any[]>((pre, key) => {
      if (kubeletParams.value[key] !== ''
        && kubeletParams.value[key] !== originKubeletParams.value[key]) {
        pre.push({
          moduleID: 'kubelet',
          flagName: key,
          origin: originKubeletParams.value[key],
          value: kubeletParams.value[key],
        });
      }
      return pre;
    }, []));
    const kubeletList = ref<any[]>([]);
    const handleGetkubeletData = async () => {
      loading.value = true;
      kubeletList.value = await $store.dispatch('clustermanager/cloudModulesParamsList', {
        $cloudID: 'tencentCloud',
        $version: '1.20.6',
        $moduleID: 'kubelet',
      });
      loading.value = false;
    };
    const keys = ref(['flagName']);
    const { searchValue, tableDataMatchSearch } = useSearch(kubeletList, keys);
    const {
      pagination,
      curPageData,
      pageChange,
      pageSizeChange,
    } = usePage(tableDataMatchSearch);
    const editInputRef = ref<any>(null);
    const activeKubeletFlagName = ref('');
    const handlekubeletMouseEnter = (index, event, row) => {
      activeKubeletFlagName.value = row.flagName;
    };
    const { proxy } = getCurrentInstance() || { proxy: null };
    const handleEditkubelet = (row) => {
      editKey.value = row.flagName;
      const $refs = proxy?.$refs || {};
      setTimeout(() => {
        ($refs.editInputRef as any)?.focus();
      }, 0);
    };
    const handleEditBlur = () => {
      editKey.value = '';
    };
    const handleReset = () => {
      kubeletParams.value = JSON.parse(JSON.stringify(originKubeletParams.value));
    };
    const handlePreview = () => {
      showPreview.value = true;
    };
    // 校验kubelet参数
    const validateKubeletParams = () => kubeletList.value.every((item) => {
      if (!kubeletParams.value[item.flagName] || !item.regex?.validator) return true;

      const regx = new RegExp(item.regex.validator);
      return regx.test(kubeletParams.value[item.flagName]);
    });

    // 添加污点
    const handleAddTaints = () => {
      (formData.value.taints as any[]).push({
        key: '',
        value: '',
        effect: 'PreferNoSchedule',
      });
    };

    // 获取标准运维任务
    const bkSopsLoading = ref(false);
    const bkSopsList = ref<any[]>([]);
    const bkSopsTemplateID = ref('');
    watch(bkSopsTemplateID, () => {
      if (!bkSopsTemplateID.value) return;
      // 清空数据
      sopsParams.value = {};
      sopsParamsList.value = [];
      handleGetSopsParams();
    });
    const handleGetbkSopsList = async () => {
      bkSopsLoading.value = true;
      bkSopsList.value = await $store.dispatch('clustermanager/bkSopsList', {
        $businessID: curProject.value.businessID,
        operator: user.value.username,
        templateSource: 'business',
        scope: 'cmdb_biz',
      });
      if (!bkSopsTemplateID.value) {
        bkSopsTemplateID.value = bkSopsList.value[0]?.templateID;
      }
      bkSopsLoading.value = false;
    };
    const handleRefreshList = async () => {
      await handleGetbkSopsList();
      await handleGetSopsParams();
    };
    const sopsParamsLoading = ref(false);
    const sopsParams = ref({});
    const isSopsParamsExitVar = computed(() => Object.values(sopsParams.value).some(value => /{{.*}}/.test(value as string)));
    const sopsParamsList = ref<any[]>([]);
    const templateUrl = ref('');
    const handleGetSopsParams = async () => {
      sopsParamsLoading.value = true;
      const data = await $store.dispatch('clustermanager/bkSopsParamsList', {
        $templateID: bkSopsTemplateID.value,
        $businessID: curProject.value.businessID,
        operator: user.value.username,
        templateSource: 'business',
        scope: 'cmdb_biz',
      });
      sopsParamsList.value = data.values;
      // 优先还原历史详情数据
      sopsParams.value = JSON.parse(JSON.stringify(
        formData.value.scaleOutExtraAddons?.plugins?.[bkSopsTemplateID.value]?.params
        || data.values.reduce((pre, item) => {
          pre[item.key] = '';
          return pre;
        }, {}),
        (key, value) => {
          if (['template_biz_id', 'template_id', 'template_user'].includes(key)) {
            return undefined;
          }
          return value;
        },
      ));
      templateUrl.value = data.templateUrl;
      sopsParamsLoading.value = false;
    };
    const handleGotoSops = () => {
      window.open(templateUrl.value);
    };
    // 调试标准运维任务
    const showDebugStatus = ref(false);
    const taskData = ref<any>({});
    const taskUrl = computed(() => {
      const [stepID] = taskData.value.stepSequence || [];
      return taskData.value?.steps?.[stepID]?.params?.taskUrl;
    });
    const handleDebugDialogClose = () => {
      taskData.value = {};
      stop();
    };
    const handlePollTask = async () => {
      taskData.value = await $store.dispatch('clustermanager/taskDetail', {
        $taskId: taskData.value.taskID,
      });
      if (['SUCCESS', 'FAILURE'].includes(taskData.value.status)) {
        stop();
      }
    };
    const { start, stop } = useInterval(handlePollTask, 5000, true);
    const handleDebug = async () => {
      const { task } = await $store.dispatch('clustermanager/bkSopsDebug', {
        businessID: String(curProject.value.businessID),
        templateID: String(bkSopsTemplateID.value),
        operator: user.value.username,
        templateSource: 'business',
        constant: {
          ...sopsParams.value,
        },
      });
      taskData.value = task || {};
      if (taskData.value.taskID) {
        showDebugStatus.value = true;
        start();
      }
    };
    // 跳转任务详情
    const handleGotoTaskDetail = () => {
      window.open(taskUrl.value);
    };


    // 创建和更新节点模板
    const btnLoading = ref(false);
    const handleCreateOrUpdate = async () => {
      const validate = await formRef.value?.validate();
      const validateKubelet = validateKubeletParams();
      if (!validate || !validateKubelet) return;

      btnLoading.value = true;
      // 后置初始化参数处理
      const data: Record<string, any> = {
        extraArgs: {
          kubelet: handleTransformParamsToKubelet(kubeletParams.value),
        },
      };
      if (postActionType.value === 'complex') {
        data.userScript = '';
        data.scaleOutExtraAddons = {
          postActions: [bkSopsTemplateID.value],
          plugins: {
            [bkSopsTemplateID.value]: {
              params: {
                template_biz_id: String(curProject.value.businessID),
                template_id: bkSopsTemplateID.value,
                template_user: user.value.username,
                ...sopsParams.value,
              },
            },
          },
        };
      } else {
        data.scaleOutExtraAddons = {};
      }

      // xss
      const cloneFormData = cloneDeep(formData.value);
      const xssDesc = xss(cloneFormData.desc);
      if (cloneFormData.desc !== xssDesc) {
        console.warn('Intercepted by XSS');
      }
      cloneFormData.desc = xssDesc;

      let result = false;
      if (isEdit.value) {
        result = await $store.dispatch('clustermanager/updateNodeTemplate', {
          $nodeTemplateId: props.nodeTemplateID,
          ...cloneFormData,
          ...data,
          updater: user.value.username,
        });
      } else {
        result = await $store.dispatch('clustermanager/createNodeTemplate', {
          ...cloneFormData,
          ...data,
          creator: user.value.username,
        });
      }
      if (result) {
        $bkMessage({
          theme: 'success',
          message: isEdit.value ? $i18n.t('generic.msg.success.edit') : $i18n.t('generic.msg.success.create'),
        });
        $router.push({ name: 'nodeTemplate' });
      }
      btnLoading.value = false;
    };
    const handleCancel = () => {
      $router.back();
    };

    // 获取详情
    const handleTransformKubeletToParams = (kubelet = '') => {
      if (!kubelet) return {};

      return kubelet.split(';').reduce((pre, current) => {
        const index = current.indexOf('=');
        const key = current.slice(0, index);
        const value = current.slice(index + 1, current.length);
        if (key) {
          pre[key] = value;
        }
        return pre;
      }, {}) || {};
    };
    const handleTransformParamsToKubelet = (params = {}) => Object.keys(params || {})
      .filter(key => params[key] !== '')
      .reduce<string[]>((pre, key) => {
      pre.push(`${key}=${params[key]}`);
      return pre;
    }, [])
      .join(';');
    const isLoading = ref(false);
    const handleGetDetail = async () => {
      if (!isEdit.value) return;

      isLoading.value = true;
      const data = await $store.dispatch('clustermanager/nodeTemplateDetail', {
        $nodeTemplateId: props.nodeTemplateID,
      });
      formData.value = {
        ...data,
        // 传给后端的结构和返回的结构不一致
        annotations: {
          values: data.annotations,
        },
      };
      // 处理标准运维相关回显参数
      postActionType.value = data.scaleOutExtraAddons?.postActions?.length ? 'complex' : 'simple';
      // eslint-disable-next-line camelcase, max-len
      bkSopsTemplateID.value = data.scaleOutExtraAddons?.plugins?.[data.scaleOutExtraAddons?.postActions?.[0]]?.params?.template_id;
      // 转换kubelet参数，便于回显
      kubeletParams.value = handleTransformKubeletToParams(formData.value?.extraArgs?.kubelet);
      // kubelet原始数据（用于diff）
      originKubeletParams.value = JSON.parse(JSON.stringify(kubeletParams.value));
      isLoading.value = false;
    };
    onMounted(() => {
      handleGetkubeletData();
      handleGetDetail();
    });
    return {
      editInputRef,
      formRef,
      isLoading,
      btnLoading,
      bkSopsTemplateID,
      bkSopsLoading,
      sopsParamsLoading,
      sopsParamsList,
      sopsParams,
      templateUrl,
      bkSopsList,
      isEdit,
      postActionType,
      loading,
      rules,
      formData,
      searchValue,
      pagination,
      curPageData,
      editKey,
      showPreview,
      kubeletParams,
      pageChange,
      pageSizeChange,
      handleAddTaints,
      handleCancel,
      handleEditkubelet,
      handleCreateOrUpdate,
      handlePreview,
      handleReset,
      handleEditBlur,
      handlekubeletMouseEnter,
      activeKubeletFlagName,
      kubeletDiffData,
      handleDebug,
      showDebugStatus,
      taskData,
      taskUrl,
      handleDebugDialogClose,
      handleGotoTaskDetail,
      handleGetbkSopsList,
      handleRefreshList,
      isSopsParamsExitVar,
      handleGotoSops,
      handleValidateForm,
    };
  },
});
</script>
<style lang="postcss" scoped>
.add-node-template {
    padding: 24px 0 24px 24px;
    height: 100%;
    overflow: auto;
    >>> .mw524 {
        max-width: 524px;
    }
    >>> .mw920 {
        max-width: 920px;
    }
    >>> .add-btn {
        font-size: 14px;
        color: #3a84ff;
        cursor: pointer;
        display: flex;
        align-items: center;
        height: 32px;
    }
    >>> .sops-wrapper {
        display: flex;
        align-items: center;
        .bcs-icon {
            color: #3a84ff;
            cursor: pointer;
        }
    }
    .node-template-aside {
        border: 1px solid #dcdee5;
        border-left: none;
        height: calc(100% - 60px);
        overflow: auto;
        background: #fff;
    }
    .mw88 {
        min-width: 88px;
    }
    .kubelet-value {
        position: relative;
        height: 32px;
        display: flex;
        align-items: center;
        .bcs-icon-edit2 {
            cursor: pointer;
            &:hover {
                color: #3a84ff;
            }
        }
        .error-tips {
            position: absolute;
            z-index: 10;
            right: 8px;
            top: 8px;
            color: #ea3636;
            cursor: pointer;
            font-size: 16px;
            display: flex;
            background-color: #fff;
        }
    }
    .kubelet {
        .left {
            font-weight: Bold;
            font-size: 14px;
        }
        .right {
            display: flex;
            align-items: center;
            min-width: 300px;
            i {
                font-size: 14px;
                cursor: pointer;
                &:hover {
                    color: #3a84ff;
                }
            }
        }
    }
    >>> .bk-sops-params {
        margin-top: 12px;
        border: 1px solid #DCDEE5;
        border-radius: 2px;
        .title {
            background: #F5F7FA;
            border-bottom: 1px solid #DCDEE5;
            height: 36px;
            padding: 0 16px;
            display: flex;
            align-items: center;
            .name {
              border-bottom: 1px dashed #979ba5;
              line-height: 20px;
            }
        }
        .content {
            padding: 20px 16px;
            &-item-label {
                padding-left: 10px;
                line-height: 1;
                .desc {
                    border-bottom: 1px dashed #979ba5;
                    display: inline-block;
                    padding-bottom: 2px;
                }
            }
        }
    }
    .bcs-icon-copy:hover {
        color: #3a84ff;
        cursor: pointer;
    }
}
>>> .task-status {
    .loading-icon {
        height: 70px;
    }
    .title {
        font-size: 20px;
        color: #313238;
        text-align: center;
    }
    .sub-title {
        text-align: center;
        font-size: 14px;
        color: #63656E;
    }
    .status-icon {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 42px;
        height: 42px;
        border-radius: 50%;
        i {
            font-weight: bold;
        }
        &.success {
            background-color: #E5F6EA;
            color: #3FC06D;
        }
        &.failure {
            background-color: #FFDDDD;
            color: #EA3636;
        }
    }
    .operator {
        display: flex;
        justify-content: center;
    }
}
>>> .template-resize {
    height: 100%;
    .bk-resize-layout-aside:after {
        width: 0;
    }
    .bk-resize-layout-aside {
        margin: 0 24px;
    }
    .bk-resize-layout-main {
        height: calc(100% - 60px);
        overflow: auto;
    }
    &.bk-resize-layout-collapsed {
        .bk-resize-layout-aside {
            border-left: none;
            margin-right: 0px;
        }
    }
}
</style>
