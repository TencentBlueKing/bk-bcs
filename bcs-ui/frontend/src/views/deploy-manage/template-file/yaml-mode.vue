<template>
  <div class="overflow-hidden" ref="contentRef">
    <bcs-resize-layout
      v-show="!renderMode || (isEdit && renderMode !== 'Helm') || (!isEdit && renderMode !== 'Helm' && !upgrade)"
      collapsible
      disabled
      :border="false"
      ref="yamlLayoutRef"
      initial-divide="230px"
      class="h-full"
      @collapse-change="handleCollapseChange">
      <div
        slot="aside"
        class="bg-[#fff] h-full overflow-y-auto overflow-x-hidden">
        <left-nav
          :list="yamlToJson"
          :active-index="activeContentIndex"
          @cellClick="({ item }) => handleAnchor(item)" />
      </div>
      <div
        slot="main"
        :class="[
          'flex flex-col',
          'shadow-[0_2px_4px_0_rgba(0,0,0,0.16)]',
          'bg-[#2E2E2E] h-full rounded-t-sm',
          isCollapse ? '' : 'ml-[16px]'
        ]">
        <yaml-content
          ref="yamlContentRefInside"
          :is-edit="isEdit"
          :value="content"
          :version="version"
          :render-mode="renderMode"
          :upgrade="upgrade"
          :content-origin="contentOrigin"
          @setContentOrigin="(val) => contentOrigin = val"
          @updateUpgrade="(val) => upgrade = val"
          @change="handleChange" />
      </div>
    </bcs-resize-layout>
    <yaml-content
      v-show="renderMode === 'Helm' || upgrade"
      ref="yamlContentRef"
      :is-edit="isEdit"
      :value="content"
      :version="version"
      :render-mode="renderMode"
      :upgrade="upgrade"
      :content-origin="contentOrigin"
      @setContentOrigin="(val) => contentOrigin = val"
      @updateUpgrade="(val) => upgrade = val"
      @change="handleChange" />
  </div>
</template>
<script setup lang="ts">
import yamljs from 'js-yaml';
import { computed, ref, watch } from 'vue';

import leftNav from './left-nav.vue';
import yamlContent from './yaml-content.vue';

const props = defineProps({
  isEdit: {
    type: Boolean,
    default: false,
  },
  value: {
    type: String,
    default: '',
  },
  version: {
    type: String,
    default: '',
  },
  renderMode: {
    type: String,
    default: '',
  },
});

const emits = defineEmits(['getUpgradeStatus', 'change']);

const upgrade  = ref(false);
const content = ref('');
const contentOrigin = ref('');
const activeContentIndex = ref(0);
const yamlToJson = ref();

const yamlLayoutRef = ref();
const watchOnce = watch(yamlToJson, () => {
  // 只有一项数据时折叠起来
  if (yamlToJson.value && yamlToJson.value.length < 2) {
    yamlLayoutRef.value?.setCollapse(true);
    yamlLayoutRef.value && (yamlLayoutRef.value.$refs.aside.style.transition = '');
  }
  watchOnce();
});

// 使用v-show，v-if时resize-layout某些值会被初始化，导致只有一个值时也展开，间隙也会消失
const yamlRef = computed(() => {
  if (!props.renderMode
    || (props.isEdit && props.renderMode !== 'Helm')
    || (!props.isEdit && props.renderMode !== 'Helm' && !upgrade.value)) {
    return yamlContentRefInside.value;
  };
  return yamlContentRef.value;
});
const yamlContentRef = ref();
const yamlContentRefInside = ref();
// 跳转到对应的yaml
const handleAnchor = (item: typeof yamlToJson.value[number]) => {
  const index = yamlToJson.value.findIndex(d => d === item);
  yamlRef.value?.setPosition(item.offset);
  activeContentIndex.value = index;
};

// 获取数据
const getData = () => yamlRef.value?.getData();

// 校验数据
const validate = async () => yamlRef.value?.validate();

const isCollapse = ref(false);
const handleCollapseChange = (value: boolean) => {
  isCollapse.value = value;
};

function handleChange(content) {
  emits('change', content);
}

watch([
  () => upgrade.value,
  () => props.renderMode,
], () => {
  emits('getUpgradeStatus', {
    isHelm: props.renderMode === 'Helm' || upgrade.value,
    upgrade: upgrade.value,
  });
});

watch(() => props.value, () => {
  // if (!props.value) return;
  content.value = props.value;
  yamlRef.value?.setValue(props.value, '');
}, { immediate: true });

// 使用watch，如果使用computed，props.value还未来得及赋值给content.value，会出现把Helm当作yaml解析的情况
watch([content, () => props.version], () => {
  // 初始化数据
  yamlToJson.value = [];
  activeContentIndex.value = 0;
  if (!props.renderMode || props.renderMode === 'Helm' || upgrade.value) return [];
  let offset = 0;
  yamlToJson.value =  yamljs.loadAll(content.value)
    .reduce<Array<{ name: string; offset: number }>>((pre, doc) => {
    const name = doc?.metadata?.name;
    if (name) {
      pre.push({
        name,
        offset,
      });
      offset += yamljs.dump(doc).length;
    }
    return pre;
  }, []);
}, { immediate: true });

watch(() => props.version, () => {
  // 触发语法校验
  content.value = '';
  // 使用setTimeout才会重新触发校验
  setTimeout(() => {
    content.value = props.value;
  });
});

watch(yamlRef, (newVal, oldVal) => {
  // 同步两个编辑器的值
  oldVal && yamlContentRef.value?.setValue(oldVal.getData());
  oldVal && yamlContentRefInside.value?.setValue(oldVal.getData());
});

defineExpose({
  getData,
  validate,
});
</script>
<style scoped lang="postcss">
/deep/ .dark-form {
  .bk-label {
    color: #B3B3B3;
    font-size: 12px;
  }
  .bk-form-input {
    background-color: #2E2E2E;
    border: 1px solid #575757;
    color: #B3B3B3;
    &:focus {
      background-color: unset !important;
    }
  }
  .bk-select {
    border: 1px solid #575757;
    color: #B3B3B3;
  }
}

/deep/ .file-editor .bk-resize-layout-aside {
  border-color: #292929;
}
</style>
