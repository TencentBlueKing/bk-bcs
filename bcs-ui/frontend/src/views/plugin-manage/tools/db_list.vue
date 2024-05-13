<template>
  <BcsContent
    :title="$t('plugin.tools.dbAtuh')"
    :desc="`(${$t('plugin.tools.cluster', { name: curCluster?.clusterName })})`">
    <div class="flex items-center justify-between mb-[16px]">
      <bcs-button theme="primary" icon="plus" @click="showCreateCrdSideslider">
        {{$t('plugin.tools.add')}}
      </bcs-button>
      <bcs-input
        class="w-[320px]"
        right-icon="bk-icon icon-search"
        clearable
        :placeholder="$t('generic.placeholder.searchName')"
        v-model.trim="searchValue">
      </bcs-input>
    </div>
    <bcs-table
      size="medium"
      :data="curPageData"
      :pagination="pagination"
      v-bkloading="{ isLoading: tableLoading }"
      @page-change="pageChange"
      @page-limit-change="pageSizeChange">
      <bcs-table-column :label="$t('generic.label.name')" show-overflow-tooltip min-width="150">
        <template #default="{ row }">
          <bcs-button text @click="showDbDetail(row)">{{ row.metadata.name || '--' }}</bcs-button>
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('k8s.namespace')" prop="metadata.namespace" min-width="100">
      </bcs-table-column>
      <bcs-table-column :label="$t('cluster.labels.createdAt')" min-width="100">
        <template #default="{ row }">
          {{ formatDate(handleGetExtData(row.metadata.uid, 'createTime')) || '--' }}
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('generic.label.updator')" min-width="100">
        <template #default="{ row }">
          {{ handleGetExtData(row.metadata.uid, 'updater') || '--' }}
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('generic.label.action')" width="150">
        <template #default="{ row }">
          <bcs-button text @click="showUpdateCrdSideslider(row)">{{$t('generic.label.update')}}</bcs-button>
          <bcs-button text class="ml-[8px]" @click="deleteCrd(row)">{{$t('generic.label.delete')}}</bcs-button>
        </template>
      </bcs-table-column>
      <template #empty>
        <BcsEmptyTableStatus
          :type="searchValue ? 'search-empty' : 'empty'"
          @clear="handleClearSearchData" />
      </template>
    </bcs-table>
    <!-- 创建 & 更新 -->
    <bcs-sideslider
      :is-show.sync="isShowCreate"
      :title="title"
      quick-close
      :width="660">
      <template #content>
        <bk-form
          :model="formData"
          :rules="rules"
          form-type="vertical"
          class="grid grid-cols-2 gap-x-[35px] p-[30px]"
          ref="formRef">
          <bk-form-item class="mt-[8px]" :label="$t('generic.label.cluster1')" required>
            <bcs-input readonly :value="curCluster?.clusterName"></bcs-input>
          </bk-form-item>
          <bk-form-item
            property="metadata.namespace"
            :label="$t('k8s.namespace')"
            error-display-type="normal"
            required>
            <NamespaceSelect :disabled="!!currentRow" :cluster-id="clusterId" v-model="formData.metadata.namespace" />
          </bk-form-item>
          <bk-form-item
            property="metadata.name"
            :label="$t('generic.label.name')"
            error-display-type="normal"
            required>
            <bcs-input :disabled="!!currentRow" v-model="formData.metadata.name"></bcs-input>
          </bk-form-item>
          <bk-form-item
            property="spec.appName"
            :label="$t('plugin.tools.biz')"
            desc-type="icon"
            :desc="$t('plugin.tools.bizTips')"
            error-display-type="normal"
            required>
            <bcs-input v-model="formData.spec.appName"></bcs-input>
          </bk-form-item>
          <bk-form-item
            property="spec.targetDb"
            :label="$t('plugin.tools.DBAddress')"
            error-display-type="normal"
            required>
            <bcs-input v-model="formData.spec.targetDb"></bcs-input>
          </bk-form-item>
          <bk-form-item
            property="spec.dbType"
            :label="$t('plugin.tools.DBType')"
            error-display-type="normal"
            required>
            <bcs-select v-model="formData.spec.dbType">
              <bcs-option id="mysql" name="mysql"></bcs-option>
              <bcs-option id="spider" name="spider"></bcs-option>
            </bcs-select>
          </bk-form-item>
          <bk-form-item
            property="spec.callUser"
            :label="$t('plugin.tools.user')"
            desc-type="icon"
            :desc="$t('plugin.tools.userTips')"
            error-display-type="normal"
            required>
            <bcs-input v-model="formData.spec.callUser"></bcs-input>
          </bk-form-item>
          <bk-form-item
            property="spec.dbName"
            :label="$t('plugin.tools.DBName')"
            desc-type="icon"
            :desc="$t('plugin.tools.DBTips')"
            error-display-type="normal"
            required>
            <bcs-input v-model="formData.spec.dbName"></bcs-input>
          </bk-form-item>
          <bk-form-item
            class="col-span-2"
            property="spec.podSelector"
            desc-type="icon"
            :desc="$t('plugin.tools.DBAuthTips')"
            :label="$t('generic.label.labelManage')"
            error-display-type="normal"
            required>
            <KeyValue v-model="formData.spec.podSelector" />
          </bk-form-item>
          <div>
            <bcs-button :loading="saving" theme="primary" @click="createOrUpdateCrd">
              {{ currentRow ? $t('generic.button.update') : $t('generic.button.create') }}
            </bcs-button>
            <bcs-button @click="isShowCreate = false">{{ $t('generic.button.cancel') }}</bcs-button>
          </div>
        </bk-form>
      </template>
    </bcs-sideslider>
    <!-- 详情 -->
    <bcs-sideslider
      :is-show.sync="isShowDetail"
      :title="currentRow && currentRow.metadata.name"
      quick-close
      :width="800">
      <template #content>
        <div class="p-[30px] text-[14px]">
          <!-- 基本信息 -->
          <div class="mb-[10px] text-[#333948]">{{$t('generic.title.basicInfo')}}</div>
          <div
            :class="[
              'flex items-center border-solid border-[1px] border-[#dfe0e5]',
              'h-[82px] text-[#737987] mb-[15px]'
            ]">
            <div class="flex-1 flex flex-col justify-center h-full p-[15px] bcs-border-right">
              <span class="mb-[10px]">{{ $t('generic.label.cluster1') }}:</span>
              <span>{{ curCluster?.clusterName || '--' }}</span>
            </div>
            <div class="flex-1 flex flex-col justify-center h-full p-[15px] bcs-border-right">
              <span class="mb-[10px]">{{ $t('k8s.namespace') }}:</span>
              <span>{{ currentRow.metadata.namespace || '--' }}</span>
            </div>
            <div class="flex-1 flex flex-col justify-center h-full p-[15px]">
              <span class="mb-[10px]">{{ $t('generic.label.name') }}:</span>
              <span>{{ currentRow.metadata.name }}</span>
            </div>
          </div>
          <!-- DB信息 -->
          <div class="mb-[10px] text-[#333948]">
            {{$t('plugin.tools.DBInfo')}}
          </div>
          <div
            :class="[
              'flex items-center border-solid border-[1px] border-[#dfe0e5]',
              'h-[82px] text-[#737987] mb-[15px]'
            ]">
            <div class="flex-1 flex flex-col justify-center h-full p-[15px] bcs-border-right">
              <span class="mb-[10px]">{{$t('plugin.tools.biz')}}：</span>
              <span>{{ currentRow.spec.appName || '--'}}</span>
            </div>
            <div class="flex-1 flex flex-col justify-center h-full p-[15px] bcs-border-right">
              <span class="mb-[10px]">{{$t('plugin.tools.DBAddress')}}：</span>
              <span>{{ currentRow.spec.targetDb || '--'}}</span>
            </div>
            <div class="flex-1 flex flex-col justify-center h-full p-[15px] bcs-border-right">
              <span class="mb-[10px]">{{$t('plugin.tools.DBType')}}：</span>
              <span>{{ currentRow.spec.dbType || '--'}}</span>
            </div>
            <div class="flex-1 flex flex-col justify-center h-full p-[15px] bcs-border-right">
              <span class="mb-[10px]">{{$t('plugin.tools.user')}}：</span>
              <span>{{ currentRow.spec.callUser || '--'}}</span>
            </div>
            <div class="flex-1 flex flex-col justify-center h-full p-[15px]">
              <span class="mb-[10px]">{{$t('plugin.tools.DBName')}}：</span>
              <span>{{ currentRow.spec.dbName || '--'}}</span>
            </div>
          </div>
          <!-- 标签 -->
          <div class="mb-[10px] text-[#333948]">
            {{$t('k8s.label')}}
          </div>
          <div>
            <bcs-tag v-for="(value, key) in currentRow.spec.podSelector" :key="key">
              {{ `${key}:${value}` }}
            </bcs-tag>
          </div>
        </div>
      </template>
    </bcs-sideslider>
  </BcsContent>
