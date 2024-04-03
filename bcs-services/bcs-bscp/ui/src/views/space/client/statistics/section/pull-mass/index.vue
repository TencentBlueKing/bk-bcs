<template>
  <div>
    <SectionTitle :title="'客户端配置拉取质量'" />
    <PullCount :bk-biz-id="bkBizId" :app-id="appId" />
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
