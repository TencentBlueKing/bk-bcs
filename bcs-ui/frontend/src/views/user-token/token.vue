<template>
  <div class="user-token">
    <div class="user-token-header">
      <span>
        <i class="bcs-icon bcs-icon-arrows-left back" @click="goBack"></i>
        <span class="title">{{$t('apiToken.text')}}</span>
      </span>
      <a class="bk-text-button help" :href="PROJECT_CONFIG.token" target="_blank">
        {{ $t('apiToken.action.link.Instructions') }}
      </a>
    </div>
    <bcs-alert type="info" class="mb15">
      <template #title>
        <div class="info-item">1. {{$t('apiToken.desc.article1')}}</div>
        <div class="info-item">2.
          <i18n path="apiToken.desc.article2">
            <a class="bk-text-button" :href="BK_IAM_HOST" target="_blank">
              {{ $t('apiToken.action.link.iam') }}
            </a>
          </i18n>
        </div>
        <div class="info-item">3. {{$t('apiToken.desc.article3')}}</div>
      </template>
    </bcs-alert>
    <bk-table :data="data" v-bkloading="{ isLoading: loading }">
      <bk-table-column :label="$t('deploy.helm.username')">
        <template #default>
          <bk-user-display-name :user-id="user.username"></bk-user-display-name>
        </template>
      </bk-table-column>
      <bk-table-column :label="$t('apiToken.text')" min-width="300">
        <template #default="{ row }">
          <div class="token-row">
            <span>{{hiddenToken ? new Array(12).fill('*').join('') : row.token}}</span>
            <i
              :class="['ml10 bcs-icon', `bcs-icon-${hiddenToken ? 'eye-slash' : 'eye'}`]"
              @click="toggleHiddenToken"></i>
            <i class="ml10 bcs-icon bcs-icon-copy" @click="handleCopyToken(row)"></i>
          </div>
        </template>
      </bk-table-column>
      <bk-table-column :label="$t('apiToken.label.expiredTime')" prop="expired_at">
        <template #default="{ row }">
          <div>{{!row.expired_at ? $t('apiToken.label.perpetual') : row.expired_at}}</div>
        </template>
      </bk-table-column>
      <bk-table-column :label="$t('generic.label.status')">
        <template #default="{ row }">
          <StatusIcon
            :status="String(row.status)"
            :status-color-map="{
              '1': 'green',
              '0': 'gray'
            }">
            {{row.status === 1 ? $t('generic.status.ready') : $t('apiToken.status.expired')}}
          </StatusIcon>
        </template>
      </bk-table-column>
      <bk-table-column :label="$t('generic.label.action')" width="150">
        <template #default="{ row }">
          <bk-button
            class="mr10"
            theme="primary"
            text
            :disabled="!row.expired_at"
            @click="handleRenewalToken(row)"
          >{{$t('apiToken.action.renewal')}}</bk-button>
          <bk-button
            theme="primary" text
            @click="handleDeleteToken(row)">{{$t('generic.button.delete')}}</bk-button>
        </template>
      </bk-table-column>
      <template #empty>
        <bcs-exception type="empty" scene="part">
          <div>{{$t('apiToken.msg.emptyDataTips')}}</div>
          <bcs-button
            class="create-token-btn" icon="plus" theme="primary"
            @click="handleCreateToken">
            {{$t('apiToken.button.newToken')}}
          </bcs-button>
        </bcs-exception>
      </template>
    </bk-table>
    <!-- 独立集群使用案例 -->
    <div class="user-token-example" v-if="data.length">
      <div class="example-item">
        <div class="title total-title">{{$t('apiToken.example.title1')}}:</div>
        <div class="code-wrapper">
          <p>kubectl:</p>
          <br>
          {{kubeConfigExample}}
        </div>
      </div>
      <div class="example-item">
        <div class="title">{{$t('apiToken.example.kubeConfigPath')}}:</div>
        <div class="code-wrapper">
          <CodeEditor
            :height="330"
            v-full-screen="{ tools: ['copy'], content: demoConfigExample }"
            readonly
            :value="demoConfigExample"
            :options="options"
            width="100%">
          </CodeEditor>
        </div>
      </div>
      <div class="example-item">
        <div class="title">{{$t('apiToken.bcsAPI')}}:</div>
        <div class="code-wrapper">
          <CodeEditor
            :height="50"
            v-full-screen="{ tools: ['copy'], content: bcsApiExample }"
            readonly
            :value="bcsApiExample"
            :options="options"
            width="100%">
          </CodeEditor>
        </div>
      </div>
    </div>

    <!-- 共享集群使用案例 -->
    <div class="user-token-example mt50" v-if="data.length && hasSharedCluster">
      <div class="example-item">
        <div class="title total-title">{{$t('apiToken.example.title2')}}:</div>
        <div class="code-wrapper">
          <p>kubectl:</p>
          <br>
          {{shareKubeConfigExample}}
        </div>
      </div>
      <div class="example-item">
        <div class="title">{{$t('apiToken.example.kubeConfigPath')}}:</div>
        <div class="code-wrapper">
          <CodeEditor
            :height="330"
            v-full-screen="{ tools: ['copy'], content: shareDemoConfigExample }"
            readonly
            :value="shareDemoConfigExample"
            :options="options"
            width="100%">
          </CodeEditor>
        </div>
      </div>
      <div class="example-item">
        <div class="title">{{$t('apiToken.bcsAPI')}}:</div>
        <div class="code-wrapper">
          <CodeEditor
            :height="50"
            v-full-screen="{ tools: ['copy'], content: shareBcsApiExample }"
            readonly
            :value="shareBcsApiExample"
            :options="options"
            width="100%">
          </CodeEditor>
        </div>
      </div>
    </div>
    <!-- 创建 or 续期 token -->
    <bcs-dialog
      v-model="showTokenDialog"
      theme="primary"
      :mask-close="false"
      header-position="left"
      :title="operateType === 'create' ? $t('apiToken.button.newToken') : $t('apiToken.title.renewToken')"
      width="640"
      :loading="updateLoading"
      @confirm="confirmUpdateTokenDialog"
      @cancel="cancelUpdateTokenDialog">
      <div class="create-token-dialog">
        <div class="title">{{$t('apiToken.title.deadLine')}}</div>
        <div class="bk-button-group">
          <bk-button
            v-for="item in timeList"
            :key="item.id"
            :class="['group-btn', { 'is-selected': item.id === active }]"
            @click="handleSelectTime(item)"
          >
            {{item.name}}
          </bk-button>
          <bcs-input
            type="number"
            :min="1"
            :max="1000"
            :precision="0"
            :show-controls="false"
            class="custom-input"
            ref="customInputRef"
            v-model="active"
            v-if="isCustomTime">
            <template slot="append">
              <div class="custom-input-append">{{$t('units.suffix.days')}}</div>
            </template>
          </bcs-input>
          <bk-button class="group-btn" v-else @click="handleCustomTime">{{$t('generic.label.custom')}}</bk-button>
        </div>
      </div>
    </bcs-dialog>
    <!-- 删除 token -->
    <bcs-dialog
      v-model="showDeleteDialog"
      theme="primary"
      header-position="left"
      :title="$t('apiToken.title.deleteToken')"
      width="640">
      <div class="delete-token-dialog">
        <div class="title">{{$t('apiToken.subTitle.deleteConfirm.text')}}:</div>
        <bcs-checkbox v-model="deleteConfirm">
          {{$t('apiToken.subTitle.deleteConfirm.desc')}}
        </bcs-checkbox>
      </div>
      <template #footer>
        <div>
          <bcs-button
            :disabled="!deleteConfirm"
            theme="primary"
            :loading="deleteLoading"
            @click="confirmDeleteToken"
          >{{$t('generic.button.confirm')}}</bcs-button>
          <bcs-button @click="cancelDeleteDialog">{{$t('generic.button.cancel')}}</bcs-button>
        </div>
      </template>
    </bcs-dialog>
  </div>
