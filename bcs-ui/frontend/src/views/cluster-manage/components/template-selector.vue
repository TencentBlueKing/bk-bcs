<template>
  <div
    class="item-node-template"
    v-bk-tooltips="{
      disabled: isTkeCluster,
      content: $t('非TKE集群不支持节点模板')
    }">
    <bcs-select
      searchable
      :clearable="false"
      placeholder=" "
      :disabled="!isTkeCluster || disabled"
      :loading="loading"
      v-model="nodeTemplateID"
      @change="handleNodeTemplateIDChange">
      <bcs-option id="" :name="$t('不使用节点模板')"></bcs-option>
      <bcs-option
        v-for="item in templateList"
        :key="item.nodeTemplateID"
        :id="item.nodeTemplateID"
        :name="item.name">
      </bcs-option>
      <template #extension>
        <span style="cursor: pointer" @click="handleGotoNodeTemplate">
          <i class="bcs-icon bcs-icon-fenxiang mr5 !text-[12px]"></i>
          {{$t('节点模板配置')}}
        </span>
      </template>
    </bcs-select>
    <template v-if="isTkeCluster">
      <span
        class="ml10 text-[12px] cursor-pointer"
        v-bk-tooltips.top="$t('刷新列表')"
        @click="handleGetNodeTemplateList">
        <i class="bcs-icon bcs-icon-reset"></i>
      </span>
      <span class="text-[12px] cursor-pointer ml15" v-if="nodeTemplateID">
        <i
          class="bcs-icon bcs-icon-yulan"
          v-bk-tooltips.top="$t('预览')"
          @click="handleShowPreview"></i>
      </span>
    </template>
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
import { defineComponent, computed, ref, onMounted, watch, toRefs } from 'vue';
import NodeTemplateDetail from '@/views/cluster-manage/node-template/node-template-detail.vue';
import $store from '@/store/index';
import $router from '@/router';
import { NODE_TEMPLATE_ID } from '@/common/constant';
import { useConfig } from '@/composables/use-app';

export default defineComponent({
  name: 'TemplateSelector',
  components: { NodeTemplateDetail },
  model: {
    prop: 'value',
    event: 'change',
  },
  props: {
    value: {
      type: String,
      default: '',
    },
    isTkeCluster: {
      type: Boolean,
      default: false,
    },
    disabled: {
      type: Boolean,
      default: false,
    },
  },
  setup(props, ctx) {
    const { isTkeCluster } = toRefs(props);
    const loading = ref(false);
    const nodeTemplateID = ref(props.value || localStorage.getItem(NODE_TEMPLATE_ID) || '');
    const templateList = ref<any[]>([]);
    const handleGetNodeTemplateList = async () => {
      loading.value = true;
      templateList.value = await $store.dispatch('clustermanager/nodeTemplateList');
      if (!isTkeCluster.value
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

    const { _INTERNAL_ } = useConfig();
    onMounted(() => {
      _INTERNAL_.value && handleGetNodeTemplateList();
      if (nodeTemplateID.value !== props.value) {
        ctx.emit('change', nodeTemplateID.value);
      }
    });
    return {
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
<style lang="postcss" scoped>
.item-node-template {
  display: flex;
  max-width: 524px;
  .bk-select {
    width: 400px;
  }
  .icon:hover {
    color: #3a84ff;
    cursor: pointer;
  }
}
</style>
