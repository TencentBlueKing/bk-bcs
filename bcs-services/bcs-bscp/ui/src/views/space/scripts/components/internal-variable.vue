<template>
  <div class="var-wrap">
    <div class="title">{{ t('内置变量') }}</div>
    <div v-for="(item, index) in varList" :key="index" class="var-item">
      <div class="var-content">
        <div class="cn-name">
          {{ item.cnName }} <InfoLine v-bk-tooltips="{ content: item.tips }" class="info-icon" />
        </div>
        <bk-overflow-title type="tips" :key="language">
          {{ language === 'shell' ? item.shellVar : item.pythonVar }}
        </bk-overflow-title>
      </div>
      <div class="copy-btn"><Copy @click="handleCopyVar(language === 'shell' ? item.shellVar : item.pythonVar)" /></div>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { InfoLine, Copy } from 'bkui-vue/lib/icon';
  import { copyToClipBoard } from '../../../../utils';
  import { Message } from 'bkui-vue';
  import { useI18n } from 'vue-i18n';

  const { t } = useI18n();

  defineProps<{
    language: string;
  }>();

  const varList = [
    {
      cnName: t('配置根目录'),
      shellVar: '${bk_bscp_temp_dir}',
      pythonVar: 'os.environn.get( \'bk_bscp_temp_dir\' )',
      tips: t('客户端配置的配置存放临时目录（temp_dir），默认值为 /data/bscp'),
    },
    {
      cnName: t('业务ID'),
      shellVar: '${bk_bscp_biz}',
      pythonVar: 'os.environn.get( \'bk_bscp_biz\' )',
      tips: t('蓝鲸配置平台上的业务ID，例如：2'),
    },
    {
      cnName: t('服务名称'),
      shellVar: '${bk_bscp_app}',
      pythonVar: 'os.environn.get( \'bk_bscp_app\' )',
      tips: t('服务配置中心上的服务名称，例如：demo_service'),
    },
    {
      cnName: t('服务配置目录'),
      shellVar: '${bk_bscp_app_temp_dir}',
      pythonVar: 'os.environn.get( \'bk_bscp_app_temp_dir\' )',
      tips: t(
        '单个客户端可使用多个服务的配置，为保证路径唯一，服务配置需存放于：配置根目录/业务ID/服务名称，服务配置存放目录 = 配置存放根目录/业务ID/服务名称',
      ),
    },
  ];

  const handleCopyVar = (text: string) => {
    copyToClipBoard(text);
    Message({
      theme: 'success',
      message: t('变量名已复制'),
    });
  };
</script>

<style scoped lang="scss">
  .var-wrap {
    padding: 8px 16px;
    width: 272px;
    height: 100%;
    line-height: 20px;
    background: #2e2e2e;
    border-top: 1px solid #181818;
    .title {
      margin-bottom: 12px;
      font-weight: 700;
      font-size: 14px;
      color: #979ba5;
    }
    .var-item {
      display: flex;
      justify-content: space-between;
      padding: 8px 16px;
      width: 240px;
      color: #8a8f99;
      font-size: 12px;
      border-top: 1px solid #000;

      &:last-child {
        border-bottom: 1px solid #000;
      }
      .var-content {
        width: 180px;
        .cn-name {
          display: flex;
          align-items: center;
          .info-icon {
            font-size: 14px;
            margin-left: 8px;
          }
        }
      }
      .copy-btn {
        display: none;
        font-size: 14px;
        line-height: 40px;
        span {
          cursor: pointer;
        }
      }
      &:hover {
        background: #292929;
        .copy-btn {
          display: block;
        }
      }
    }
  }
</style>