</template>
<script lang="ts">
import clusterDemoConfig from 'text-loader?modules!./cluster-demo.yaml';
import shareClusterDemoConfig from 'text-loader?modules!./share-cluster-demo.yaml';
import { computed, defineComponent, onMounted, ref } from 'vue';

import $bkMessage from '@/common/bkmagic';
import { copyText, renderTemplate } from '@/common/util';
import CodeEditor from '@/components/monaco-editor/new-editor.vue';
import StatusIcon from '@/components/status-icon';
import { useAppData, useCluster } from '@/composables/use-app';
import fullScreen from '@/directives/full-screen';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $store from '@/store';

export default defineComponent({
  name: 'UserToken',
  components: { StatusIcon, CodeEditor },
  directives: {
    'full-screen': fullScreen,
  },
  setup() {
    const { clusterList } = useCluster();
    const hasSharedCluster = computed(() => clusterList.value.some(item => !!item.is_shared));
    const goBack = () => {
      $router.back();
    };
    // editor options
    const options = ref({
      roundedSelection: false,
      scrollBeyondLastLine: false,
      renderLineHighlight: 'none',
    });
    // 用户信息
    const user = computed(() => $store.state.user);
    // 特性开关
    const { flagsMap } = useAppData();
    // 使用案例
    const projectID = computed(() => $store.getters.curProjectId);
    const projectCode = computed(() => $store.getters.curProjectCode);
    const kubeConfigExample = ref('kubectl --kubeconfig=/root/.kube/demo_config get node');

    const demoConfigExample = ref(renderTemplate(clusterDemoConfig, {
      username: user.value.username,
      token: `\${${$i18n.t('apiToken.text')}}`,
      bcs_api_host: window.BCS_API_HOST,
      projectID: projectID.value,
    }));
    const apiExample = 'curl -X GET -H "Authorization: Bearer ${token}" -H "accept: application/json" "${bcs_api_host}/clusters/${cluster_id}/version"';
    const bcsApiExample = ref(renderTemplate(apiExample, {
      token: `\${${$i18n.t('apiToken.text')}}`,
      bcs_api_host: window.BCS_API_HOST,
      projectID: projectID.value,
    }));

    const shareKubeConfigExample = ref('kubectl --kubeconfig=/root/.kube/demo_config get all -n <namespace>');
    const shareDemoConfigExample = ref(renderTemplate(shareClusterDemoConfig, {
      username: user.value.username,
      token: `\${${$i18n.t('apiToken.text')}}`,
      bcs_api_host: window.BCS_API_HOST,
      projectCode: projectCode.value,
    }));
    const shareApiExample = 'curl -X GET -H "Authorization: Bearer ${token}" -H "accept: application/json" "${bcs_api_host}/projects/${projectCode}/clusters/${cluster_id}/version"';
    const shareBcsApiExample = ref(renderTemplate(shareApiExample, {
      token: `\${${$i18n.t('apiToken.text')}}`,
      bcs_api_host: window.BCS_API_HOST,
      projectCode: projectCode.value,
    }));

    const timeList = ref([
      {
        id: -1,
        name: $i18n.t('apiToken.label.perpetual'),
      },
      {
        id: 30,
        name: $i18n.t('units.time.nDays', { num: 30 }),
      },
      {
        id: 3 * 30,
        name: $i18n.t('units.time.nDays', { num: 3 * 30 }),
      },
      {
        id: 6 * 30,
        name: $i18n.t('units.time.nDays', { num: 6 * 30 }),
      },
      {
        id: 365,
        name: $i18n.t('units.time.nDays', { num: 365 }),
      },
    ]);
    const active = ref(-1);
    const activeTimestamp = computed(() => (active.value === -1 ? active.value : active.value * 24 * 60 * 60));

    // 自定义时间
    const isCustomTime = ref(false);
    const customInputRef = ref<any>(null);
    const handleSelectTime = (item) => {
      active.value = item.id;
      isCustomTime.value = false;
    };
    const handleCustomTime = () => {
      isCustomTime.value = true;
      setTimeout(() => {
        customInputRef.value.focus();
        active.value = 1;
      }, 0);
    };

    const hiddenToken = ref(true);
    // 隐藏Token
    const toggleHiddenToken = () => {
      hiddenToken.value = !hiddenToken.value;
    };
    // 复制Token
    const handleCopyToken = (row) => {
      copyText(row.token);
      $bkMessage({
        theme: 'success',
        message: $i18n.t('generic.msg.success.copy'),
      });
    };
    // token操作
    const operateType = ref<'create' | 'edit'>('create');
    const showTokenDialog = ref(false);
    const showDeleteDialog = ref(false);
    // 新建token事件
    const handleCreateToken = () => {
      showTokenDialog.value = true;
      operateType.value = 'create';
    };
    // 取消更新token事件
    const cancelUpdateTokenDialog = () => {
      curEditRow.value = null;
      active.value = -1;
      isCustomTime.value = false;
    };
    const curEditRow = ref<any>(null);
    // 续期Token事件
    const handleRenewalToken = (row) => {
      showTokenDialog.value = true;
      operateType.value = 'edit';
      curEditRow.value = row;
    };
    // 删除Token事件
    const handleDeleteToken = async (row) => {
      showDeleteDialog.value = true;
      curEditRow.value = row;
    };
    // 删除确认复选框
    const deleteConfirm = ref(false);
    // 取消删除事件
    const cancelDeleteDialog = () => {
      curEditRow.value = null;
      deleteConfirm.value = false;
      showDeleteDialog.value = false;
    };
    // token列表
    const loading = ref(false);
    const data = ref([]);
    const getTokenList = async () => {
      loading.value = true;
      data.value = await $store.dispatch('token/getTokens', {
        $username: user.value.username,
      });
      loading.value = false;
    };

    // 创建或者续期Token
    const updateLoading = ref(false);
    const confirmUpdateTokenDialog = async () => {
      updateLoading.value = true;
      if (operateType.value === 'create') {
        const result = await $store.dispatch('token/createToken', {
          username: user.value.username,
          expiration: activeTimestamp.value,
        });
        result && $bkMessage({
          theme: 'success',
          message: $i18n.t('generic.msg.success.create'),
        });
      } else if (operateType.value === 'edit' && curEditRow.value) {
        const result = await $store.dispatch('token/updateToken', {
          $token: curEditRow.value.token,
          expiration: activeTimestamp.value,
        });
        result && $bkMessage({
          theme: 'success',
          message: $i18n.t('apiToken.msg.renewOK'),
        });
      }
      updateLoading.value = false;
      showTokenDialog.value = false;
      cancelUpdateTokenDialog();
      getTokenList();
    };
    const deleteLoading = ref(false);
    const confirmDeleteToken = async () => {
      deleteLoading.value = true;
      const result = await $store.dispatch('token/deleteToken', {
        $token: curEditRow.value.token,
      });
      deleteLoading.value = false;
      showDeleteDialog.value = false;
      cancelDeleteDialog();
      result && $bkMessage({
        theme: 'success',
        message: $i18n.t('generic.msg.success.delete'),
      });
      getTokenList();
    };

    onMounted(() => {
      getTokenList();
    });
    return {
      options,
      deleteLoading,
      updateLoading,
      loading,
      data,
      deleteConfirm,
      showTokenDialog,
      showDeleteDialog,
      timeList,
      active,
      isCustomTime,
      customInputRef,
      user,
      hiddenToken,
      operateType,
      goBack,
      handleRenewalToken,
      handleDeleteToken,
      handleCreateToken,
      confirmUpdateTokenDialog,
      handleSelectTime,
      handleCustomTime,
      toggleHiddenToken,
      handleCopyToken,
      confirmDeleteToken,
      cancelDeleteDialog,
      cancelUpdateTokenDialog,
      kubeConfigExample,
      demoConfigExample,
      bcsApiExample,
      shareKubeConfigExample,
      shareDemoConfigExample,
      shareBcsApiExample,
      BK_IAM_HOST: window.BK_IAM_HOST,
      hasSharedCluster,
      flagsMap,
    };
  },
});
</script>
<style lang="postcss" scoped>
.user-token {
    padding: 16px 84px;
    overflow: auto;
}
.info-item {
    line-height: 20px;
}
.user-token-header {
    font-size: 16px;
    height: 32px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 12px;
    .back {
        cursor: pointer;
        font-weight: 700;
        color: #3a84ff;
    }
    .title {
        margin-left: 8px;
    }
    .help {
        font-size: 14px;
    }
}
.token-row {
    display: flex;
    align-items: center;
    i {
        cursor: pointer;
    }
}
.create-token-btn {
    margin-top: 16px;
}
.create-token-dialog {
    padding: 0 4px 24px 4px;
    .title {
        font-size: 14px;
        margin-bottom: 14px;
    }
    .bk-button-group {
        display: flex;
    }
    .group-btn {
        min-width: 80px;
    }
    .custom-input {
        width: 100px;
        margin-left: -1px;
        >>> input {
            padding: 0 4px !important;
        }
        &-append {
            width: 28px;
            height: 32px;
            font-size: 12px;
            display: flex;
            align-items: center;
            justify-content: center;
        }
    }
}
.delete-token-dialog {
    padding: 0 4px 24px 4px;
    .title {
        font-size: 14px;
        font-weight: 700;
        margin-bottom: 14px
    }
}
.user-token-example {
    margin-top: 24px;
    .example-item {
        margin-bottom: 24px;
        .title {
            margin-bottom: 12px;
            font-weight: 400;
            text-align: left;
            color: #313238;
            font-size: 14px;
        }
        .total-title {
            font-weight: 700;
        }
        .code-wrapper {
            font-size: 14px;
        }
    }
}
</style>
