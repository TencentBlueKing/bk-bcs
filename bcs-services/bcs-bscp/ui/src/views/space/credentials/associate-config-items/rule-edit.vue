<template>
  <div class="rule-edit">
    <div class="head">
      <p class="title">配置关联规则</p>
      <bk-popover
        placement="bottom"
        theme="light"
        trigger="click"
        ext-cls="view-rule-wrap"
      >
        <span class="view-rule">查看规则示例</span>
        <template #content>
            <ViewRuleExample />
        </template>
      </bk-popover>
    </div>
    <div class="rules-edit-area">
      <div v-for="(rule, index) in localRules" class="rule-list" :key="index">
        <div :class="['rule-item', { 'is-error': !rule.isRight }]">
          <bk-input
            v-model="rule.content"
            class="rule-input"
            placeholder="请填写"
            :disabled="rule.type === 'del'"
            @input="handleInput(index)"
            @blur="handleRuleContentChange(index)"
          >
            <template #suffix>
              <div
                v-if="rule.type"
                v-bk-tooltips="{
                  disabled: rule.type !== 'modify',
                  content: `${rule.original} -> ${rule.content}`,
                }"
                :class="`status-tag ${rule.type}`"
              >
                {{ RULE_TYPE_MAP[rule.type] }}
              </div>
            </template>
          </bk-input>
          <div class="action-btns">
            <i
              v-if="rule.type === 'del'"
              v-bk-tooltips="'撤销本次删除'"
              class="bk-bscp-icon icon-revoke revoke-icon"
              @click="handleRevoke(index)"
            >
            </i>
            <template v-else>
              <i v-if="localRules.length > 1" class="bk-bscp-icon icon-reduce" @click="handleDeleteRule(index)"></i>
              <i style="margin-left: 10px" class="bk-bscp-icon icon-add" @click="handleAddRule(index)"></i>
            </template>
          </div>
        </div>
        <div class="error-info" v-if="!rule.isRight"><span>输入的规则有误，请重新确认</span></div>
      </div>
    </div>
    <!-- <div class="preview-btn">预览匹配结果</div> -->
  </div>
</template>
<script setup lang="ts">
import { ref, watch } from 'vue';
import { ICredentialRule, IRuleEditing, IRuleUpdateParams } from '../../../../../types/credential';
import ViewRuleExample from './view-rule-example.vue';

const props = defineProps<{
  rules: ICredentialRule[];
}>();

const emits = defineEmits(['change']);

const RULE_TYPE_MAP: { [key: string]: string } = {
  new: '新增',
  del: '删除',
  modify: '修改',
};

const localRules = ref<IRuleEditing[]>([]);

const transformRulesToEditing = (rules: ICredentialRule[]) => {
  const rulesEditing: IRuleEditing[] = [];
  rules.forEach((item) => {
    const { id, spec } = item;
    rulesEditing.push({ id, type: '', content: spec.credential_scope, original: spec.credential_scope, isRight: true });
  });
  return rulesEditing;
};

watch(
  () => props.rules,
  (val) => {
    if (val.length === 0) {
      localRules.value = [{ id: 0, type: 'new', content: '', original: '', isRight: true }];
    } else {
      localRules.value = transformRulesToEditing(val);
    }
  },
  { immediate: true },
);

const handleAddRule = (index: number) => {
  localRules.value.splice(index + 1, 0, { id: 0, type: 'new', content: '', original: '', isRight: true });
};

const handleDeleteRule = (index: number) => {
  const rule = localRules.value[index];
  if (rule.id) {
    rule.type = 'del';
  } else {
    localRules.value.splice(index, 1);
  }
  updateRuleParams();
};

const handleRevoke = (index: number) => {
  const rule = localRules.value[index];
  const { content, original } = rule;
  rule.type = content === original ? '' : 'modify';
  updateRuleParams();
};

const validateRule = (rule: string) => {
  if (rule.length < 2) {
    return false;
  }
  const paths = rule.split('/');
  return paths.length > 1 && paths.every(path => path.length > 0);
};

// 产品逻辑：没有检测到输入错误时：鼠标失焦后检测；如果检测到错误时：输入框只要有内容变化就要检测
const handleInput = (index: number) => {
  const rule = localRules.value[index];
  if (!rule.isRight) {
    rule.isRight = validateRule(rule.content);
  }
};

const handleRuleContentChange = (index: number) => {
  const rule = localRules.value[index];
  localRules.value[index].isRight = validateRule(rule.content);
  if (rule.id) {
    rule.type = rule.content === rule.original ? '' : 'modify';
  }
  updateRuleParams();
};

const updateRuleParams = () => {
  const params: IRuleUpdateParams = {
    add_scope: [],
    del_id: [],
    alter_scope: [],
  };
  localRules.value.forEach((item) => {
    const { id, type, content } = item;
    switch (type) {
      case 'new':
        if (content) {
          params.add_scope.push(content);
        }
        break;
      case 'del':
        params.del_id.push(id);
        break;
      case 'modify':
        params.alter_scope.push({ id, scope: content });
        break;
    }
  });
  emits('change', params);
};

const handleRuleValidate = () => {
  localRules.value.forEach((item) => {
    item.isRight = validateRule(item.content);
  });
  return localRules.value.some(item => !item.isRight);
};

defineExpose({ handleRuleValidate });
</script>
<style lang="scss" scoped>
.head {
  display: flex;
  justify-content: space-between;
  .view-rule {
    font-size: 12px;
    color: #3A84FF;
    cursor: pointer;
  }
}
.title {
  position: relative;
  margin: 0 0 6px;
  line-height: 20px;
  font-size: 12px;
  color: #63656e;
  &:after {
    position: absolute;
    top: 0;
    width: 14px;
    color: #ea3636;
    text-align: center;
    content: '*';
  }
}
.rule-list {
  margin-bottom: 24px;
}
.rule-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  .rule-input {
    position: relative;
    width: 312px;
    .status-tag {
      position: absolute;
      top: 4px;
      right: 8px;
      padding: 0 8px;
      line-height: 22px;
      font-size: 12px;
      border-radius: 2px;
      &.new {
        background: #e4faf0;
        color: #14a568;
      }
      &.del {
        background: #feebea;
        color: #ea3536;
      }
      &.modify {
        background: #fff1db;
        color: #fe9c00;
      }
    }
  }
  .action-btns {
    width: 38px;
    color: #979ba5;
    font-size: 14px;
    text-align: right;
    > i {
      cursor: pointer;
      &:hover {
        color: #3a84ff;
      }
    }
  }
}
.is-error {
  .rule-input {
    border-color: #ea3636 !important;
  }
}
.error-info {
  margin: 4px 0 6px;
  height: 16px;
  color: #ea3636;
  font-size: 12px;
  line-height: 16px;
}
.preview-btn {
  margin-top: 16px;
  padding: 5px 0;
  width: 100%;
  text-align: center;
  font-size: 14px;
  color: #3a84ff;
  border: 1px solid #3a84ff;
  border-radius: 2px;
  cursor: pointer;
}
</style>

<style lang="scss">
.view-rule-wrap {
  padding: 16px !important;
}
</style>
