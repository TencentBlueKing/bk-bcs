<template>
  <bk-sideslider
    ref="sideSliderRef"
    width="640"
    :title="t('查看配置文件')"
    :quick-close="true"
    :is-show="props.show"
    @closed="close"
    @shown="setEditorHeight">
    <bk-loading :loading="detailLoading" class="config-loading-container">
      <bk-tab v-model:active="activeTab" type="card-grid" ext-cls="view-config-tab">
        <bk-tab-panel name="content" :label="t('配置文件信息')">
          <bk-form label-width="100" form-type="vertical">
            <bk-form-item :label="t('配置文件名称')">{{ configDetail.name }}</bk-form-item>
            <bk-form-item :label="t('配置文件路径')">{{ configDetail.path }}</bk-form-item>
            <bk-form-item :label="t('配置文件描述')">
              <div class="memo">{{ configDetail.memo || configDetail.revision_memo || '--' }}</div>
            </bk-form-item>
            <bk-form-item :label="t('配置文件内容')">
              <div v-if="configDetail.file_type === 'binary'" class="binary-file-card" @click="handleDownloadFile">
                <div class="basic-info">
                  <TextFill class="file-icon" />
                  <div class="content">
                    <div class="name">{{ configDetail.name }}</div>
                    <div class="time">{{ datetimeFormat(configDetail.update_at || configDetail.create_at) }}</div>
                  </div>
                  <div class="size">{{ byteUnitConverse(Number(configDetail.byte_size)) }}</div>
                </div>
              </div>
              <ConfigContentEditor
                v-else
                :content="content as string"
                :editable="false"
                :show-tips="false"
                :height="editorHeight"
                :variables="variables" />
            </bk-form-item>
          </bk-form>
        </bk-tab-panel>
        <bk-tab-panel name="meta" :label="t('元数据')">
          <ConfigContentEditor
            language="json"
            :content="JSON.stringify(configDetail, null, 2)"
            :editable="false"
            :show-tips="false" />
        </bk-tab-panel>
      </bk-tab>
    </bk-loading>
    <section class="action-btns">
      <bk-button @click="close">{{ t('关闭') }}</bk-button>
    </section>
  </bk-sideslider>
