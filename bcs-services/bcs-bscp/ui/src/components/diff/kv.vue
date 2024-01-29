<template>
  <section class="kv-diff">
    <div class="left-version-content">
      <div class="content-box">
        <bk-overflow-title type="tips">{{ base }}</bk-overflow-title>
      </div>
    </div>
    <div class="right-version-content">
      <div class="content-box">
        <bk-overflow-title type="tips">{{ current }}</bk-overflow-title>
      </div>
    </div>
  </section>
</template>

<script lang="ts" setup>
defineProps<{
  current: string;
  base: string;
}>();
</script>

<style scoped lang="scss">
.kv-diff {
  display: flex;
  align-items: center;
  height: 100%;
  background: #f5f7fa;
}
.left-version-content,
.right-version-content {
  padding: 15px;
  width: 50%;
  height: 100%;
}
.content-box {
  width: 435px;
  height: 52px;
  padding-left: 15px;
  line-height: 52px;
  background: #ffffff;
  border: 1px solid #c4c6cc;
  border-radius: 2px;
  font-size: 12px;
}
</style>
