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
            <bk-form-item :label="t('配置文件绝对路径')">{{ fileAP() }}</bk-form-item>
            <bk-form-item :label="t('配置文件描述')">
              <div class="memo">{{ configDetail.memo || configDetail.revision_memo || '--' }}</div>
            </bk-form-item>
            <bk-form-item :label="t('配置文件内容')">
              <bk-loading
                v-if="configDetail.file_type === 'binary'"
                mode="spin"
                theme="primary"
                :opacity="0.6"
                size="mini"
                :title="t('文件下载中，请稍后')"
                :loading="fileDownloading"
                class="file-down-loading">
                <div class="binary-file-card" @click="handleDownloadFile">
                  <div class="basic-info">
                    <TextFill class="file-icon" />
                    <div class="content">
                      <div class="name">{{ configDetail.name }}</div>
                      <div class="time">{{ datetimeFormat(configDetail.update_at || configDetail.create_at) }}</div>
                    </div>
                    <div class="size">{{ byteUnitConverse(Number(configDetail.byte_size)) }}</div>
                  </div>
                </div>
              </bk-loading>
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
      <bk-button theme="primary" @click="emits('openEdit')">{{ t('编辑') }}</bk-button>
      <bk-button @click="close">{{ t('关闭') }}</bk-button>
    </section>
  </bk-sideslider>
</template>
<script setup lang="ts">
  import { nextTick, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { storeToRefs } from 'pinia';
  import { TextFill } from 'bkui-vue/lib/icon';
  import ConfigContentEditor from '../../../../../service/detail/config/components/config-content-editor.vue';
  import { getTemplateConfigMeta, downloadTemplateContent } from '../../../../../../../api/template';
  import { byteUnitConverse, datetimeFormat } from '../../../../../../../utils/index';
  import { fileDownload } from '../../../../../../../utils/file';
  import { IVariableEditParams } from '../../../../../../../../types/variable';
  import { IFileConfigContentSummary } from '../../../../../../../../types/config';
  import useTemplateStore from '../../../../../../../store/template';

  const { currentTemplateSpace } = storeToRefs(useTemplateStore());

  interface IConfigMeta {
    template_space_id?: number;
    template_space_name?: string;
    template_set_id?: number;
    template_set_name?: string;
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

  const { t } = useI18n();

  const props = defineProps<{
    id: number;
    spaceId: string;
    show: Boolean;
  }>();

  const emits = defineEmits(['update:show', 'openEdit']);

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
  const sideSliderRef = ref();
  const editorHeight = ref(0);
  const fileDownloading = ref(false);

  watch(
    () => props.show,
    (val) => {
      if (val) {
        getTemplateDetail();
        content.value = '';
        activeTab.value = 'content';
        variables.value = [];
        fileDownloading.value = false;
      }
    },
  );

  // 配置文件绝对路径
  const fileAP = () => {
    const { path, name } = configDetail.value;
    if (path.endsWith('/')) {
      return `${path}${name}`;
    }
    return `${path}/${name}`;
  };

  // 获取模板配置详情，非文件类型配置文件内容下载内容，文件类型手动点击时再下载
  const getTemplateDetail = async () => {
    try {
      detailLoading.value = true;
      const res = await getTemplateConfigMeta(props.spaceId, props.id);
      configDetail.value = {
        ...res.data.detail,
        create_at: datetimeFormat(res.data.detail.create_at),
      };
      if (configDetail.value.file_type === 'binary') {
        content.value = {
          name: configDetail.value.name,
          signature: configDetail.value.signature,
          size: String(configDetail.value.byte_size),
        };
      } else {
        const configContent = await downloadTemplateContent(
          props.spaceId,
          currentTemplateSpace.value,
          configDetail.value.signature,
        );
        content.value = String(configContent);
      }
    } catch (e) {
      console.error(e);
    } finally {
      detailLoading.value = false;
    }
  };

  const handleDownloadFile = async () => {
    if (fileDownloading.value) return;
    fileDownloading.value = true;
    const { signature, name } = content.value as IFileConfigContentSummary;
    const res = await downloadTemplateContent(props.spaceId, props.id, signature, true);
    fileDownload(res, name);
    fileDownloading.value = false;
  };

  const setEditorHeight = () => {
    nextTick(() => {
      const el = sideSliderRef.value.$el.querySelector('.config-loading-container');
      editorHeight.value = el.offsetHeight > 510 ? el.offsetHeight - 400 : 300;
    });
  };

  const close = () => {
    emits('update:show', false);
  };
</script>
<style lang="scss" scoped>
  .config-loading-container {
    height: calc(100vh - 101px);
    .config-form-wrapper {
      padding: 20px 40px;
      height: 100%;
    }
  }
  .view-config-tab {
    height: 100%;
    :deep(.bk-tab-header) {
      padding: 8px 24px 0;
      background: #eaebf0;
    }
    :deep(.bk-tab-content) {
      padding: 24px 40px;
      height: calc(100% - 48px);
      box-shadow: none;
      overflow: auto;
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

  .file-down-loading {
    :deep(.bk-loading-indicator) {
      align-items: center;
      flex-direction: row;
      .bk-loading-title {
        margin-top: 0px;
        margin-left: 8px;
        color: #979ba5;
        font-size: 12px;
      }
    }
  }
</style>
