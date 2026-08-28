<template>
  <div class="detail p30">
    <div class="detail-title">{{ $t('generic.title.basicInfo') }}</div>
    <div class="detail-content basic-info">
      <div class="basic-info-item">
        <label>UID</label>
        <span class="bcs-ellipsis">{{ data.uid || '--' }}</span>
      </div>
      <div class="basic-info-item">
        <label>{{ $t('generic.label.status') }}</label>
        <span class="bcs-ellipsis">{{ data.status || '--' }}</span>
      </div>
      <div class="basic-info-item">
        <label>{{ $t('cluster.labels.createdAt') }}</label>
        <span>{{ data.createTime ? timeZoneTransForm(data.createTime, false) : '--' }}</span>
      </div>
    </div>
    <div class="detail-title mt-[20px]">{{ $t('logCollector.label.configInfo') }}</div>
    <div class="detail-content basic-info">
      <div class="basic-info-item">
        <label>{{ $t('metrics.cpuUsage') }}</label>
        <span class="bcs-ellipsis" v-if="data.quota">
          {{ data.cpuUseRate.toFixed(2) * 100 }}%
          （{{ `${formatQuotaQuantity(data.used ? data.used.cpuLimits : '0', 'cpu')}${$t('units.suffix.cores')}` }}
          / {{ `${formatQuotaQuantity(data.quota ? data.quota.cpuLimits : '0', 'cpu')}${$t('units.suffix.cores')}` }}）
        </span>
        <span class="bcs-ellipsis" v-else>{{ $t('dashboard.ns.tips.notEnabledNamespaceQuota') }}</span>
      </div>
      <div class="basic-info-item">
        <label>{{ $t('metrics.memUsage') }}</label>
        <span class="bcs-ellipsis" v-if="data.quota">
          {{ data.memoryUseRate.toFixed(2) * 100 }}%
          （{{ `${formatQuotaQuantity(data.used ? data.used.memoryLimits : '0', 'mem')}Gi` }}
          / {{ `${formatQuotaQuantity(data.quota ? data.quota.memoryLimits : '0', 'mem')}Gi` }}）
        </span>
        <span class="bcs-ellipsis" v-else>{{ $t('dashboard.ns.tips.notEnabledNamespaceQuota') }}</span>
      </div>
    </div>
    <bcs-tab class="mt20" type="card" :label-height="42">
      <bcs-tab-panel name="label" :label="$t('k8s.label')">
        <bk-table :data="data.labels">
          <bk-table-column label="Key" prop="key"></bk-table-column>
          <bk-table-column label="Value" prop="value"></bk-table-column>
        </bk-table>
      </bcs-tab-panel>
      <bcs-tab-panel name="annotations" :label="$t('k8s.annotation')">
        <bk-table :data="data.annotations">
          <bk-table-column label="Key" prop="key"></bk-table-column>
          <bk-table-column label="Value" prop="value"></bk-table-column>
        </bk-table>
      </bcs-tab-panel>
      <bcs-tab-panel name="config" :label="$t('generic.label.var')">
        <bk-table :data="data.variables">
          <bk-table-column label="Key" prop="key"></bk-table-column>
          <bk-table-column label="Value" prop="value"></bk-table-column>
        </bk-table>
      </bcs-tab-panel>
      <bcs-tab-panel
        name="otherQuotas"
        :label="$t('dashboard.ns.label.otherQuotas')"
        v-if="editable || (data.otherQuotas && data.otherQuotas.length)">
        <div class="other-quota-toolbar" v-if="editable">
          <bcs-button theme="primary" icon="plus" @click="handleCreateQuota">
            {{ $t('dashboard.ns.action.createOtherQuota') }}
          </bcs-button>
        </div>
        <bk-table class="other-quota-table" :data="data.otherQuotas || []">
          <bk-table-column
            :label="$t('generic.label.name')"
            prop="name"
            min-width="120"
            show-overflow-tooltip>
          </bk-table-column>
          <bk-table-column :label="$t('dashboard.ns.label.cpuRequests')" min-width="105">
            <template #default="{ row }">
              <div class="quota-usage">
                <span>{{ formatQuotaUsage(row, 'cpuRequests', 'cpu') }}</span>
                <span>{{ formatUsageRate(row, 'cpuRequests') }}</span>
              </div>
            </template>
          </bk-table-column>
          <bk-table-column :label="$t('dashboard.ns.label.cpuLimits')" min-width="105">
            <template #default="{ row }">
              <div class="quota-usage">
                <span>{{ formatQuotaUsage(row, 'cpuLimits', 'cpu') }}</span>
                <span>{{ formatUsageRate(row, 'cpuLimits') }}</span>
              </div>
            </template>
          </bk-table-column>
          <bk-table-column :label="$t('dashboard.ns.label.memoryRequests')" min-width="120">
            <template #default="{ row }">
              <div class="quota-usage">
                <span>{{ formatQuotaUsage(row, 'memoryRequests', 'mem') }}</span>
                <span>{{ formatUsageRate(row, 'memoryRequests') }}</span>
              </div>
            </template>
          </bk-table-column>
          <bk-table-column :label="$t('dashboard.ns.label.memoryLimits')" min-width="120">
            <template #default="{ row }">
              <div class="quota-usage">
                <span>{{ formatQuotaUsage(row, 'memoryLimits', 'mem') }}</span>
                <span>{{ formatUsageRate(row, 'memoryLimits') }}</span>
              </div>
            </template>
          </bk-table-column>
          <bk-table-column
            :label="$t('generic.label.action')"
            width="90"
            v-if="editable">
            <template #default="{ row }">
              <bk-button text class="mr-[8px]" @click="handleEditQuota(row)">
                {{ $t('generic.button.edit') }}
              </bk-button>
              <bk-button text @click="handleDeleteQuota(row)">
                {{ $t('generic.button.delete') }}
              </bk-button>
            </template>
          </bk-table-column>
        </bk-table>
      </bcs-tab-panel>
    </bcs-tab>

    <bcs-dialog
      v-model="quotaDialog.isShow"
      :title="quotaDialog.isEdit
        ? $t('dashboard.ns.title.editOtherQuota')
        : $t('dashboard.ns.title.createOtherQuota')"
      :width="650">
      <bk-form
        :label-width="120"
        v-bkloading="{ isLoading: quotaDialog.loading }"
        :model="quotaDialog.form">
        <bk-form-item
          :label="$t('dashboard.ns.label.quotaName')"
          property="name"
          :rules="nameRules"
          error-display-type="normal"
          required>
          <bcs-input
            v-model="quotaDialog.form.name"
            class="w-[410px]"
            :disabled="quotaDialog.isEdit"
            :maxlength="63">
          </bcs-input>
        </bk-form-item>
        <bk-form-item
          label="CPU"
          property="quota"
          :rules="quotaRules"
          error-display-type="normal"
          required>
          <div class="flex items-center">
            <span class="mr-[10px] text-[12px] text-[#979ba5]">Request</span>
            <bcs-input
              v-model="quotaDialog.form.quota.cpuRequests"
              class="w-[150px] mr-[20px]"
              type="text">
              <div class="group-text" slot="append">{{ $t('units.suffix.cores') }}</div>
            </bcs-input>
            <span class="mr-[10px] text-[12px] text-[#979ba5]">Limit</span>
            <bcs-input
              v-model="quotaDialog.form.quota.cpuLimits"
              class="w-[150px]"
              type="text">
              <div class="group-text" slot="append">{{ $t('units.suffix.cores') }}</div>
            </bcs-input>
          </div>
        </bk-form-item>
        <bk-form-item
          label="Memory"
          error-display-type="normal"
          required>
          <div class="flex items-center">
            <span class="mr-[10px] text-[12px] text-[#979ba5]">Request</span>
            <bcs-input
              v-model="quotaDialog.form.quota.memoryRequests"
              class="w-[150px] mr-[20px]"
              type="text">
              <div class="group-text" slot="append">GiB</div>
            </bcs-input>
            <span class="mr-[10px] text-[12px] text-[#979ba5]">Limit</span>
            <bcs-input
              v-model="quotaDialog.form.quota.memoryLimits"
              class="w-[150px]"
              type="text">
              <div class="group-text" slot="append">GiB</div>
            </bcs-input>
          </div>
        </bk-form-item>
      </bk-form>
      <div slot="footer">
        <bcs-button
          theme="primary"
          class="mr5"
          :loading="quotaDialog.loading"
          @click="handleSaveQuota">
          {{ $t('generic.button.confirm') }}
        </bcs-button>
        <bcs-button :disabled="quotaDialog.loading" @click="handleCancelQuota">
          {{ $t('generic.button.cancel') }}
        </bcs-button>
      </div>
    </bcs-dialog>
  </div>
