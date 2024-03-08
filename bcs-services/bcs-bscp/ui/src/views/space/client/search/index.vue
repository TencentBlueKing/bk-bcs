<template>
  <section class="client-search-page">
    <div class="header">
      <ClientHeader title="客户端查询" />
    </div>
    <div class="content">
      <bk-button style="margin-bottom: 16px">批量重试</bk-button>
      <bk-table :data="tableData" :border="['outer', 'row']" :pagination="pagination">
        <bk-table-column type="selection" :min-width="40" :width="40"></bk-table-column>
        <bk-table-column label="UID" :width="254" prop="attachment.uid"></bk-table-column>
        <bk-table-column label="IP" :width="120" prop="spec.ip"></bk-table-column>
        <bk-table-column
          label="客户端标签"
          :width="296"
          :explain="{ content:'', head: tableTips.clientTag, }"></bk-table-column>
        <bk-table-column label="当前配置版本" :width="140"></bk-table-column>
        <bk-table-column label="最近一次拉取配置状态" :width="168"></bk-table-column>
        <bk-table-column label="附加信息" :width="244" :explain="{ content: tableTips.information }"></bk-table-column>
        <bk-table-column label="在线状态" :width="94" :explain="{ content: tableTips.status }"></bk-table-column>
        <bk-table-column label="首次连接时间" :width="154" prop="spec.first_connect_time"></bk-table-column>
        <bk-table-column label="最后心跳时间" :width="154" prop="spec.last_heartbeat_time"></bk-table-column>
        <bk-table-column label="CPU资源占用(当前/最大)" :width="174"></bk-table-column>
        <bk-table-column label="内容资源占用(当前/最大)" :width="170"></bk-table-column>
        <bk-table-column label="客户端组件类型" :width="128" prop="spec.client_type"></bk-table-column>
        <bk-table-column label="客户端组件版本" :width="128" prop="spec.client_version"></bk-table-column>
        <bk-table-column label="操作" :width="106" fixed="right"></bk-table-column>
      </bk-table>
    </div>
  </section>
</template>

<script lang="ts" setup>
  import { ref } from 'vue';
  import ClientHeader from '../components/client-header.vue';

  const pagination = ref({
    count: 200,
    current: 1,
    limit: 10,
  });

  const tableTips = {
    clientTag: '客户端标签与服务分组配合使用实现服务配置灰度发布场景',
    information: '主要用于记录客户端非标识性元数据，例如客户端用途等附加信息（标识性元数据使用客户端标签）',
    status:
      '客户端每 15 秒会向服务端发送一次心跳数据，如果服务端连续3个周期没有接收到客户端心跳数据，视客户端为离线状态',
  };

  const tableData = [
    {
      id: 1526,
      spec: {
        client_version: 'v1.0.0',
        ip: '172.23.19.151',
        labels: '{}',
        annotations: '{}',
        first_connect_time: '2024-02-21T02:32:59.926833Z',
        last_heartbeat_time: '2024-02-21T03:41:34.485218Z',
        online_status: 'offline',
        resource: {
          cpu_usage: 0,
          cpu_max_usage: 1.9995992403202343,
          memory_usage: '23810048',
          memory_max_usage: '32169984',
        },
        current_release_id: 0,
        target_release_id: 22,
        release_change_status: 'Failed',
        release_change_failed_reason: 'DownloadFailed',
        failed_detail_reason:
          'open metadata.json failed, err: open /data/bscp/2/test-demo/metadata.json: no such file or directory',
        client_type: 'sdk',
      },
      attachment: {
        uid: '70574c0cbb8d48139dc71943d7326fc1',
        biz_id: 2,
        app_id: 1,
      },
      message_type: '',
    },
  ];
</script>

<style scoped lang="scss">
  .header {
    height: 120px;
    padding: 40px 120px 0 40px;
    background: #eff5ff;
  }
  .content {
    padding: 24px;
  }
</style>
