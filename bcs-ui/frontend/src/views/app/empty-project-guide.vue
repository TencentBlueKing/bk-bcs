<template>
  <bk-exception type="403">
    <span>{{$t('无项目权限')}}</span>
    <div class="text-subtitle">{{$t('你没有相应项目的访问权限，请前往申请相关项目权限')}}</div>
    <div style="display: flex;align-items: center; justify-content: center;">
      <a class="bk-text-button text-wrap" @click="handleGotoIAM">{{$t('去申请')}}</a>
      <a
        class="bk-text-button text-wrap"
        v-authority="{
          actionId: 'project_create',
          permCtx: {}
        }"
        @click="handleGotoProjectManage">{{$t('创建项目')}}</a>
    </div>
  </bk-exception>
</template>
<script lang="ts">
import { defineComponent, onMounted, ref } from '@vue/composition-api';
import { projectViewPerms } from '@/api/base';

export default defineComponent({
  name: 'ProjectGuide',
  setup(props, ctx) {
    const { $router, $route } = ctx.root;
    const iamUrl = ref<string>(window.BK_IAM_APP_URL);
    function handleGotoIAM() {
      window.open(iamUrl.value);
    }
    function handleGotoProjectManage() {
      if (window.REGION === 'ieod') {
        window.open(`${window.DEVOPS_HOST}/console/pm`);
      } else {
        if ($route.name === 'projectManage') return;
        $router.push({
          name: 'projectManage',
        });
      }
    }
    onMounted(async () => {
      const data = await projectViewPerms();
      // eslint-disable-next-line camelcase
      iamUrl.value = data.perms?.apply_url;
    });
    return {
      handleGotoIAM,
      handleGotoProjectManage,
    };
  },
});
</script>
<style lang="postcss" scoped>
>>> .text-subtitle {
  color: #979BA5;
  font-size: 14px;
  text-align: center;
  margin-top: 14px;
}
>>> .text-wrap {
  display: flex;
  align-items: center;
  justify-content: center;
  color: #3A84FF;
  font-size: 14px;
  margin-top: 12px;
}
</style>
