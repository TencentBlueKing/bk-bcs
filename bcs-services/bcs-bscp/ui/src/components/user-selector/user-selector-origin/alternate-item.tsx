// @ts-nocheck
import { computed, defineComponent, toRefs, withModifiers } from 'vue';

import RenderAvatar from './render-avatar';
import RenderList from './render-list';
import tooltips from './tooltips';

export default defineComponent({
  name: 'AlternateItem',
  directives: {
    tooltips,
  },
  // props: ['selector', 'user', 'keyword', 'index'],
  props: {
    selector: {
      type: Object,
    },
    user: {
      type: Object,
    },
    keyword: {
      type: String,
    },
    index: {
      type: Number,
    },
  },
  setup(props) {
    const { selector, user, keyword } = toRefs(props);
    const disabled = computed(() => selector.value.disabledUsers.includes(user.value.username));
    const getItemContent = () => {
      const [nameWithoutDomain, domain] = user.value.username.split('@');
      let displayText = nameWithoutDomain;
      if (keyword.value) {
        displayText = displayText.replace(
          new RegExp(keyword.value, 'g'),
          `<span>${keyword.value}</span>`,
        );
      }
      const displayUsername = selector.value.displayDomain && domain
        ? `${displayText}@${domain}`
        : displayText;

      const displayName = user.value.display_name;
      if (displayName) {
        return `${displayUsername}(${displayName})`;
      }

      return displayUsername;
    };
    const getTitle = () => selector.value.getDisplayText(user.value);

    return {
      disabled,
      getItemContent,
      getTitle,
    };
  },
  render() {
    return (
      <li
        class={[
          'alternate-item',
          this.index === this.selector.highlightIndex ? 'highlight' : '',
          this.disabled && !this.selector.renderList ? 'disabled' : '',
        ]}
        onClick={(e) => e.stopPropagation()}
        onMousedown={withModifiers(() => this.selector.handleUserMousedown(this.user, this.disabled), ['left', 'stop'])}
        onMouseup={withModifiers(() => this.selector.handleUserMouseup(this.user, this.disabled), ['left', 'stop'])}>
        {
          this.selector.renderList
            ? <>
              <RenderList
                selector={this.selector}
                keyword={this.keyword}
                user={this.user}
                disabled={this.disabled}
              >
              </RenderList>
            </>
            : <>
              {
                this.selector.tagType === 'avatar'
                  ? <>
                    <RenderAvatar
                      class="item-avatar"
                      user={this.user}
                      urlMethod={this.selector.avatarUrl}>
                    </RenderAvatar>
                  </>
                  : null
              }
              {
                this.selector.displayListTips && this.user.category_name
                  ? <>
                    <span
                      class="item-folder"
                      v-tooltips={{
                        placement: 'right',
                        interactive: true,
                        theme: 'light list-item-tips',
                        content: this.user.category_name,
                        offset: [0, 18],
                      }}>
                      { this.user.category_name }
                    </span>
                  </>
                  : null
              }
              <span
                class="item-name"
                title={this.getTitle()}
                v-html={this.getItemContent()}
              ></span>
            </>
        }
      </li>
    );
  },
});
