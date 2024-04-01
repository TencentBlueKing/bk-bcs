<template>
  <bk-dialog
    :title="`${t('上线版本')}-${versionData.spec.name}`"
    ext-cls="release-version-dialog"
    :is-show="props.show"
    :esc-close="false"
    :quick-close="false"
    :is-loading="pending"
    @closed="handleClose"
    @confirm="handleConfirm">
    <bk-form class="form-wrapper" form-type="vertical" ref="formRef" :rules="rules" :model="localVal">
      <template v-if="props.releaseType === 'all'">
        <div v-if="excludeGroups.length > 0" class="exclude-groups">
          <p class="tips">
            {{ t('确认上线后，以下分组') }}
            <span class="em">{{ t('以外') }}</span>
            {{ t('的客户端实例将上线当前版本配置') }}
          </p>
          <div class="group-list-wrapper">
            <div v-for="group in excludeGroups" class="group-item" :key="group.id">
              <div class="name">{{ group.name }}</div>
              <div v-if="group.rules.length > 0" class="rules">
                <bk-overflow-title type="tips">
                  <span v-for="(rule, index) in group.rules" :key="index" class="rule">
                    <span v-if="index > 0"> & </span>
                    <rule-tag class="tag-item" :rule="rule" />
                  </span>
                </bk-overflow-title>
              </div>
            </div>
          </div>
        </div>
        <p v-else class="tips">
          {{ t('确认上线后，') }}
          <span class="em">{{ t('全部') }}</span>
          {{ t('客户端实例将上线当前版本配置') }}
        </p>
      </template>
      <bk-form-item
        v-if="previewData.length > 0"
        :label="
          props.releaseType === 'select'
            ? t('确认上线后，以下分组的客户端实例将上线当前版本配置')
            : t('以下分组将变更版本')
        ">
        <div class="group-list-wrapper">
          <div v-for="previewGroup in previewData" class="release-section" :key="previewGroup.id">
            <div class="section-header" @click="previewGroup.fold = !previewGroup.fold">
              <span class="angle-icon">
                <AngleRight v-if="previewGroup.fold" />
                <AngleDown v-else />
              </span>
              <div :class="['version-type-marking', previewGroup.type]">
                【{{ TYPE_MAP[previewGroup.type as keyof typeof TYPE_MAP] }}】
              </div>
              <span v-if="previewGroup.type === 'modify'" class="release-name">
                {{ previewGroup.name }} <ArrowsRight class="arrow-icon" /> {{ versionData.spec.name }}
              </span>
            </div>
            <div v-show="!previewGroup.fold" class="group-list">
              <div v-for="group in previewGroup.children" class="group-item" :key="group.id">
                <div class="name">{{ group.name }}</div>
                <div v-if="group.desc" class="desc">{{ group.desc }}</div>
                <div v-if="group.rules.length > 0" class="rules">
                  <bk-overflow-title type="tips">
                    <span v-for="(rule, index) in group.rules" :key="index" class="rule">
                      <span v-if="index > 0"> & </span>
                      <rule-tag class="tag-item" :rule="rule" />
                    </span>
                  </bk-overflow-title>
                </div>
              </div>
            </div>
          </div>
        </div>
      </bk-form-item>
      <bk-form-item :label="t('上线说明')" property="memo">
        <bk-input v-model="localVal.memo" type="textarea" :placeholder="t('请输入')" :maxlength="200" :resize="true" />
      </bk-form-item>
    </bk-form>
    <template #footer>
      <div class="dialog-footer">
        <bk-button theme="primary" :loading="pending" @click="handleConfirm">{{ t('确定上线') }}</bk-button>
        <bk-button :disabled="pending" @click="handleClose">{{ t('取消') }}</bk-button>
      </div>
    </template>
  </bk-dialog>
