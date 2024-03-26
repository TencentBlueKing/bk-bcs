<template>
  <div class="rule-edit">
    <div class="head">
      <p class="title">{{ t('配置关联规则') }}</p>
    </div>
    <div class="rules-edit-area">
      <div v-for="(rule, index) in localRules" class="rule-list" :key="index">
        <div :class="['rule-item', { 'rule-error': !rule.isRight }, { 'service-error': !rule.isSelectService }]">
          <bk-select
            v-model="rule.app"
            class="service-select"
            :filterable="true"
            :input-search="false"
            :disabled="rule.type === 'del'"
            :placeholder="t('请选择服务')"
            @change="handleSelectApp(index)">
            <bk-option v-for="app in appList" :id="app" :key="app.id" :name="app.spec.name" />
          </bk-select>
          <div style="width: 10px">/</div>
          <bk-input
            v-model="rule.content"
            class="rule-input"
            :placeholder="inputPlaceholder(rule)"
            :disabled="rule.type === 'del'"
            @input="handleRuleContentChange(index)">
            <template #suffix>
              <div
                v-if="rule.type"
                v-bk-tooltips="{
                  disabled: rule.type !== 'modify',
                  content: `${rule.original} -> ${rule.content}`,
                }"
                :class="`status-tag ${rule.type}`">
                {{ RULE_TYPE_MAP[rule.type] }}
              </div>
            </template>
          </bk-input>
          <div class="action-btns">
            <i
              v-if="rule.type === 'del'"
              v-bk-tooltips="t('撤销本次删除')"
              class="bk-bscp-icon icon-revoke revoke-icon"
              @click="handleRevoke(index)">
            </i>
            <template v-else>
              <i v-if="localRules.length > 1" class="bk-bscp-icon icon-reduce" @click="handleDeleteRule(index)"></i>
              <i style="margin-left: 10px" class="bk-bscp-icon icon-add" @click="handleAddRule(index)"></i>
            </template>
          </div>
          <div
            :class="['preview', { 'preview-mode': previewRule?.index === index }, { 'need-preview': rule.needPreview }]"
            @click="handlePreviewRule(rule, index)">
            <span>预览</span><Arrows-Right class="arrow-icon" />
          </div>
        </div>
        <div class="error-info" v-if="!rule.isRight || !rule.isSelectService">
          <span v-if="!rule.isSelectService">{{ t('请选择服务') }}</span>
          <span v-else-if="!rule.isRight" class="rule-error">{{ t('输入的规则有误，请重新确认') }}</span>
        </div>
      </div>
    </div>
    <!-- <div class="preview-btn">预览匹配结果</div> -->
  </div>
