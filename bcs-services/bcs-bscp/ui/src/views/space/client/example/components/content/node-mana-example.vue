<template>
  <section class="node-mana-container">
    <form-option ref="fileOptionRef" label-name="服务标签" @update-option-data="getOptionData" />
    <div class="node-content">
      <span class="node-label">{{ $t('示例预览') }}</span>
      <div class="top-tip">
        {{ $t('节点管理插件客户端需要在') }}
        <span class="em" @click="linkTo(linkUrl.nodeManaUrl)">
          {{ $t('节点管理平台') }}<share class="tip-icon-share" />
        </span>
        {{ $t('部署“bkbscp (bscp服务配置分发和热更新)”插件，部署详情请参考产品白皮书：') }}
        <span class="em" @click="linkTo(linkUrl.clientNode)">
          {{ $t('《客户端配置》-“节点管理插件客户端拉取配置”章节') }}<share class="tip-icon-share" />
        </span>
      </div>
      <div class="preview-content">
        <bk-form label-width="145">
          <bk-form-item :label="$t('服务：')">
            <div class="service-content">
              <div class="service-item">
                <span :class="['item-label', { 'item-label--en': locale === 'en' }]"> {{ $t('标签') }}： </span>
                <ul class="bk-form-content" v-if="optionData.labelArr.length">
                  <li class="label-li" v-for="(item, index) in optionData.labelArr" :key="index">
                    <div class="label-content">
                      <span class="label-key">key</span>
                      <div class="label-value">{{ item.key }}</div>
                      <copy-shape class="icon-shape" v-show="item.key" @click="copyText(item.key as string)" />
                    </div>
                    <div class="label-content">
                      <span class="label-key">value</span>
                      <div class="label-value">{{ item.value }}</div>
                      <copy-shape class="icon-shape" v-show="item.value" @click="copyText(item.value as string)" />
                    </div>
                  </li>
                </ul>
                <div v-else class="bk-form-content">--</div>
              </div>
              <div class="service-item">
                <span :class="['item-label', { 'item-label--en': locale === 'en' }]"> {{ $t('服务名称') }}： </span>
                <span class="bk-form-content">
                  <span
                    class="content-em"
                    v-if="basicInfo!.serviceName.value"
                    @click="copyText(basicInfo!.serviceName.value)">
                    {{ basicInfo!.serviceName.value }} <copy-shape class="icon-shape" />
                  </span>
                </span>
              </div>
            </div>
          </bk-form-item>
          <bk-form-item label="feedAddr：">
            <span class="content-em" @click="copyText(feedAddr!)">
              {{ feedAddr }} <copy-shape class="icon-shape" />
            </span>
          </bk-form-item>
          <bk-form-item :label="$t('业务ID：')">
            <span class="content-em" v-if="bizId" @click="copyText(bizId)">
              {{ bizId }} <copy-shape class="icon-shape" />
            </span>
          </bk-form-item>
          <bk-form-item :label="$t('临时目录：')">
            <span class="content-em" v-show="optionData.tempDir" @click="copyText(optionData.tempDir)">
              {{ optionData.tempDir }} <copy-shape class="icon-shape" />
            </span>
          </bk-form-item>
          <bk-form-item :label="$t('全局标签：')">
            <span class="">
              {{ $t('(全局标签与服务标签参数一样，常用于按标签进行灰度发布；不同的是全局标签可供多个服务共用)') }}
            </span>
          </bk-form-item>
          <bk-form-item :label="`${$t('客户端密钥')}:`">
            <span class="content-em" v-if="optionData.clientKey" @click="copyText(optionData.clientKey)">
              {{ optionData.privacyCredential }} <copy-shape class="icon-shape" />
            </span>
          </bk-form-item>
        </bk-form>
      </div>
    </div>
  </section>
</template>