</template>
<script setup lang="ts">
  import { ref, watch } from 'vue';
  import { storeToRefs } from 'pinia';
  import { useI18n } from 'vue-i18n';
  import { AngleDown, AngleRight, ArrowsRight } from 'bkui-vue/lib/icon';
  import { publishVersion } from '../../../../../../api/config';
  import { IGroupToPublish, IGroupPreviewItem } from '../../../../../../../types/group';
  import { IConfigVersion } from '../../../../../../../types/config';
  import useConfigStore from '../../../../../../store/config';
  import { aggregatePreviewData, aggregateExcludedData } from '../hooks/aggregate-groups';
  import RuleTag from '../../../../groups/components/rule-tag.vue';

  const versionStore = useConfigStore();
  const { versionData } = storeToRefs(versionStore);

  const { t } = useI18n();

  interface IFormData {
    groups: number[];
    all: boolean;
    memo: string;
  }

  interface IModifyReleasePreviewItem extends IGroupPreviewItem {
    fold: boolean;
  }

  const TYPE_MAP = {
    plain: t('首次上线'),
    modify: t('变更版本'),
    retain: t('保留版本'),
  };

  const props = withDefaults(
    defineProps<{
      show: boolean;
      bkBizId: string;
      appId: number;
      versionList: IConfigVersion[];
      groupList: IGroupToPublish[];
      releaseType: string;
      releasedGroups?: number[];
      groups: IGroupToPublish[];
    }>(),
    {
      releasedGroups: () => [],
    },
  );

  const emits = defineEmits(['confirm', 'update:show']);

  const localVal = ref<IFormData>({
    groups: [],
    all: false,
    memo: '',
  });
  const previewData = ref<IModifyReleasePreviewItem[]>([]);
  const excludeGroups = ref<IGroupToPublish[]>([]);
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

  watch(
    () => props.show,
    (val) => {
      if (val) {
        const previewList = aggregatePreviewData(
          props.groups,
          props.groupList,
          props.releasedGroups,
          props.releaseType,
          versionData.value.id,
        );
        previewData.value = previewList.map((item) => ({ ...item, fold: false }));
        const excludeList = aggregateExcludedData(
          props.groups,
          props.groupList,
          props.releaseType,
          versionData.value.id,
        );
        const list: IGroupToPublish[] = [];
        excludeList.forEach((item) => {
          list.push(...item.children);
        });
        excludeGroups.value = list;
      }
    },
  );

  watch(
    () => props.groups,
    () => {
      localVal.value.groups = props.groups.map((item) => item.id);
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
        if (excludeGroups.value.length > 0) {
          params.all = false;
        } else {
          params.all = true;
          params.groups = [];
        }
      }
      const resp = await publishVersion(props.bkBizId, props.appId, versionData.value.id, params);
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
  .exclude-groups {
    margin-bottom: 16px;
    .tips {
      display: flex;
      align-items: center;
      margin: 0 0 8px;
      font-size: 12px;
      line-height: 20px;
      .em {
        font-weight: 700;
        color: #ff9c01;
      }
    }
  }
  .group-list-wrapper {
    padding: 8px;
    max-height: 320px;
    border: 1px solid #dcdee5;
    border-radius: 2px;
    overflow: auto;
    .release-section {
      margin-bottom: 8px;
    }
    .section-header {
      display: flex;
      align-items: center;
      font-size: 12px;
      color: #63656e;
      cursor: pointer;
      &:hover {
        .angle-icon {
          color: #3a84ff;
        }
      }
      .angle-icon {
        font-size: 18px;
        line-height: 1;
      }
      .version-type-marking {
        &.modify {
          color: #ff9c01;
        }
      }
      .release-name {
        display: inline-flex;
        align-items: center;
        .arrow-icon {
          font-size: 20px;
          color: #ff9c01;
        }
      }
    }
  }
  .group-item {
    display: flex;
    align-items: center;
    white-space: nowrap;
    overflow: hidden;
    &:not(:last-child) {
      margin-bottom: 8px;
    }
    .name {
      padding: 0 10px;
      height: 22px;
      line-height: 22px;
      font-size: 12px;
      color: #63656e;
      background: #f0f1f5;
      border-radius: 2px;
    }
    .desc {
      font-size: 12px;
      color: #979ba5;
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
