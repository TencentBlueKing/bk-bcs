<template>
  <bk-sideslider
    title="关联配置文件"
    width="400"
    :is-show="props.show"
    :before-close="handleBeforeClose"
    @closed="handleClose"
  >
    <section class="associate-config-items">
      <div :class="['rules-wrapper', { 'edit-mode': isRuleEdit }]">
        <RuleEdit
          v-if="isRuleEdit"
          ref="ruleEdit"
          :id="props.id"
          :rules="rules"
          @change="handleRuleChange"
          @formChange="isFormChange = true"/>
        <RuleView v-else :rules="rules" @edit="isRuleEdit = true" />
      </div>
      <!-- <div class="results-wrapper">
        <MatchingResult />
      </div> -->
    </section>
    <div class="action-btns">
      <bk-button v-if="isRuleEdit" theme="primary" :loading="pending" @click="handleSave"> 保存 </bk-button>
      <bk-button
        v-else
        v-cursor="{ active: !props.hasManagePerm }"
        :class="{ 'bk-button-with-no-perm': !props.hasManagePerm }"
        theme="primary"
        @click="handleOpenEdit"
      >
        编辑规则
      </bk-button>
      <bk-button @click="handleClose">{{ isRuleEdit ? '取消' : '关闭' }}</bk-button>
    </div>
  </bk-sideslider>
</template>
<script setup lang="ts">
import { ref, watch } from 'vue';
import { storeToRefs } from 'pinia';
import useGlobalStore from '../../../../store/global';
import { getCredentialScopes, updateCredentialScopes } from '../../../../api/credentials';
import { ICredentialRule, IRuleUpdateParams } from '../../../../../types/credential';
import useModalCloseConfirmation from '../../../../utils/hooks/use-modal-close-confirmation';
// import MatchingResult from './matching-result.vue'
import RuleView from './rule-view.vue';
import RuleEdit from './rule-edit.vue';

const { spaceId } = storeToRefs(useGlobalStore());

const props = defineProps<{
  show: boolean;
  id: number;
  permCheckLoading: boolean;
  hasManagePerm: boolean;
}>();

const emits = defineEmits(['close', 'refresh', 'applyPerm']);

const loading = ref(true);
const rules = ref<ICredentialRule[]>([]);
const ruleChangeParams = ref<IRuleUpdateParams>({
  add_scope: [],
  del_id: [],
  alter_scope: [],
});
const isRuleEdit = ref(false);
const isFormChange = ref(false);
const pending = ref(false);
const ruleEdit = ref();
watch(
  () => props.show,
  (val) => {
    if (val) {
      loadRules();
      ruleChangeParams.value = {
        add_scope: [],
        del_id: [],
        alter_scope: [],
      };
    }
  },
);

const loadRules = async () => {
  loading.value = true;
  const res = await getCredentialScopes(spaceId.value, props.id);
  rules.value = res.details;
  loading.value = false;
};

const handleOpenEdit = () => {
  if (props.permCheckLoading || !props.hasManagePerm) {
    emits('applyPerm');
  }
  isRuleEdit.value = true;
};

const handleRuleChange = (val: IRuleUpdateParams) => {
  ruleChangeParams.value = Object.assign({}, ruleChangeParams.value, val);
};

const handleSave = async () => {
  if (ruleEdit.value.handleRuleValidate()) return;
  pending.value = true;
  try {
    await updateCredentialScopes(spaceId.value, props.id, ruleChangeParams.value);
    ruleChangeParams.value = {
      add_scope: [],
      del_id: [],
      alter_scope: [],
    };
    isRuleEdit.value = false;
    loadRules();
    emits('refresh');
  } catch (e) {
    console.error(e);
  } finally {
    pending.value = false;
  }
};

const handleBeforeClose = async () => {
  if (isRuleEdit.value && isFormChange.value) {
    const result = await useModalCloseConfirmation();
    return result;
  }
  return true;
};

const handleClose = () => {
  isRuleEdit.value = false;
  pending.value = false;
  emits('close');
};
</script>
<style lang="scss" scoped>
.associate-config-items {
  display: flex;
  align-items: flex-start;
  height: calc(100vh - 101px);
}
.rules-wrapper {
  padding: 16px 24px;
  width: 400px;
  height: 100%;
  background: #ffffff;
  overflow: auto;
  &.edit-mode {
    padding-right: 16px;
  }
}
.results-wrapper {
  padding: 16px 24px;
  width: 560px;
  height: 100%;
  background: #f5f7fa;
  overflow: auto;
}
.action-btns {
  border-top: 1px solid #dcdee5;
  padding: 8px 24px;
  .bk-button {
    margin-right: 8px;
    min-width: 88px;
  }
}
</style>
