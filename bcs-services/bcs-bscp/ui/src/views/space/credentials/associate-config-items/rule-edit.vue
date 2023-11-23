<template>
  <div class="rule-edit">
    <p class="title">配置关联规则</p>
    <div class="rules-edit-area">
      <div v-for="(rule, index) in localRules" class="rule-list" :key="index">
        <div class="rule-item">
          <bk-input
            v-model="rule.content"
            class="rule-input"
            placeholder="请填写"
            :disabled="rule.type === 'del'"
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
        <div class="error-info"><span v-if="!rule.isRight">输入的规则有误，请重新确认</span></div>
      </div>
      <div class="tips">
        <div>填写格式为“服务名称/配置路径”，支持通配符，常用方法：</div>
        <div>关联myservice服务下所有的配置(包含子目录):myservice/**</div>
        <div>关联myservice服务/etc目录下所有的配置(不含子目录):myservice/etc/*</div>
        <div>关联myservice服务/etc/nginx/nginx.conf文件:myservice/etc/nginx/nginx.conf</div>
      </div>
    </div>
    <!-- <div class="preview-btn">预览匹配结果</div> -->
  </div>
</template>
<script setup lang="ts">
import { ref, watch } from 'vue';
import { ICredentialRule, IRuleEditing, IRuleUpdateParams } from '../../../../../types/credential';

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

const testRule = (rule: string) => /\/([A-Za-z0-9]+[A-Za-z0-9-_.]*\/?)*$/.test(rule);

const handleRuleContentChange = (index: number) => {
  const rule = localRules.value[index];
  localRules.value[index].isRight = testRule(rule.content);
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
</script>
<style lang="scss" scoped>
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
    > i {
      cursor: pointer;
      &:hover {
        color: #3a84ff;
      }
    }
  }
}
.error-info {
  height: 16px;
  color: #ea3636;
  font-size: 12px;
  line-height: 16px;
}
.tips {
  margin: 8px 0 0;
  line-height: 20px;
  color: #979ba5;
  font-size: 12px;
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
