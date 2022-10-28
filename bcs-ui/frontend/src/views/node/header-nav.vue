<template>
  <div class="header-nav">
    <bcs-breadcrumb separator-class="bk-icon icon-angle-right">
      <template slot="prefix">
        <bcs-icon class="back-icon" type="arrows-left" size="24px" @click="goBack"></bcs-icon>
      </template>
      <bcs-breadcrumb-item
        v-for="(item, index) in list"
        :key="index"
        :to="item.link">
        <span class="bcs-ellipsis" :title="item.title">{{item.title}}</span>
      </bcs-breadcrumb-item>
    </bcs-breadcrumb>
    <slot></slot>
  </div>
</template>
<script lang="ts">
import { defineComponent } from '@vue/composition-api';
import $router from '@/router/index';

export default defineComponent({
  name: 'HeaderNav',
  props: {
    list: {
      type: Array,
      default: () => [],
    },
  },
  setup() {
    const goBack = () => {
      $router.back();
    };
    return {
      goBack,
    };
  },
});
</script>

<style lang="postcss" scoped>
.header-nav {
    display: flex;
    align-items: center;
    justify-content: space-between;
    height: 60px;
    border-bottom: 1px solid #dde4eb;
    background: #fff;
    box-shadow: 4px 0 3px rgb(0 0 0 / 10%);
    padding: 0 24px 0 20px;
    font-size: 16px;
    .back-icon {
        cursor: pointer;
        font-weight: 700;
        color: #3a84ff;
    }
    >>> .bk-breadcrumb-goback {
        margin-right: 0;
    }
    >>> .bk-breadcrumb-item {
        font-size: 16px;
    }
    >>> .bk-breadcrumb-item {
        display: flex;
        align-items: center;
        justify-content: center;
    }
}
</style>
