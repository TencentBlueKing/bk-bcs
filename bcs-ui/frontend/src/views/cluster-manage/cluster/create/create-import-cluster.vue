<template>
  <section class="create-import-cluster bcs-content-wrapper">
    <BkForm
      :label-width="labelWidth"
      :model="importClusterInfo"
      :rules="formRules"
      class="import-form"
      ref="importFormRef">
      <BkFormItem :label="$t('集群名称')" property="clusterName" error-display-type="normal" required>
        <bk-input :maxlength="64" v-model="importClusterInfo.clusterName"></bk-input>
      </BkFormItem>
      <BkFormItem :label="$t('导入方式')">
        <bk-radio-group class="btn-group" v-model="importClusterInfo.importType">
          <bk-radio class="btn-group-first" value="kubeconfig">{{$t('kubeconfig')}}</bk-radio>
          <bk-radio value="provider">{{$t('云服务商')}}</bk-radio>
        </bk-radio-group>
      </BkFormItem>
      <BkFormItem :label="$t('集群环境')" required v-if="$INTERNAL">
        <bk-radio-group class="btn-group" v-model="importClusterInfo.environment">
          <bk-radio class="btn-group-first" value="debug">
            {{ $t('测试环境') }}
          </bk-radio>
          <bk-radio value="prod">
            {{ $t('正式环境') }}
          </bk-radio>
        </bk-radio-group>
      </BkFormItem>
      <BkFormItem :label="$t('集群描述')">
        <bk-input v-model="importClusterInfo.description" type="textarea"></bk-input>
      </BkFormItem>
      <template v-if="importClusterInfo.importType === 'provider'">
        <BkFormItem
          :label="$t('云服务商')"
          property="provider"
          error-display-type="normal"
          required>
          <bcs-select
            :loading="templateLoading"
            class="w640"
            v-model="importClusterInfo.provider"
            :clearable="false">
            <bcs-option
              v-for="item in availableTemplateList"
              :key="item.cloudID"
              :id="item.cloudID"
              :name="item.name">
            </bcs-option>
          </bcs-select>
        </BkFormItem>
        <BkFormItem :label="$t('云凭证')" property="accountID" error-display-type="normal" required>
          <div style="display: flex;">
            <bcs-select
              :loading="accountsLoading"
              class="w640"
              :clearable="false"
              searchable
              v-model="importClusterInfo.accountID">
              <bcs-option
                v-for="item in accountsList"
                :key="item.account.accountID"
                :id="item.account.accountID"
                :name="item.account.accountName"
                v-authority="{
                  clickable: webAnnotations.perms[item.account.accountID]
                    && webAnnotations.perms[item.account.accountID].cloud_account_use,
                  actionId: 'cloud_account_use',
                  resourceName: item.account.accountName,
                  disablePerms: true,
                  newPerms: true,
                  permCtx: {
                    project_id: item.account.projectID,
                    account_id: item.account.accountID,
                    operator: user.username
                  }
                }">
              </bcs-option>
              <template slot="extension">
                <div class="extension-item" style="cursor: pointer" @click="handleGotoCloudToken">
                  <i class="bk-icon icon-plus-circle"></i>
                  <span>{{ $t('新增凭证') }}</span>
                </div>
              </template>
            </bcs-select>
            <span
              class="refresh ml10"
              v-bk-tooltips="$t('刷新列表')"
              @click="handleGetCloudAccounts">
              <i class="bcs-icon bcs-icon-reset"></i>
            </span>
          </div>
        </BkFormItem>
        <BkFormItem :label="$t('所属区域')" property="region" error-display-type="normal" required>
          <bcs-select
            :loading="regionLoading"
            class="w640"
            searchable
            :clearable="false"
            v-model="importClusterInfo.region">
            <bcs-option
              v-for="item in regionList"
              :key="item.region"
              :id="item.region"
              :name="item.regionName">
            </bcs-option>
          </bcs-select>
        </BkFormItem>
        <BkFormItem :label="$t('TKE集群ID')" property="cloudID" error-display-type="normal" required>
          <bcs-select
            class="w640" searchable
            :clearable="false"
            :loading="clusterLoading"
            v-model="importClusterInfo.cloudID">
            <bcs-option
              v-for="item in clusterList"
              :key="item.clusterID"
              :id="item.clusterID"
              :name="item.clusterName">
            </bcs-option>
          </bcs-select>
        </BkFormItem>
      </template>
      <BkFormItem
        :label="$t('集群kubeconfig')"
        property="yaml"
        error-display-type="normal"
        required
        v-else>
        <bk-button class="mb10">
          <input
            class="file-input"
            accept="*"
            type="file"
            @change="handleFileChange">
          {{$t('文件导入')}}
        </bk-button>
        <CodeEditor
          class="cube-config"
          lang="yaml"
          width="100%"
          :height="480"
          :options="{ lineNumbers: 'off' }"
          ref="codeEditorRef"
          v-model="importClusterInfo.yaml"
        ></CodeEditor>
      </BkFormItem>
      <BkFormItem class="mt16" v-if="importClusterInfo.importType === 'kubeconfig'">
        <bk-button
          theme="primary"
          :loading="testLoading"
          @click="handleTest">{{$t('kubeconfig可用性测试')}}</bk-button>
      </BkFormItem>
      <BkFormItem class="mt32">
        <span
          v-bk-tooltips="{
            content: $t('请先测试kubeconfig可用性'),
            disabled: isTestSuccess || importClusterInfo.importType === 'provider'
          }">
          <bk-button
            class="btn"
            theme="primary"
            :loading="loading"
            :disabled="!isTestSuccess && importClusterInfo.importType === 'kubeconfig'"
            @click="handleImport"
          >{{$t('导入')}}</bk-button>
        </span>
        <bk-button
          class="btn ml5"
          @click="handleCancel"
        >{{$t('取消')}}</bk-button>
      </BkFormItem>
    </BkForm>
  </section>
