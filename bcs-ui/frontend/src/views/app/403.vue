<template>
  <bk-exception :type="type">
    <span>{{$t('iam.title.perms')}}</span>
    <bk-table :data="tableData" class="mt25" v-bkloading="{ isLoading }">
      <bk-table-column :label="$t('iam.label.system')" prop="system" min-width="150">
        {{ name }}
      </bk-table-column>
      <bk-table-column :label="$t('iam.label.action')" prop="auth" min-width="220">
        <template #default="{ row }">
          {{ actionsMap[row.action_id] || '--' }}
        </template>
      </bk-table-column>
      <bk-table-column :label="$t('iam.label.resource')" prop="resource" min-width="220">
        <template #default>
          {{ projectCode }}
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
import { defineComponent, onBeforeMount, PropType, ref } from 'vue';

import useProjects from '../project-manage/project/use-project';

import $bkMessage from '@/common/bkmagic';
import $router from '@/router';
import actionsMap from '@/views/app/actions-map';
import usePlatform from '@/composables/use-platform';

interface IPerms {
  action_list: Array<{
    action_id: string
    resource_type: string
  }>
  apply_url: string
}

export default defineComponent({
  name: 'AuthForbidden',
  props: {
    type: {
      type: String,
      default: '403',
    },
    perms: {
      type: Object as PropType<IPerms>,
    },
    fromRoute: {
      type: String,
      default: '',
    },
    projectCode: {
      type: String,
      required: true,
    },
  },
  setup(props) {
    const { fetchProjectInfo } = useProjects();
    const { config } = usePlatform();
    const name = ref(config.i18n.name);
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

    const handleGetProjectPerms = async () => {
      if (props.perms) {
        handleSetPermsData(props.perms.apply_url, props.perms.action_list);
        return;
      };

      isLoading.value = true;
      const { code, web_annotations, message } = await fetchProjectInfo({
        $projectId: props.projectCode,
      });
      isLoading.value = false;

      if (code === 0) {  // 有权限
        const { href } = $router.resolve({
          name: 'clusterMain',
          params: {
            projectCode: props.projectCode,
          },
        });
        window.location.href = href;
      } else if (code === 40403) { // 无权限
        const perms: IPerms = web_annotations?.perms;
        handleSetPermsData(perms.apply_url, perms.action_list);
      } else {
        $bkMessage({
          theme: 'error',
          message,
        });
      }
    };

    onBeforeMount(async () => {
      // 查询权限信息
      handleGetProjectPerms();
    });
    return {
      name,
      handleGotoIAM,
      actionsMap,
      tableData,
      isLoading,
      href,
    };
  },
});
</script>
