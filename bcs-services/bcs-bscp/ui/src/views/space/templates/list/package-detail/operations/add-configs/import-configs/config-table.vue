<template>
  <div style="margin-bottom: 16px">
    <div class="title">
      <div class="title-content" @click="emits('changeExpand')">
        <DownShape :class="['fold-icon', { fold: !expand }]" />
        <div class="title-text">
          {{ isExsitTable ? t('已存在配置文件') : t('新建配置文件') }} <span>({{ tableData.length }})</span>
        </div>
      </div>
    </div>
    <div class="table-container">
      <table class="table" v-show="expand">
        <thead>
          <tr>
            <th class="th-cell name">{{ t('配置文件绝对路径') }}</th>
            <th class="th-cell type">{{ t('配置文件格式') }}</th>
            <th class="th-cell memo">
              <div class="th-cell-edit">
                <span>{{ t('配置文件描述') }}</span>
                <bk-popover
                  ext-cls="popover-wrap"
                  theme="light"
                  trigger="manual"
                  placement="bottom"
                  :is-show="batchSet.isShowMemoPop">
                  <edit-line class="edit-line" @click="batchSet.isShowMemoPop = true" />
                  <template #content>
                    <div class="pop-wrap" v-click-outside="() => (batchSet.isShowMemoPop = false)">
                      <div class="pop-content">
                        <div class="pop-title">{{ t('批量设置配置文件描述') }}</div>
                        <bk-input v-model="batchSet.memo" :placeholder="t('请输入')"></bk-input>
                      </div>
                      <div class="pop-footer">
                        <div class="button">
                          <bk-button
                            theme="primary"
                            style="margin-right: 8px; width: 80px"
                            size="small"
                            @click="handleConfirmPop('memo')">
                            {{ t('确定') }}
                          </bk-button>
                          <bk-button size="small" @click="batchSet.isShowMemoPop = false">{{ t('取消') }}</bk-button>
                        </div>
                      </div>
                    </div>
                  </template>
                </bk-popover>
              </div>
            </th>
            <th class="th-cell privilege">
              <div class="th-cell-edit">
                <span class="required">{{ t('文件权限') }}</span>
                <bk-popover
                  ext-cls="popover-wrap"
                  theme="light"
                  trigger="manual"
                  placement="bottom"
                  :is-show="batchSet.isShowPrivilege">
                  <edit-line class="edit-line" @click="batchSet.isShowPrivilege = true" />
                  <template #content>
                    <div class="pop-wrap privilege-wrap" v-click-outside="() => (batchSet.isShowPrivilege = false)">
                      <div class="pop-content">
                        <div class="pop-title">{{ t('批量设置文件权限') }}</div>
                        <bk-input
                          v-model="batchSet.privilege"
                          :placeholder="t('请输入')"
                          style="width: 184px; margin-bottom: 16px"
                          @blur="testPrivilegeInput(batchSet.privilege)"></bk-input>
                        <span class="error-tip" style="margin-left: 10px" v-if="batchSet.isShowPrivilegeError">
                          {{ t('只能输入三位 0~7 数字且文件own必须有读取权限') }}
                        </span>
                        <div class="privilege-select-panel">
                          <div v-for="(item, index) in PRIVILEGE_GROUPS" class="group-item" :key="index" :label="item">
                            <div class="header">{{ item }}</div>
                            <div class="checkbox-area">
                              <bk-checkbox-group
                                class="group-checkboxs"
                                :model-value="privilegeGroupsValue(batchSet.privilege)[index]"
                                @change="handleSelectPrivilege(index, $event)">
                                <bk-checkbox size="small" :label="4" :disabled="index === 0">{{ t('读') }}</bk-checkbox>
                                <bk-checkbox size="small" :label="2">{{ t('写') }}</bk-checkbox>
                                <bk-checkbox size="small" :label="1">{{ t('执行') }}</bk-checkbox>
                              </bk-checkbox-group>
                            </div>
                          </div>
                        </div>
                      </div>
                      <div class="pop-footer">
                        <div class="button">
                          <bk-button
                            theme="primary"
                            style="margin-right: 8px; width: 80px"
                            size="small"
                            @click="handleConfirmPop('privilege')">
                            {{ t('确定') }}
                          </bk-button>
                          <bk-button size="small" @click="handleCancelPop">{{ t('取消') }}</bk-button>
                        </div>
                      </div>
                    </div>
                  </template>
                </bk-popover>
              </div>
            </th>
            <th class="th-cell user">
              <div class="th-cell-edit">
                <span class="required">{{ t('用户') }}</span>
                <bk-popover
                  ext-cls="popover-wrap"
                  theme="light"
                  trigger="manual"
                  placement="bottom"
                  :is-show="batchSet.isShowUserPop">
                  <edit-line class="edit-line" @click="batchSet.isShowUserPop = true" />
                  <template #content>
                    <div class="pop-wrap" v-click-outside="() => (batchSet.isShowUserPop = false)">
                      <div class="pop-content">
                        <div class="pop-title">{{ t('批量设置用户') }}</div>
                        <bk-input v-model="batchSet.user" :placeholder="t('请输入')"></bk-input>
                      </div>
                      <div class="pop-footer">
                        <div class="button">
                          <bk-button
                            theme="primary"
                            style="margin-right: 8px; width: 80px"
                            size="small"
                            @click="handleConfirmPop('user')">
                            {{ t('确定') }}
                          </bk-button>
                          <bk-button size="small" @click="handleCancelPop">{{ t('取消') }}</bk-button>
                        </div>
                      </div>
                    </div>
                  </template>
                </bk-popover>
              </div>
            </th>
            <th class="th-cell user-group">
              <div class="th-cell-edit">
                <span class="required">{{ t('用户组') }}</span>
                <bk-popover
                  ext-cls="popover-wrap"
                  theme="light"
                  trigger="manual"
                  placement="bottom"
                  :is-show="batchSet.isShowUserGroupPop">
                  <edit-line class="edit-line" @click="batchSet.isShowUserGroupPop = true" />
                  <template #content>
                    <div class="pop-wrap" v-click-outside="() => (batchSet.isShowUserGroupPop = false)">
                      <div class="pop-content">
                        <div class="pop-title">{{ t('批量设置用户组') }}</div>
                        <bk-input v-model="batchSet.user_group" :placeholder="t('请输入')"></bk-input>
                      </div>
                      <div class="pop-footer">
                        <div class="button">
                          <bk-button
                            theme="primary"
                            style="margin-right: 8px; width: 80px"
                            size="small"
                            @click="handleConfirmPop('user_group')">
                            {{ t('确定') }}
                          </bk-button>
                          <bk-button size="small" @click="handleCancelPop">{{ t('取消') }}</bk-button>
                        </div>
                      </div>
                    </div>
                  </template>
                </bk-popover>
              </div>
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(item, index) in data" :key="item.name + item.path">
            <td class="not-editable td-cell">
              <bk-overflow-title type="tips">
                {{ item.path + item.name }}
              </bk-overflow-title>
            </td>
            <td class="not-editable td-cell">
              <bk-overflow-title type="tips">
                {{ item.file_type === 'text' ? t('文本') : t('二进制') }}
              </bk-overflow-title>
            </td>
            <td class="td-cell-editable" :class="{ change: isContentChange(item.id, 'memo') }">
              <bk-input v-model="item.memo" :placeholder="t('请输入')"></bk-input>
            </td>
            <td class="td-cell-editable" :class="{ change: isContentChange(item.id, 'privilege') }">
              <div class="perm-input">
                <bk-input
                  v-model="item.privilege"
                  :placeholder="t('请输入')"
                  @blur="handlePrivilegeInputBlur(item)" />
                <bk-popover ext-cls="privilege-select-popover" theme="light" trigger="click" placement="bottom">
                  <div class="perm-panel-trigger">
                    <i class="bk-bscp-icon icon-configuration-line"></i>
                  </div>
                  <template #content>
                    <div class="privilege-select-panel">
                      <div v-for="(group, i) in PRIVILEGE_GROUPS" class="group-item" :key="i" :label="item">
                        <div class="header">{{ group }}</div>
                        <div class="checkbox-area">
                          <bk-checkbox-group
                            class="group-checkboxs"
                            :model-value="privilegeGroupsValue(item.privilege)[i]"
                            @change="handleSelectPrivilege(i, $event, item)">
                            <bk-checkbox size="small" :label="4" :disabled="i === 0">{{ t('读') }}</bk-checkbox>
                            <bk-checkbox size="small" :label="2">{{ t('写') }}</bk-checkbox>
                            <bk-checkbox size="small" :label="1">{{ t('执行') }}</bk-checkbox>
                          </bk-checkbox-group>
                        </div>
                      </div>
                    </div>
                  </template>
                </bk-popover>
              </div>
            </td>
            <td class="td-cell-editable" :class="{ change: isContentChange(item.id, 'user') }">
              <bk-input v-model="item.user" :placeholder="t('请输入')"></bk-input>
            </td>
            <td class="td-cell-editable" :class="{ change: isContentChange(item.id, 'user_group') }">
              <bk-input v-model="item.user_group" :placeholder="t('请输入')"></bk-input>
            </td>
            <td class="td-cell-delete">
              <i class="bk-bscp-icon icon-reduce delete-icon" @click="handleDeleteConfig(index)"></i>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { ref, computed, watch, onMounted } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { DownShape, EditLine } from 'bkui-vue/lib/icon';
  import { IConfigImportItem } from '../../../../../../../../../types/config';
  import { cloneDeep, isEqual } from 'lodash';
  import Message from 'bkui-vue/lib/message';

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
  const batchSet = ref({
    memo: '',
    privilege: '644',
    user: '',
    user_group: '',
    isShowMemoPop: false,
    isShowUserPop: false,
    isShowUserGroupPop: false,
    isShowPrivilege: false,
    isShowPrivilegeError: false,
  });

  const data = ref<IConfigImportItem[]>([]);
  const initData = ref<IConfigImportItem[]>([]);

  const props = withDefaults(
    defineProps<{
      tableData: IConfigImportItem[];
      isExsitTable: boolean;
      expand: boolean;
    }>(),
    {},
  );

  const emits = defineEmits(['change', 'changeExpand']);

  onMounted(() => {
    data.value = cloneDeep(props.tableData);
    initData.value = cloneDeep(props.tableData);
  });

  watch(
    () => props.tableData,
    () => {
      data.value = cloneDeep(props.tableData);
      initData.value = cloneDeep(props.tableData);
    },
  );

  watch(
    () => data.value,
    () => {
      if (isEqual(data.value, initData.value)) {
        return;
      }
      emits('change', data.value);
    },
    { deep: true },
  );

  // 将权限数字拆分成三个分组配置
  const privilegeGroupsValue = computed(() => (privilege: string) => {
    const data: { [index: string]: number[] } = { 0: [], 1: [], 2: [] };
    if (privilege.length > 0) {
      const valArr = privilege.split('').map((i) => parseInt(i, 10));
      valArr.forEach((item, index) => {
        data[index as keyof typeof data] = PRIVILEGE_VALUE_MAP[item as keyof typeof PRIVILEGE_VALUE_MAP];
      });
    }
    return data;
  });

  const testPrivilegeInput = (privilege: string) => {
    const val = String(privilege);
    const own = parseInt(privilege[0], 10);
    if (/^[0-7]{3}$/.test(val) && own >= 4) {
      batchSet.value.isShowPrivilegeError = false;
    } else {
      batchSet.value.privilege = '644';
      batchSet.value.isShowPrivilegeError = true;
    }
  };

  // 选择文件权限
  const handleSelectPrivilege = (index: number, val: number[], item?: IConfigImportItem) => {
    let groupsValue;
    if (item) {
      groupsValue = { ...privilegeGroupsValue.value(item.privilege) };
    } else {
      groupsValue = { ...privilegeGroupsValue.value(batchSet.value.privilege) };
    }

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
    if (item) {
      item.privilege = newVal;
    } else {
      batchSet.value.privilege = newVal;
      batchSet.value.isShowPrivilegeError = false;
    }
  };

  const handleConfirmPop = (prop: string) => {
    if (prop === 'memo') {
      data.value.forEach((item) => {
        item.memo = batchSet.value.memo;
      });
    }
    if (prop === 'user') {
      data.value.forEach((item) => {
        item.user = batchSet.value.user;
      });
    }
    if (prop === 'privilege') {
      if (batchSet.value.isShowPrivilegeError) return;
      data.value.forEach((item) => {
        item.privilege = batchSet.value.privilege;
      });
    }
    if (prop === 'user_group') {
      data.value.forEach((item) => {
        item.user_group = batchSet.value.user_group;
      });
    }
    handleCancelPop();
  };

  const handleCancelPop = () => {
    batchSet.value = {
      memo: '',
      privilege: '644',
      user: '',
      user_group: '',
      isShowMemoPop: false,
      isShowUserPop: false,
      isShowUserGroupPop: false,
      isShowPrivilege: false,
      isShowPrivilegeError: false,
    };
  };

  const handleDeleteConfig = (index: number) => {
    console.log(index);
    data.value = data.value.filter((item, i) => i !== index);
  };

  // 权限输入框失焦后，校验输入是否合法，如不合法回退到上次输入
  const handlePrivilegeInputBlur = (item: IConfigImportItem) => {
    const val = item.privilege;
    const own = parseInt(val[0], 10);
    if (!/^[0-7]{3}$/.test(val) || own < 4) {
      item.privilege = '644';
      Message({
        message: t('只能输入三位 0~7 数字且文件own必须有读取权限'),
        theme: 'error',
      });
    }
  };

  // 判断内容是否改变
  const isContentChange = (id: number, key: string) => {
    if (!props.isExsitTable) return;
    const newConfig = data.value.find((config) => config.id === id);
    const oldConfig = initData.value.find((config) => config.id === id);
    return newConfig![key as keyof IConfigImportItem] !== oldConfig![key as keyof IConfigImportItem];
  };
