<template>
  <section :class="['service-card', { 'no-view-perm': !props.service.permissions.view }]" @click="handleCardClick">
    <div class="card-content-wrapper">
      <div class="card-head">
        <bk-tag class="type-tag">{{ isFileType ? t('文件型') : t('键值型') }}</bk-tag>
        <div class="service-name">
          <bk-overflow-title type="tips">
            {{ props.service.spec?.name }}
          </bk-overflow-title>
        </div>
      </div>
      <span class="del-btn"><Del @click="handleDeleteItem" /></span>
      <div class="service-alias">
        <bk-overflow-title type="tips">
          {{ props.service.spec?.alias }}
        </bk-overflow-title>
      </div>
      <div class="service-config">
        <div class="config-info">
          <span class="bk-bscp-icon icon-configuration-line"></span>
          <span>{{ props.service.config?.count }}{{ isFileType ? t('个配置文件') : t('个配置项') }}</span>
        </div>
        <div class="time-info">
          <span class="bk-bscp-icon icon-time-2" v-bk-tooltips="{ content: t('最新上线'), placement: 'top' }"></span>
          <template v-if="props.service.config && props.service.config.update_at">
            {{ datetimeFormat(props.service.config.update_at) }}
          </template>
          <template v-else>{{ t('未更新') }}</template>
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
        <div v-else class="apply-btn">{{ t('申请服务权限') }}</div>
      </div>
    </div>
  </section>
</template>
<script setup lang="ts">
  import { computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { useRoute, useRouter } from 'vue-router';
  import { storeToRefs } from 'pinia';
  import { Del } from 'bkui-vue/lib/icon';
  import useGlobalStore from '../../../../../store/global';
  import { IAppItem } from '../../../../../../types/app';
  import { IPermissionQueryResourceItem } from '../../../../../../types/index';
  import { datetimeFormat } from '../../../../../utils/index';

  const { showApplyPermDialog, permissionQuery } = storeToRefs(useGlobalStore());

  const { t } = useI18n();
  const route = useRoute();
  const router = useRouter();

  const props = defineProps<{
    service: IAppItem;
  }>();

  const emits = defineEmits(['edit', 'update', 'delete']);

  const isFileType = computed(() => props.service.spec.config_type === 'file');

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

  const handleDeleteItem = () => {
    if (props.service.permissions.delete) {
      emits('delete', props.service);
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

  const openPermApplyDialog = (query: { resources: IPermissionQueryResourceItem[] }) => {
    permissionQuery.value = query;
    showApplyPermDialog.value = true;
  };

  const goToDetail = () => {
    router.push({ name: 'service-config', params: { spaceId: route.params.spaceId, appId: props.service.id } });
  };
</script>
<style lang="scss" scoped>
  .service-card {
    position: relative;
    width: 304px;
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
      position: relative;
      padding-top: 22px;
      color: #313238;
      line-height: 22px;
      .type-tag {
        position: absolute;
        top: 0;
        left: 0;
        margin: 0;
      }
      .service-name {
        margin-top: 4px;
        padding: 0 20px;
        font-size: 14px;
        font-weight: bold;
        line-height: 22px;
      }
    }
    .service-alias {
      padding: 0 20px;
      font-size: 12px;
      line-height: 20px;
      color: #63656e;
    }
    .service-config {
      display: flex;
      align-items: end;
      justify-content: space-between;
      margin: 12px 8px 16px;
      padding: 0 8px;
      font-size: 12px;
      background: #f5f7fa;
      border-radius: 2px;
      color: #979ba5;
      line-height: 28px;
      .config-info {
        flex-shrink: 0;
      }
      .time-info {
        flex-shrink: 0;
      }
      .bk-bscp-icon {
        font-size: 14px;
        margin-right: 5px;
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
</style>