</template>
<script setup lang="ts">
import { ref } from 'vue';

import useCustomCrdList from './use-custom-crd';

import { formatDate } from '@/common/util';
import BcsContent from '@/components/layout/Content.vue';
import NamespaceSelect from '@/components/namespace-selector/namespace-select.vue';
import $i18n from '@/i18n/i18n-setup';
import KeyValue from '@/views/cluster-manage/components/key-value.vue';

const props = defineProps({
  clusterId: {
    type: String,
    default: '',
    required: true,
  },
});

// 表单数据
const initFormData = {
  metadata: {
    name: '',
    namespace: '',
  },
  spec: {
    appName: '',
    callUser: '',
    dbName: '',
    dbType: '',
    // operator: '',
    podSelector: {},
    targetDb: '',
  },
};
const formData = ref({
  ...initFormData,
});

// hooks
const {
  curCluster,
  currentRow,
  curPageData,
  pagination,
  tableLoading,
  saving,
  searchValue,
  isShowCreate,
  title,
  formRef,
  pageChange,
  pageSizeChange,
  handleGetExtData,
  showCreateCrdSideslider,
  showUpdateCrdSideslider,
  createOrUpdateCrd,
  deleteCrd,
  handleClearSearchData,
} = useCustomCrdList({
  $crd: 'bcsdbprivconfigs.bkbcs.tencent.com',
  $apiVersion: 'bkbcs.tencent.com/v1',
  $kind: 'BcsDbPrivConfig',
  clusterId: props.clusterId,
  formData,
  initFormData,
});