</script>

<style scoped lang="scss">
  .title {
    height: 28px;
    background: #eaebf0;
    border-radius: 2px 2px 0 0;

    .title-content {
      display: flex;
      align-items: center;
      height: 100%;
      margin-left: 10px;
      cursor: pointer;
      .fold-icon {
        margin-right: 8px;
        font-size: 14px;
        color: #979ba5;
        transition: transform 0.2s ease-in-out;
        &.fold {
          transform: rotate(-90deg);
        }
      }
      .title-text {
        font-weight: 700;
        font-size: 12px;
        color: #63656e;
        span {
          font-size: 12px;
          color: #979ba5;
        }
      }
    }
  }
  .table {
    width: 100%;
    border-collapse: collapse;
    border: 1px solid #dcdee5;
    font-size: 12px;
    line-height: 20px;
    table-layout: fixed;
    overflow: auto;
    .th-cell {
      padding-left: 16px;
      height: 42px;
      color: #313238;
      background: #fafbfd;
      border: 1px solid #dcdee5;
      text-align: left;
      .th-cell-edit {
        display: flex;
        justify-content: space-between;
        padding-right: 16px;
        .edit-line {
          color: #3a84ff;
          cursor: pointer;
          text-align: right;
        }
        .required {
          position: relative;
          &::before {
            position: absolute;
            top: 0;
            left: -14px;
            width: 14px;
            color: #ea3636;
            text-align: center;
            content: '*';
          }
        }
      }
    }
    .name {
      width: 300px;
    }
    .type {
      width: 100px;
    }
    .memo {
      width: 199px;
    }
    .privilege {
      width: 102px;
    }
    .user {
      width: 78px;
    }
    .user-group {
      width: 91px;
    }
    .not-editable {
      background-color: #f5f7fa;
    }
    .td-cell {
      border: 1px solid #dcdee5;
      padding-left: 16px;
    }
    .td-cell-editable {
      padding: 0;
      border: 1px solid #dcdee5;
      :deep(.bk-input) {
        height: 42px;
        .bk-input--text {
          padding-left: 16px;
        }
        &:not(.is-focused) {
          border: none;
        }
      }
    }
    .change {
      :deep(.bk-input--text) {
        background-color: #fff3e1;
      }
    }
    .td-cell-delete {
      border: 1px solid #dcdee5;
      width: 50px;
      text-align: center;
      .delete-icon {
        cursor: pointer;
        font-size: 14px;
        color: gray;
      }
    }
  }
  .pop-wrap {
    width: 300px;
    .pop-content {
      padding: 16px;
      .pop-title {
        line-height: 24px;
        font-size: 16px;
        padding-bottom: 10px;
      }
    }

    .pop-footer {
      position: relative;
      height: 42px;
      background: #fafbfd;
      border-top: 1px solid #dcdee5;
      .button {
        position: absolute;
        right: 16px;
        top: 50%;
        transform: translateY(-50%);
      }
    }
  }
  .perm-input {
    position: relative;
    display: flex;
    align-items: center;
    :deep(.bk-input) {
      .bk-input--number-control {
        display: none;
      }
    }
    .perm-panel-trigger {
      position: absolute;
      right: 0;
      width: 32px;
      height: 40px;
      line-height: 42px;
      text-align: center;
      color: #3a84ff;
      cursor: pointer;
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
  .error-tip {
    color: red;
  }
  .privilege-wrap {
    width: auto;
  }
</style>

<style lang="scss">
  .popover-wrap {
    padding: 0 !important;
  }
</style>
