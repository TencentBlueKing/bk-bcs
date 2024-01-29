<template>
  <div
    class="flex items-center"
    v-bk-tooltips="{
      disabled: supportNodeTemplate,
      content: $t('cluster.nodeTemplate.tips.tkeClusterCanNotUse')
    }">
    <bcs-select
      class="flex-1"
      searchable
      :clearable="false"
      placeholder=" "
      :disabled="!supportNodeTemplate || disabled"
      :loading="loading"
      v-model="nodeTemplateID"
      @change="handleNodeTemplateIDChange">
      <bcs-option id="" :name="$t('cluster.nodeTemplate.msg.notUseTemplate')"></bcs-option>
      <bcs-option
        v-for="item in templateList"
        :key="item.nodeTemplateID"
        :id="item.nodeTemplateID"
        :name="item.name">
      </bcs-option>
      <SelectExtension
        slot="extension"
        :link-text="$t('cluster.nodeTemplate.title.templateConfig')"
        @link="handleGotoNodeTemplate"
        @refresh="handleGetNodeTemplateList" />
    </bcs-select>
    <span class="text-[12px] cursor-pointer ml10" v-if="nodeTemplateID && supportNodeTemplate && showPreview">
      <i
        class="bcs-icon bcs-icon-yulan"
        v-bk-tooltips.top="$t('generic.title.preview')"
        @click="handleShowPreview"></i>
    </span>
    <!-- 节点模板详情 -->
    <bcs-sideslider
      :is-show.sync="showDetail"
      :title="currentTemplate.name"
      quick-close
      :width="800">
      <div slot="content">
        <NodeTemplateDetail :data="currentTemplate"></NodeTemplateDetail>
      </div>
    </bcs-sideslider>
  </div>
</template>
<script lang="ts">
import { computed, defineComponent, onMounted, ref, watch } from 'vue';

import { NODE_TEMPLATE_ID } from '@/common/constant';
import $router from '@/router';
import $store from '@/store/index';
import SelectExtension from '@/views/cluster-manage/add/common/select-extension.vue';
import NodeTemplateDetail from '@/views/cluster-manage/node-template/node-template-detail.vue';

export default defineComponent({
  name: 'TemplateSelector',
  components: { NodeTemplateDetail, SelectExtension },
  model: {
    prop: 'value',
    event: 'change',
  },
  props: {
    value: {
      type: String,
      default: '',
    },
    provider: {
      type: String,
      default: '',
    },
    disabled: {
      type: Boolean,
      default: false,
    },
    showPreview: {
      type: Boolean,
      default: true,
    },
  },
  setup(props, ctx) {
    const supportNodeTemplate = computed(() => ['tencentCloud', 'tencentPublicCloud'].includes(props.provider));
    const loading = ref(false);
    const nodeTemplateID = ref(props.value || localStorage.getItem(NODE_TEMPLATE_ID) || '');
    const templateList = ref<any[]>([]);
    const handleGetNodeTemplateList = async () => {
      loading.value = true;
      templateList.value = await $store.dispatch('clustermanager/nodeTemplateList');
      if (!supportNodeTemplate.value
        || !templateList.value.find(item => item.nodeTemplateID === nodeTemplateID.value)
      ) {
        nodeTemplateID.value = '';
        ctx.emit('change', nodeTemplateID.value);
      }
      loading.value = false;
    };
    const handleGotoNodeTemplate = () => {
      const location = $router.resolve({ name: 'nodeTemplate' });
      window.open(location.href);
    };
    const showDetail = ref(false);
    const currentTemplate = computed(() => templateList.value
      .find(item => item.nodeTemplateID === nodeTemplateID.value) || {});

    watch(currentTemplate, () => {
      ctx.emit('template-change', currentTemplate.value);
    }, { deep: true, immediate: true });

    const handleShowPreview = () => {
      showDetail.value = true;
    };
    // todo 框架支持数据持久化
    const handleNodeTemplateIDChange = (value) => {
      localStorage.setItem(NODE_TEMPLATE_ID, value);
      ctx.emit('change', value);
    };

    onMounted(() => {
      handleGetNodeTemplateList();
      if (nodeTemplateID.value !== props.value) {
        ctx.emit('change', nodeTemplateID.value);
      }
    });
    return {
      supportNodeTemplate,
      loading,
      nodeTemplateID,
      currentTemplate,
      templateList,
      handleNodeTemplateIDChange,
      handleGetNodeTemplateList,
      handleGotoNodeTemplate,
      showDetail,
      handleShowPreview,
    };
  },
});
</script>
