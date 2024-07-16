<template>
  <BcsContent
    title="Polaris"
    :desc="`(${$t('plugin.tools.cluster', { name: customCrdList.curCluster?.clusterName })})`">
    <div class="flex items-center justify-between mb-[16px]">
      <bcs-button
        class="mr-[5px] min-w-[115px]" theme="primary" icon="plus" @click="customCrdList.showCreateCrdSideslider">
        {{$t('plugin.tools.create')}}
      </bcs-button>
      <div class="flex justify-end">
        <ClusterSelect v-model="propClusterId" cluster-type="all" @change="handleClusterChange"></ClusterSelect>
        <NamespaceSelect
          :cluster-id="propClusterId"
          class="w-[250px] ml-[5px] mr-[5px]"
          :clearable="true"
          v-model="ns"
          @change="handleGetList">
        </NamespaceSelect>
        <bcs-input
          class="w-[320px]"
          right-icon="bk-icon icon-search"
          clearable
          :placeholder="$t('generic.placeholder.searchName')"
          v-model.trim="customCrdList.searchValue">
        </bcs-input>
      </div>
    </div>
    <bcs-table
      size="large"
      :data="customCrdList.curPageData"
      :pagination="customCrdList.pagination"
      v-bkloading="{ isLoading: customCrdList.tableLoading }"
      @page-change="customCrdList.pageChange"
      @page-limit-change="customCrdList.pageSizeChange">
      <bcs-table-column :label="$t('plugin.tools._ruleName')" show-overflow-tooltip min-width="100">
        <template #default="{ row }">
          {{ row.metadata.name || '--' }}
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('plugin.tools.cluster_ns')" min-width="220">
        <template #default="{ row }">
          <p class="bcs-ellipsis leading-[24px]" :title="customCrdList.curCluster?.clusterName">
            <span>{{ $t('generic.label.cluster1') }}：</span>
            <span>{{ customCrdList.curCluster?.clusterName }}</span>
          </p>
          <p class="bcs-ellipsis leading-[24px]" :title="row.namespace">
            <span>{{ $t('k8s.namespace') }}：</span>
            <span> {{ row.metadata.namespace }}</span>
          </p>
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('plugin.tools.polarisInfo')" min-width="180">
        <template #default="{ row }">
          <p class="bcs-ellipsis leading-[24px]">
            <span class="inline-flex min-w-[62px]">{{ $t('generic.label.name') }}：</span>
            <span>{{ row.spec && row.spec.polaris ? row.spec.polaris.name : '--' }}</span>
          </p>
          <p class="bcs-ellipsis leading-[24px]">
            <span class="inline-flex min-w-[62px]">{{ $t('k8s.namespace') }}：</span>
            <span>{{ row.spec && row.spec.polaris ? row.spec.polaris.namespace : '--' }}</span>
          </p>
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('deploy.templateset.service')" min-width="160">
        <template #default="{ row }">
          <div class="flex flex-col">
            <div
              v-for="(service, index) in row.spec.services"
              :key="index"
              :class="index <= row.spec.services.length ? 'mb-[10px]' : ''">
              <p>- name: {{ service.name }}</p>
              <p class="pl-[10px]">port: <span>{{ service.port }}</span></p>
              <p class="pl-[10px]">
                direct:
                {{ service.direct ? `${$t('units.boolean.true')}` : `${$t('units.boolean.false')}` }}
              </p>
            </div>
          </div>
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('plugin.tools.ipPortAndWeight')" min-width="180">
        <template #default="{ row }">
          <div
            class="flex flex-col"
            v-if="row.status
              && row.status.syncStatus
              && row.status.syncStatus.lastRemoteInstances
              && row.status.syncStatus.lastRemoteInstances.length">
            <div
              v-for="(remote, remoteIndex) in row.status.syncStatus.lastRemoteInstances"
              :key="remoteIndex"
              class="bcs-ellipsis leading-[24px]">
              {{ remote.ip || '--' }}: {{ remote.port || '--' }} {{ remote.weight || '--' }}
            </div>
          </div>
          <div v-else>--</div>
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('generic.label.status')" width="270">
        <template #default="{ row }">
          <div v-if="row.status && row.status.syncStatus">
            <p class="bcs-ellipsis leading-[24px]" :title="row.status.syncStatus.state">
              {{ $t('plugin.tools.syncStatus') }}：{{ row.status.syncStatus.state || '--' }}
            </p>
            <p class="bcs-ellipsis leading-[24px]" :title="row.status.syncStatus.lastSyncLatency">
              {{ $t('plugin.tools.syncTime') }}：{{ row.status.syncStatus.lastSyncLatency || '--' }}
            </p>
            <p class="bcs-ellipsis leading-[24px]" :title="row.status.syncStatus.lastSyncTime">
              {{ $t('plugin.tools.lastSyncTime') }}：{{ formatDate(row.status.syncStatus.lastSyncTime) || '--' }}
            </p>
          </div>
          <div v-else>--</div>
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('projects.operateAudit.record')" min-width="240">
        <template #default="{ row }">
          <p class="bcs-ellipsis leading-[24px]">
            {{ $t('generic.label.updator') }}：<span>
              {{ customCrdList.handleGetExtData(row.metadata.uid, 'updater') || '--' }}
            </span>
          </p>
          <p class="bcs-ellipsis leading-[24px]">
            {{ $t('cluster.labels.updatedAt') }}：
            {{
              row.status && row.status.applyStatus
                ? (formatDate(row.status.applyStatus.lastApplyTime) || '--')
                : '--'
            }}
          </p>
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('generic.label.action')" width="150">
        <template #default="{ row }">
          <bcs-button
            text
            @click="customCrdList.showUpdateCrdSideslider(row)">
            {{$t('generic.label.update')}}
          </bcs-button>
          <bcs-button
            text
            class="ml-[8px]"
            @click="customCrdList.deleteCrd(row)">
            {{$t('generic.label.delete')}}
          </bcs-button>
        </template>
      </bcs-table-column>
      <template #empty>
        <BcsEmptyTableStatus
          :type="customCrdList.searchValue ? 'search-empty' : 'empty'"
          @clear="customCrdList.handleClearSearchData" />
      </template>
    </bcs-table>
    <!-- 创建 & 更新 -->
    <bcs-sideslider
      :is-show.sync="customCrdList.isShowCreate"
      :title="customCrdList.title"
      quick-close
      :width="800">
      <template #content>
        <bk-form
          :model="formData"
          :rules="rules"
          form-type="vertical"
          class="p-[30px]"
          ref="formRef">
          <bk-form-item
            :label="$t('plugin.tools._ruleName')"
            property="metadata.name"
            error-display-type="normal"
            required>
            <bcs-input
              :disabled="!!customCrdList.currentRow"
              clearable
              class="w-[calc(50%-16px)]"
              :maxlength="64"
              v-model="formData.metadata.name">
            </bcs-input>
          </bk-form-item>
          <div class="flex items-start mt-[8px]">
            <bk-form-item class="flex-1 mr-[32px]" :label="$t('generic.label.cluster1')" required>
              <bcs-input readonly :value="customCrdList.curCluster?.clusterName"></bcs-input>
            </bk-form-item>
            <bk-form-item
              class="flex-1 !mt-0"
              :label="$t('k8s.namespace')"
              property="metadata.namespace"
              error-display-type="normal"
              required>
              <NamespaceSelect
                :disabled="!!customCrdList.currentRow"
                :cluster-id="propClusterId" v-model="formData.metadata.namespace" />
            </bk-form-item>
          </div>
          <!-- Polaris信息 -->
          <bk-form-item :label="$t('plugin.tools.polarisInfo')" class="mt-[8px]">
            <div class="bcs-border p-[20px] bg-[#fafbfd]">
              <div class="flex items-start">
                <bk-form-item
                  class="flex-1 mr-[30px]"
                  :label="$t('generic.label.name')"
                  property="spec.polaris.name"
                  error-display-type="normal"
                  required>
                  <bcs-input
                    :disabled="!!customCrdList.currentRow"
                    :placeholder="$t('plugin.tools.allowNumLettersSymbols')"
                    :maxlength="128"
                    v-model="formData.spec.polaris.name">
                  </bcs-input>
                </bk-form-item>
                <bk-form-item
                  class="flex-1 !mt-0"
                  :label="$t('k8s.namespace')"
                  property="spec.polaris.namespace"
                  error-display-type="normal"
                  required>
                  <bcs-select
                    :disabled="!!customCrdList.currentRow"
                    :class="!customCrdList.currentRow ? 'bg-[#fff]' : ''"
                    searchable
                    v-model="formData.spec.polaris.namespace">
                    <bcs-option
                      v-for="item in polarisNameSpaceList"
                      :key="item.id"
                      :id="item.id"
                      :name="item.name">
                    </bcs-option>
                  </bcs-select>
                </bk-form-item>
              </div>
              <bcs-checkbox
                :disabled="!!customCrdList.currentRow" class="mt-[20px]" v-model="showToken" @change="toggleShowToken">
                {{ $t('plugin.tools.polaris') }}
              </bcs-checkbox>
              <bk-form-item
                class="mt-[8px]"
                label="Token"
                desc-type="icon"
                :desc="$t('plugin.tools.createSvc')"
                property="spec.polaris.token"
                error-display-type="normal"
                required
                v-if="showToken">
                <bcs-input class="w-[calc(50%-15px)]" v-model="formData.spec.polaris.token"></bcs-input>
              </bk-form-item>
            </div>
          </bk-form-item>
          <!-- 关联Service -->
          <bk-form-item
            property="spec.services"
            error-display-type="normal"
            :label="$t('deploy.templateset.service')"
            class="mt-[8px]">
            <div class="bcs-border p-[20px] bg-[#fafbfd]">
              <div
                v-for="item, index in formData.spec.services"
                :key="index"
                class="bcs-border relative px-[15px] py-[20px] bg-[#fff] mb-[10px]">
                <div class="flex items-center">
                  <bk-form-item class="flex-1 mr-[26px]" :label="$t('plugin.tools._serviceName')" required>
                    <bcs-input v-model="item.name"></bcs-input>
                  </bk-form-item>
                  <bk-form-item class="flex-1 !mt-0" :label="$t('deploy.helm.port')" required>
                    <bk-input type="number" :min="1" :max="65535" v-model="item.port"></bk-input>
                  </bk-form-item>
                </div>
                <div class="flex items-center mt-[10px]">
                  <bcs-checkbox class="flex-1 mr-[26px]" v-model="item.direct">
                    {{ $t('plugin.tools.pod') }}
                    <i
                      class="bk-icon icon-question-circle cursor-pointer text-[#979ba5] text-[16px]"
                      v-bk-tooltips.top="$t('plugin.tools.nodePort')">
                    </i>
                  </bcs-checkbox>
                  <bk-form-item class="flex-1" :label="$t('plugin.tools.weight')" required>
                    <bk-input type="number" :min="0" :max="100" v-model="item.weight"></bk-input>
                  </bk-form-item>
                </div>
                <i
                  :class="[
                    'absolute right-[5px] top-[5px] bk-icon icon-close3-shape',
                    'cursor-pointer text-[#979ba5] hover:text-[#3a84ff]'
                  ]"
                  @click="handleDeleteService(index)">
                </i>
              </div>
              <div
                :class="[
                  'bcs-border border-dashed flex items-center justify-center text-[14px] h-[42px] bg-[#fff]',
                  'cursor-pointer hover:border-[#3a84ff] hover:border-solid hover:text-[#3a84ff]',
                  !!formData.spec.services.length ? 'mt-[20px]' : ''
                ]"
                @click="handleAddService">
                <i class="bcs-icon bcs-icon-plus mr-[4px]"></i>
                {{ $t('plugin.tools.clickToAdd') }}
              </div>
            </div>
          </bk-form-item>
          <div class="mt-[25px]">
            <bcs-button :loading="customCrdList.saving" theme="primary" @click="customCrdList.createOrUpdateCrd">
              {{ customCrdList.currentRow ? $t('generic.button.update') : $t('generic.button.create') }}
            </bcs-button>
            <bcs-button @click="customCrdList.isShowCreate = false">{{ $t('generic.button.cancel') }}</bcs-button>
          </div>
        </bk-form>
      </template>
    </bcs-sideslider>
  </BcsContent>
