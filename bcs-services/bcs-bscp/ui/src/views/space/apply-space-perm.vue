<script setup lang="ts">
  import { storeToRefs } from 'pinia';
  import { useGlobalStore } from '../../store/global';

  const { applyPermUrl } = storeToRefs(useGlobalStore())

  const handleApply = () => {
    window.open(applyPermUrl.value, '__blank')
  }

</script>
<template>
    <div class="apply-space-perm">
      <bk-exception type="403" scene="part">
        <div class="apply-description">
          <h4 class="title">无业务权限</h4>
          <p class="tips">你没有相应业务的访问权限，请前往申请相关业务权限</p>
          <bk-button text theme="primary" @click="handleApply">申请权限</bk-button>
        </div>
      </bk-exception>
    </div>
</template>
<style lang="scss" scoped>
  .apply-space-perm {
    padding-top: 100px;
    height: 100%;
    font-size: 12px;
  }
  .title {
    margin: 0 0 8px;
    font-size: 14px;
    color: #63656e;
    line-height: 22px;
  }
  .tips {
    margin: 0 0 8px;
    font-size: 12px;
    color: #979ba5;
  }
</style>