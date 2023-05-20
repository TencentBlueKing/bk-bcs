<script setup lang="ts">
  import { ref, onMounted, watch } from 'vue'
  import { useRoute, useRouter } from 'vue-router'
  import { ECategoryType, IAllCategoryGroupItem } from '../../../../../../types/group';
  import { getSpaceGroupList } from '../../../../../api/group';

  const route = useRoute()
  const router = useRouter()

  const props = withDefaults(defineProps<{
    size: string,
    appId: number,
    multiple: boolean,
    value: string|number|number[]
  }>(), {
    size: 'default'
  })
  const emits = defineEmits(['change'])

  const groupList = ref<IAllCategoryGroupItem[]>([])
  const groupListLoading = ref(false)
  const groups = ref<string|number|number[]>(props.multiple ? [] : '')

  watch(() => props.value, (val: string|number|number[]) => {
    groups.value = props.multiple ? [ ...<number[]>val ] : val
  }, { immediate: true })

  onMounted(() => {
    getGroupList()
  })

  // 获取全部分组列表
  const getGroupList = async() => {
    groupListLoading.value = true
    const query = {
      mode: ECategoryType.Custom,
      start: 0,
      limit: 200,
    }
    const res = await getSpaceGroupList(<string>route.params.spaceId)
    groupList.value = res.details
    groupListLoading.value = false
  }

  const handleGoToCreateGroup = () => {
    const { spaceId, appId } = route.params
    const targetRoute = router.resolve({ name: 'service-group', params: { spaceId, appId } })
    window.open(targetRoute.fullPath, '__blank')
  }

</script>
<template>
  <bk-select :value="groups" :size="props.size" :loading="groupListLoading" multiple-mode="tag" :multiple="props.multiple" @change="emits('change', $event)">
    <bk-option-group v-for="category in groupList" collapsible :key="category.group_category_id" :label="category.group_category_name">
      <bk-option v-for="group in category.groups" :key="group.id" :value="group.id" :label="group.name"></bk-option>
    </bk-option-group>
    <div class="selector-extensition" slot="extension">
      <div class="content" @click="handleGoToCreateGroup">
        <i class="bk-bscp-icon icon-add create-group-icon"></i>
        服务管理
      </div>
    </div>
  </bk-select>
</template>
<style lang="scss" scoped>
  .selector-extensition {
    .content {
      height: 40px;
      line-height: 40px;
      text-align: center;
      border-top: 1px solid #dcdee5;
      background: #fafbfd;
      font-size: 12px;
      color: #63656e;
      cursor: pointer;
      &:hover {
        color: #3a84ff;
        .create-group-icon {
          color: #3a84ff;
        }
      }
    }
    .create-group-icon {
      font-size: 14px;
      color: #979ba5;
    }
  }
</style>