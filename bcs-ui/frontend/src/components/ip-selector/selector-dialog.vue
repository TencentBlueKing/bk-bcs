<template>
  <bcs-dialog
    class="selector-dialog"
    :mask-close="false"
    :close-icon="false"
    :esc-close="false"
    :value="modelValue"
    :width="dialogWidth"
    :auto-close="false"
    @value-change="handleValueChange"
    @confirm="handleConfirm">
    <Selector
      ref="selector"
      :key="selectorKey"
      :height="dialogHeight"
      :ip-list="ipList"
      :disabled-ip-list="disabledIpList"
      :cloud-id="cloudId"
      :region="region"
      :vpc="vpc"
      v-if="modelValue"
      @change="handleIpSelectorChange"
    />
  </bcs-dialog>
</template>
<script lang="ts">
import { defineComponent, ref, toRefs, watch, onMounted, PropType } from 'vue';
import Selector from './ip-selector-bcs.vue';
import $bkMessage from '@/common/bkmagic';
import $i18n from '@/i18n/i18n-setup';

export default defineComponent({
  name: 'SelectorDialog',
  components: {
    Selector,
  },
  model: {
    prop: 'modelValue',
    event: 'change',
  },
  props: {
    modelValue: {
      type: Boolean,
      default: false,
    },
    // 回显IP列表
    ipList: {
      type: Array,
      default: () => ([]),
    },
    disabledIpList: {
      type: Array as PropType<Array<string|{ip: string, tips: string}>>,
      default: () => [],
    },
    cloudId: {
      type: String,
      default: '',
    },
    region: {
      type: String,
      default: '',
    },
    // 集群VPC
    vpc: {
      type: Object,
      default: () => ({}),
    },
  },

  setup(props, ctx) {
    const { emit } = ctx;
    const { modelValue } = toRefs(props);
    const dialogWidth = ref(1200);
    const dialogHeight = ref(600);

    const selectorKey = ref(String(new Date().getTime()));
    watch(modelValue, () => {
      selectorKey.value = String(new Date().getTime());
    });
    const handleValueChange = (value: boolean) => {
      emit('change', value);
    };

    const handleIpSelectorChange = (data) => {
      emit('nodes-change', data);
    };

    const selector = ref<any>();
    const handleConfirm = () => {
      const data = selector.value?.handleGetData() || [];
      if (!data.length) {
        $bkMessage({
          theme: 'error',
          message: $i18n.t('请选择服务器'),
        });
        return;
      }
      emit('confirm', data);
    };

    onMounted(() => {
      dialogWidth.value = document.body.clientWidth < 1650 ? 1200 : document.body.clientWidth - 650;
      dialogHeight.value = document.body.clientHeight < 1000 ? 460 : document.body.clientHeight - 320;
    });

    return {
      selector,
      selectorKey,
      dialogWidth,
      dialogHeight,
      handleValueChange,
      handleConfirm,
      handleIpSelectorChange,
    };
  },
});
</script>
<style lang="postcss" scoped>
.selector-dialog {
    >>> .bk-dialog {
        top: 100px;
    }
    >>> .bk-dialog-tool {
        display: none;
    }
    >>> .bk-dialog-body {
        padding: 0;
    }
    >>> .bk-dialog-footer {
        border-top: none;
    }
}
</style>
