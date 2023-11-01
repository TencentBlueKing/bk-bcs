<template>
  <div class="title">
    <div class="title-content" @click="expand = !expand">
      <DownShape :class="['fold-icon', { fold: !expand }]" />
      <div class="title-text">
        {{ headText }} <span>({{ tableData.length }})</span>
      </div>
    </div>
  </div>
  <table class="table" v-if="expand">
    <thead>
      <tr>
        <th class="th-cell name">配置项名称</th>
        <th class="th-cell path">配置项路径</th>
        <th class="th-cell type">配置项格式</th>
        <th class="th-cell memo th-cell-edit">
          <span>配置项描述</span>
          <bk-popover
            ext-cls="popover-wrap"
            theme="light"
            trigger="click"
            placement="bottom"
            :is-show="batchSet.isShowMemoPop"
          >
            <div @click="batchSet.isShowMemoPop = true"><edit-line class="edlit-line" /></div>
            <template #content>
              <div class="pop-wrap">
                <div class="pop-content">
                  <div class="pop-title">批量设置配置项描述</div>
                  <bk-input v-model="batchSet.memo"></bk-input>
                </div>
                <div class="pop-footer">
                  <div class="button">
                    <bk-button
                      theme="primary"
                      style="margin-right: 8px; width: 80px"
                      size="small"
                      @click="handleConfirmPop('memo')">确定</bk-button>
                    <bk-button size="small" @click="batchSet.isShowMemoPop = false">取消</bk-button>
                  </div>
                </div>
              </div>
            </template>
          </bk-popover>
        </th>
        <th class="th-cell permishion th-cell-edit">
          <span>文件权限</span>
          <edit-line class="edlit-line" />
        </th>
        <th class="th-cell user th-cell-edit">
          <span>用户</span>
          <bk-popover
            ext-cls="popover-wrap"
            theme="light"
            trigger="click"
            placement="bottom"
            :is-show="batchSet.isShowUserPop"
          >
            <div @click="batchSet.isShowUserPop = true"><edit-line class="edlit-line" /></div>
            <template #content>
              <div class="pop-wrap">
                <div class="pop-content">
                  <div class="pop-title">批量设置用户</div>
                  <bk-input v-model="batchSet.user"></bk-input>
                </div>
                <div class="pop-footer">
                  <div class="button">
                    <bk-button
                      theme="primary"
                      style="margin-right: 8px; width: 80px"
                      size="small"
                      @click="handleConfirmPop('user')">确定</bk-button>
                    <bk-button size="small" @click="handleCancelPop">取消</bk-button>
                  </div>
                </div>
              </div>
            </template>
          </bk-popover>
        </th>
        <th class="th-cell user-group th-cell-edit">
          <span>用户组</span>
          <bk-popover
            ext-cls="popover-wrap"
            theme="light"
            trigger="click"
            placement="bottom"
            :is-show="batchSet.isShowUserGroupPop"
          >
            <div @click="batchSet.isShowUserGroupPop = true"><edit-line class="edlit-line" /></div>
            <template #content>
              <div class="pop-wrap">
                <div class="pop-content">
                  <div class="pop-title">批量设置用户组</div>
                  <bk-input v-model="batchSet.user_group"></bk-input>
                </div>
                <div class="pop-footer">
                  <div class="button">
                    <bk-button
                      theme="primary"
                      style="margin-right: 8px; width: 80px"
                      size="small"
                      @click="handleConfirmPop('user_group')">确定</bk-button>
                    <bk-button size="small" @click="handleCancelPop">取消</bk-button>
                  </div>
                </div>
              </div>
            </template>
          </bk-popover>
        </th>
      </tr>
    </thead>
    <tbody>
      <tr v-for="(item, index) in data" :key="index">
        <td class="not-editable td-cell">{{ item.name }}</td>
        <td class="not-editable td-cell">{{ item.path }}</td>
        <td class="not-editable td-cell">{{ item.file_type }}</td>
        <td class="td-cell-editable"><bk-input v-model="item.memo"></bk-input></td>
        <td class="td-cell-editable"><bk-input v-model="item.privilege"></bk-input></td>
        <td class="td-cell-editable"><bk-input v-model="item.user"></bk-input></td>
        <td class="td-cell-editable"><bk-input v-model="item.user_group"></bk-input></td>
      </tr>
    </tbody>
  </table>
</template>

<script lang="ts" setup>
import { ref, computed } from 'vue';
import { DownShape, EditLine } from 'bkui-vue/lib/icon';
import { IConfigImport } from '../../../../../../../../../types/config';
const expand = ref(true);
const batchSet = ref({
  memo: '',
  privilege: '',
  user: '',
  user_group: '',
  isShowMemoPop: false,
  isShowUserPop: false,
  isShowUserGroupPop: false,
});
const props = withDefaults(
  defineProps<{
    tableData: IConfigImport[];
    headText: string;
  }>(),
  {},
);

// const emits = defineEmits(['update:tableData']);
const data = computed(() => props.tableData.map((item) => {
  item.isEdit = false;
  return item;
}));

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
    privilege: '',
    user: '',
    user_group: '',
    isShowMemoPop: false,
    isShowUserPop: false,
    isShowUserGroupPop: false,
  };
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
  .th-cell {
    position: relative;
    padding-left: 16px;
    height: 42px;
    font-weight: normal;
    color: #313238;
    text-align: left;
    background: #fafbfd;
    border: 1px solid #dcdee5;
    .edlit-line {
      color: #3a84ff;
      position: absolute;
      right: 16px;
      top: 50%;
      transform: translateY(-50%);
      cursor: pointer;
    }
  }
  .name {
    width: 136px;
  }
  .path {
    width: 163px;
  }
  .type {
    width: 87px;
  }
  .memo {
    width: 125px;
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
      border: none;
      .bk-input--text {
        padding-left: 16px;
      }
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
</style>

<style lang="scss">
.popover-wrap {
  padding: 0 !important;
}
</style>
