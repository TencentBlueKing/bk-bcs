<template>
  <bk-popover
    v-if="isFileType"
    ref="buttonRef"
    theme="light batch-operation-button-popover"
    placement="bottom-end"
    trigger="click"
    width="108"
    :arrow="false"
    @after-show="isPopoverOpen = true"
    @after-hidden="isPopoverOpen = false">
    <bk-button :disabled="props.selectedIds.length === 0" :class="['batch-set-btn', { 'popover-open': isPopoverOpen }]">
      {{ t('批量操作') }}
      <AngleDown class="angle-icon" />
    </bk-button>
    <template #content>
      <div class="operation-item" @click="handleOpenBantchEditPerm">
        {{ t('批量修改权限') }}
      </div>
      <div class="operation-item" @click="handleOpenBantchDelet">
        {{ t('批量删除') }}
      </div>
    </template>
  </bk-popover>
  <bk-button
    v-else
    class="batch-delete-btn"
    :disabled="props.selectedIds.length === 0"
    @click="isBatchDeleteDialogShow = true">
    {{ t('批量删除') }}
  </bk-button>
  <DeleteConfirmDialog
    v-model:isShow="isBatchDeleteDialogShow"
    :title="t('确认删除所选的 {n} 项配置项？', { n: props.selectedIds.length })"
    :pending="batchDeletePending"
    @confirm="handleBatchDeleteConfirm">
    <div>
      {{
        t('已生成版本中存在的配置项，可以通过恢复按钮撤销删除，新增且未生成版本的配置项，将无法撤销删除，请谨慎操作。')
      }}
    </div>
  </DeleteConfirmDialog>
  <bk-dialog
    :is-show="isBatchEditPermDialogShow"
    :title="$t('批量修改权限')"
    :theme="'primary'"
    quick-close
    ext-cls="batch-edit-perm-dialog"
    :width="640"
    @confirm="handleConfirm"
    @closed="isBatchEditPermDialogShow = false">
    <div class="selected-tag">
      {{ `${t('已选')} ` }} <span class="count">{{ props.selectedIds.length }}</span> {{ `${t('个配置项')}` }}
    </div>
    <bk-form form-type="vertical" class="user-settings">
      <bk-form-item :label="t('文件权限')">
        <div class="perm-input">
          <bk-popover
            ext-cls="privilege-tips-wrap"
            theme="light"
            trigger="manual"
            placement="top"
            :is-show="showPrivilegeErrorTips">
            <bk-input
              v-model="privilegeInputVal"
              type="number"
              :placeholder="t('保持不变')"
              @blur="handlePrivilegeInputBlur" />
            <template #content>
              <div>{{ t('只能输入三位 0~7 数字') }}</div>
              <div class="privilege-tips-btn-area">
                <bk-button text theme="primary" @click="showPrivilegeErrorTips = false">{{ t('我知道了') }}</bk-button>
              </div>
            </template>
          </bk-popover>
          <bk-popover ext-cls="privilege-select-popover" theme="light" trigger="click" placement="bottom">
            <div class="perm-panel-trigger">
              <i class="bk-bscp-icon icon-configuration-line"></i>
            </div>
            <template #content>
              <div class="privilege-select-panel">
                <div v-for="(item, index) in PRIVILEGE_GROUPS" class="group-item" :key="index" :label="item">
                  <div class="header">{{ item }}</div>
                  <div class="checkbox-area">
                    <bk-checkbox-group
                      class="group-checkboxs"
                      :model-value="privilegeGroupsValue[index]"
                      @change="handleSelectPrivilege(index, $event)">
                      <bk-checkbox size="small" :label="4" :disabled="true">
                        {{ t('读') }}
                      </bk-checkbox>
                      <bk-checkbox size="small" :label="2">{{ t('写') }}</bk-checkbox>
                      <bk-checkbox size="small" :label="1">{{ t('执行') }}</bk-checkbox>
                    </bk-checkbox-group>
                  </div>
                </div>
              </div>
            </template>
          </bk-popover>
        </div>
      </bk-form-item>
      <bk-form-item :label="t('用户')">
        <bk-input v-model="localVal.user" :placeholder="t('保持不变')"></bk-input>
      </bk-form-item>
      <bk-form-item :label="t('用户组')">
        <bk-input v-model="localVal.user_group" :placeholder="t('保持不变')"></bk-input>
      </bk-form-item>
    </bk-form>
  </bk-dialog>
