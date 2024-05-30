<template>
  <div class="scripts-menu">
    <MenuList
      v-if="!loading"
      :title="t('前/后置脚本')"
      :value="selected"
      :list="scriptDetailList"
      @selected="selectScript" />
  </div>
</template>
<script lang="ts" setup>
  import { ref, onMounted, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { useRoute } from 'vue-router';
  import { storeToRefs } from 'pinia';
  import useServiceStore from '../../../../../../../../store/service';
  import { getConfigScript } from '../../../../../../../../api/config';
  import { getDiffType } from '../../../../../../../../utils/index';
  import MenuList from './menu-list.vue';

  const { t } = useI18n();
  const route = useRoute();
  const bkBizId = ref(String(route.params.spaceId));
  const { appData } = storeToRefs(useServiceStore());

  const props = defineProps<{
    currentVersionId: number;
    baseVersionId: number;
    actived: boolean;
  }>();

  const emits = defineEmits(['selected']);

  const scriptDetailList = ref([
    {
      id: 'pre',
      name: t('前置脚本'),
      type: '',
      current: {
        language: '',
        content: '',
      },
      base: {
        language: '',
        content: '',
      },
    },
    {
      id: 'post',
      name: t('后置脚本'),
      type: '',
      current: {
        language: '',
        content: '',
      },
      base: {
        language: '',
        content: '',
      },
    },
  ]);
  const selected = ref();
  const loading = ref(true);

  watch(
    () => props.baseVersionId,
    () => {
      initData();
      if (typeof selected.value === 'string') {
        selectScript(selected.value);
      }
    },
  );

  watch(
    () => props.actived,
    (val) => {
      if (!val) {
        selected.value = undefined;
      }
    },
    {
      immediate: true,
    },
  );

  onMounted(async () => {
    await getScriptDetail(props.currentVersionId, 'current');
    // 选择基准版本后才计算变更状态
    if (props.baseVersionId) {
      await getScriptDetail(props.baseVersionId, 'base');
      updateDiff();
    }
    initData(true);
  });

  const initData = async (needGetCrt = false) => {
    loading.value = true;
    if (needGetCrt) {
      await getScriptDetail(props.currentVersionId, 'current');
    }
    // 选择基准版本后才计算变更状态
    if (props.baseVersionId) {
      await getScriptDetail(props.baseVersionId, 'base');
      updateDiff();
    }
    loading.value = false;
  };

  const getScriptDetail = async (id: number, type: 'current' | 'base') => {
    const scriptSetting = await getConfigScript(bkBizId.value, appData.value.id as number, id);
    const { pre_hook, post_hook } = scriptSetting;
    scriptDetailList.value[0][type] = {
      language: pre_hook.type,
      content: pre_hook.content,
    };
    scriptDetailList.value[1][type] = {
      language: post_hook.type,
      content: post_hook.content,
    };
  };

  // 计算前置脚本或后置脚本差异
  const updateDiff = async () => {
    scriptDetailList.value[0].type = getDiffType(
      scriptDetailList.value[0].base.content,
      scriptDetailList.value[0].current.content,
    );
    scriptDetailList.value[1].type = getDiffType(
      scriptDetailList.value[1].base.content,
      scriptDetailList.value[1].current.content,
    );
  };

  const selectScript = (id: string) => {
    const script = id === 'pre' ? scriptDetailList.value[0] : scriptDetailList.value[1];
    const { base, current } = script;
    const diffData = { id, contentType: 'text', base, current };
    selected.value = id;
    emits('selected', diffData);
  };
</script>
