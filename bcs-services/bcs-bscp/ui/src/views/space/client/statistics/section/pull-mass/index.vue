<template>
  <div>
    <SectionTitle :title="$t('客户端配置拉取质量')" />
    <PullCount :bk-biz-id="bkBizId" :app-id="appId" :title="$t('客户端拉取配置文件次数统计')" :is-duplicates="false" />
    <PullCount :bk-biz-id="bkBizId" :app-id="appId" :title="$t('已拉取配置文件的客户端数量统计')" :is-duplicates="true" />
    <section class="pull-wrap">
      <PullSuccess :bk-biz-id="bkBizId" :app-id="appId" />
      <div class="pull-error">
        <PullError :bk-biz-id="bkBizId" :app-id="appId" />
      </div>
    </section>
  </div>
</template>

<script lang="ts" setup>
  import SectionTitle from '../../components/section-title.vue';
  import PullCount from './pull-count.vue';
  import PullError from './pull-error.vue';
  import PullSuccess from './pull-success.vue';

  defineProps<{
    bkBizId: string;
    appId: number;
  }>();
</script>

<style scoped lang="scss">
  .pull-wrap {
    display: flex;
    justify-content: space-between;
    .pull-error {
      flex-grow: 1;
    }
  }
</style>
