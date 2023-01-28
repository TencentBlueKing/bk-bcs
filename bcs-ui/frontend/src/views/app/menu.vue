<template>
  <ul class="bk-menu">
    <li class="bk-menu-item" v-for="(item, itemIndex) in list" :key="itemIndex">
      <div class="line" v-if="item.type === 'line'"></div>
      <div
        :class="['bk-menu-title-wrapper', item.disable, { selected: selected === item.id }]"
        @click="handleItemClick(item)" v-else>
        <i :class="['bcs-icon left-icon', item.icon]"></i>
        <div class="bk-menu-title">
          <span>
            {{item.name}}
            <bcs-tag theme="danger" v-if="item.new">NEW</bcs-tag>
          </span>
        </div>
        <i
          :class="['bcs-icon right-icon bcs-icon-angle-down',
                   openedMenu.includes(item.id) ? 'selected' : 'bcs-icon-angle-down']"
          v-if="item.children && item.children.length"></i>
      </div>
      <CollapseTransition>
        <ul v-show="openedMenu.includes(item.id)">
          <li class="bk-menu-child-item" v-for="(child, childIndex) in (item.children || [])" :key="childIndex">
            <div
              :class="['bk-menu-child-title-wrapper', { selected: selected === child.id }]"
              @click="handleChildClick(child, item)">
              {{child.name}}
            </div>
          </li>
        </ul>
      </CollapseTransition>
    </li>
  </ul>
</template>

<script lang="ts">
import { defineComponent, ref, toRefs, watch, PropType } from '@vue/composition-api';
import CollapseTransition from './collapse-transition';
import { IMenuItem } from '@/store/menu';

export default defineComponent({
  name: 'SideMenu',
  components: { CollapseTransition },
  props: {
    // 菜单列表
    list: {
      type: Array as PropType<IMenuItem[]>,
      default: () => ([]),
    },
    // 选中菜单的ID
    selected: {
      type: String,
      default: '',
    },
  },
  setup(props, ctx) {
    const { emit } = ctx;

    const openedMenu = ref<string[]>([]); // 展开的菜单项
    const { selected, list } = toRefs(props);
    watch(selected, () => {
      // 如果是子菜单选中时默认展开父级
      const parent = list.value.find((item: IMenuItem) => !!item.children?.some(child => child.id === selected.value));
      if (parent) {
        const exit = openedMenu.value.some(id => id === parent.id);
        !exit && openedMenu.value.push(parent.id);
      }
    }, { immediate: true });

    const handleItemClick = (item) => {
      if (item.children?.length) {
        const index = openedMenu.value.findIndex(id => id === item.id);
        if (index > -1) {
          openedMenu.value.splice(index, 1);
        } else {
          openedMenu.value.push(item.id);
        }
      } else {
        emit('change', item);
      }
    };
    const handleChildClick = (child, item) => {
      // 点击子菜单时折叠其他菜单项
      openedMenu.value = [item.id];
      emit('change', child);
    };

    return {
      openedMenu,
      handleItemClick,
      handleChildClick,
    };
  },
});
</script>

<style lang="postcss" scoped>
@import '@/css/variable.css';

.collapse-transition {
    -webkit-transition: .2s height ease-in-out, .2s padding-top ease-in-out, .2s padding-bottom ease-in-out;
    -moz-transition: .2s height ease-in-out, .2s padding-top ease-in-out, .2s padding-bottom ease-in-out;
    transition: .2s height ease-in-out, .2s padding-top ease-in-out, .2s padding-bottom ease-in-out;
}

.bk-menu {
    position: relative;
}

.bk-menu-item {
    cursor: pointer;
    .line {
        height: 1px;
       border-bottom: 1px solid #f0f1f5;
       /* opacity: 0.5; */
    }
}

.bk-menu-child-item {
    &:hover {
        color: $primaryColor;
    }
}

.bk-menu-title-wrapper {
    height: 48px;
    line-height: 48px;
    font-size: 14px;
    padding: 0 40px 0 25px;
    position: relative;

    &:hover {
        color: $primaryColor;

        .left-icon {
            color: $primaryColor;
        }
    }

    &.hide {
        display: none;
    }

    &.disable {
        cursor: not-allowed;
        color: #c3cdd7;
        .left-icon {
            cursor: not-allowed;
            color: #c3cdd7
        }
    }

    &.selected {
        background-color: $primaryLightColor;
        color: $primaryColor;
        .left-icon {
            color: $primaryColor;
        }
    }

    &.child-selected {
        font-weight: 700;
    }

    .biz-badge {
        position: absolute;
        right: 20px;
        top: 17px;
    }

    .left-icon {
        vertical-align: middle;
        font-size: 20px;
        position: absolute;
        top: 14px;
        color: $fontWeightColor;

        &.selected {
            color: $primaryColor;
        }

        &.disable {
            cursor: not-allowed;
            color: #c3cdd7;
        }
    }

    .right-icon {
        position: absolute;
        right: 20px;
        top: 17px;
        font-size: 12px;
        -webkit-transition: transform linear .2s;
        transition: transform linear .2s;

        &.selected {
            color: $primaryColor;
            -webkit-transform: rotate(180deg);
            transform: rotate(180deg);
        }
    }

    .bk-menu-title {
        margin-left: 40px;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
        font-weight: normal;
    }
}

.bk-menu-child-title-wrapper {
    font-size: 14px;
    height: 36px;
    line-height: 36px;
    padding-left: 65px;
    position: relative;

    &.selected {
        background-color: $primaryLightColor;
        color: $primaryColor;
    }
}

</style>
