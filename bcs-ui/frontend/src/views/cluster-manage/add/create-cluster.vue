<template>
  <section class="create-form-cluster bcs-content-wrapper !overflow-auto">
    <bk-form :label-width="labelWidth" :model="basicInfo" :rules="basicDataRules" ref="basicFormRef">
      <bk-form-item :label="$t('cluster.labels.name')" property="clusterName" error-display-type="normal" required>
        <bk-input :maxlength="64" class="w640" v-model="basicInfo.clusterName"></bk-input>
      </bk-form-item>
      <bk-form-item
        :label="$t('cluster.create.label.provider')" property="provider" error-display-type="normal" required>
        <bcs-select :loading="templateLoading" class="w640" v-model="basicInfo.provider" :clearable="false">
          <bcs-option
            v-for="item in availableTemplateList"
            :key="item.cloudID"
            :id="item.cloudID"
            :name="item.name">
          </bcs-option>
        </bcs-select>
      </bk-form-item>
      <bk-form-item
        :label="$t('generic.label.version')"
        property="clusterBasicSettings.version"
        error-display-type="normal"
        required>
        <bcs-select class="w640" v-model="basicInfo.clusterBasicSettings.version" searchable :clearable="false">
          <bcs-option v-for="item in versionList" :key="item" :id="item" :name="item"></bcs-option>
        </bcs-select>
      </bk-form-item>
      <bk-form-item :label="$t('cluster.create.label.desc1')">
        <bk-input v-model="basicInfo.description" type="textarea"></bk-input>
      </bk-form-item>
      <bk-form-item :label="$t('cluster.create.label.params.text')" ref="extraInfoRef" v-show="expanded">
        <KeyValue
          class="w700"
          :show-footer="false"
          :show-header="false"
          :key-advice="keyAdvice"
          ref="keyValueRef"
        ></KeyValue>
      </bk-form-item>
      <bk-form-item>
        <div class="action">
          <i :class="['bk-icon', expanded ? 'icon-angle-double-up' : 'icon-angle-double-down']"></i>
          <span @click="toggleSettings">
            {{ expanded ? $t('cluster.create.button.putAwayConfig') : $t('cluster.create.button.expandConfig')}}
          </span>
        </div>
      </bk-form-item>
      <bk-form-item
        :label="$t('cluster.create.label.chooseMaster')"
        property="ipList"
        error-display-type="normal"
        required>
        <bk-button icon="plus" @click="handleShowIpSelector">
          {{$t('generic.ipSelector.action.selectHost')}}
        </bk-button>
        <bk-table class="ip-list mt10" :data="basicInfo.ipList" v-if="basicInfo.ipList.length">
          <bk-table-column type="index" :label="$t('cluster.create.label.index')" width="60"></bk-table-column>
          <bk-table-column :label="$t('generic.ipSelector.label.innerIp')" prop="bk_host_innerip"></bk-table-column>
          <bk-table-column :label="$t('generic.ipSelector.label.idc')" prop="idc_name"></bk-table-column>
          <bk-table-column
            :label="$t('generic.ipSelector.label.serverModel')" prop="svr_device_class"></bk-table-column>
          <bk-table-column :label="$t('generic.label.action')" width="100">
            <template #default="{ row }">
              <bcs-button text @click="handleRemoveServer(row)">{{$t('cluster.create.button.remove')}}</bcs-button>
            </template>
          </bk-table-column>
        </bk-table>
      </bk-form-item>
      <bk-form-item>
        <bk-button
          class="btn"
          theme="primary"
          @click="handleShowConfirmDialog">{{$t('generic.button.confirm')}}</bk-button>
        <bk-button class="btn" @click="handleCancel">{{$t('generic.button.cancel')}}</bk-button>
      </bk-form-item>
    </bk-form>
    <bcs-dialog
      v-model="confirmDialog"
      theme="primary"
      header-position="left"
      :title="$t('cluster.create.button.confirmCreateCluster.text')"
      width="640">
      <div class="create-cluster-dialog">
        <div class="title">{{$t('cluster.create.button.confirmCreateCluster.doc.title')}}:</div>
        <bcs-checkbox-group class="confirm-wrapper" v-model="createConfirm">
          <bcs-checkbox
            value="2"
            class="mt10">
            {{$t('cluster.create.button.confirmCreateCluster.doc.article0')}}
          </bcs-checkbox>
        </bcs-checkbox-group>
      </div>
      <template #footer>
        <div>
          <bcs-button
            :disabled="createConfirm.length < 1"
            theme="primary"
            :loading="loading"
            @click="handleCreateCluster"
          >{{$t('cluster.create.button.confirmCreateCluster.text2')}}</bcs-button>
          <bcs-button @click="confirmDialog = false">{{$t('cluster.create.button.cancel')}}</bcs-button>
        </div>
      </template>
    </bcs-dialog>
    <IpSelector
      :show-dialog="showIpSelector"
      :ip-list="basicInfo.ipList"
      @confirm="handleChooseServer"
      @cancel="showIpSelector = false">
    </IpSelector>
  </section>
