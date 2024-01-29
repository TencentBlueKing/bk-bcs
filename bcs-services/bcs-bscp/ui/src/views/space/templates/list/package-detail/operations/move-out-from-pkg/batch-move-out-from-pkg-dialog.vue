<template>
  <bk-dialog
    ext-cls="move-out-configs-dialog"
    :title="t('批量移出当前套餐')"
    :confirm-text="t('确定移出')"
    :cancel-text="t('取消')"
    :width="480"
    :is-show="props.show"
    :esc-close="false"
    :quick-close="false"
    :is-loading="pending"
    @confirm="handleConfirm"
    @closed="close"
  >
    <div class="selected-mark">
      {{t('已选')}} <span class="num">{{ props.value.length }}</span> {{t('个配置文件')}}
    </div>
    <p class="tips">{{t('以下服务配置的未命名版本中引用此套餐的内容也将更新')}}</p>
    <div class="service-table">
      <bk-loading style="min-height: 100px" :loading="loading">
        <bk-table :data="citedList" :max-height="maxTableHeight">
          <bk-table-column :label="t('所在模板套餐')" prop="template_set_name"></bk-table-column>
          <bk-table-column :label="t('使用此套餐的服务')">
            <template #default="{ row }">
              <div v-if="row.app_id" class="app-info" @click="goToConfigPageImport(row.app_id)">
                <div v-overflow-title class="name-text">{{ row.app_name }}</div>
                <LinkToApp class="link-icon" :id="row.app_id" />
              </div>
            </template>
          </bk-table-column>
        </bk-table>
      </bk-loading>
    </div>
  </bk-dialog>
</template>
<script lang="ts" setup>
import { ref, computed, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import { storeToRefs } from 'pinia';
import Message from 'bkui-vue/lib/message';
import useGlobalStore from '../../../../../../../store/global';
import useTemplateStore from '../../../../../../../store/template';
import { ITemplateConfigItem, IPackagesCitedByApps } from '../../../../../../../../types/template';
import { moveOutTemplateFromPackage, getUnNamedVersionAppsBoundByPackages } from '../../../../../../../api/template';
import LinkToApp from '../../../components/link-to-app.vue';

const { spaceId } = storeToRefs(useGlobalStore());
const { packageList, currentTemplateSpace, currentPkg } = storeToRefs(useTemplateStore());
const { t } = useI18n();

const props = defineProps<{
  show: boolean;
  currentPkg: number;
  value: ITemplateConfigItem[];
}>();

const emits = defineEmits(['update:show', 'movedOut']);

const router = useRouter();

const loading = ref(false);
const citedList = ref<IPackagesCitedByApps[]>([]);
const pending = ref(false);

const maxTableHeight = computed(() => {
  const windowHeight = window.innerHeight;
  return windowHeight * 0.6 - 200;
});

watch(
  () => props.show,
  () => {
    getCitedData();
  },
);

const goToConfigPageImport = (id: number) => {
  const { href } = router.resolve({
    name: 'service-config',
    params: { appId: id },
    query: { pkg_id: currentTemplateSpace.value },
  });
  window.open(href, '_blank');
};

const getCitedData = async () => {
  loading.value = true;
  const params = {
    start: 0,
    all: true,
  };
  const res = await getUnNamedVersionAppsBoundByPackages(
    spaceId.value,
    currentTemplateSpace.value,
    [props.currentPkg],
    params,
  );
  citedList.value = res.details;
  loading.value = false;
};

const handleConfirm = async () => {
  const pkg = packageList.value.find(item => item.id === currentPkg.value);
  if (!pkg) return;

  try {
    pending.value = true;
    const ids = props.value.map(item => item.id);
    await moveOutTemplateFromPackage(spaceId.value, currentTemplateSpace.value, ids, [currentPkg.value as number]);
    emits('movedOut');
    close();
    Message({
      theme: 'success',
      message: t('配置文件移出套餐成功'),
    });
  } catch (e) {
    console.log(e);
  } finally {
    pending.value = false;
  }
};

const close = () => {
  emits('update:show', false);
};
</script>
<style lang="scss" scoped>
.selected-mark {
  display: inline-block;
  margin-bottom: 16px;
  padding: 0 12px;
  height: 32px;
  line-height: 32px;
  border-radius: 16px;
  font-size: 12px;
  color: #63656e;
  background: #f0f1f5;
  .num {
    color: #3a84ff;
  }
}
.app-info {
  display: flex;
  align-items: center;
  overflow: hidden;
  cursor: pointer;
  .name-text {
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
  }
  .link-icon {
    flex-shrink: 0;
    margin-left: 10px;
  }
}
</style>
