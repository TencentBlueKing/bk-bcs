<template>
  <div>
    <div class="sops-wrapper">
      <bcs-select
        :loading="bkSopsLoading"
        :clearable="false"
        class="max-w-[524px]"
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
    <div class="bk-sops-params max-w-[600px]" v-bkloading="{ isLoading: sopsParamsLoading }">
      <div class="title">
        <span v-bk-tooltips.top="{ content: $t('cluster.nodeTemplate.sops.title.taskArgs.tips') }" class="name">
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
        <bcs-button
          theme="primary"
          outline
          @click="handleBeforeDebug">
          {{$t('cluster.nodeTemplate.sops.button.debug.text')}}
        </bcs-button>
      </div>
    </div>
    <!-- 任务参数 -->
    <bcs-dialog :title="$t('cluster.nodeTemplate.sops.title.confirmInnerVar')" v-model="showTaskParams">
      <div v-bkloading="{ isLoading: templateValuesLoading }">
        <div
          class="content-item mb15"
          v-for="item in variableList"
          :key="item.key">
          <div class="content-item-label">
            <span>{{item.key}}</span>
          </div>
          <div class="relative">
            <bcs-input
              behavior="simplicity"
              :disabled="item.disabled"
              v-model="variableParams[item.key]">
            </bcs-input>
            <i
              v-if="!variableParams[item.key]"
              v-bk-tooltips="$t('generic.validate.required')"
              class="text-[red] absolute top-[10px] right-[5px] bk-icon icon-exclamation-circle-shape"></i>
          </div>
        </div>
      </div>
      <template #footer>
        <div>
          <bcs-button theme="primary" @click="handleDebug">{{ $t('generic.button.confirm1') }}</bcs-button>
          <bcs-button @click="showTaskParams = false">{{ $t('generic.button.cancel') }}</bcs-button>
        </div>
      </template>
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
            <span
              v-bk-tooltips="{
                content: $t('cluster.nodeTemplate.sops.status.failed.tips'),
                disabled: !!taskUrl
              }">
              <bcs-button
                class="mw88"
                theme="primary"
                :disabled="!taskUrl"
                @click="handleGotoTaskDetail"
              >{{$t('cluster.nodeTemplate.sops.status.running.detailBtn')}}</bcs-button>
            </span>
            <bcs-button
              class="mw88 ml5"
              theme="primary"
              @click="handleBeforeDebug">{{$t('cluster.nodeTemplate.sops.button.retry')}}</bcs-button>
          </div>
        </template>
      </div>
    </bcs-dialog>
  </div>
</template>
<script lang="ts">
import { computed, defineComponent, onMounted, ref, set, toRefs, watch } from 'vue';

import useInterval from '@/composables/use-interval';
import $store from '@/store/index';

export default defineComponent({
  name: 'BkSops',
  props: {
    addons: {
      type: Object,
      default: () => null,
    },
    actionsKey: {
      type: String,
      required: true,
      default: 'preActions',
      // validator(value: string) {
      //   return [
      //     'preActions',
      //     'postActions',
      //   ].indexOf(value) > -1;
      // },
    },
    clusterId: {
      type: String,
      default: '',
    },
    allowSkipWhenFailed: {
      type: Boolean,
      default: false,
    },
  },
  setup(props) {
    const { addons, actionsKey, clusterId, allowSkipWhenFailed } = toRefs(props);
    const curProject = computed(() => $store.state.curProject);
    const user = computed(() => $store.state.user);

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
      // 详情数据
      sopsParams.value = JSON.parse(JSON.stringify(
        addons.value?.plugins?.[bkSopsTemplateID.value]?.params
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
    const showTaskParams = ref(false);
    const variableParams = ref({});
    const variableList = ref<any[]>([]);
    const templateValuesLoading = ref(false);
    const templateValues = ref<any[]>();
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
    const handleBeforeDebug = async () => {
      variableParams.value = {};
      variableList.value = [];
      const variable = Object.keys(sopsParams.value).filter(key => /^{{.+}}$/.test(sopsParams.value[key]));
      if (variable.length) {
        // 有内置变量，默认渲染
        showTaskParams.value = true;
        templateValuesLoading.value = true;
        templateValues.value =  await $store.dispatch('clustermanager/bkSopsTemplatevalues', {
          clusterID: clusterId.value,
          operator: user.value.username,
        });
        templateValuesLoading.value = false;
        variable.forEach((key) => {
          const template = templateValues.value?.find(item => item.refer === sopsParams.value[key]);
          set(variableParams.value, key, template?.value);
          variableList.value.push({
            key,
            disabled: !!template?.value,
          });
        });
        // 排序
        variableList.value = variableList.value.sort(item => (item.disabled ? 0 : -1));
      } else {
        handleDebug();
      }
    };
    const handleDebug = async () => {
      // 校验数据
      const keys = Object.keys(variableParams.value);
      if (keys.length && keys.some(key => !variableParams.value[key])) return;

      const { task } = await $store.dispatch('clustermanager/bkSopsDebug', {
        businessID: String(curProject.value.businessID),
        templateID: String(bkSopsTemplateID.value),
        operator: user.value.username,
        templateSource: 'business',
        constant: {
          ...sopsParams.value,
          ...variableParams.value,
        },
      });
      taskData.value = task || {};
      if (taskData.value.taskID) {
        showTaskParams.value = false;
        showDebugStatus.value = true;
        start();
      }
    };
    // 跳转任务详情
    const handleGotoTaskDetail = () => {
      window.open(taskUrl.value);
    };

    const bkSopsData = computed(() => ({
      [actionsKey.value]: [bkSopsTemplateID.value],
      plugins: {
        [bkSopsTemplateID.value]: {
          params: {
            template_biz_id: String(curProject.value.businessID),
            template_id: bkSopsTemplateID.value,
            template_user: user.value.username,
            ...sopsParams.value,
          },
          allowSkipWhenFailed: allowSkipWhenFailed.value,
        },
      },
    }));

    onMounted(() => {
      // eslint-disable-next-line camelcase, max-len
      bkSopsTemplateID.value = addons.value?.plugins?.[addons.value?.[actionsKey.value]?.[0]]?.params?.template_id;
      handleGetbkSopsList();
    });

    return {
      variableList,
      variableParams,
      templateValuesLoading,
      showTaskParams,
      bkSopsTemplateID,
      bkSopsLoading,
      sopsParamsLoading,
      sopsParamsList,
      sopsParams,
      templateUrl,
      bkSopsList,
      handleDebug,
      showDebugStatus,
      taskData,
      taskUrl,
      bkSopsData,
      handleDebugDialogClose,
      handleGetbkSopsList,
      handleRefreshList,
      isSopsParamsExitVar,
      handleGotoSops,
      handleGotoTaskDetail,
      handleBeforeDebug,
    };
  },
});
</script>
<style lang="postcss" scoped>
.sops-wrapper {
  display: flex;
  align-items: center;
  .bcs-icon {
      color: #3a84ff;
      cursor: pointer;
  }
}
.bk-sops-params {
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
.task-status {
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
</style>