</template>
<script lang="ts">
import { computed, defineComponent, onMounted, ref, watch } from 'vue';

import $bkMessage from '@/common/bkmagic';
import IpSelector from '@/components/ip-selector/ip-selector.vue';
import KeyValue from '@/components/key-value.vue';
import useFormLabel from '@/composables/use-form-label';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $store from '@/store';

export default defineComponent({
  name: 'CreateCluster',
  components: {
    IpSelector,
    KeyValue,
  },
  setup() {
    const basicFormRef = ref<any>(null);
    const basicInfo = ref<{
      clusterName: string;
      description: string;
      provider: string;
      clusterBasicSettings: {
        version: string;
      };
      ipList: any[];
    }>({
      clusterName: '',
      description: '',
      provider: '',
      clusterBasicSettings: {
        version: '',
      },
      ipList: [],
    });
    const templateList = ref<any[]>([]);
    const availableTemplateList = computed(() => templateList.value
      .filter(item => !item?.confInfo?.disableCreateCluster));
    const versionList = computed(() => {
      const cloud = templateList.value.find(item => item.cloudID === basicInfo.value.provider);
      return cloud?.clusterManagement.availableVersion || [];
    });
    const basicDataRules = ref({
      clusterName: [
        {
          required: true,
          message: $i18n.t('generic.validate.required'),
          trigger: 'blur',
        },
      ],
      provider: [
        {
          required: true,
          message: $i18n.t('generic.validate.required'),
          trigger: 'blur',
        },
      ],
      'clusterBasicSettings.version': [
        {
          required: true,
          message: $i18n.t('generic.validate.required'),
          trigger: 'blur',
        },
      ],
      ipList: [
        {
          message: $i18n.t('cluster.create.validate.masterNum2'),
          validator(val) {
            return val.length % 2 !== 0;
          },
          trigger: 'blur',
        },
      ],
    });
    const showIpSelector = ref(false);
    const confirmDialog = ref(false);
    const createConfirm = ref([]);
    const loading = ref(false);
    const expanded = ref(false);
    const templateLoading = ref(false);
    const keyAdvice = ref([
      {
        name: 'DOCKER_LIB',
        desc: $i18n.t('cluster.create.label.params.dockerPath'),
        default: '',
      },
      {
        name: 'DOCKER_VERSION',
        desc: $i18n.t('cluster.create.label.params.dockerVersion'),
        default: '19.03.9',
      },
      {
        name: 'KUBELET_LIB',
        desc: $i18n.t('cluster.create.label.params.kubeletPath'),
        default: '',
      },
      {
        name: 'K8S_VER',
        desc: $i18n.t('cluster.create.label.clusterVersion'),
        default: '',
      },
      {
        name: 'K8S_SVC_CIDR',
        desc: $i18n.t('cluster.create.label.params.k8sServiceCIDR'),
        default: '',
      },
      {
        name: 'K8S_POD_CIDR',
        desc: $i18n.t('cluster.create.label.params.k8sPodCIDR'),
        default: '',
      },
    ]);

    const handleShowIpSelector = () => {
      showIpSelector.value = true;
    };
    const handleChooseServer = (data) => {
      basicInfo.value.ipList = data;
      showIpSelector.value = false;
    };
    const handleShowConfirmDialog = async () => {
      const result = await basicFormRef.value?.validate();
      // todo bcs兼容1.0组件，对表单组件做了包装，导致 keyValueRef 取不到
      const validateKey = basicFormRef.value?.$refs?.extraInfoRef?.$refs?.keyValueRef?.validate();
      if (!result || !validateKey) return;
      confirmDialog.value = true;
    };
    watch(confirmDialog, (value) => {
      if (!value) {
        createConfirm.value = [];
      }
    });
    const curProject = computed(() => $store.state.curProject);
    const user = computed(() => $store.state.user);
    const handleCreateCluster = async () => {
      loading.value = true;
      confirmDialog.value = false;
      loading.value = false;
      const extraInfo = basicFormRef.value?.$refs?.extraInfoRef?.$refs?.keyValueRef?.labels || {};
      const params = {
        environment: 'prod',
        projectID: curProject.value.project_id,
        businessID: String(curProject.value.businessID),
        engineType: 'k8s',
        isExclusive: true,
        clusterType: 'single',
        creator: user.value.username,
        manageType: 'INDEPENDENT_CLUSTER',
        clusterName: basicInfo.value?.clusterName,
        description: basicInfo.value?.description,
        provider: basicInfo.value?.provider,
        region: 'default',
        vpcID: '',
        clusterBasicSettings: basicInfo.value?.clusterBasicSettings,
        networkType: 'overlay',
        extraInfo: {
          create_cluster: Object.keys(extraInfo)
            .reduce<any[]>((pre, key) => {
            pre.push(`${key}=${extraInfo[key]}`);
            return pre;
          }, [])
            .join(';'),
        },
        networkSettings: {},
        master: basicInfo.value.ipList.map((item: any) => item.bk_host_innerip),
      };
      const result = await $store.dispatch('clustermanager/createCluster', params);
      if (result) {
        $bkMessage({
          theme: 'success',
          message: $i18n.t('generic.msg.success.create'),
        });
        $router.push({ name: 'clusterMain' });
      }
    };
    const handleCancel = () => {
      $router.back();
    };
    const toggleSettings = () => {
      expanded.value = !expanded.value;
    };
    const handleGetTemplateList = async () => {
      templateLoading.value = true;
      templateList.value = await $store.dispatch('clustermanager/fetchCloudList');
      templateLoading.value = false;
    };
    const handleRemoveServer = async (row) => {
      const index = basicInfo.value.ipList.findIndex(item => item?.cloudArea?.id === row?.cloudArea?.id
                    && item.ip === row.ip);
      if (index > -1) {
        basicInfo.value.ipList.splice(index, 1);
      }
    };
    const { labelWidth, initFormLabelWidth } = useFormLabel();
    onMounted(() => {
      handleGetTemplateList();
      initFormLabelWidth(basicFormRef.value);
    });
    return {
      labelWidth,
      keyAdvice,
      templateLoading,
      expanded,
      loading,
      availableTemplateList,
      versionList,
      basicFormRef,
      basicInfo,
      showIpSelector,
      basicDataRules,
      confirmDialog,
      createConfirm,
      handleChooseServer,
      handleShowIpSelector,
      handleShowConfirmDialog,
      handleCreateCluster,
      handleCancel,
      toggleSettings,
      handleGetTemplateList,
      handleRemoveServer,
    };
  },
});
</script>
<style lang="postcss" scoped>
.create-form-cluster {
  padding: 24px;
  /deep/ .bk-textarea-wrapper {
      width: 640px;
      height: 80px;
      background-color: #fff;
  }
  /deep/ .w640 {
      width: 640px;
      background: #fff;
  }
  /deep/ .w700 {
      width: 705px;
  }
  /deep/ .btn {
      width: 88px;
  }
  /deep/ .ip-list {
      max-width: 1000px;
  }
  /deep/ .action {
      i {
          font-size: 20px;
      }
      span {
          cursor: pointer;
      }
      align-items: center;
      color: #3a84ff;
      display: flex;
      font-size: 14px;
  }
}
.create-cluster-dialog {
  padding: 0 4px 24px 4px;
  .title {
      font-size: 14px;
      font-weight: 700;
      margin-bottom: 14px
  }
  .confirm-wrapper {
      display: flex;
      flex-direction: column;
  }
}
</style>
