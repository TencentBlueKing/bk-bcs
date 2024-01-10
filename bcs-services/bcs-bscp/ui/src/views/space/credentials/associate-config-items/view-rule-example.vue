<template>
  <div class="rule-title">文件型配置，以选择服务myservice为例:</div>
  <div style="margin-bottom: 16px;">
    <div v-for="file in rule.file" class="rule-wrap" :key="file.info">
    <div class="example">{{ file.example }}</div>
    <div class="info">{{ file.info }}</div>
  </div>
  </div>
  <div class="rule-title">键值（KV）型配置，以选择服务myservice为例:</div>
  <div v-for="kv in rule.kv" class="rule-wrap" :key="kv.info">
    <div class="example">{{ kv.example }}</div>
    <div class="info">{{ kv.info }}</div>
  </div>
</template>

<script lang="ts" setup>
const rule = {
  file: [
    {
      example: '**',
      info: '关联myservice服务下所有的配置(包含子目录)',
    },
    {
      example: 'etc/*',
      info: '关联myservice服务/etc目录下所有的配置(不含子目录)',
    },
    {
      example: 'etc/nginx/nginx.conf',
      info: '关联myservice服务/etc/nginx/nginx.conf文件',
    },
  ],
  kv: [
    {
      example: '*',
      info: '关联myservice服务下所有配置项',
    },
    {
      example: 'demo_*',
      info: '关联myservice服务下所有以demo_开头的配置项',
    },
  ],
};
</script>

<style scoped lang="scss">
.rule-title {
  height: 20px;
  margin-bottom: 6px;
  font-size: 12px;
  color: #979ba5;
  line-height: 20px;
}
.rule-wrap {
  width: 325px;
  height: 56px;
  padding: 8px 16px;
  background: #f5f7fa;
  border-radius: 2px;
  &:not(:last-child) {
    margin-bottom: 8px;
  }
  .example {
    height: 20px;
    font-size: 12px;
    color: #313238;
    line-height: 20px;
  }
  .info {
    height: 20px;
    font-size: 12px;
    color: #979ba5;
    line-height: 20px;
  }
}
</style>
