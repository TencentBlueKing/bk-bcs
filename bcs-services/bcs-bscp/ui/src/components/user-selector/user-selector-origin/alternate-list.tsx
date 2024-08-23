// @ts-nocheck
/* eslint-disable */
import { hideAll } from 'tippy.js';
import {
  type ComponentPublicInstance,
  computed,
  defineComponent,
  getCurrentInstance,
  type HTMLAttributes,
  nextTick,
  ref,
  watch,
  withModifiers,
} from 'vue';

import AlternateItem from './alternate-item';
import instanceStore from './instance-store';

export default defineComponent({
  setup() {
    const { proxy } = getCurrentInstance();
    instanceStore.setInstance('alternateContent', 'alternateList', proxy);

    const selector = ref(null);
    const keyword = ref('');
    const next = ref(true);
    const loading = ref(true);
    const matchedUsers = ref([]);
    const wrapperStyle = computed(() => {
      const style: any = {};
      if (selector.value?.panelWidth) {
        style.width = `${parseInt(selector.value.panelWidth, 10)}px`;
      }
      return style;
    });
    const listStyle = computed(() => {
      const style = {
        'max-height': '192px',
      };
      if (selector.value) {
        const maxHeight = parseInt(selector.value.listScrollHeight, 10);
        if (!isNaN(maxHeight)) {
          style['max-height'] = `${maxHeight}px`;
        }
      }
      return style;
    });
    const getIndex = (index: number, childIndex = 0) => {
      let flattenedIndex = 0;
      matchedUsers.value.slice(0, index).forEach((user) => {
        if (user.hasOwnProperty('children')) {
          flattenedIndex += user.children.length;
        } else {
          flattenedIndex += 1;
        }
      });
      return flattenedIndex + childIndex;
    };

    const handleScroll = () => {
      hideAll({ exclude: selector.value.inputRef, duration: 0 });
      if (loading.value || !next.value) {
        return false;
      }
      const list = alternateList.value;
      const threshold = 32;
      if (list.scrollTop + list.clientHeight > list.scrollHeight - threshold) {
        selector.value.search(keyword.value, next.value);
      }
    };

    watch(keyword, () => {
      alternateItem.value = [];
      nextTick(() => {
        alternateList.value.scrollTop = 0;
      });
    });

    const alternateListContainer = ref(null);
    const alternateList = ref(null);
    const alternateItem = ref([]);

    const setRef = (el: HTMLElement | ComponentPublicInstance | HTMLAttributes) => {
      alternateItem.value.push(el);
    };

    return {
      selector,
      keyword,
      next,
      loading,
      matchedUsers,
      wrapperStyle,
      listStyle,
      getIndex,
      handleScroll,
      alternateListContainer,
      alternateList,
      alternateItem,
      setRef,
    };
  },
  render() {
    return (
      <div ref="alternateListContainer"
        class={[
          'user-selector-alternate-list-wrapper',
          this.selector?.displayListTips ? 'has-folder' : '',
          this.loading ? 'is-loading' : '',
        ]}
        style={this.wrapperStyle}
      >
        <ul
          class="alternate-list"
          ref="alternateList"
          style={this.listStyle}
          onScroll={() => void this.handleScroll()}
        >
          {
            this.matchedUsers.map((user, index) => {
              if (user.hasOwnProperty('children')) {
                return <>
                  <li
                    class="alternate-group"
                    onClick={(e) => e.stopPropagation()}
                    onMousedown={withModifiers((): any => void this.selector.handleGroupMousedown(), ['left', 'stop'])}
                    onMouseup={withModifiers((): any => void this.selector.handleGroupMouseup(), ['left', 'stop'])}
                  >
                    { `${user.display_name}(${user.children.length})` }
                  </li>
                  {
                    user.children.map((child: any, childIndex: number) => <>
                      <AlternateItem
                        // ref="alternateItem"
                        ref={(el: HTMLElement | ComponentPublicInstance | HTMLAttributes) => void this.setRef(el)}
                        index={this.getIndex(index, childIndex)}
                        selector={this.selector}
                        user={child}
                        keyword={this.keyword} />
                    </>)
                  }
                </>;
              }
              return <>
                <AlternateItem
                  // ref="alternateItem"
                  ref={(el: HTMLElement | ComponentPublicInstance | HTMLAttributes) => void this.setRef(el)}
                  selector={this.selector}
                  user={user}
                  index={this.getIndex(index)}
                  keyword={this.keyword} />
              </>;
            })
          }
        </ul>
        {
           (!this.loading && !this.matchedUsers.length)
             ? <>
              <p class="alternate-empty" >
                { this.selector.emptyText }
              </p>
             </>
             : null
        }
      </div>
    );
  },
});
