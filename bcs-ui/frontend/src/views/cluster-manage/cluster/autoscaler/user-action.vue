<template>
  <div v-bkloading="{ isLoading: bkSopsLoading }">
    <template v-if="!bkSopsTemplateID">
      <template v-if="script">
        <p class="text-[12px] text-[#313238] mb-[10px]">{{$t('cluster.ca.nodePool.detail.title.bashContent')}}</p>
        <pre class="px-[16px] py-[8px] bg-[#F4F4F7] rounded-sm text-[12px]">{{script}}</pre>
      </template>
      <bcs-exception type="empty" scene="part" v-else> </bcs-exception>
    </template>
    <template v-else>
      <p class="text-[12px] text-[#313238] mb-[10px]">{{$t('cluster.ca.nodePool.detail.title.sopsName')}}</p>
      <div class="text-[12px] mb-[12px]">{{bkSopsTemplateName}}</div>
      <p class="text-[12px] text-[#313238] mb-[10px]">{{$t('cluster.ca.nodePool.detail.title.sopsParams')}}</p>
      <bcs-table :data="sopsParamsData">
        <bcs-table-column :label="$t('cluster.ca.nodePool.detail.label.params')" prop="key"></bcs-table-column>
        <bcs-table-column :label="$t('generic.label.value')" prop="value">
          <template #default="{ row }">
            {{row.value || '--'}}
          </template>
        </bcs-table-column>
      </bcs-table>
    </template>
  </div>
</template>
<script lang="ts">
import { computed, defineComponent, ref, toRefs, watch } from 'vue';

import $store from '@/store/index';

export default defineComponent({
  name: 'UserAction',
  props: {
    script: {
      type: String,
      default: '',
    },
    addons: {
      type: Object,
      default: () => null,
    },
    actionsKey: {
      type: String,
      default: '',
      // validator(value: string) {
      //   return [
      //     'preActions',
      //     'postActions',
      //   ].indexOf(value) > -1;
      // },
    },
  },
  setup(props) {
    const { addons, actionsKey } = toRefs(props);
    const curProject = computed(() => $store.state.curProject);
    const user = computed(() => $store.state.user);

    const bkSopsLoading = ref(false);
    const sopsParamsData = ref<any>([]);
    const bkSopsTemplateID = ref('');
    const bkSopsList = ref<any[]>([]);
    const bkSopsTemplateName = computed(() => bkSopsList.value
      ?.find(item => item.templateID === bkSopsTemplateID.value)?.templateName);
    const handleGetbkSopsList = async () => {
      bkSopsLoading.value = true;
      bkSopsList.value = await $store.dispatch('clustermanager/bkSopsList', {
        $businessID: curProject.value.cc_app_id,
        operator: user.value.username,
        templateSource: 'business',
        scope: 'cmdb_biz',
      });
      bkSopsLoading.value = false;
    };

    watch(addons, () => {
      handleSetSopsData();
    }, { immediate: true });

    function handleSetSopsData() {
      if (!addons.value) return;
      // eslint-disable-next-line camelcase, max-len
      bkSopsTemplateID.value = addons.value?.plugins?.[addons.value?.[actionsKey.value]?.[0]]?.params?.template_id;
      const obj = JSON.parse(JSON.stringify(
        addons.value?.plugins?.[bkSopsTemplateID.value]?.params || {},
        (key, value) => {
          if (['template_biz_id', 'template_id', 'template_user'].includes(key)) {
            return undefined;
          }
          return value;
        },
      ));
      sopsParamsData.value = Object.keys(obj).map(key => ({ key, value: obj[key] }));
      bkSopsTemplateID.value && handleGetbkSopsList();
    };

    return {
      bkSopsLoading,
      bkSopsTemplateID,
      bkSopsTemplateName,
      sopsParamsData,
      bkSopsList,
    };
  },
});
</script>
