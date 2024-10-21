<template>
  <bk-sideslider width="640" quick-close :is-show="props.show" :before-close="handleBeforeClose" @closed="close">
    <template #header>
      <div class="service-edit-head">
        <span class="title">{{ isViewMode ? t('服务属性') : t('编辑服务') }}</span>
        <bk-button v-if="isViewMode" class="edit-entry-btn" theme="primary" @click="isViewMode = false">
          {{ t('编辑') }}
        </bk-button>
      </div>
    </template>
    <div class="service-edit-wrapper">
      <bk-form v-if="isViewMode" label-width="100">
        <bk-form-item :label="t('form_服务名称')">{{ serviceData!.spec.name }}</bk-form-item>
        <bk-form-item :label="t('form_服务别名')">{{ serviceData!.spec.alias }}</bk-form-item>
        <bk-form-item :label="t('服务描述')">
          {{ serviceData!.spec.memo || '--' }}
        </bk-form-item>
        <bk-form-item :label="t('数据格式')">
          {{ serviceData!.spec.config_type === 'file' ? t('文件型') : t('键值型') }}
        </bk-form-item>
        <bk-form-item v-if="serviceData!.spec.config_type !== 'file'" :label="t('数据类型')">
          {{ serviceData!.spec.data_type === 'any' ? t('任意类型') : serviceData!.spec.data_type }}
        </bk-form-item>
        <bk-form-item>
          <template #label>
            <div :class="{ 'offset-y': serviceData!.spec.is_approve }">
              {{ t('上线审批') }}
            </div>
          </template>
          <ul class="approval-info" v-if="serviceData!.spec.is_approve">
            <li class="approval-li">
              <div class="approval-hd">指定审批人</div>
              <div class="approval-bd">
                <bk-member-selector
                  v-model="selectionsApprover"
                  :api="getApproverListApi()"
                  :is-error="!selectionsApprover.length"
                  type="info" />
              </div>
            </li>
            <li class="approval-li">
              <div class="approval-hd">审批方式</div>
              <div class="approval-bd">{{ serviceData!.spec.approve_type === 'OrSign' ? t('或签') : t('会签') }}</div>
            </li>
          </ul>
          <span v-else>未开启</span>
        </bk-form-item>
        <bk-form-item :label="t('创建者')">
          {{ serviceData?.revision.creator }}
        </bk-form-item>
        <bk-form-item :label="t('创建时间')">
          {{ datetimeFormat(serviceData!.revision.create_at) }}
        </bk-form-item>
        <bk-form-item :label="t('更新者')">
          {{ serviceData!.revision.reviser }}
        </bk-form-item>
        <bk-form-item :label="t('更新时间')">
          {{ datetimeFormat(serviceData!.revision.update_at) }}
        </bk-form-item>
      </bk-form>
      <SearviceForm
        v-else
        ref="formCompRef"
        :form-data="serviceEditForm"
        :approver-api="getApproverListApi()"
        @change="handleChange"
        :editable="true" />
    </div>
    <div v-if="!isViewMode" class="service-edit-footer">
      <bk-button
        v-cursor="{ active: !props.service.permissions.update }"
        theme="primary"
        :class="{ 'bk-button-with-no-perm': !props.service.permissions.update }"
        :loading="pending"
        @click="handleEditConfirm">
        {{ t('保存') }}
      </bk-button>
      <bk-button @click="isViewMode = true">{{ t('取消') }}</bk-button>
    </div>
  </bk-sideslider>
