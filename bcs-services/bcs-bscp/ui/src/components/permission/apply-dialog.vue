<template>
  <bk-dialog
    ext-cls="version-compare-dialog"
    title=""
    :width="768"
    :is-show="show"
    :esc-close="false"
    :quick-close="false"
    @closed="handleClose"
  >
    <div class="permission-modal">
      <div class="permission-header">
        <span class="title-icon">
          <img :src="LockIcon" alt="permission-lock" class="lock-img" />
        </span>
        <h3>{{ t('该操作需要以下权限') }}</h3>
      </div>
      <table class="permission-table table-header">
        <thead>
          <tr>
            <th width="40%">{{ t('需要申请的权限') }}</th>
            <th width="60%">{{ t('关联的资源实例') }}</th>
          </tr>
        </thead>
      </table>
      <div class="table-content">
        <table class="permission-table">
          <tbody>
            <template v-if="resources.length > 0">
              <tr v-for="(resource, index) in resources" :key="index">
                <td width="40%">{{ resource.action_name }}</td>
                <td width="60%">
                  <p class="resource-type-item">
                    {{ resource.resource_name }}
                  </p>
                </td>
              </tr>
            </template>
            <tr v-else>
              <td class="no-data" colspan="2">{{ t('无数据') }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
    <template #footer>
      <div class="button-group">
        <bk-button theme="primary" :disabled="loading" @click="handleSubmitClick">
          {{ clicked ? t('已申请') : t('去申请') }}
        </bk-button>
        <bk-button @click="handleClose">{{ t('取消') }}</bk-button>
      </div>
    </template>
  </bk-dialog>
</template>
<script setup lang="ts">
import { ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { storeToRefs } from 'pinia';
import useGlobalStore from '../../store/global';
import { IPermissionResource } from '../../../types/index';
import { permissionCheck } from '../../api/index';
import LockIcon from '../../assets/lock-radius.svg';

const globalStore = useGlobalStore();
const { showApplyPermDialog, permissionQuery } = storeToRefs(globalStore);
const { t } = useI18n();

const show = ref(false);
const loading = ref(false);
const url = ref('');
const resources = ref<IPermissionResource[]>([]);
const clicked = ref(false);

watch(
  () => showApplyPermDialog.value,
  (val) => {
    clicked.value = false;
    show.value = val;
    if (val) {
      getPermUrl();
    }
  },
);

const getPermUrl = async () => {
  loading.value = true;
  const res = await permissionCheck(permissionQuery.value);
  resources.value = res.resources;
  url.value = res.apply_url;
  loading.value = false;
};
const goToIAM = () => {
  window.open(url.value, '__blank');
  clicked.value = true;
};

const handleSubmitClick = () => {
  if (clicked.value) {
    window.location.reload();
  } else {
    goToIAM();
  }
};

const handleClose = () => {
  showApplyPermDialog.value = false;
  permissionQuery.value = {
    resources: [],
  };
};
</script>
<style lang="scss" scoped>
.permission-modal {
  margin-bottom: 40px;
  .permission-header {
    text-align: center;
    .title-icon {
      display: inline-block;
    }
    .lock-img {
      width: 120px;
    }
    h3 {
      margin: 6px 0 24px;
      color: #63656e;
      font-size: 20px;
      font-weight: normal;
      line-height: 1;
    }
  }
  .permission-table {
    width: 100%;
    color: #63656e;
    border-bottom: 1px solid #e7e8ed;
    border-collapse: collapse;
    table-layout: fixed;
    th,
    td {
      padding: 12px 18px;
      font-size: 12px;
      text-align: left;
      border-bottom: 1px solid #e7e8ed;
      word-break: break-all;
    }
    th {
      color: #313238;
      background: #f5f6fa;
    }
  }
  .table-content {
    max-height: 260px;
    border-bottom: 1px solid #e7e8ed;
    border-top: none;
    overflow: auto;
    .permission-table {
      border-top: none;
      border-bottom: none;
      td:last-child {
        border-right: none;
      }
      tr:last-child td {
        border-bottom: none;
      }
      .resource-type-item {
        padding: 0;
        margin: 0;
      }
    }
    .no-data {
      padding: 30px;
      text-align: center;
      color: #999999;
    }
  }
}
.button-group {
  .bk-button {
    margin-left: 7px;
  }
}
</style>