</template>
<script setup lang="ts">
  import { ref, watch, computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { ICredentialRule, IRuleEditing, IRuleUpdateParams, IPreviewRule } from '../../../../../types/credential';
  import { IAppItem } from '../../../../../types/app';
  import { ArrowsRight } from 'bkui-vue/lib/icon';

  const { t } = useI18n();
  const props = defineProps<{
    rules: ICredentialRule[];
    appList: IAppItem[];
    previewRule: IPreviewRule | null;
  }>();

  const emits = defineEmits(['change', 'formChange', 'update:previewRule', 'trigger-save-btn-disabled']);

  const RULE_TYPE_MAP: { [key: string]: string } = {
    new: t('新增'),
    del: t('删除'),
    modify: t('修改'),
  };

  const localRules = ref<IRuleEditing[]>([]);

  const inputPlaceholder = computed(() => (rule: IRuleEditing) => {
    if (!rule.app) return ' ';
    if (rule.app.spec.config_type === 'file') return t('请输入文件路径');
    return t('请输入配置项名称');
  });

  const needPreview = computed(() => localRules.value.some((rule) => rule.needPreview));

  const transformRulesToEditing = (rules: ICredentialRule[]) => {
    const rulesEditing: IRuleEditing[] = [];
    rules.forEach((item) => {
      const {
        id,
        spec: { app, scope },
      } = item;
      const selectApp = props.appList.find((appItem) => appItem.spec.name === app);
      rulesEditing.push({
        id,
        type: '',
        content: scope.slice(1),
        original: scope.slice(1),
        isRight: true,
        app: selectApp || null,
        originalApp: app,
        isSelectService: true,
        needPreview: false,
      });
    });
    return rulesEditing;
  };

  watch(
    () => props.rules,
    (val) => {
      if (val.length === 0) {
        localRules.value = [
          {
            id: 0,
            type: 'new',
            content: '',
            original: '',
            isRight: true,
            app: null,
            originalApp: '',
            isSelectService: true,
            needPreview: false,
          },
        ];
      } else {
        localRules.value = transformRulesToEditing(val);
      }
    },
    { immediate: true },
  );

  watch(
    () => needPreview.value,
    () => {
      emits('trigger-save-btn-disabled', needPreview.value);
    },
  );

  const handleAddRule = (index: number) => {
    localRules.value.splice(index + 1, 0, {
      id: 0,
      type: 'new',
      content: '',
      original: '',
      isRight: true,
      app: null,
      originalApp: '',
      isSelectService: true,
      needPreview: false,
    });
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
    const { content, original, app, originalApp } = rule;
    rule.type = content === original && app?.spec.name === originalApp ? '' : 'modify';
    updateRuleParams();
  };

  const validateRule = (rule: IRuleEditing) => {
    // 文件型 需要忽略前导/进行校验
    if (rule.app?.spec.config_type === 'file') {
      const validateContent = rule.content[0] === '/' ? rule.content.slice(1) : rule.content;
      if (!validateContent.length) {
        return false;
      }
      const paths = validateContent.split('/');
      return !!paths.length && paths.every((path) => path.length > 0);
    }
    // 键值型
    return !!rule.content.length;
  };

  const handleSelectApp = (index: number) => {
    const rule = localRules.value[index];
    localRules.value[index].isSelectService = !!localRules.value[index].app;
    if (rule.id) {
      const isRuleChange = rule.content !== rule.original || rule.app?.spec.name !== rule.originalApp;
      rule.type = isRuleChange ? 'modify' : '';
      rule.needPreview = isRuleChange;
    }
    updateRuleParams();
  };

  const handleRuleContentChange = (index: number) => {
    const rule = localRules.value[index];
    localRules.value[index].isRight = validateRule(rule);
    if (rule.id) {
      const isRuleChange = rule.content !== rule.original || rule.app?.spec.name !== rule.originalApp;
      rule.type = isRuleChange ? 'modify' : '';
      rule.needPreview = isRuleChange;
    }
    updateRuleParams();
    emits('formChange');
  };

  const updateRuleParams = () => {
    const params: IRuleUpdateParams = {
      add_scope: [],
      del_id: [],
      alter_scope: [],
    };
    localRules.value.forEach((item) => {
      const { id, type, content, app } = item;
      switch (type) {
        case 'new':
          if (content) {
            params.add_scope.push({ app: app!.spec.name, scope: content[0] === '/' ? content : `/${content}` });
          }
          break;
        case 'del':
          params.del_id.push(id);
          break;
        case 'modify':
          params.alter_scope.push({ id, scope: content[0] === '/' ? content : `/${content}`, app: app!.spec.name });
          break;
      }
    });
    emits('change', params);
    emits('formChange');
  };

  const handleRuleValidate = () => {
    localRules.value.forEach((item) => {
      item.isRight = validateRule(item);
      item.isSelectService = !!item.app;
    });
    return localRules.value.some((item) => !item.isRight || !item.isSelectService);
  };

  const handlePreviewRule = (rule: IRuleEditing, index: number) => {
    // 规则为新增或修改 先进行校验
    if (rule.type) {
      rule.isSelectService = !!rule.app;
      rule.isRight = validateRule(rule);
    }
    let previewRule: IPreviewRule | null;
    // 规则错误取消预览
    if (!rule.isRight || !rule.isSelectService) {
      previewRule = null;
    } else {
      previewRule = {
        id: rule.id,
        appName: rule.app!.spec.name,
        scopeContent: `/${rule.content}`,
        index,
      };
      rule.needPreview = false;
    }
    emits('update:previewRule', previewRule);
  };

  defineExpose({ handleRuleValidate });
</script>
<style lang="scss" scoped>
  .head {
    display: flex;
    justify-content: space-between;
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
    .service-select {
      width: 180px;
      margin-right: 8px;
    }
    .rule-input {
      position: relative;
      width: 248px;
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
      color: #979ba5;
      font-size: 14px;
      text-align: right;
      padding: 0 17px 0 9px;
      > i {
        cursor: pointer;
        &:hover {
          color: #3a84ff;
        }
      }
    }
    .preview {
      position: relative;
      display: flex;
      align-items: center;
      justify-content: space-evenly;
      width: 52px;
      height: 24px;
      background: #f0f1f5;
      border-radius: 2px;
      color: #979ba5;
      font-size: 12px;
      .arrow-icon {
        font-size: 16px;
      }
      &.preview-mode {
        background: #e1ecff;
        color: #3a84ff;
      }
      &.need-preview::after {
        position: absolute;
        content: '';
        width: 6px;
        height: 6px;
        background: #ff9c01;
        border-radius: 50%;
        right: -3px;
        top: -3px;
      }
    }
  }
  .rule-error {
    .rule-input {
      border-color: #ea3636;
    }
  }
  .service-error {
    .service-select {
      :deep(.bk-input) {
        border-color: #ea3636;
      }
    }
  }
  .error-info {
    position: relative;
    margin: 4px 0 6px;
    height: 16px;
    color: #ea3636;
    font-size: 12px;
    line-height: 16px;
    .rule-error {
      position: absolute;
      left: 200px;
    }
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