</template>
<script setup lang="ts">
  import { ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { storeToRefs } from 'pinia';
  import useGlobalStore from '../../../../../store/global';
  import { updateApp, getApproverListApi } from '../../../../../api/index';
  import { getKvList } from '../../../../../api/config';
  import { datetimeFormat } from '../../../../../utils/index';
  import { IAppItem } from '../../../../../../types/app';
  import { IServiceEditForm } from '../../../../../../types/service';
  import useModalCloseConfirmation from '../../../../../utils/hooks/use-modal-close-confirmation';
  import SearviceForm from './service-form.vue';
  import { IConfigKvType } from '../../../../../../types/config';
  import { InfoBox, Message } from 'bkui-vue';
  import BkMemberSelector from '../../../../../components/user-selector/index.vue';

  const { showApplyPermDialog, permissionQuery } = storeToRefs(useGlobalStore());

  const { t } = useI18n();

  const props = defineProps<{
    show: boolean;
    service: IAppItem;
  }>();

  const emits = defineEmits(['update:show', 'reload']);

  const isFormChange = ref(false);
  const isViewMode = ref(true);
  const serviceEditForm = ref<IServiceEditForm>({
    name: '',
    alias: '',
    config_type: 'file',
    data_type: 'any',
    memo: '',
    is_approve: true,
    approver: '',
    approve_type: 'OrSign',
  });
  const serviceData = ref<IAppItem>();
  const pending = ref(false);
  const formCompRef = ref();
  const selectionsApprover = ref<string[]>([]);

  watch(
    () => props.show,
    (val) => {
      if (val) {
        isFormChange.value = false;
        isViewMode.value = true;
        const { spec } = props.service;
        const { name, memo, config_type, data_type, alias, is_approve, approver, approve_type } = spec;
        serviceEditForm.value = {
          name,
          memo,
          config_type,
          data_type,
          alias,
          is_approve,
          approver,
          approve_type,
        };
        if (approver) {
          selectionsApprover.value = approver.split(',');
        }
        serviceData.value = props.service;
      }
    },
  );

  const openPermApplyDialog = () => {
    permissionQuery.value = {
      resources: [
        {
          biz_id: props.service.biz_id,
          basic: {
            type: 'app',
            action: 'update',
            resource_id: props.service.id,
          },
        },
      ],
    };
    showApplyPermDialog.value = true;
  };

  const handleChange = (val: IServiceEditForm) => {
    isFormChange.value = true;
    serviceEditForm.value = val;
  };

  const handleEditConfirm = async () => {
    if (!props.service.permissions.update) {
      openPermApplyDialog();
      return;
    }
    await formCompRef.value.validate();
    const { id, biz_id } = props.service;
    const dataType = serviceEditForm.value.data_type;
    if (dataType !== 'any') {
      const configList = await getKvList(String(biz_id), id as number, { all: true, start: 0 });
      const res = configList.details.some((config: IConfigKvType) => config.spec.kv_type !== dataType);
      if (res) {
        InfoBox({
          infoType: 'danger',
          'ext-cls': 'info-box-style',
          title: t('调整服务数据类型{n}失败', { n: dataType }),
          subTitle: t('该服务下存在非{n}类型的配置项，如需修改，请先调整该服务下的所有配置项数据类型为{n}', {
            n: dataType,
          }),
          dialogType: 'confirm',
          confirmText: t('我知道了'),
        });
        return;
      }
    }
    if (serviceEditForm.value.approver) {
      selectionsApprover.value = serviceEditForm.value.approver.split(',');
    }
    const data = {
      id,
      biz_id,
      ...serviceEditForm.value,
    };

    const res = await updateApp({ id, biz_id, data });
    serviceData.value = res;
    Message({
      theme: 'success',
      message: t('服务属性编辑成功'),
    });
    emits('reload');
    isViewMode.value = true;
  };

  const handleBeforeClose = async () => {
    if (!isViewMode.value && isFormChange.value) {
      const result = await useModalCloseConfirmation();
      return result;
    }
    return true;
  };

  const close = () => {
    emits('update:show', false);
  };
</script>
<style lang="scss" scoped>
  .service-edit-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding-right: 24px;
    width: 100%;
    .edit-entry-btn {
      min-width: 64px;
    }
  }
  .service-edit-wrapper {
    padding: 20px 24px;
    height: calc(100vh - 101px);
    font-size: 12px;
    :deep(.bk-form-item) {
      .bk-form-label,
      .bk-form-content {
        line-height: 16px;
        font-size: 12px;
      }
      .bk-form-label {
        color: #979ba5;
      }
      .bk-form-content {
        color: #63656e;
      }
    }
    .content-edit {
      position: relative;
      // padding-right: 16px;
      .edit-icon {
        display: none;
        position: absolute;
        right: -20px;
        top: -3px;
        font-size: 22px;
        color: #979ba5;
        cursor: pointer;
        &:hover {
          color: #3a84ff;
        }
        &.no-edit-perm {
          color: #c4c6cc;
        }
      }
      &:hover .edit-icon {
        display: block;
      }
    }
  }
  .service-edit-footer {
    padding: 8px 24px;
    height: 48px;
    width: 100%;
    background: #fafbfd;
    border-top: 1px solid #dcdee5;
    box-shadow: none;
    button {
      margin-right: 8px;
      min-width: 88px;
    }
  }
  .offset-y {
    transform: translateY(10px);
  }
  .approval-info {
    padding: 8px 16px;
    background-color: #f5f7fa;
    .approval-li {
      font-size: 12px;
      line-height: 20px;
      & + .approval-li {
        margin-top: 16px;
      }
    }
    .approval-hd {
      word-wrap: break-word;
      word-break: break-all;
      color: #979ba5;
    }
    .approval-bd {
      margin-top: 4px;
      color: #313238;
    }
  }
</style>
