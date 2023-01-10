<script setup lang="ts">
</script>
<template>
  <section class="config-list-table">
    <bk-table :border="['outer']">
      <bk-table-column label="配置项名称" :sort="true"></bk-table-column>
      <bk-table-column label="配置预览"></bk-table-column>
      <bk-table-column label="配置格式"></bk-table-column>
      <bk-table-column label="创建人"></bk-table-column>
      <bk-table-column label="修改人"></bk-table-column>
      <bk-table-column label="修改时间" :sort="true"></bk-table-column>
      <bk-table-column label="变更状态"></bk-table-column>
      <bk-table-column label="操作"></bk-table-column>
    </bk-table>
  </section>
</template>