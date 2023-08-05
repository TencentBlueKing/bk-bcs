<script lang="ts" setup>
  import { ref, watch } from 'vue';
  import { storeToRefs } from 'pinia'
  import { Plus, Search, Ellipsis } from 'bkui-vue/lib/icon'
  import { useGlobalStore } from '../../../../../store/global'
  import { useTemplateStore } from '../../../../../store/template'
  import { getTemplatePackageList } from '../../../../../api/template'
  import { ITemplatePackageItem } from '../../../../../../types/template'

  const { spaceId } = storeToRefs(useGlobalStore())
  const templateStore = useTemplateStore()
  const { currentTemplateSpace } = storeToRefs(templateStore)

  // const props = defineProps<{
  //   templateSpaceId: number;
  //   templateId: number;
  // }>()

  const loading = ref(false)
  const list = ref<ITemplatePackageItem[]>([])

  watch([() => spaceId.value, () => currentTemplateSpace.value], ([newSpaceId, newTemplateSpace]) => {
    console.log(newSpaceId, newTemplateSpace)
    getList()
  })

  const getList = async () => {
    loading.value = true
    const params = {
      start: 0,
      limit: 100
    }
    const res = await getTemplatePackageList(spaceId.value, currentTemplateSpace.value, params)
    list.value = res.details
    templateStore.$patch((state) => {
      state.packageList = res.details
      if (res.details.length > 0) {
        state.currentPkg = res.details[0].id
      }
    })
    loading.value = false
  }

</script>
<template>
  <div class="package-list-comp">
    <div class="search-wrapper">
      <div class="create-btn" v-bk-tooltips="'新建模板套餐'">
        <Plus />
      </div>
      <div class="search-input">
        <bk-input placeholder="搜索模板套餐">
          <template #suffix>
            <Search class="search-icon" />
          </template>
        </bk-input>
      </div>
    </div>
    <ul class="package-list">
      <li v-for="pkg in list" class="package-item active" :key="pkg.id">
        <div class="pkg-wrapper">
          <div class="mark-icon">
            <i class="bk-bscp-icon icon-folder"></i>
          </div>
          <div class="text">
            <span class="name">{{ pkg.spec.name }}</span>
            <span class="num">{{ pkg.spec.template_ids.length }}</span>
          </div>
          <Ellipsis class="action-more-icon" />
        </div>
      </li>
    </ul>
    <ul class="other-package-list">
      <li class="package-item">
        <div class="pkg-wrapper">
          <div class="mark-icon">
            <i class="bk-bscp-icon icon-app-store"></i>
          </div>
          <div class="text">
            <span class="name">全部配置项</span>
            <span class="num">30</span>
          </div>
          <Ellipsis class="action-more-icon" />
        </div>
      </li>
      <li class="package-item">
        <div class="pkg-wrapper">
          <div class="mark-icon">
            <i class="bk-bscp-icon icon-empty"></i>
          </div>
          <div class="text">
            <span class="name">未指定套餐</span>
            <span class="num">30</span>
          </div>
          <Ellipsis class="action-more-icon" />
        </div>
      </li>
    </ul>
  </div>
</template>
<style lang="scss" scoped>
  .package-list-comp {
    padding-top: 12px;
    height: calc(100% - 58px);
    .search-wrapper {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 0 16px;
    }
    .create-btn {
      display: inline-flex;
      align-items: center;
      justify-content: center;
      margin-right: 8px;
      width: 32px;
      height: 32px;
      font-size: 24px;
      color: #c4c6cc;
      border: 1px solid #c4c6cc;
      border-radius: 2px;
      cursor: pointer;
      &:hover {
        color: #3a84ff;
        border-color: #3a84ff;
      }
    }
    .search-input {
      width: calc(100% - 40px);
    }
    .search-icon {
      margin-right: 10px;
      color: #979ba5;
    }
    .package-list {
      padding-top: 16px;
      height: calc(100% - 104px);
    }
    .other-package-list {
      padding-top: 8px;
      border-top: 1px solid #dcdee5;
    }
    .package-item {
      padding: 8px 16px;
      cursor: pointer;
      &.active {
        background: #e1ecff;
        .mark-icon {
          color: #3a84ff;
        }
        .text {
          .name {
            color: #3a84ff;
          }
          .num {
            background: #a3c5fd;
            color: #ffffff;
          }
        }
      }
      .pkg-wrapper {
        display: flex;
        align-items: center;
      }
      .mark-icon {
        display: flex;
        align-items: center;
        width: 12px;
        height: 12px;
        color: #c4c6cc;
        font-size: 12px;
        .icon-folder {
          transform-origin: 0 50%;
          transform: scale(0.8);
        }
        .icon-empty {
          transform-origin: 0 50%;
          transform: scale(0.7);
        }
      }
      .text {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 0 4px;
        width: calc(100% - 30px);
        .name {
          font-size: 12px;
          color: #63656e;
        }
        .num {
          padding: 0 8px;
          color: #979ba5;
          height: 16px;
          line-height: 16px;
          font-size: 12px;
          background: #f0f1f5;
          border-radius: 2px;
        }
      }
      .action-more-icon {
        display: flex;
        align-items: center;
        justify-content: center;
        transform: rotate(90deg);
        width: 16px;
        height: 16px;
        color: #979ba5;
        border-radius: 50%;
        cursor: pointer;
        &:hover {
          background: rgba(99, 101, 110, 0.1);
          color: #3a84ff;
        }
      }
    }
  }
</style>