// 表单校验
const rules = ref({
  'metadata.name': [{
    required: true,
    message: $i18n.t('generic.validate.required'),
    trigger: 'blur',
  }],
  'metadata.namespace': [{
    required: true,
    message: $i18n.t('generic.validate.required'),
    trigger: 'blur',
  }],
  'spec.appName': [{
    required: true,
    message: $i18n.t('generic.validate.required'),
    trigger: 'blur',
  }],
  'spec.callUser': [{
    required: true,
    message: $i18n.t('generic.validate.required'),
    trigger: 'blur',
  }],
  'spec.dbName': [{
    required: true,
    message: $i18n.t('generic.validate.required'),
    trigger: 'blur',
  }],
  'spec.dbType': [{
    required: true,
    message: $i18n.t('generic.validate.required'),
    trigger: 'blur',
  }],
  'spec.targetDb': [{
    required: true,
    message: $i18n.t('generic.validate.required'),
    trigger: 'blur',
  }],
  'spec.podSelector': [{
    validator: () => !!Object.keys(formData.value.spec.podSelector).length,
    message: $i18n.t('generic.validate.required'),
    trigger: 'blur',
  }],
});

// 详情页
const isShowDetail = ref(false);
const showDbDetail = (row) => {
  currentRow.value = row;
  isShowDetail.value = true;
};
</script>