</template>
<script setup lang="ts">
import { ref, watch } from 'vue';

import useCustomCrdList from './use-custom-crd';

import { formatDate } from '@/common/util';
import ClusterSelect from '@/components/cluster-selector/cluster-select.vue';
import BcsContent from '@/components/layout/Content.vue';
import NamespaceSelect from '@/components/namespace-selector/namespace-select.vue';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $store from '@/store';

interface IPolarisFormData {
  metadata: {
    name: string
    namespace: string
  }
  spec: {
    polaris: {
      name: string
      namespace: string
      operator: string
      token: string
    },
    services: Array<{
      'direct': boolean
      'generationMode'?: string
      'name': string
      'namespace': string
      'port'?: number
      'weight': number
    }>,
  }
}

const props = defineProps({
  clusterId: {
    type: String,
    default: '',
    required: true,
  },
});

const polarisNameSpaceList = [
  {
    id: 'Production',
    name: 'Production',
  },
  {
    id: 'Pre-release',
    name: 'Pre-release',
  },
  {
    id: 'Test',
    name: 'Test',
  },
  {
    id: 'Development',
    name: 'Development',
  },
];
// 表单数据
const initFormData = {
  metadata: {
    name: '',
    namespace: '',
  },
  spec: {
    polaris: {
      name: '',
      namespace: '',
      operator: '',
      token: '',
    },
    services: [],
  },
};
const formData = ref<IPolarisFormData>({
  ...initFormData,
});