</template>
<script setup lang="ts">
  import { nextTick, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { storeToRefs } from 'pinia';
  import { TextFill } from 'bkui-vue/lib/icon';
  import ConfigContentEditor from '../../components/config-content-editor.vue';
  import {
    getConfigItemDetail,
    getReleasedConfigItemDetail,
    downloadConfigContent,
  } from '../../../../../../../api/config';
  import {
    getTemplateVersionsDetailByIds,
    getTemplateVersionDetail,
    downloadTemplateContent,
  } from '../../../../../../../api/template';
  import { byteUnitConverse, datetimeFormat } from '../../../../../../../utils/index';
  import { fileDownload } from '../../../../../../../utils/file';
  import { IVariableEditParams } from '../../../../../../../../types/variable';
  import { IFileConfigContentSummary } from '../../../../../../../../types/config';
  import { getReleasedAppVariables } from '../../../../../../../api/variable';
  import useConfigStore from '../../../../../../../store/config';

  interface IConfigMeta {
    name: string;
    path: string;
    file_mode: string;
    file_type: string;
    memo?: string;
    revision_memo?: string;
    revision_version?: string;
    byte_size: string;
    origin_byte_size?: string;
    signature: string;
    origin_signature?: string;
    md5: string;
    create_at: string;
    creator: string;
    update_at?: string;
    reviser?: string;
    user: string;
    user_group: string;
    privilege: string;
  }

  const { versionData } = storeToRefs(useConfigStore());
  const { t } = useI18n();

  const props = defineProps<{
    bkBizId: string;
    appId: number;
    id: number;
    versionId: number;
    type: string; // 取值为config/template，分别表示非模板套餐下配置文件和模板套餐下配置文件
    show: Boolean;
  }>();

  const emits = defineEmits(['update:show']);

  const detailLoading = ref(true);
  const activeTab = ref('content');
  const configDetail = ref<IConfigMeta>({
    name: '',
    path: '',
    file_mode: '',
    file_type: '',
    byte_size: '',
    signature: '',
    md5: '',
    create_at: '',
    creator: '',
    user: '',
    user_group: '',
    privilege: '',
  });
  const content = ref<string | IFileConfigContentSummary>('');
  const variables = ref<IVariableEditParams[]>([]);
  const variablesLoading = ref(false);
  const tplSpaceId = ref(0);
  const sideSliderRef = ref();
  const editorHeight = ref(0);

  watch(
    () => props.show,
    (val) => {
      if (val) {
        getDetailData();
        content.value = '';
        activeTab.value = 'content';
        variables.value = [];
      }
    },
  );

  const getDetailData = async () => {
    detailLoading.value = true;
    if (props.type === 'config') {
      getConfigDetail();
    } else if (props.type === 'template') {
      getTemplateDetail();
    }
    // 未命名版本id为0，不需要展示变量替换
    if (props.versionId) {
      getVariableList();
    }
  };

  // 获取非模板套餐下配置文件详情配置，非文件类型配置文件内容下载内容，文件类型手动点击时再下载
  const getConfigDetail = async () => {
    try {
      if (versionData.value.id) {
        const res = await getReleasedConfigItemDetail(props.bkBizId, props.appId, versionData.value.id, props.id);
        const { content, memo } = res.config_item.commit_spec;
        const { byte_size, origin_byte_size, signature, origin_signature, md5 } = content;
        const { create_at, creator, update_at, reviser } = res.config_item.revision;
        const { name, path, file_type, file_mode, permission } = res.config_item.spec;
        const { user, user_group, privilege } = permission;
        configDetail.value = {
          name,
          path,
          file_type,
          file_mode,
          memo,
          byte_size,
          origin_byte_size,
          signature,
          origin_signature,
          md5,
          create_at,
          creator,
          update_at,
          reviser,
          user,
          user_group,
          privilege,
        };
      } else {
        const res = await getConfigItemDetail(props.bkBizId, props.id, props.appId);
        const { create_at, creator, update_at, reviser } = res.config_item.revision;
        const { name, memo, path, file_type, file_mode, permission } = res.config_item.spec;
        const { user, user_group, privilege } = permission;
        const { byte_size, signature, md5 } = res.content;
        configDetail.value = {
          name,
          path,
          file_type,
          file_mode,
          memo,
          byte_size,
          signature,
          md5,
          create_at,
          creator,
          update_at,
          reviser,
          user,
          user_group,
          privilege,
        };
      }
      const signature = versionData.value.id
        ? (configDetail.value.origin_signature as string)
        : configDetail.value.signature;
      if (configDetail.value.file_type === 'binary') {
        content.value = { name: configDetail.value.name, size: configDetail.value.byte_size, signature };
      } else {
        const configContent = await downloadConfigContent(props.bkBizId, props.appId, signature);
        content.value = String(configContent);
      }
    } catch (e) {
      console.error(e);
    } finally {
      detailLoading.value = false;
    }
  };

  // 获取模板配置详情，非文件类型配置文件内容下载内容，文件类型手动点击时再下载
  const getTemplateDetail = async () => {
    try {
      detailLoading.value = true;
      let template_space_id;
      if (versionData.value.id) {
        const res = await getTemplateVersionDetail(props.bkBizId, props.appId, versionData.value.id, props.id);
        configDetail.value = { ...res.detail };
        template_space_id = res.detail.template_space_id;
      } else {
        const res = await getTemplateVersionsDetailByIds(props.bkBizId, [props.id]);
        const { revision, spec } = res.details[0];
        const { creator, create_at } = revision;
        const { content_spec, file_mode, file_type, name, revision_memo, revision_version, path, permission } = spec;
        const { byte_size, signature, md5 } = content_spec;
        const { user, user_group, privilege } = permission;
        template_space_id = res.details[0].attachment?.template_space_id;
        configDetail.value = {
          name,
          path,
          file_type,
          file_mode,
          revision_memo,
          revision_version,
          byte_size,
          signature,
          md5,
          create_at,
          creator,
          user,
          user_group,
          privilege,
        };
      }

      tplSpaceId.value = template_space_id;
      const signature = versionData.value.id
        ? (configDetail.value.origin_signature as string)
        : configDetail.value.signature;
      if (configDetail.value.file_type === 'binary') {
        content.value = {
          name: configDetail.value.name,
          signature,
          size: String(configDetail.value.byte_size),
        };
      } else {
        const configContent = await downloadTemplateContent(props.bkBizId, template_space_id, signature);
        content.value = String(configContent);
      }
    } catch (e) {
      console.error(e);
    } finally {
      detailLoading.value = false;
    }
  };

  const getVariableList = async () => {
    variablesLoading.value = true;
    const res = await getReleasedAppVariables(props.bkBizId, props.appId, props.versionId);
    variables.value = res.details;
    variablesLoading.value = false;
  };

  const handleDownloadFile = async () => {
    const { signature, name } = content.value as IFileConfigContentSummary;
    const getContent = props.type === 'template' ? downloadTemplateContent : downloadConfigContent;
    const res = await getContent(props.bkBizId, props.id, signature);
    fileDownload(res, name);
  };

  const setEditorHeight = () => {
    nextTick(() => {
      const el = sideSliderRef.value.$el.querySelector('.config-loading-container');
      editorHeight.value = el.offsetHeight > 510 ? el.offsetHeight - 310 : 300;
    });
  };

  const close = () => {
    emits('update:show', false);
  };
</script>
<style lang="scss" scoped>
  .config-loading-container {
    height: calc(100vh - 101px);
    overflow: auto;
    .config-form-wrapper {
      padding: 20px 40px;
      height: 100%;
    }
  }
  .view-config-tab {
    :deep(.bk-tab-header) {
      padding: 8px 24px 0;
      background: #eaebf0;
    }
    :deep(.bk-tab-content) {
      padding: 24px 40px;
      box-shadow: none;
    }
    :deep(.bk-form-label) {
      color: #979ba5;
      font-size: 12px;
    }
    :deep(.bk-form-content) {
      color: #313238;
      font-size: 12px;
    }
    .memo {
      line-height: 20px;
      white-space: pre-wrap;
      word-break: break-word;
    }
  }
  .binary-file-card {
    padding: 12px 16px;
    background: #ffffff;
    font-size: 12px;
    border: 1px solid #c4c6cc;
    border-radius: 2px;
    .basic-info {
      display: flex;
      align-items: center;
      justify-content: space-between;
    }
    .file-icon {
      margin-right: 17px;
      font-size: 28px;
      color: #63656e;
    }
    .content {
      flex: 1;
      .name {
        color: #63656e;
        line-height: 20px;
        cursor: pointer;
        &:hover {
          color: #3a84ff;
        }
      }
      .time {
        margin-top: 2px;
        color: #979ba5;
        line-height: 16px;
      }
    }
    .size {
      color: #63656e;
      font-weight: 700;
    }
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