</template>
<script lang="ts">
/* eslint-disable camelcase */
import { defineComponent, ref } from 'vue';

import {
  formatQuotaQuantity,
  isQuotaFormValueValid,
  QuotaField,
  quotaToFormValues,
  QuotaValues,
  serializeQuotaFormValues,
} from './other-quota';

import { createOtherQuota, deleteOtherQuota, updateOtherQuota } from '@/api/modules/project';
import $bkMessage from '@/common/bkmagic';
import { timeZoneTransForm } from '@/common/util';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import $i18n from '@/i18n/i18n-setup';

const createQuotaForm = () => ({
  name: '',
  quota: {
    cpuRequests: '0',
    cpuLimits: '0',
    memoryRequests: '0',
    memoryLimits: '0',
  },
});

export default defineComponent({
  name: 'NamespaceDetail',
  props: {
    data: {
      type: Object,
      default: () => ({}),
    },
    clusterId: {
      type: String,
      default: '',
    },
    editable: {
      type: Boolean,
      default: false,
    },
  },
  emits: ['refresh'],
  setup(props, { emit }) {
    const formatQuotaUsage = (row, field: QuotaField, type: 'cpu' | 'mem') => {
      const hard = row.quota?.[field];
      if (!hard) return '--';

      const unit = type === 'cpu' ? $i18n.t('units.suffix.cores') : 'Gi';
      const used = formatQuotaQuantity(row.used?.[field] || '0', type);
      const total = formatQuotaQuantity(hard, type);
      return `${used}/${total} ${unit}`;
    };
    const formatUsageRate = (row, field: QuotaField) => {
      if (!row.quota?.[field]) return '';
      const rate = Number(row.usageRate?.[field] || 0) * 100;
      return `${rate.toFixed(2)}%`;
    };

    const originalQuota = ref<Partial<QuotaValues>>();
    const quotaDialog = ref({
      isShow: false,
      isEdit: false,
      loading: false,
      form: createQuotaForm(),
    });
    const validateQuotaName = () => {
      const { name } = quotaDialog.value.form;
      return name !== props.data.name && /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/.test(name);
    };
    const validateQuotaResource = () => {
      const { quota } = quotaDialog.value.form;
      const values = Object.values(quota);
      return values.every(isQuotaFormValueValid);
    };
    const nameRules = [{
      validator: validateQuotaName,
      message: $i18n.t('dashboard.ns.validate.otherQuotaName'),
      trigger: 'blur',
    }];
    const quotaRules = [{
      validator: validateQuotaResource,
      message: $i18n.t('dashboard.ns.validate.otherQuotaResource'),
      trigger: 'blur',
    }];
    const handleCreateQuota = () => {
      originalQuota.value = undefined;
      quotaDialog.value = {
        isShow: true,
        isEdit: false,
        loading: false,
        form: createQuotaForm(),
      };
    };
    const handleEditQuota = (row) => {
      originalQuota.value = { ...row.quota };
      quotaDialog.value = {
        isShow: true,
        isEdit: true,
        loading: false,
        form: {
          name: row.name,
          quota: quotaToFormValues(row.quota || {}),
        },
      };
    };
    const handleCancelQuota = () => {
      if (quotaDialog.value.loading) return;
      quotaDialog.value.isShow = false;
    };
    const handleSaveQuota = async () => {
      if (!validateQuotaName()) {
        $bkMessage({
          theme: 'error',
          message: $i18n.t('dashboard.ns.validate.otherQuotaName'),
        });
        return;
      }
      if (!validateQuotaResource()) {
        $bkMessage({
          theme: 'error',
          message: $i18n.t('dashboard.ns.validate.otherQuotaResource'),
        });
        return;
      }

      const { form, isEdit } = quotaDialog.value;
      const quota = serializeQuotaFormValues(form.quota, originalQuota.value);
      quotaDialog.value.loading = true;
      const request = isEdit ? updateOtherQuota : createOtherQuota;
      const result = await request({
        $clusterId: props.clusterId,
        $namespace: props.data.name,
        ...(isEdit ? { $quotaName: form.name } : { quotaName: form.name }),
        quota,
      }).then(() => true)
        .catch(() => false);
      quotaDialog.value.loading = false;
      if (!result) return;

      $bkMessage({
        theme: 'success',
        message: $i18n.t(isEdit ? 'generic.msg.success.update' : 'generic.msg.success.create'),
      });
      quotaDialog.value.isShow = false;
      emit('refresh');
    };
    const handleDeleteQuota = (row) => {
      $bkInfo({
        type: 'warning',
        clsName: 'custom-info-confirm',
        title: $i18n.t('generic.title.confirmDelete1', { name: row.name }),
        subTitle: $i18n.t('dashboard.ns.tips.deleteOtherQuotaWarning'),
        defaultInfo: true,
        confirmFn: async () => {
          const result = await deleteOtherQuota({
            $clusterId: props.clusterId,
            $namespace: props.data.name,
            $quotaName: row.name,
          }).then(() => true)
            .catch(() => false);
          if (!result) return;

          $bkMessage({
            theme: 'success',
            message: $i18n.t('generic.msg.success.delete'),
          });
          emit('refresh');
        },
      });
    };

    return {
      formatQuotaQuantity,
      formatQuotaUsage,
      formatUsageRate,
      handleCancelQuota,
      handleCreateQuota,
      handleDeleteQuota,
      handleEditQuota,
      handleSaveQuota,
      nameRules,
      quotaDialog,
      quotaRules,
      timeZoneTransForm,
    };
  },
});
</script>
<style lang="postcss" scoped>
.detail {
  font-size: 14px;
  /deep/ .bk-tab-label-item {
      background-color: #FAFBFD;
      border-bottom: 1px solid #dcdee5;
      line-height: 41px !important;
      height: 41px;
      &.active {
          border-bottom: none;
      }
  }
  /deep/ .bk-tab-label-wrapper {
      overflow: unset !important;
  }
  &-title {
      margin-bottom: 10px;
      color: #313238;
  }
  &-content {
      &.basic-info {
          border: 1px solid #dfe0e5;
          border-radius: 2px;
          .basic-info-item {
              display: flex;
              align-items: center;
              height: 32px;
              padding: 0 15px;
              &:nth-of-type(even) {
                  background: #F7F8FA;
              }
              label {
                  line-height: 32px;
                  border-right: 1px solid #dfe0e5;
                  width: 200px;
              }
              span {
                  padding: 0 15px;
                  flex: 1;
                  overflow: hidden;
                  text-overflow: ellipsis;
                  white-space: normal;
                  word-break: break-all;
                  display: -webkit-box;
                  -webkit-line-clamp: 1;
                  -webkit-box-orient: vertical;
              }
          }
      }
  }
}
.other-quota-toolbar {
  padding: 12px 0;
}
.other-quota-table {
  /deep/ .cell {
    padding: 0 8px;
  }
}
.quota-usage {
  display: flex;
  flex-direction: column;
  line-height: 20px;
  white-space: nowrap;
  span:last-child {
    color: #979ba5;
    font-size: 12px;
  }
}
</style>
