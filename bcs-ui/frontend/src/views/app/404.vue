<template>
  <bk-exception
    class="h-full flex items-center justify-center my-px"
    type="404">
    <span>{{$t('页面不存在')}}</span>
  </bk-exception>
</template>
<script>
export default {
  name: 'NotFound',
};
</script>