</template>
<script lang="ts">
import { defineComponent, ref, computed, onMounted, watch } from 'vue';
import useFormLabel from '@/composables/use-form-label';
import BkForm from 'bk-magic-vue/lib/form';
import BkFormItem from 'bk-magic-vue/lib/form-item';
import CodeEditor from '@/components/monaco-editor/new-editor.vue';
import $router from '@/router';
import $store from '@/store';
import $i18n from '@/i18n/i18n-setup';
import $bkMessage from '@/common/bkmagic';

export default defineComponent({
  name: 'CreateImportCluster',
  components: {
    CodeEditor,
    BkForm,
    BkFormItem,
  },
  setup() {
    const importClusterInfo = ref({
      importType: 'kubeconfig',
      clusterName: '',
      environment: 'prod',
      description: '',
      yaml: '',
      provider: '',
      region: '',
      accountID: '',
      cloudID: '',
    });
    const isTestSuccess = ref(false);
    const testLoading = ref(false);
    const loading = ref(false);
    const importFormRef = ref<any>(null);
    const formRules = ref({
      provider: [
        {
          required: true,
          message: $i18n.t('必填项'),
          trigger: 'blur',
        },
      ],
      yaml: [
        {
          required: true,
          message: $i18n.t('必填项'),
          trigger: 'blur',
        },
      ],
      clusterName: [
        {
          required: true,
          message: $i18n.t('必填项'),
          trigger: 'blur',
        },
      ],
      region: [
        {
          required: true,
          message: $i18n.t('必填项'),
          trigger: 'blur',
        },
      ],
      accountID: [
        {
          required: true,
          message: $i18n.t('必填项'),
          trigger: 'blur',
        },
      ],
      cloudID: [
        {
          required: true,
          message: $i18n.t('必填项'),
          trigger: 'blur',
        },
      ],
    });

    // 导入文件
    const codeEditorRef = ref<any>(null);
    const handleFileChange = (event) => {
      const [file] = event.target.files;
      if (!file) return;
      const reader = new FileReader();
      reader.readAsText(file, 'UTF-8');
      reader.onload = (e) => {
        importClusterInfo.value.yaml = e?.target?.result as string || '';
        codeEditorRef.value?.update();
      };
    };

    const handleCancel = () => {
      $router.back();
    };

    const curProject = computed(() => $store.state.curProject);
    const user = computed(() => $store.state.user);

    // 云服务商
    const templateList = ref<any[]>([]);
    const availableTemplateList = computed(() => templateList.value
      .filter(item => item.enable === 'true' && !item?.confInfo?.disableImportCluster));
    const templateLoading = ref(false);
    const handleGetCloudList = async () => {
      templateLoading.value = true;
      templateList.value = await $store.dispatch('clustermanager/fetchCloudList');
      templateLoading.value = false;
    };

    // 区域列表
    const regionList = ref<any[]>([]);
    const regionLoading = ref(false);
    const getRegionList = async () => {
      if (!importClusterInfo.value.provider || !importClusterInfo.value.accountID) return;
      regionLoading.value = true;
      regionList.value = await $store.dispatch('clustermanager/cloudRegionByAccount', {
        $cloudId: importClusterInfo.value.provider,
        accountID: importClusterInfo.value.accountID,
      });
      regionLoading.value = false;
    };

    // 云账户信息
    const webAnnotations = ref({ perms: {} });
    const accountsLoading = ref(false);
    const accountsList = ref<any[]>([]);
    const handleGetCloudAccounts = async () => {
      if (!importClusterInfo.value.provider) return;
      accountsLoading.value = true;
      const res = await $store.dispatch('clustermanager/cloudAccounts', {
        $cloudId: importClusterInfo.value.provider,
        projectID: curProject.value.project_id,
        operator: user.value.username,
      });
      accountsList.value = res.data;
      webAnnotations.value = res.web_annotations || { perms: {} };
      accountsLoading.value = false;
    };

    // 集群列表
    const clusterLoading = ref(false);
    const clusterList = ref<any[]>([]);
    const handleGetClusterList = async () => {
      if (!importClusterInfo.value.region || !importClusterInfo.value.provider) return;

      clusterLoading.value = true;
      clusterList.value = await $store.dispatch('clustermanager/cloudClusterList', {
        region: importClusterInfo.value.region,
        $cloudId: importClusterInfo.value.provider,
        accountID: importClusterInfo.value.accountID,
      });
      clusterLoading.value = false;
    };

    watch(() => importClusterInfo.value.importType, (type) => {
      type === 'provider' && !templateList.value.length && handleGetCloudList();
    });
    watch(() => importClusterInfo.value.provider, () => {
      importClusterInfo.value.accountID = '';
      handleGetCloudAccounts();
      importClusterInfo.value.region = '';
      getRegionList();
      importClusterInfo.value.cloudID = '';
      handleGetClusterList();
    });
    watch(() => importClusterInfo.value.accountID, () => {
      importClusterInfo.value.region = '';
      getRegionList();
      importClusterInfo.value.cloudID = '';
      handleGetClusterList();
    });
    watch(() => importClusterInfo.value.region, () => {
      importClusterInfo.value.cloudID = '';
      handleGetClusterList();
    });
    watch(() => importClusterInfo.value.yaml, () => {
      isTestSuccess.value = false;
    });

    // 可用性测试
    const handleTest = async () => {
      testLoading.value = true;
      const result = await $store.dispatch('clustermanager/kubeConfig', {
        kubeConfig: importClusterInfo.value.yaml,
      });
      if (result) {
        isTestSuccess.value = true;
        $bkMessage({
          theme: 'success',
          message: $i18n.t('测试成功'),
        });
      }
      testLoading.value = false;
    };
    // 集群导入
    const handleImport = async () => {
      const validate = await importFormRef.value.validate();
      if (!validate) return;

      loading.value = true;
      const params = {
        clusterName: importClusterInfo.value.clusterName,
        description: importClusterInfo.value.description,
        projectID: curProject.value.project_id,
        businessID: String(curProject.value.cc_app_id),
        provider: importClusterInfo.value.importType === 'kubeconfig'
          ? 'bluekingCloud' : importClusterInfo.value.provider, // importClusterInfo.value.provider,
        region: importClusterInfo.value.importType === 'kubeconfig'
          ? 'default' : importClusterInfo.value.region,
        environment: importClusterInfo.value.environment,
        engineType: 'k8s',
        isExclusive: true,
        clusterType: 'single',
        manageType: 'INDEPENDENT_CLUSTER',
        creator: user.value.username,
        cloudMode: {
          cloudID: importClusterInfo.value.importType === 'kubeconfig'
            ? '' : importClusterInfo.value.cloudID,
          kubeConfig: importClusterInfo.value.yaml,
        },
        networkType: 'overlay',
        accountID: importClusterInfo.value.importType === 'kubeconfig'
          ? '' : importClusterInfo.value.accountID,
      };
      const result = await $store.dispatch('clustermanager/importCluster', params);
      loading.value = false;
      if (result) {
        $bkMessage({
          theme: 'success',
          message: $i18n.t('导入成功'),
        });
        $router.push({ name: 'home' });
      }
    };
    const handleGotoCloudToken = () => {
      const data = $router.resolve({ name: 'tencentCloud' });
      window.open(data.href, '_blank');
    };

    const { labelWidth, initFormLabelWidth } = useFormLabel();
    onMounted(() => {
      initFormLabelWidth(importFormRef.value);
    });
    return {
      regionLoading,
      templateLoading,
      user,
      codeEditorRef,
      clusterLoading,
      accountsLoading,
      accountsList,
      isTestSuccess,
      testLoading,
      labelWidth,
      availableTemplateList,
      importClusterInfo,
      importFormRef,
      loading,
      formRules,
      handleFileChange,
      handleImport,
      handleCancel,
      regionList,
      getRegionList,
      clusterList,
      handleTest,
      handleGotoCloudToken,
      handleGetCloudAccounts,
      webAnnotations,
    };
  },
});
</script>
<style lang="postcss" scoped>
/deep/ .mt16 {
  margin-top: 16px;
}
/deep/ .mt32 {
  margin-top: 32px;
}
.create-import-cluster {
  padding: 24px;
  /deep/ .w640 {
      width: 640px;
      background: #fff;
  }
  /deep/ .bk-input-text {
      width: 640px;
  }
  /deep/ .bk-textarea-wrapper {
      width: 640px;
      height: 80px;
  }
  /deep/ .cube-config {
      max-width: 1000px;
  }
  /deep/ .btn {
      width: 88px;
  }
  /deep/ .file-input {
      position: absolute;
      opacity: 0;
      left: 0;
      top: 0;
      width: 100%;
      height: 100%;
  }
}
.import-cluster-dialog {
  padding: 0 4px 24px 4px;
  .title {
      font-size: 14px;
      font-weight: 700;
      margin-bottom: 14px
  }
}
/deep/ .btn-group-first {
  min-width: 100px;
}
.refresh {
  font-size: 12px;
  color: #3a84ff;
  cursor: pointer;
}
</style>
