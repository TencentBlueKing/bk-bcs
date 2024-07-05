<template>
  <section class="node-mana-container">
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
              <span class="bk-form-content">
                {{ $t('(通常用于按标签进行灰度发布，支持设置多个标签；如果不需要按标签灰度，可不设置)') }}
              </span>
            </div>
            <div class="service-item">
              <span :class="['item-label', { 'item-label--en': locale === 'en' }]"> {{ $t('服务名称') }}： </span>
              <span class="bk-form-content">
                <span class="content-em" @click="copyText(serviceName!)">
                  {{ serviceName }} <copy-shape class="icon-shape" />
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
          <span class="content-em" @click="copyText(bizId)"> {{ bizId }} <copy-shape class="icon-shape" /> </span>
        </bk-form-item>
        <bk-form-item :label="$t('临时目录：')">
          <span class="content-em" @click="copyText('/data/bscp')">/data/bscp <copy-shape class="icon-shape" /></span>
        </bk-form-item>
        <bk-form-item :label="$t('全局标签：')">
          {{ $t('(全局标签与服务标签参数一样，常用于按标签进行灰度发布；不同的是全局标签可供多个服务共用)') }}
        </bk-form-item>
        <bk-form-item :label="$t('服务密钥：')">
          {{ $t('(即客户端密钥，需填写与此服务配置关联过的实际客户端密钥)') }}
        </bk-form-item>
      </bk-form>
    </div>
  </section>
</template>

<script lang="ts" setup>
  import { ref, Ref, inject } from 'vue';
  import { useRoute } from 'vue-router';
  import { Share, CopyShape } from 'bkui-vue/lib/icon';
  import { copyToClipBoard } from '../../../../../../utils/index';
  import BkMessage from 'bkui-vue/lib/message';
  import { useI18n } from 'vue-i18n';

  const { t, locale } = useI18n();
  const route = useRoute();

  const linkUrl = {
    nodeManaUrl: `${(window as any).BK_NODE_HOST}/#/plugin-manager/rule`,
    clientNode: 'https://bk.tencent.com/docs/markdown/ZH/BSCP/1.29/UserGuide/Function/client_configuration.md',
  };

  const bizId = ref(String(route.params.spaceId));
  const feedAddr = ref((window as any).FEED_ADDR);
  const serviceName = inject<Ref<string>>('serviceName');

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
    .top-tip {
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
    margin-top: 17px;
    padding: 24px 0;
    // height: 500px;
    background-color: #f5f7fa;
    .icon-shape {
      font-size: 12px;
      color: #3a84ff;
      visibility: hidden;
    }
  }
  .service-content {
    width: 600px;
    padding: 7px 17px;
    border: 1px solid #dcdee5;
    border-radius: 2px;
    background-color: #fafbfd;
  }
  .service-item {
    display: flex;
    justify-content: flex-start;
    align-items: flex-start;
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
      font-size: 12px;
      color: #979ba5;
      .content-em {
        color: #313238;
        cursor: pointer;
        &:hover {
          // background-color: red;
          .icon-shape {
            visibility: visible;
            vertical-align: middle;
          }
        }
      }
    }
  }
</style>
