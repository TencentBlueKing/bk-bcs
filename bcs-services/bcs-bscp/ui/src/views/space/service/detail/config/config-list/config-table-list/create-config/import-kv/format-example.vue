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
      <template v-if="format === 'text'">
        <template v-for="item in textFormat" :key="item.formatTite">
          <div class="formate">
            <div>{{ item.formatTitle }}</div>
            <div>{{ item.formatContent }}</div>
          </div>
          <div class="example">
            <div v-for="exampleList in item.example" :key="exampleList.title">
              <div>{{ exampleList.title }}</div>
              <div v-for="(example, index) in exampleList.list" :key="index" class="text-example">
                <span>{{ example.key }}</span>
                <span class="type">{{ example.type }}</span>
                <span v-if="example.secret_type">{{ example.secret_type }}</span>
                <span>{{ example.value }}</span>
                <span>{{ example.secret_hidden }}</span>
              </div>
            </div>
          </div>
        </template>
      </template>
      <div v-else-if="format === 'json'">
        <div>JSON {{ $t('格式') }}:</div>
        <div>{{ `{“key”: {“kv_type”: ${$t('数据类型')}, “value”: ${$t('配置项值')} \}\}` }}</div>
      </div>
      <div v-else>
        <div>YAML {{ $t('格式') }}:</div>
        <div>key {{ $t('数据类型') }} value {{ $t('描述') }}</div>
      </div>
      <div v-if="format !== 'text'" class="example">
        <div>{{ $t('示例') }}:</div>
        <bk-input v-model="copyContent" type="textarea" :read-only="true" :resize="false" />
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

  const textFormat = [
    {
      formatTitle: t('普通文本格式：'),
      formatContent: t('key 数据类型 value 描述（可选）'),
      example: [
        {
          title: t('示例：'),
          list: [
            {
              key: 'string_key',
              type: 'string',
              secret_type: '',
              value: 'string_value',
              secret_hidden: '',
            },
            {
              key: 'number_key',
              type: 'number',
              secret_type: '',
              value: 100,
              secret_hidden: '',
            },
          ],
        },
      ],
    },
    {
      formatTitle: t('敏感文本格式：'),
      formatContent: t('key 数据类型 凭证类型 value 是否可见 描述（可选）'),
      example: [
        {
          title: t('密码示例：'),
          list: [
            {
              key: 'user_name',
              type: 'secret',
              secret_type: 'password',
              value: 'password_value',
              secret_hidden: 'visible',
            },
          ],
        },
        {
          title: t('API密钥示例：'),
          list: [
            {
              key: 'api_key_name',
              type: 'secret',
              secret_type: 'secret_key',
              value: 'api_key_value',
              secret_hidden: 'invisible',
            },
          ],
        },
        {
          title: t('访问令牌示例：'),
          list: [
            {
              key: 'access_token_name',
              type: 'secret',
              secret_type: 'token',
              value: 'access_token_value',
              secret_hidden: 'invisible',
            },
          ],
        },
      ],
    },
  ];

  /* eslint-disable */
  const copyContent = computed(() => {
    if (props.format === 'text') {
      return `string_key string strign_value
number_key number 100
user_name secret password password_value visible
api_key_name secret secret_key api_key_value invisible
access_token_name secret token access_token_value invisible`;
    }
    if (props.format === 'json') {
      return `{
  "string_demo": {
    "kv_type": "string",
    "value": "blueking"
  },
  "number_demo": {
    "kv_type": "number",
    "value": 12345
  },
  "text_demo": {
    "kv_type": "text",
    "value": "text"
  },
  "json_demo": {
    "kv_type": "json",
    "value": "{\\"name\\": \\"John Doe\\", \\"age\\": 30, \\"city\\": \\"New York\\", \\"hobbies\\": [\\"reading\\", \\"travelling\\", \\"sports\\"]}"
  },
  "xml_demo": {
    "kv_type": "xml",
    "value": "<person>\\n  <name>John Doe</name>\\n  <age>30</age>\\n  <city>New York</city>\\n  <hobbies>\\n    <hobby>reading</hobby>\\n    <hobby>travelling</hobby>\\n    <hobby>sports</hobby>\\n  </hobbies>\\n</person>"
  },
  "yaml_demo": {
    "kv_type": "yaml",
    "value": "name: John Doe\\nage: 30\\ncity: New York\\nhobbies:\\n  - reading\\n  - travelling\\n  - sports"
  }
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
          "name": "bk",
          "aaa": 18
       }
xml_key:
    kv_type: xml
    value: |-
       <xml> xml_value </xml>`;
  });
  /* eslint-enable */

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
      color: #c4c6cc;
      font-size: 13px;
      .formate {
        margin-top: 16px;
      }
      .text-example {
        display: flex;
        gap: 8px;
      }
      .example {
        margin-top: 13px;
        .type {
          color: #ff9c01;
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
