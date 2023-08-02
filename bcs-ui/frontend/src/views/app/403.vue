<template>
  <bk-exception :type="type">
    <span>{{$t('iam.title.perms')}}</span>
    <bk-table :data="tableData" class="mt25" v-bkloading="{ isLoading }">
      <bk-table-column :label="$t('iam.label.system')" prop="system" min-width="150">
        {{ $t('bcs.name') }}
      </bk-table-column>
      <bk-table-column :label="$t('iam.label.action')" prop="auth" min-width="220">
        <template #default="{ row }">
          {{ actionsMap[row.action_id] || '--' }}
        </template>
      </bk-table-column>
      <bk-table-column :label="$t('iam.label.resource')" prop="resource" min-width="220">
        <template #default="{ row }">
          {{ row.resource_name || '--' }}
        </template>
      </bk-table-column>
    </bk-table>
    <bk-button
      theme="primary"
      class="mt25"
      :disabled="!href"
      @click="handleGotoIAM"
    >{{$t('iam.button.apply')}}</bk-button>
  </bk-exception>
</template>
<script lang="ts">
import { PropType, defineComponent, onBeforeMount, ref } from 'vue';
import { userPermsByAction } from '@/api/base';
import actionsMap from '@/views/app/actions-map';

export default defineComponent({
  name: 'AuthForbidden',
  props: {
    type: {
      type: String,
      default: '403',
    },
    actionId: {
      type: String,
      default: '',
    },
    resourceName: {
      type: String,
      default: '',
    },
    permCtx: {
      type: [Object, String],
      default: () => ({}),
    },
    // 接口返回的权限数据
    perms: {
      type: [Object, String] as PropType<{
        action_list: Array<{
          action_id: string
          resource_type: string
        }>
        apply_url: string
      }>,
      default: () => null,
    },
    fromRoute: {
      type: String,
      default: '',
    },
  },
  setup(props) {
    const tableData = ref<any[]>([]);
    const href = ref('');
    const isLoading = ref(false);
    const handleGotoIAM = () => {
      window.open(href.value);
    };

    // 设置权限信息
    const handleSetPermsData = (url, data) => {
      href.value = url;
      tableData.value = data;
    };

    const handleGetPermsData = async () => {
      if (!props.actionId) return;
      isLoading.value = true;
      const data = await userPermsByAction({
        $actionId: [props.actionId],
        perm_ctx: typeof props.permCtx === 'string'
          ? JSON.parse(props.permCtx)
          : props.permCtx,
      }).catch(() => ({}));
      isLoading.value = false;
      if (data?.perms?.[props.actionId] && props.fromRoute) {
        // 有权限跳回原来界面
        window.location.href = props.fromRoute;
      } else {
        // 无权限
        handleSetPermsData(data?.perms?.apply_url, [{
          resource_name: props.resourceName,
          action_id: props.actionId,
        }]);
      }
    };

    onBeforeMount(async () => {
      if (props.perms) {
        // 已经返回权限信息
        const { apply_url, action_list } = typeof props.perms === 'string'
          ? JSON.parse(props.perms)
          : props.perms;
        handleSetPermsData(apply_url, action_list.map(item => ({
          ...item,
          resource_name: props.resourceName,
        })));
      } else {
        // 查询权限信息
        handleGetPermsData();
      }
    });
    return {
      handleGotoIAM,
      actionsMap,
      tableData,
      isLoading,
      href,
    };
  },
});
</script>
