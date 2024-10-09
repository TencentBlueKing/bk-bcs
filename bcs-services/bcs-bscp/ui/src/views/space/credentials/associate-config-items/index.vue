<template>
  <bk-sideslider
    :title="sideSliderTitle"
    :width="960"
    :is-show="props.show"
    :before-close="handleBeforeClose"
    @closed="handleClose">
    <template #header>
      <div class="header-wrapper">
        <span>{{ sideSliderTitle }}</span>
        <bk-popover
          v-if="!props.isExampleMode"
          placement="bottom-start"
          theme="light"
          trigger="click"
          ext-cls="view-rule-wrap">
          <span class="view-rule">{{ t('查看规则示例') }}</span>
          <template #content>
            <ViewRuleExample />
          </template>
        </bk-popover>
      </div>
    </template>
    <section class="associate-config-items">
      <div :class="['rules-wrapper', { 'edit-mode': isRuleEdit }]">
        <RuleEdit
          v-if="isRuleEdit"
          v-model:preview-rule="previewRule"
          ref="ruleEdit"
          :id="props.id"
          :rules="rules"
          :app-list="appList"
          :is-example-mode="props.isExampleMode"
          @change="handleRuleChange"
          @form-change="isFormChange = true"
          @trigger-save-btn-disabled="saveBtnDisabled = $event" />
        <RuleView v-else v-model:preview-rule="previewRule" :rules="rules" @edit="isRuleEdit = true" />
      </div>
      <div v-if="rules.length || isRuleEdit" class="results-wrapper">
        <MatchingResult :rule="previewRule" :bk-biz-id="spaceId" />
      </div>
    </section>
    <div class="action-btns">
      <bk-button
        v-if="isRuleEdit"
        theme="primary"
        :loading="pending"
        :disabled="saveBtnDisabled"
        v-bk-tooltips="{ content: '请先预览所有关联规则修改结果后，才能保存', disabled: !saveBtnDisabled }"
        @click="handleSave">
        {{ t('保存') }}
      </bk-button>
      <bk-button
        v-else
        v-cursor="{ active: !props.hasManagePerm }"
        :class="{ 'bk-button-with-no-perm': !props.hasManagePerm }"
        theme="primary"
        @click="handleOpenEdit">
        {{ t('编辑规则') }}
      </bk-button>
      <bk-button @click="handleClose">{{ isRuleEdit ? t('取消') : t('关闭') }}</bk-button>
    </div>
  </bk-sideslider>
</template>
<script setup lang="ts">
  import { ref, watch, onMounted, computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { storeToRefs } from 'pinia';
  import useGlobalStore from '../../../../store/global';
  import { getCredentialScopes, updateCredentialScopes } from '../../../../api/credentials';
  import { ICredentialRule, IRuleUpdateParams, IPreviewRule } from '../../../../../types/credential';
  import useModalCloseConfirmation from '../../../../utils/hooks/use-modal-close-confirmation';
  import { getAppList } from '../../../../api/index';
  import { IAppItem } from '../../../../../types/app';
  import MatchingResult from './matching-result.vue';
  import RuleView from './rule-view.vue';
  import RuleEdit from './rule-edit.vue';
  import { Message } from 'bkui-vue';
  import ViewRuleExample from './view-rule-example.vue';

  const { spaceId } = storeToRefs(useGlobalStore());
  const { t } = useI18n();

  const props = withDefaults(
    defineProps<{
      show: boolean;
      id: number;
      permCheckLoading: boolean;
      hasManagePerm: boolean;
      isExampleMode?: boolean;
      exampleRules?: ICredentialRule[];
    }>(),
    {
      isExampleMode: false, // 配置示例模式(无密钥id)
      exampleRules: () => [],
    },
  );

  const emits = defineEmits(['close', 'refresh', 'applyPerm', 'sendExampleRules']);

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
  const appList = ref<IAppItem[]>([]);
  const previewRule = ref<IPreviewRule | null>(null);
  const saveBtnDisabled = ref(false);

  onMounted(async () => {
    const resp = await getAppList(spaceId.value, { start: 0, all: true });
    appList.value = resp.details;
  });

  const sideSliderTitle = computed(() => (props.isExampleMode ? t('配置文件筛选规则') : t('关联服务配置')));

  watch(
    () => props.show,
    (val) => {
      if (val) {
        // 配置示例无需载入密钥关联的规则
        if (props.isExampleMode) {
          rules.value = props.exampleRules.length ? props.exampleRules : [];
        } else {
          loadRules();
        }
        ruleChangeParams.value = {
          add_scope: [],
          del_id: [],
          alter_scope: [],
        };
      }
      previewRule.value = null;
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
    if (props.isExampleMode) {
      // 配置示例不需要调用接口
      emits('sendExampleRules', ruleChangeParams.value);
      ruleChangeParams.value = {
        add_scope: [],
        del_id: [],
        alter_scope: [],
      };
      isRuleEdit.value = false;
    } else {
      try {
        pending.value = true;
        await updateCredentialScopes(spaceId.value, props.id, ruleChangeParams.value);
        ruleChangeParams.value = {
          add_scope: [],
          del_id: [],
          alter_scope: [],
        };
        isRuleEdit.value = false;
        loadRules();
        emits('refresh');
        Message({
          theme: 'success',
          message: t('编辑规则成功'),
        });
      } catch (e) {
        console.error(e);
      } finally {
        pending.value = false;
      }
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
    flex: 1;
    padding: 16px 24px;
    height: 100%;
    background: #ffffff;
    overflow: auto;
    &.edit-mode {
      padding-right: 16px;
    }
  }
  .results-wrapper {
    padding: 16px 24px;
    width: 360px;
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
  .header-wrapper {
    .view-rule {
      margin-left: 24px;
      font-size: 12px;
      color: #3a84ff;
      cursor: pointer;
    }
  }
</style>
