<template>
  <div class="example-wrap">
    <div class="header">
      <span>{{ $t('格式示例') }}</span>
      <copy-shape
        class="icon"
        v-bk-tooltips="{
          content: $t('复制示例内容'),
          placement: 'top',
          extCls: 'copy-example-content',
        }"
        @click="handleCopyText" />
    </div>
    <div class="content">
      <div class="format">
        <div v-if="format === 'text'">
          <div>{{ $t('文本格式') }}:</div>
          <div>key {{ $t('数据类型') }} value {{ $t('描述') }}</div>
        </div>
        <div v-else-if="format === 'json'">
          <div>JSON {{ $t('格式') }}:</div>
          <div>{{ `{“key”: {“kv_type”: ${$t('数据类型')}, “value”: ${$t('配置项值')} \}\}` }}</div>
        </div>
        <div v-else>
          <div>YAML {{ $t('格式') }}:</div>
          <div>key {{ $t('数据类型') }} value {{ $t('描述') }}</div>
        </div>
      </div>
      <div class="example">
        <div>{{ $t('示例') }}:</div>
        <div v-if="format === 'text'">
          <div class="data">
            <span class="key">string_key</span>
            <span class="type">string</span>
            <span class="value">strign_value</span>
          </div>
          <div class="data">
            <span class="key">number_key</span>
            <span class="type">number</span>
            <span class="value">100</span>
          </div>
        </div>
        <bk-input v-else v-model="copyContent" type="textarea" :read-only="true" :resize="false" />
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { computed } from 'vue';
  import { CopyShape } from 'bkui-vue/lib/icon';
  import { copyToClipBoard } from '../../../../../../../../../utils';
  import { Message } from 'bkui-vue';
  import { useI18n } from 'vue-i18n';

  const { t } = useI18n();
  const props = defineProps<{
    format: string;
  }>();

  const copyContent = computed(() => {
    if (props.format === 'text') {
      return `string_key string strign_value
number_key number 100`;
    }
    if (props.format === 'json') {
      return `{
    "string_key": {"kv_type": "string", "value": "string_value"},
    "number_key": {"kv_type": "number", "value": 100},
    "text_key": {"kv_type": "text", "value": "line1\\nline2"},
    "json_key": {"kv_type": "json", "value": "{'name': 'bk', 'age': 18}"},
    "xml_key": {"kv_type": "xml", "value": "<xml>\\n xml_value\\n</xml>"},
    "yaml_key": {"kv_type": "yaml", "value": "def:\\n name:bk\\n age:18"}
}`;
    }
    return `string_key:
    kv_type: string
    value: string_value
number_key:
    kv_type: number
    value: 100
text_key:
    kv_type: text
    value: |-
       line1
       line2
json_key:
    kv_type: json
    value: |-
       {
          “name”: “bk”
          “age”: 18
       }
xml_key:
    kv_type: xml
    value: |-
       <xml>\n xml_value\n</xml>`;
  });

  // 复制
  const handleCopyText = () => {
    copyToClipBoard(copyContent.value);
    Message({
      theme: 'success',
      message: t('示例内容已复制'),
    });
  };
</script>

<style scoped lang="scss">
  .example-wrap {
    width: 520px;
    background: #2e2e2e;
    padding: 0 16px;
    border-top: 1px solid #000;
    .header {
      display: flex;
      align-items: center;
      gap: 16px;
      padding: 8px 0 12px 0;
      font-weight: 700;
      font-size: 14px;
      color: #979ba5;
      border-bottom: 1px solid #000;
      .icon {
        cursor: pointer;
      }
    }
    .content {
      padding-top: 16px;
      color: #c4c6cc;
      font-size: 13px;
      .example {
        margin-top: 13px;
        .type {
          color: #ff9c01;
          margin: 0 8px;
        }
      }
    }
  }
  :deep(.bk-textarea) {
    border: none;
    box-shadow: none !important;
    height: 270px;
    color: #c4c6cc;
    textarea {
      height: 100%;
      background-color: #2e2e2e;
      font-size: 13px;
    }
  }
</style>

<style>
  .copy-example-content {
    background-color: #000000 !important;
  }
</style>