const propClusterId = ref<string>(props.clusterId);

const ns = ref<string>('');

const getParams = () => {
  const data: IPolarisFormData = JSON.parse(JSON.stringify(formData.value));
  data.spec.polaris.operator = $store.state.user?.username;
  data.spec.services = data.spec.services.map(item => ({
    ...item,
    namespace: data.metadata.namespace,
  }));
  return data;
};

// 表单校验
const rules = ref({
  'metadata.name': [{
    required: true,
    message: $i18n.t('generic.validate.required'),
    trigger: 'blur',
  }, {
    trigger: 'blur',
    message: $i18n.t('plugin.tools.ruleCharacterCriteria'),
    validator: () => /^([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]$/.test(formData.value.metadata.name),
  }],
  'metadata.namespace': [{
    required: true,
    message: $i18n.t('generic.validate.required'),
    trigger: 'blur',
  }],
  'spec.polaris.name': [{
    required: true,
    message: $i18n.t('generic.validate.required'),
    trigger: 'blur',
  }, {
    trigger: 'blur',
    message: $i18n.t('plugin.tools.polarisInfoNameCriteria'),
    validator: () => /^[\w-.:]{1,128}$/.test(formData.value.spec.polaris.name),
  }],
  'spec.polaris.namespace': [{
    required: true,
    message: $i18n.t('generic.validate.required'),
    trigger: 'blur',
  }],
  'spec.polaris.token': [{
    message: $i18n.t('generic.validate.required'),
    trigger: 'blur',
    validator: () => {
      if (showToken.value) {
        return !!formData.value.spec.polaris.token;
      }
      return true;
    },
  }],
  'spec.services': [{
    message: $i18n.t('generic.validate.required'),
    trigger: 'blur',
    validator: () => formData.value.spec.services?.length
      && formData.value.spec.services?.every(item => item.name && item.port && item.weight >= 0),
  }],
});

// hooks

const customCrdList = ref(useCustomCrdList({
  $crd: 'polarisconfigs.tkex.tencent.com',
  $apiVersion: 'tkex.tencent.com/v1',
  $kind: 'PolarisConfig',
  clusterId: propClusterId.value,
  formData,
  initFormData,
  getParams,
  $namespace: ns.value || undefined,
}));

const showToken = ref(false);
watch(() => customCrdList.value.currentRow, () => {
  showToken.value = !!customCrdList.value.currentRow?.spec?.polaris?.token;
});
watch(propClusterId, () => {
  $router.push({
    params: {
      clusterId: propClusterId.value,
    },
  });
});
const toggleShowToken = () => {
  formData.value.spec.polaris.token = '';
};

// 添加service
const handleAddService = () => {
  formData.value.spec.services.push({
    direct: false,
    generationMode: '',
    name: '',
    namespace: '',
    port: undefined,
    weight: 0,
  });
};
// 删除service
const handleDeleteService = (index: number) => {
  formData.value.spec.services.splice(index, 1);
};
const handleClusterChange = () => {
  ns.value = '';
  handleGetList();
};
const handleGetList = () => {
  customCrdList.value.handleGetCrdList({
    $crd: 'polarisconfigs.tkex.tencent.com',
    $clusterId: propClusterId.value,
    $category: 'custom_objects',
    namespace: ns.value || undefined,
  });
};
</script>
