<script setup lang="ts">
  import { ref, watch, onMounted } from 'vue'
  import { Ellipsis } from 'bkui-vue/lib/icon'
  import { getConfigVersionList } from '../../../../api/config'

  interface IVersionItem {
    id: number;
    attachment: object;
    revision: object;
    spec: {
      name: string
    };
  }

  const props = defineProps<{
    bkBizId: string,
    appId: number
  }>()

  const versionListLoading = ref(false)
  const versionList = ref<IVersionItem[]>([])

  watch(() => props.appId, () => {
    getVersionList()
  })

  onMounted(() => {
    getVersionList()
  })

  const getVersionList = async() => {
    try {
      versionListLoading.value = true
      const res = await getConfigVersionList(props.bkBizId, props.appId)
      console.log(res)
      versionList.value = res.data.details
    } catch (e) {
      console.error(e)
    } finally {
      versionListLoading.value = false
    }
  }
</script>
<template>
  <section class="version-container">
    <section v-if="versionList.length > 0" class="version-steps">
      <section v-for="(version, index) in versionList" class="version-item" :key="version.id">
        <div class="dot-line">
          <div :class="['dot', { first: index === 0, last: index === versionList.length - 1 }]"></div>
        </div>
        <div class="version-name">{{ version.spec.name }}</div>
        <bk-dropdown class="action-area">
          <Ellipsis class="action-more-icon" />
          <template #content>
            <bk-dropdown-menu placement="bottom-end">
              <bk-dropdown-item>版本对比</bk-dropdown-item>
              <bk-dropdown-item>废弃</bk-dropdown-item>
            </bk-dropdown-menu>
          </template>
        </bk-dropdown>
      </section>
    </section>
    <bk-exception
      v-else
      style="margin-top: 100px;"
      description="没有版本数据"
      type="empty"
      scene="part" />
  </section>
</template>

<style lang="scss" scoped>
  .version-steps {
    padding: 16px 0;
    overflow: auto;
  }
  .version-item {
    position: relative;
    padding: 0 40px 0 48px;
    cursor: pointer;
    &.active {
      background: #e1ecff;
    }
    &:hover {
      background: #e1ecff;
    }
    .dot-line {
      position: absolute;
      left: 28px;
      top: 0;
      display: flex;
      align-items: center;
      height: 100%;
      .dot {
        width: 8px;
        height: 8px;
        border-radius: 50%;
        border: 1px solid #c4c6cc;
        background: #f0f1f5;
        &:not(.first):before {
          position: absolute;
          top: 0;
          left: 4px;
          content: '';
          width: 1px;
          height: 16px;
          background: #dcdee5;
        }
        &:not(.last):after {
          position: absolute;
          bottom: 0;
          left: 4px;
          content: '';
          width: 1px;
          height: 16px;
          background: #dcdee5;
        }
      }
    }
  }
  .version-name {
    height: 40px;
    line-height: 40px;
    font-size: 12px;
    color: #313238;
    text-align: left;
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
  }
  .action-area {
    position: absolute;
    right: 13px;
    top: 9px;
    line-height: 22px;
    .action-more-icon {
      display: flex;
      align-items: center;
      justify-content: center;
      transform: rotate(90deg);
      width: 22px;
      height: 22px;
      color: #979ba5;
      border-radius: 50%;
      cursor: pointer;
      &:hover {
        background: rgba(99, 101, 110, 0.1);
        color: #3a84ff;
      }
    }
  }
</style>