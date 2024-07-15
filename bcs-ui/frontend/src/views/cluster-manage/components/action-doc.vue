<template>
  <div>
    <div class="title">
      {{title}}
    </div>
    <div class="content-wrapper">
      <div class="content">
        <div class="content-item">
          <div class="label">{{$t('cluster.nodeTemplate.variable.title')}}</div>
          <bcs-table class="mt15" :data="configList">
            <bcs-table-column :label="$t('cluster.nodeTemplate.variable.label.var.text')">
              <template #default="{ row }">
                <span
                  v-bk-tooltips.top="{
                    content: $t('cluster.nodeTemplate.variable.label.var.tips', { name: row.refer })
                  }"
                  @click="handleCopyVar(row)">
                  {{row.refer}}
                </span>
              </template>
            </bcs-table-column>
            <bcs-table-column
              :label="$t('cluster.nodeTemplate.variable.label.meaning')"
              prop="desc"
              show-overflow-tooltip
            ></bcs-table-column>
          </bcs-table>
        </div>
        <BcsMd class="mt15" :code="type === 'default' ? postActionDescMd : autoscalerScriptsMd"></BcsMd>
      </div>
    </div>
  </div>
</template>
<script lang="ts">
import { computed, defineComponent, onMounted, ref } from 'vue';

import autoscalerScriptsMd from '../autoscaler/autoscaler-scripts.md';
import autoscalerScriptsMdEn from '../autoscaler/autoscaler-scripts-en.md';
import postActionDescMd from '../node-template/postaction-desc.md';
import postActionDescMdEn from '../node-template/postaction-desc-en.md';

import $bkMessage from '@/common/bkmagic';
import { copyText } from '@/common/util';
import BcsMd from '@/components/bcs-md/index.vue';
import $i18n from '@/i18n/i18n-setup';
import $store from '@/store/index';

export default defineComponent({
  name: 'ActionDoc',
  components: { BcsMd },
  props: {
    title: {
      type: String,
      default: '',
    },
    type: {
      type: String,
      default: 'default',
    },
  },
  setup() {
    // 配置说明
    const configLoading = ref(false);
    const configList = ref([]);
    const isEn = computed(() => $store.state.isEn);
    const handleGetConfigList = async () => {
      configLoading.value = true;
      configList.value = await $store.dispatch('clustermanager/bkSopsTemplatevalues');
      configLoading.value = false;
    };
    const handleCopyVar = (row) => {
      copyText(row.refer);
      $bkMessage({
        theme: 'success',
        message: $i18n.t('generic.msg.success.copy'),
      });
    };

    onMounted(() => {
      handleGetConfigList();
    });

    return {
      postActionDescMd: isEn.value ? postActionDescMdEn : postActionDescMd,
      autoscalerScriptsMd: isEn.value ? autoscalerScriptsMdEn : autoscalerScriptsMd,
      configList,
      handleCopyVar,
      handleGetConfigList,
    };
  },
});
</script>
<style lang="postcss" scoped>
.title {
  height: 52px;
  padding: 0 16px;
  font-size: 16px;
  color: #313238;
  display: flex;
  align-items: center;
  box-shadow: inset 0 -1px 0 0 #DCDEE5;
}
.content-wrapper {
  max-height: calc(100% - 52px);
  overflow: auto;
}
.content {
  padding: 16px 0;
  .content-item {
      padding: 0 24px;
      .label {
          font-weight: 600;
          line-height: 1.25;
          font-size: 1em;
          color: #24292e;
      }
  }
  >>> .bcs-md-preview {
      overflow: hidden;
  }
}
</style>