</template>
<script lang="ts" setup>
  import { ref, computed, watch } from 'vue';
  import { AngleDown } from 'bkui-vue/lib/icon';
  import { useI18n } from 'vue-i18n';
  import Message from 'bkui-vue/lib/message';
  import { batchDeleteServiceConfigs, batchDeleteKv, batchAddConfigList } from '../../../../../../../api/config';
  import DeleteConfirmDialog from '../../../../../../../components/delete-confirm-dialog.vue';
  import { IConfigItem } from '../../../../../../../../types/config';
  const { t } = useI18n();

  const PRIVILEGE_GROUPS = [t('属主（own）'), t('属组（group）'), t('其他人（other）')];
  const PRIVILEGE_VALUE_MAP = {
    0: [],
    1: [1],
    2: [2],
    3: [1, 2],
    4: [4],
    5: [1, 4],
    6: [2, 4],
    7: [1, 2, 4],
  };

  const props = defineProps<{
    bkBizId: string;
    appId: number;
    selectedIds: number[];
    isFileType: boolean; // 是否为文件型配置
    selectedItems: IConfigItem[];
  }>();

  const emits = defineEmits(['deleted']);

  const batchDeletePending = ref(false);
  const isBatchDeleteDialogShow = ref(false);
  const isBatchEditPermDialogShow = ref(false);
  const isPopoverOpen = ref(false);
  const buttonRef = ref();
  const privilegeInputVal = ref('');
  const showPrivilegeErrorTips = ref(false);
  const localVal = ref({
    privilege: '',
    user: '',
    user_group: '',
  });

  // 将权限数字拆分成三个分组配置
  const privilegeGroupsValue = computed(() => {
    const data: { [index: string]: number[] } = { 0: [], 1: [], 2: [] };
    if (typeof localVal.value.privilege === 'string' && localVal.value.privilege.length > 0) {
      const valArr = localVal.value.privilege.split('').map((i) => parseInt(i, 10));
      valArr.forEach((item, index) => {
        data[index as keyof typeof data] = PRIVILEGE_VALUE_MAP[item as keyof typeof PRIVILEGE_VALUE_MAP];
      });
    }
    return data;
  });

  watch(
    () => isBatchEditPermDialogShow.value,
    (val) => {
      if (val) {
        localVal.value = {
          privilege: '',
          user: '',
          user_group: '',
        };
        privilegeInputVal.value = '';
      }
    },
  );

  const handleBatchDeleteConfirm = async () => {
    batchDeletePending.value = true;
    if (props.isFileType) {
      await batchDeleteServiceConfigs(props.bkBizId, props.appId, props.selectedIds);
    } else {
      await batchDeleteKv(props.bkBizId, props.appId, props.selectedIds);
    }
    Message({
      theme: 'success',
      message: props.isFileType ? t('批量删除配置文件成功') : t('批量删除配置项成功'),
    });
    batchDeletePending.value = false;
    isBatchDeleteDialogShow.value = false;
    emits('deleted');
  };

  const handleOpenBantchEditPerm = () => {
    buttonRef.value.hide();
    isBatchEditPermDialogShow.value = true;
  };

  const handleOpenBantchDelet = () => {
    buttonRef.value.hide();
    isBatchDeleteDialogShow.value = true;
  };

  // 权限输入框失焦后，校验输入是否合法，如不合法回退到上次输入
  const handlePrivilegeInputBlur = () => {
    const val = String(privilegeInputVal.value);
    if (/^[0-7]{3}$/.test(val)) {
      localVal.value.privilege = val;
      showPrivilegeErrorTips.value = false;
    } else {
      privilegeInputVal.value = String(localVal.value.privilege);
      showPrivilegeErrorTips.value = true;
    }
  };

  // 选择文件权限
  const handleSelectPrivilege = (index: number, val: number[]) => {
    const groupsValue = { ...privilegeGroupsValue.value };
    groupsValue[index] = val;
    const digits = [];
    for (let i = 0; i < 3; i++) {
      let sum = 0;
      if (groupsValue[i].length > 0) {
        sum = groupsValue[i].reduce((acc, crt) => acc + crt, 0);
      }
      digits.push(sum);
    }
    const newVal = digits.join('');
    privilegeInputVal.value = newVal;
    localVal.value.privilege = newVal;
  };

  const handleConfirm = async () => {
    const editConfigList = props.selectedItems.map((item) => {
      const { id, spec, commit_spec } = item;
      return {
        id,
        ...spec,
        privilege: localVal.value.privilege || spec.permission.privilege,
        user: localVal.value.user || spec.permission.user,
        user_group: localVal.value.user_group || spec.permission.user_group,
        byte_size: commit_spec.content.byte_size,
        sign: commit_spec.content.signature,
      };
    });

    try {
      await batchAddConfigList(props.bkBizId, props.appId, editConfigList, false);
      Message({
        theme: 'success',
        message: t('配置文件权限批量修改成功'),
      });
      isBatchEditPermDialogShow.value = false;
    } catch (error) {
      console.error(error);
    }
    emits('deleted');
  };
