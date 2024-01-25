<template>
  <bk-dialog
    :title="t('上线版本')"
    ext-cls="release-version-dialog"
    :is-show="props.show"
    :esc-close="false"
    :quick-close="false"
    :is-loading="pending"
    @closed="handleClose"
    @confirm="handleConfirm">
    <bk-form class="form-wrapper" form-type="vertical" ref="formRef" :rules="rules" :model="localVal">
      <bk-form-item :label="t('本次上线分组')">
        <div v-for="group in groupsToBePreviewed" class="group-item" :key="group.id">
          <div class="name">{{ group.name }}</div>
          <default-group-rules-popover
            v-if="group.id === 0 && excludedGroups.length > 0"
            :excluded-groups="excludedGroups" />
          <div v-if="group.rules.length > 0" class="rules">
            <bk-overflow-title type="tips">
              <span v-for="(rule, index) in group.rules" :key="index" class="rule">
                <span v-if="index > 0"> & </span>
                <rule-tag class="tag-item" :rule="rule" />
              </span>
            </bk-overflow-title>
          </div>
        </div>
        <template v-if="groupsToBePreviewed.length === 0">--</template>
      </bk-form-item>
      <bk-form-item :label="t('上线说明')" property="memo">
        <bk-input v-model="localVal.memo" type="textarea" :placeholder="t('请输入')" :maxlength="200" :resize="true" />
      </bk-form-item>
    </bk-form>
    <template #footer>
      <div class="dialog-footer">
        <bk-button theme="primary" :loading="pending" @click="handleConfirm">{{t('确定上线')}}</bk-button>
        <bk-button :disabled="pending" @click="handleClose">{{t('取消')}}</bk-button>
      </div>
    </template>
  </bk-dialog>
</template>
<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { publishVersion } from '../../../../../../api/config';
import { IGroupToPublish } from '../../../../../../../types/group';
import RuleTag from '../../../../groups/components/rule-tag.vue';
import DefaultGroupRulesPopover from './default-group-rules-popover.vue';

interface IFormData {
  groups: number[];
  all: boolean;
  memo: string;
}

const { t } = useI18n();

const props = defineProps<{
  show: boolean;
  bkBizId: string;
  appId: number;
  releaseId: number | null;
  releaseType: string;
  groups: IGroupToPublish[];
}>();

const emits = defineEmits(['confirm', 'update:show']);

const localVal = ref<IFormData>({
  groups: [],
  all: false,
  memo: '',
});
const pending = ref(false);
const formRef = ref();
const rules = {
  memo: [
    {
      validator: (value: string) => value.length <= 200,
      message: t('最大长度200个字符'),
    },
  ],
};

// 只展示已上线的分组
const groupsToBePreviewed = computed(() => props.groups.filter((group) => {
  const { id, release_id } = group;
  // 过滤掉当前版本已上线分组
  if (release_id === props.releaseId) {
    return false;
  }
  return id === 0 || release_id > 0 || props.releaseType === 'select';
}));

// 默认分组对应的排除分组
const excludedGroups = computed(() => props.groups.filter(group => group.release_id > 0 && group.id > 0));

watch(
  () => props.groups,
  () => {
    localVal.value.groups = props.groups.map(item => item.id);
  },
  { immediate: true },
);

const handleClose = () => {
  emits('update:show', false);
  localVal.value = {
    groups: [],
    all: false,
    memo: '',
  };
};

const handleConfirm = async () => {
  try {
    pending.value = true;
    await formRef.value.validate();
    const params = { ...localVal.value };
    // 全部实例上线，只需要将all置为true
    if (props.releaseType === 'all') {
      params.groups = [];
      params.all = true;
    }
    const resp = await publishVersion(props.bkBizId, props.appId, props.releaseId as number, params);
    handleClose();
    // 目前组件库dialog关闭自带250ms的延迟，所以这里延时300ms
    setTimeout(() => {
      emits('confirm', resp.data.have_credentials as boolean);
    }, 300);
  } catch (e) {
    console.error(e);
    // InfoBox({
    // // @ts-ignore
    //   infoType: "danger",
    //   title: '版本上线失败',
    //   subTitle: e.response.data.error.message,
    //   confirmText: '重试',
    //   onConfirm () {
    //     handleConfirm()
    //   }
    // })
  } finally {
    pending.value = false;
  }
};
</script>
<style lang="scss" scoped>
.form-wrapper {
  padding-bottom: 24px;
  :deep(.bk-form-label) {
    font-size: 12px;
  }
}
.group-item {
  display: flex;
  align-items: center;
  margin-bottom: 8px;
  white-space: nowrap;
  overflow: hidden;
  .name {
    padding: 0 10px;
    height: 22px;
    line-height: 22px;
    font-size: 12px;
    color: #63656e;
    background: #f0f1f5;
    border-radius: 2px;
  }
  .rules {
    margin-left: 8px;
    font-size: 12px;
    line-height: 22px;
    color: #c4c6cc;
    overflow: hidden;
  }
}
.dialog-footer {
  .bk-button {
    margin-left: 8px;
  }
}
</style>
<style lang="scss">
.release-version-dialog.bk-dialog-wrapper .bk-dialog-header {
  padding-bottom: 20px;
}
</style>
