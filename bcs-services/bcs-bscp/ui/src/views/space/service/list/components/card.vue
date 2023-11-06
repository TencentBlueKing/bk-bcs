<template>
  <section :class="['service-card', { 'no-view-perm': !props.service.permissions.view }]" @click="handleCardClick">
    <div class="card-content-wrapper">
      <div class="card-head">{{ props.service.spec?.name }}</div>
      <span class="del-btn"><Del @click="handleDeleteItem" /></span>
      <div class="service-config">
        <div class="config-info">
          <span class="bk-bscp-icon icon-configuration-line"></span>
          {{ props.service.config?.count }}个配置文件
        </div>
        <div class="time-info">
          <span class="bk-bscp-icon icon-time-2"></span>
          <template v-if="props.service.config && props.service.config.update_at">
            {{ datetimeFormat(props.service.config.update_at) }}
          </template>
          <template v-else>未更新</template>
        </div>
      </div>
      <div class="card-footer">
        <template v-if="props.service.permissions.view">
          <bk-button size="small" text @click="emits('edit', props.service)">
            {{ t('服务属性') }}
          </bk-button>
          <span class="divider-middle"></span>
          <bk-button size="small" text @click="goToDetail">
            {{ t('配置管理') }}
          </bk-button>
        </template>
        <div v-else class="apply-btn">申请服务权限</div>
      </div>
    </div>
  </section>
  <bk-dialog
    ext-cls="delete-service-dialog"
    v-model:is-show="isShowDeleteDialog"
    :theme="'primary'"
    :dialog-type="'operation'"
    header-align="center"
    footer-align="center"
    @value-change="dialogInputStr = ''"
    :draggable="false"
  >
    <div class="dialog-content">
      <div class="dialog-title">确认删除此服务？</div>
      <div>删除的服务<span>无法找回</span>,请谨慎操作！</div>
      <div class="dialog-input">
        <div class="dialog-info">
          请输入服务名<span>{{ service.spec.name }}</span
          >以确认删除
        </div>
        <bk-input v-model="dialogInputStr" />
      </div>
    </div>
    <template #footer>
      <div class="dialog-footer">
        <bk-button
          theme="danger"
          style="margin-right: 20px"
          :disabled="dialogInputStr !== service.spec.name"
          @click="handleDeleteService"
          >删除</bk-button
        >
        <bk-button @click="isShowDeleteDialog = false">取消</bk-button>
      </div>
    </template>
  </bk-dialog>
</template>
<script setup lang="ts">
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';
import { storeToRefs } from 'pinia';
import { Del } from 'bkui-vue/lib/icon';
import useGlobalStore from '../../../../../store/global';
import { IAppItem } from '../../../../../../types/app';
import { IPermissionQueryResourceItem } from '../../../../../../types/index';
import { deleteApp } from '../../../../../api';
import { datetimeFormat } from '../../../../../utils/index';

const { showApplyPermDialog, permissionQuery } = storeToRefs(useGlobalStore());

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const isShowDeleteDialog = ref(false);
const dialogInputStr = ref('');

const props = defineProps<{
  service: IAppItem;
}>();

const emits = defineEmits(['edit', 'update']);

const handleDeleteItem = () => {
  if (props.service.permissions.delete) {
    isShowDeleteDialog.value = true;
  } else {
    const query = {
      resources: [
        {
          biz_id: props.service.biz_id,
          basic: {
            type: 'app',
            action: 'delete',
            resource_id: props.service.id,
          },
        },
      ],
    };
    openPermApplyDialog(query);
  }
};

const handleCardClick = () => {
  if (props.service.permissions.view) {
    return;
  }
  const query = {
    resources: [
      {
        biz_id: props.service.biz_id,
        basic: {
          type: 'app',
          action: 'view',
          resource_id: props.service.id,
        },
      },
    ],
  };
  openPermApplyDialog(query);
};

const openPermApplyDialog = (query: { resources: IPermissionQueryResourceItem[] }) => {
  permissionQuery.value = query;
  showApplyPermDialog.value = true;
};

const goToDetail = () => {
  router.push({ name: 'service-config', params: { spaceId: route.params.spaceId, appId: props.service.id } });
};

const handleDeleteService = async () => {
  await deleteApp(props.service.id as number, props.service.biz_id);
  emits('update');
  isShowDeleteDialog.value = false;
};
</script>
<style lang="scss" scoped>
.service-card {
  position: relative;
  width: 304px;
  height: 143px;
  padding: 0px 8px 16px 8px;
  &.no-view-perm {
    cursor: pointer;
    .card-content-wrapper {
      background: #fafbfd;
      .del-btn {
        display: none !important;
      }
    }
    .card-head {
      color: #c4c6cc;
      &::before {
        background: #dcdee5;
      }
    }
    .service-config {
      color: #c4c6cc;
    }
    &:hover {
      .apply-btn {
        color: #3a84ff;
      }
    }
  }
  .card-content-wrapper {
    height: 100%;
    background: #ffffff;
    border: 1px solid #dcdee5;
    border-radius: 2px;
    text-align: left;
    &:hover {
      .del-btn {
        display: block;
      }
    }
  }
  .del-btn {
    display: none;
    position: absolute;
    right: 18px;
    top: 18px;
    color: #979ba5;
    cursor: pointer;
    z-index: 1;
    &:hover {
      color: #3a84ff;
    }
  }
  .card-head {
    margin-top: 16px;
    position: relative;
    height: 22px;
    font-weight: Bold;
    font-size: 14px;
    color: #313238;
    line-height: 22px;
    text-align: left;
    padding: 0 50px 0 16px;
    display: flex;
    align-items: center;

    &::before {
      content: '';
      position: absolute;
      left: 0;
      top: 3px;
      width: 4px;
      height: 16px;
      background: #699df4;
      border-radius: 0 2px 2px 0;
    }
  }
  .service-config {
    padding: 0 16px;
    height: 33px;
    font-size: 12px;
    color: #979ba5;
    line-height: 20px;
    margin: 4px 0 12px 0;
    display: flex;
    align-items: end;
    .config-info {
      width: 85px;
    }
    .time-info {
      padding-left: 10px;
    }
  }
  .card-footer {
    height: 40px;
    border-top: solid 1px #f0f1f5;
    display: flex;
    justify-content: center;
    width: 100%;
    font-size: 12px;
    :deep(.bk-button) {
      width: 50%;
      &.is-text {
        color: #979ba5;
      }
      &:hover {
        color: #3a84ff;
      }
    }
    .divider-middle {
      display: inline-block;
      width: 1px;
      height: 100%;
      background: #f0f1f5;
      margin: 0 16px;
    }
    .apply-btn {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 100%;
      height: 100%;
      color: #979ba5;
    }
  }
}
.dialog-content {
  text-align: center;
  margin: 10px 0 20px;
  span {
    color: red;
  }
  .dialog-title {
    margin: 10px;
    font-size: 24px;
    color: #121213;
  }
  .dialog-input {
    margin-top: 10px;
    text-align: start;
    padding: 20px;
    background-color: #f4f7fa;
    .dialog-info {
      margin-bottom: 5px;
      span {
        color: #121213;
        font-weight: 600;
      }
    }
  }
}
.dialog-footer {
  .bk-button {
    width: 100px;
  }
}
</style>

<style lang="scss">
.delete-service-dialog {
  .bk-modal-header {
    display: none;
  }
  .bk-modal-footer {
    border: none !important;
    background-color: #fff !important;
  }
}
</style>