<script lang="ts" setup>
  import { ref, Ref, inject } from 'vue';
  import { useRoute } from 'vue-router';
  import { Share, CopyShape } from 'bkui-vue/lib/icon';
  import { copyToClipBoard } from '../../../../../../utils/index';
  import BkMessage from 'bkui-vue/lib/message';
  import FormOption from '../form-option.vue';
  import { useI18n } from 'vue-i18n';
  // import { cloneDeep } from 'lodash';

  interface labelItem {
    key: String;
    value: String;
  }

  const { t, locale } = useI18n();
  const route = useRoute();
  const basicInfo = inject<{ serviceName: Ref<string>; serviceType: Ref<string> }>('basicInfo');

  const linkUrl = {
    nodeManaUrl: `${(window as any).BK_NODE_HOST}/#/plugin-manager/rule`,
    clientNode: (window as any).CLIENT_CONFIGURATION_DOC,
  };

  const keyValidateReg = new RegExp(
    '^[a-z0-9A-Z]([-_a-z0-9A-Z]*[a-z0-9A-Z])?((\\.|\\/)[a-z0-9A-Z]([-_a-z0-9A-Z]*[a-z0-9A-Z])?)*$',
  );
  const valueValidateReg = new RegExp(/^(?:-?\d+(\.\d+)?|[A-Za-z0-9]([-A-Za-z0-9_.]*[A-Za-z0-9])?)$/);
  const sysDirectories: string[] = ['/bin', '/boot', '/dev', '/lib', '/lib64', '/proc', '/run', '/sbin', '/sys'];

  const fileOptionRef = ref();
  const bizId = ref(String(route.params.spaceId));
  const feedAddr = ref((window as any).GRPC_ADDR);
  // fileOption组件传递过来的数据汇总
  const optionData = ref({
    clientKey: '',
    privacyCredential: '',
    labelArr: [] as labelItem[],
    tempDir: '',
  });

  const getOptionData = async (data: any) => {
    let labelArr = [];
    let tempDir = data.tempDir;
    // 标签展示方式加工
    if (data.labelArr.length) {
      labelArr = data.labelArr.map((item: string) => {
        // 转换字符串
        let [key, value] = item.replace(/"/g, '').split(':');
        // 与其他模板同样校验
        key = keyValidateReg.test(key) ? key : '';
        value = valueValidateReg.test(value) ? value : '';
        return { key, value };
      });
    }
    // 临时目录展示方式加工
    if (tempDir) {
      if (sysDirectories.some((dir) => tempDir === dir || tempDir.startsWith(`${dir}/`))) {
        tempDir = '';
      }
      if (!tempDir.startsWith('/') || tempDir.endsWith('/')) {
        tempDir = '';
      }
      const parts = tempDir.split('/').slice(1);
      parts.some((part: string) => {
        if (part.startsWith('.') || !/^[\u4e00-\u9fa5A-Za-z0-9.\-_#%,@^+=\\[\]{}]+$/.test(part)) {
          tempDir = '';
          return true;
        }
        return false;
      });
    }
    optionData.value = {
      ...data,
      tempDir,
      labelArr,
    };
  };

  const linkTo = (url: string) => {
    window.open(url, '__blank');
  };
  const copyText = (copyVal: string) => {
    copyToClipBoard(copyVal);
    BkMessage({
      theme: 'success',
      message: t('复制成功'),
    });
  };
</script>

<style scoped lang="scss">
  .node-mana-container {
    .node-content {
      margin-top: 24px;
      padding-top: 12px;
      border-top: 1px solid #dcdee5;
    }
    .node-label {
      font-weight: 700;
      font-size: 14px;
      letter-spacing: 0;
      line-height: 22px;
      color: #63656e;
    }
    .top-tip {
      margin-top: 8px;
      font-size: 12px;
      color: #63656e;
    }
    .em {
      color: #3a84ff;
      cursor: pointer;
    }
    .tip-icon-share {
      margin: -1px 4px 0;
      vertical-align: middle;
    }
  }
  .preview-content {
    margin-top: 13px;
    padding: 24px 0;
    // height: 500px;
    background-color: #f5f7fa;
    .icon-shape {
      font-size: 12px;
      color: #3a84ff;
      visibility: hidden;
    }
    :deep(.bk-form) {
      .bk-form-item {
        margin: 0;
        &:first-child {
          margin-bottom: 8px;
        }
      }
      .bk-form-label {
        padding-right: 0;
        font-size: 12px;
      }
      .bk-form-content {
        padding-left: 17px;
        font-size: 12px;
        color: #979ba5;
        .content-em {
          color: #313238;
          cursor: pointer;
          &:hover {
            .icon-shape {
              visibility: visible;
              vertical-align: middle;
            }
          }
        }
      }
    }
  }
  .label-li {
    & + .label-li {
      margin-top: 24px;
    }
    .label-content {
      display: flex;
      justify-content: flex-start;
      align-items: center;
      color: #63656e;
      & + .label-content {
        margin-top: 8px;
      }
      &:hover {
        .icon-shape {
          visibility: visible;
          vertical-align: middle;
        }
      }
      .icon-shape {
        margin-left: 10px;
        cursor: pointer;
      }
    }
    .label-key {
      margin-right: 16px;
      width: 31px;
      height: 30px;
      text-align: right;
    }
    .label-value {
      padding: 0 8px;
      width: 240px;
      height: 30px;
      line-height: 30px;
      border: 1px solid #dcdee5;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
  }
  .service-content {
    width: 600px;
    padding: 7px 17px;
    border-radius: 2px;
    background-color: #fafbfd;
  }
  .service-item {
    display: flex;
    justify-content: flex-start;
    align-items: flex-start;
    & + .service-item {
      margin-top: 8px;
    }
    .item-label {
      flex-shrink: 0;
      width: 60px;
      font-size: 12px;
      white-space: nowrap;
      text-align: right;
      color: #63656e;
      line-height: 32px;
      &--en {
        width: 91px;
      }
    }
    .item-content {
      font-size: 12px;
    }
  }
</style>
