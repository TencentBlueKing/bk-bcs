<script setup lang="ts">
  import { useI18n } from "vue-i18n"
  import { useRoute, useRouter } from 'vue-router'
  import { Del } from 'bkui-vue/lib/icon'
  import { InfoBox } from 'bkui-vue/lib'
  import { IAppItem } from '../../../../../../types/app'
  import { deleteApp } from "../../../../../api";


  const { t } = useI18n()
  const route = useRoute()
  const router = useRouter()

  const props = defineProps<{
    service: IAppItem
  }>()

  const emits = defineEmits(['edit', 'update'])

  const handleDeleteItem = () => {
    InfoBox({
      title: `确认是否删除服务 ${props.service.spec.name}?`,
      infoType: "danger",
      headerAlign: "center" as const,
      footerAlign: "center" as const,
      onConfirm: async () => {
        await deleteApp(<number>props.service.id, props.service.biz_id);
        emits('update')
      },
    } as any);
  }

  const goToDetail = () => {
    router.push({ name: 'service-config', params: { spaceId: route.params.spaceId, appId: props.service.id } })
  }
</script>
<template>
  <section class="service-card">
    <div class="card-content-wrapper">
      <div class="card-head">{{ props.service.spec?.name }}</div>
      <span class="del-btn"><Del @click="handleDeleteItem" /></span>
      <div class="service-config">
        <div class="config-info">
          <span class="bk-bscp-icon icon-configuration-line"></span>
          {{ props.service.config?.count }}个配置项
        </div>
        <div class="time-info">
          <span class="bk-bscp-icon icon-time-2"></span>
          {{ props.service.config?.update_at || '未更新' }}
        </div>
      </div>
      <div class="card-footer">
        <bk-button
          size="small"
          text
          @click="emits('edit', props.service)">
          {{ t("服务属性") }}
        </bk-button>
        <span class="divider-middle"></span>
        <bk-button size="small" text @click="goToDetail">
          {{t("配置管理")}}
        </bk-button>
      </div>
    </div>
  </section>
</template>
<style lang="scss" scoped>
  .service-card {
    position: relative;
    width: 304px;
    height: 165px;
    padding: 0px 8px 16px 8px;
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
        content: "";
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
      height: 55px;
      font-size: 12px;
      color: #979ba5;
      line-height: 20px;
      margin: 4px 0 12px 0;
      display: flex;
      align-items: end;
      .config-info {
        width: 80px;
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
    }
  }
</style>
