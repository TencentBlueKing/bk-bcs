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
import { defineComponent, onBeforeMount, ref } from 'vue';
import { userPermsByAction } from '@/api/modules/user-manager';
import actionsMap from '@/views/app/actions-map';
import $router from '@/router';

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
      if (data?.perms?.[props.actionId]) {
        // 有权限跳回原来界面
        if (props.fromRoute) {
          window.location.href = props.fromRoute;
        } else {
          const { href } = $router.resolve({ name: 'home' });
          window.location.href = href;
        }
      } else {
        // 无权限
        handleSetPermsData(data?.perms?.apply_url, [{
          resource_name: props.resourceName,
          action_id: props.actionId,
        }]);
      }
    };

    onBeforeMount(async () => {
      // 查询权限信息
      handleGetPermsData();
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