</script>

<style lang="scss" scoped>
  .batch-set-btn {
    min-width: 108px;
    height: 32px;
    margin-left: 8px;
    &.popover-open {
      .angle-icon {
        transform: rotate(-180deg);
      }
    }
    .angle-icon {
      font-size: 20px;
      transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    }
  }
  .user-settings {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 16px;
  }
  .perm-input {
    display: flex;
    align-items: center;
    width: 172px;
    :deep(.bk-input) {
      width: 140px;
      border-right: none;
      border-top-right-radius: 0;
      border-bottom-right-radius: 0;
      .bk-input--number-control {
        display: none;
      }
    }
    .perm-panel-trigger {
      width: 32px;
      height: 32px;
      text-align: center;
      background: #fafcfe;
      color: #3a84ff;
      border: 1px solid #3a84ff;
      cursor: pointer;
      &.disabled {
        color: #dcdee5;
        border-color: #dcdee5;
        cursor: not-allowed;
      }
    }
  }
  .privilege-select-panel {
    display: flex;
    align-items: top;
    border: 1px solid #dcdee5;
    .group-item {
      .header {
        padding: 0 16px;
        height: 42px;
        line-height: 42px;
        color: #313238;
        font-size: 12px;
        background: #fafbfd;
        border-bottom: 1px solid #dcdee5;
      }
      &:not(:last-of-type) {
        .header,
        .checkbox-area {
          border-right: 1px solid #dcdee5;
        }
      }
    }
    .checkbox-area {
      padding: 10px 16px 12px;
      background: #ffffff;
      &:not(:last-child) {
        border-right: 1px solid #dcdee5;
      }
    }
    .group-checkboxs {
      font-size: 12px;
      .bk-checkbox ~ .bk-checkbox {
        margin-left: 16px;
      }
      :deep(.bk-checkbox-label) {
        font-size: 12px;
      }
    }
  }
  .selected-tag {
    display: inline-block;
    height: 32px;
    background: #f0f1f5;
    line-height: 32px;
    border-radius: 16px;
    padding: 0 12px;
    margin: 8px 0px 16px;
    .count {
      color: #3a84ff;
    }
  }
</style>

<style lang="scss">
  .batch-operation-button-popover.bk-popover.bk-pop2-content {
    padding: 4px 0;
    border: 1px solid #dcdee5;
    box-shadow: 0 2px 6px 0 #0000001a;
    .operation-item {
      padding: 0 12px;
      min-width: 58px;
      height: 32px;
      line-height: 32px;
      color: #63656e;
      font-size: 12px;
      cursor: pointer;
      &:hover {
        background: #f5f7fa;
      }
    }
  }
</style>
