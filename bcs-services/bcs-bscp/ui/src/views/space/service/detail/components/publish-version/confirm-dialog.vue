<template>
  <bk-dialog
    title="上线版本"
    ext-cls="release-version-dialog"
    :is-show="props.show"
    :esc-close="false"
    :quick-close="false"
    :is-loading="pending"
    @closed="handleClose"
    @confirm="handleConfirm"
  >
    <bk-form class="form-wrapper" form-type="vertical" ref="formRef" :rules="rules" :model="localVal">
      <bk-form-item label="上线分组">
        <div v-for="group in props.groups" class="group-item" :key="group.id">
          <div class="name">{{ group.name }}</div>
          <div class="rules">
            <bk-overflow-title type="tips">
              <span v-for="(rule, index) in group.rules" :key="index" class="rule">
                <span v-if="index > 0"> & </span>
                <rule-tag class="tag-item" :rule="rule" />
              </span>
            </bk-overflow-title>
          </div>
        </div>
      </bk-form-item>
      <bk-form-item label="上线说明" property="memo">
        <bk-input v-model="localVal.memo" type="textarea" :maxlength="200" :resize="true"></bk-input>
      </bk-form-item>
    </bk-form>
    <template #footer>
      <div class="dialog-footer">
        <bk-button theme="primary" :loading="pending" @click="handleConfirm">确定上线</bk-button>
        <bk-button :disabled="pending" @click="handleClose">取消</bk-button>
      </div>
    </template>
  </bk-dialog>
</template>
<script setup lang="ts">
import { ref, watch } from 'vue';
import { publishVersion } from '../../../../../../api/config';
import { IGroupToPublish } from '../../../../../../../types/group';
import RuleTag from '../../../../groups/components/rule-tag.vue';

interface IFormData {
  groups: number[];
  all: boolean;
  memo: string;
}

const props = defineProps<{
  show: boolean;
  bkBizId: string;
  appId: number;
  releaseId: number | null;
  groupType: string;
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
      message: '最大长度200个字符',
    },
  ],
};

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
    if (props.groupType === 'all') {
      params.groups = [];
      params.all = true;
    }
    await publishVersion(props.bkBizId, props.appId, props.releaseId as number, params);
    handleClose();
    // 目前组件库dialog关闭自带250ms的延迟，所以这里延时300ms
    setTimeout(() => {
      emits('confirm');
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
