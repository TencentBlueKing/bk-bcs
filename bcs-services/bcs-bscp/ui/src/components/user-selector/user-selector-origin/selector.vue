<template>
  <!-- eslint-disable -->
  <div
    class="user-selector"
    v-if="isSelector"
    v-bind="$attrs"
    :style="{
      height: fixedHeight ? selectorHeight + 'px' : 'auto',
    }"
    @mousedown="shouldUpdate = false"
    @click="focus"
  >
    <div class="user-selector-layout">
      <div
        class="user-selector-container"
        ref="containerRef"
        :class="{
          focus: isFocus,
          disabled: disabled,
          placeholder: !localValue.length && !isFocus,
          'is-fast-clear': fastClear,
          'has-avatar': tagType === 'avatar',
          'is-loading': loading,
          'is-flex-height': !fixedHeight,
        }"
        :style="containerStyle"
        :data-placeholder="placeholder"
        @mousewheel="handleContainerScroll"
      >
        <template v-if="multiple || !isFocus">
          <!-- :ref="getRefSetter('selected')" -->
          <span
            v-for="(user, index) in localValueUsers"
            class="user-selector-selected"
            :key="user.username"
            @click.stop
            @mousedown.left.stop="handleSelectedMousedown($event, index)"
            @mouseup.left.stop="handleSelectedMouseup($event, index)"
            @mouseenter="handleSelectedMouseenter($event, user)"
            @mouseleave="handleSelectedMouseleave($event, user)"
          >
            <template v-if="renderTag">
              <render-tag
                :user="user"
                :username="user.username"
                :index="index"
              ></render-tag>
            </template>

            <template v-else>
              <render-avatar
                class="user-selector-selected-avatar"
                v-if="tagType === 'avatar'"
                :user="user"
                :url-method="avatarUrl">
              </render-avatar>
              <span class="user-selector-selected-value">
                {{ getDisplayText(user) }}
              </span>
            </template>

            <i
              class="user-selector-selected-clear bk-biz-components-icon bk-biz-icon-close"
              v-if="tagClearable && tagType === 'tag'"
              @mouseup.left.stop
              @mousedown.left.stop="handleRemoveMouseDown"
              @click.stop.prevent="handleRemoveSelected(user, index)">
            </i>
          </span>
        </template>
        <span
          ref="inputRef"
          class="user-selector-input"
          spellcheck="false"
          contenteditable
          v-show="isFocus"
          @click.stop
          @input="handleInput($event)"
          @blur="handleBlur"
          @paste.prevent.stop="handlePaste($event)"
          @keydown="handleKeydown($event)"
        >
        </span>
      </div>
      <i
        class="user-selector-clear bk-biz-components-icon bk-biz-icon-close-circle-shape"
        v-if="fastClear && !disabled && localValue.length"
        @click.stop="handleFastClear"
      >
      </i>
    </div>
  </div>
  <span class="user-selector user-selector-info" v-else>
    {{ userInfo }}
  </span>
</template>
<script>
/* eslint-disable */
import { throttle } from 'lodash';
import Tippy from 'tippy.js';
import {
  computed,
  createApp,
  defineComponent,
  getCurrentInstance,
  nextTick,
  onMounted,
  provide,
  ref,
  toRefs,
  watch,
} from 'vue';

import AlternateList from './alternate-list';
import instanceStore from './instance-store';
import RenderAvatar from './render-avatar';
import RenderTag from './render-tag';
import request from './request';

import 'tippy.js/dist/tippy.css';
import 'tippy.js/themes/light.css';
import '@icon-cool/bk-icon-bk-biz-components';
export default defineComponent({
  name: 'BkUserSelector',
  components: {
    RenderTag,
    RenderAvatar,
  },
  props: {
    modelValue: {
      type: Array,
      default: () => [],
    },
    placeholder: {
      type: String,
      default: '请输入',
    },
    disabled: {
      type: Boolean,
      default: false,
    },
    multiple: {
      type: Boolean,
      default: true,
    },
    exclude: {
      type: Boolean,
      default: true,
    },
    focusRowLimit: {
      type: Number,
      default: 6,
    },
    defaultAlternate: {
      type: [String, Array, Function],
      validator(value) {
        return (
          value === 'history'
          || typeof value === 'function'
          || value instanceof Array
        );
      },
    },
    searchFromDefaultAlternate: {
      type: Boolean,
      default: true,
    },
    historyKey: String,
    historyLabel: {
      type: String,
      default: '最近选择',
    },
    historyRecord: {
      type: Number,
      default: 5,
    },
    displayListTips: Boolean,
    fuzzySearchMethod: Function,
    exactSearchMethod: Function,
    emptyText: {
      type: String,
      default: '无匹配人员',
    },
    tagClearable: {
      type: Boolean,
      default: true,
    },
    fastClear: Boolean,
    renderList: Function,
    renderTag: Function,
    displayTagTips: Boolean,
    tagTipsContent: Function,
    tagTipsDelay: {
      type: Number,
      default: 300,
    },
    tagType: {
      type: String,
      default: 'tag',
      validator(value) {
        return ['tag', 'avatar'].includes(value);
      },
    },
    avatarUrl: {
      type: Function,
      default: () => null,
    },
    fixedHeight: {
      type: Boolean,
      default: true,
    },
    disabledUsers: {
      type: Array,
      default: () => [],
    },
    listScrollHeight: [Number, String],
    api: String,
    searchLimit: {
      type: Number,
      default: 20,
    },
    pasteFormatter: {
      type: Function,
      default(value) {
        return value.replace(/\(.*/, '');
      },
    },
    pasteValidator: Function,
    panelWidth: {
      type: [Number, String],
      validator(value) {
        const pixel = parseInt(value, 10);
        return pixel >= 190;
      },
    },
    displayDomain: {
      type: Boolean,
      default: true,
    },
    type: {
      type: String,
      default: 'selector',
      validator(value) {
        return ['selector', 'info'].includes(value);
      },
    },
  },
  emits: [
    'update:modelValue',
    'change',
    'remove-selected',
    'select-user',
    'keydown',
    'focus',
    'blur',
    'clear',
  ],
  setup(props, ctx) {
    const {
      modelValue,
      disabled,
      multiple,
      exclude,
      focusRowLimit,
      defaultAlternate,
      searchFromDefaultAlternate,
      historyKey,
      historyLabel,
      historyRecord,
      fuzzySearchMethod,
      exactSearchMethod,
      displayTagTips,
      tagTipsContent,
      tagTipsDelay,
      fixedHeight,
      disabledUsers,
      api,
      searchLimit,
      pasteFormatter,
      pasteValidator,
      displayDomain,
      type,
    } = toRefs(props);

    const search = async (value, next) => {
      try {
        const popoverInstance = getPopoverInstance();
        getAlternateContent();
        popoverInstance.setContent(alternateContent.value.$refs.alternateListContainer);
        showPopover();
        alternateContent.value.loading = !!value;
        const { results: users, next: nextPage } = await new Promise(async (resolve, _reject) => {
          if (value) {
            const promise = [(fuzzySearchMethod.value || defaultFuzzySearchMethod)(value, next)];
            if (searchFromDefaultAlternate.value) {
              promise.push(getDefaultAlternateData(value));
            }
            const [fuzzySearchData, defaultAlternateData] = await Promise.all(promise);
            if (defaultAlternateData) {
              fuzzySearchData.results.unshift(...defaultAlternateData.results);
            }
            resolve(fuzzySearchData);
          } else {
            const defaultAlternateData = getDefaultAlternateData();
            resolve(defaultAlternateData);
          }
        });

        if (!isFocus.value) {
          return;
        }
        const { matched, flattened } = filterUsers(users);
        if (!value && !flattened.length) {
          hidePopover();
          return;
        }

        matchedUsers.value = next ? [...matchedUsers.value, ...matched] : matched;
        flattenedUsers.value = next ? [...flattenedUsers.value, ...flattened] : flattened;
        highlightIndex.value = flattened.length && !!inputValue.value ? 0 : -1;

        alternateContent.value.next = nextPage;
        alternateContent.value.keyword = value;
        alternateContent.value.matchedUsers = matchedUsers.value;
        alternateContent.value.loading = false;
      } catch (e) {
        if (e.type === 'reset') {
          return;
        }
        matchedUsers.value = [];
        flattenedUsers.value = [];
        console.error(e);
      }
    };

    const containerRef = ref(null);
    const inputRef = ref(null);
    const { proxy } = getCurrentInstance();

    provide('parentSelector', proxy);

    const selectorHeight = ref(32);
    const singleRowHeight = ref(30);
    const inputValue = ref('');
    const inputIndex = ref(0);
    const highlightIndex = ref(-1);
    const shouldUpdate = ref(true);
    const isFocus = ref(false);
    const overflowTagIndex = ref(null);
    const currentUsers = ref([]);
    const matchedUsers = ref([]);
    const flattenedUsers = ref([]);
    const scheduleSearch = throttle(search, 800, { leading: false });
    const popoverInstance = ref(null);
    const alternateContent = ref(null);
    const selectedTipsTimer = ref({});
    const overflowTagNode = ref(null);
    const loading = ref(false);
    const isSelector = computed(() => type.value === 'selector');
    const containerStyle = computed(() => {
      const style = {};
      if (isFocus.value) {
        style.maxHeight = fixedHeight.value
          ? `${focusRowLimit.value * singleRowHeight.value}px`
          : 'auto';
      } else if (fixedHeight.value) {
        style.height = `${singleRowHeight.value}px`;
      }
      return style;
    });

    // const localValue = computed({
    //   get() {
    //     return [...modelValue.value];
    //   },
    //   set(value) {
    //     ctx.emit('update:modelValue', value);
    //     ctx.emit('change', value);
    //   },
    // });
    const localValue = ref([]);
    localValue.value = [...modelValue.value];

    const localValueUsers = computed(() => localValue.value.map((username) => {
      const user = currentUsers.value.find(user => user.username === username);
      return user || { username };
    }));

    const userInfo = computed(() => localValueUsers.value
      .map(user => getDisplayText(user))
      .join(';'));

    const getCurrentUsers = async () => {
      try {
        if (api.value) {
          currentUsers.value = await request.scheduleExactSearch(
            api.value,
            localValue.value,
          );
        } else if (exactSearchMethod.value) {
          currentUsers.value = await exactSearchMethod.value(localValue.value);
        } else {
          console.warn('No exact search method has been set');
        }
      } catch (error) {
        console.error(error);
      }
    };
    const getDefaultAlternateData = async (keyword) => {
      let users = [];
      const isMatch = (user, keyword) => user.username.toLowerCase().indexOf(keyword.toString().toLowerCase())
        > -1;
      if (defaultAlternate.value === 'history') {
        users = [
          { display_name: historyLabel.value, children: getHistoryUsers() },
        ];
      } else if (defaultAlternate.value instanceof Array) {
        users = defaultAlternate.value;
      } else if (typeof defaultAlternate.value === 'function') {
        users = await defaultAlternate.value();
      }
      if (keyword) {
        const filterResult = [];
        users.forEach((user) => {
          if (user.hasOwnProperty('children')) {
            const children = user.children.filter(child => isMatch(child, keyword));
            if (children.length) {
              filterResult.push({
                ...user,
                children,
              });
            }
          } else if (isMatch(user, keyword)) {
            filterResult.push(user);
          }
        });
        users = filterResult;
      }
      return Promise.resolve({ results: users, next: false });
    };
    const filterUsers = (users) => {
      const matched = [];
      const flattened = [];
      users.forEach((user) => {
        if (user.hasOwnProperty('children')) {
          const children = user.children.filter(child => !flattened.some(flattenedUser => flattenedUser.username === child.username));
          if (multiple.value) {
            const unexistUser = children.filter(child => !localValue.value.includes(child.username));
            if (unexistUser.length) {
              user.children = unexistUser;
              matched.push(user);
              flattened.push(...unexistUser);
            }
          } else {
            matched.push(user);
            flattened.push(...children);
          }
          return;
        }
        const exist = localValue.value.includes(user.username);
        const repeat = flattened.some(flattenedUser => flattenedUser.username === user.username);
        if ((!multiple.value || !exist) && !repeat) {
          matched.push(user);
          flattened.push(user);
        }
      });
      return {
        matched,
        flattened,
      };
    };
    const defaultFuzzySearchMethod = async (value, next) => {
      if (api.value) {
        const params = {
          app_code: 'bk-magicbox',
          page: next || 1,
          page_size: searchLimit.value,
          fuzzy_lookups: value,
        };
        const { count, results } = await request.fuzzySearch(api.value, params);
        const nextPage = count > params.page * params.page_size ? params.page + 1 : false;

        return {
          next: nextPage,
          results,
        };
      }

      if (!fuzzySearchMethod.value) {
        console.warn('No fuzzy search method has been set');
        return Promise.resolve({ next: false, results: [] });
      }
      return fuzzySearchMethod.value(value, next);
    };
    const getUserTips = async (instance, username) => {
      try {
        const contentElement = document.createElement('span');
        if (typeof tagTipsContent.value === 'function') {
          const content = await tagTipsContent.value(username);
          contentElement.innerHTML = content;
        } else {
          const user = await (
            exactSearchMethod.value || defaultExactSearchMethod
          )(username);
          contentElement.innerHTML = user
            ? user.category_name
            : 'Non existing user';
        }
        instance.setContent(contentElement);
      } catch (e) {
        console.error(e);
        instance.setContent(e.message);
      }
    };
    const defaultExactSearchMethod = (value) => {
      if (api.value) {
        return request.exactSearch(api.value, value);
      }
      if (!exactSearchMethod.value) {
        console.warn('No exact search method has been set');
        return Promise.resolve({});
      }
      return exactSearchMethod.value(value);
    };
    const getPopoverInstance = () => {
      if (!popoverInstance.value) {
        popoverInstance.value = Tippy(inputRef.value, {
          theme: 'light user-selector-popover',
          appendTo: document.body,
          trigger: 'manual',
          placement: 'bottom-start',
          offset: [0, 5],
          arrow: false,
          hideOnClick: false,
          content: '',
          interactive: true,
          onHide: () => {
            handlePopoverHide();
          },
          onShow: () => isFocus.value,
        });
      }
      return popoverInstance.value;
    };
    const getAlternateContent = () => {
      if (!alternateContent.value) {
        const alternateContentApp = createApp(AlternateList);
        const alternateContentContainer = document.createElement('div');
        alternateContentApp.mount(alternateContentContainer);
        alternateContent.value = instanceStore.getInstance('alternateContent', 'alternateList');
        // alternateContent.value.selector = proxy;
        alternateContent.value.selector = proxy;
        // document.body.appendChild(alternateContentContainer);
      }
      // return alternateContent;
    };
    const getHistoryUsers = () => {
      try {
        if (historyKey.value) {
          const users = JSON.parse(window.localStorage.getItem(historyKey.value)) || [];
          return users.filter(user => !disabledUsers.value.includes(user.username));
        }
        throw new Error('History key not provide');
      } catch (e) {
        console.error(e);
        return [];
      }
    };
    const updateHistoryUsers = (user) => {
      if (historyKey.value) {
        try {
          const histories = getHistoryUsers();
          const exist = histories.findIndex(history => history.username === user.username);
          if (exist > -1) {
            histories.splice(exist, 1);
          }
          Array.isArray(user)
            ? histories.unshift(...user)
            : histories.unshift(user);
          const newHistories = histories
            .filter(history => !disabledUsers.value.includes(history.username))
            .slice(0, historyRecord.value);
          window.localStorage.setItem(
            historyKey.value,
            JSON.stringify(newHistories),
          );
        } catch (e) {
          console.error(e);
        }
      }
    };
    const updatePopover = () => {
      popoverInstance.value?.popperInstance?.update();
    };
    const showPopover = () => {
      updatePopover();
      popoverInstance.value?.show(0);
    };
    const hidePopover = () => {
      popoverInstance.value?.hide(0);
    };
    const handlePopoverHide = () => {
      nextTick(() => {
        matchedUsers.value = [];
        flattenedUsers.value = [];
        alternateContent.value.matchedUsers = [];
      });
    };
    const getDisplayText = (user) => {
      const isObject = typeof user === 'object';
      let displayText = isObject ? user.username : user;
      displayText = displayDomain.value
        ? displayText
        : displayText.replace(/@.*/, '');
      if (isObject && user.display_name) {
        displayText += `(${user.display_name})`;
      }
      return displayText;
    };
    const focus = () => {
      if (disabled.value) {
        return false;
      }
      clearOverflowTimer();
      inputIndex.value = localValue.value.length;
      if (!multiple.value && modelValue.value.length) {
        inputValue.value = getDisplayText(modelValue.value[0]);
        inputRef.value.innerHTML = inputValue.value;
        moveInput(0, { selectRange: true });
      } else {
        moveInput(0);
      }
    };
    const handleContainerScroll = (event) => {
      popoverInstance.value?.state.isVisible && event.preventDefault();
    };
    const handleSelectedMousedown = (_event, _index) => {
      if (disabled.value) {
        return false;
      }
      shouldUpdate.value = false;
    };
    const handleSelectedMouseup = (event, index) => {
      if (disabled.value) {
        return false;
      }
      if (multiple.value) {
        const $referenceTarget = event.target;
        const { offsetWidth } = $referenceTarget;
        const eventX = event.offsetX;
        inputIndex.value = eventX > offsetWidth / 2 ? index + 1 : index;
        moveInput(0);
      } else {
        inputValue.value = getDisplayText(modelValue.value[0]);
        inputRef.value.innerHTML = inputValue.value;
        moveInput(0, { selectRange: true });
      }
    };
    const handleSelectedMouseenter = (event, { username }) => {
      if (!displayTagTips.value) {
        return false;
      }
      const target = event.currentTarget;
      if (target._user_tips_) {
        return false;
      }
      selectedTipsTimer.value[username] = setTimeout(() => {
        target._user_tips_ = Tippy(target, {
          theme: 'light small-arrow user-selected-tips',
          offset: [0, 5],
          appendTo: document.body,
          arrow: true,
          content: 'loading...',
          placement: 'top',
          interactive: true,
          onShow: (instance) => {
            getUserTips(instance, username);
          },
        });
        target._user_tips_.show();
        delete selectedTipsTimer.value[username];
      }, tagTipsDelay.value);
    };
    const handleSelectedMouseleave = (event, { username }) => {
      if (displayTagTips.value) {
        selectedTipsTimer.value[username]
          && clearTimeout(selectedTipsTimer.value[username]);
      }
    };
    const handleRemoveMouseDown = () => {
      shouldUpdate.value = false;
    };
    const handleRemoveSelected = ({ username }, index) => {
      if (disabled.value) {
        return false;
      }
      const lv = [...localValue.value];
      lv.splice(index, 1);
      localValue.value = lv;
      reset();
      if (isFocus.value) {
        moveInput(index >= inputIndex.value ? 0 : -1);
      } else {
        handleBlur();
      }
      ctx.emit('remove-selected', username);
    };
    const handleUserMousedown = (_user, _disabled) => {
      shouldUpdate.value = false;
    };
    const handleUserMouseup = (user, disabled) => {
      // debugger;
      if (disabled || disabled.value) {
        moveInput(0);
        return false;
      }
      updateHistoryUsers(user);
      currentUsers.value.push(user);
      if (multiple.value) {
        const lv = [...localValue.value];
        lv.splice(inputIndex.value, 0, user.username);
        localValue.value = lv;
        setTimeout(() => {
          moveInput(1);
          setSelection({ reset: true });
          search();
        }, 0);
      } else {
        localValue.value = [user.username];
        reset();
        handleBlur();
      }
      ctx.emit('select-user', user);
    };
    const handleGroupMousedown = () => {
      shouldUpdate.value = false;
    };
    const handleGroupMouseup = () => {
      moveInput(0);
    };

    const setSelection = (option = {}) => {
      if (option.reset) {
        reset();
      }
      isFocus.value = true;
      shouldUpdate.value = true;
      nextTick(() => {
        const $input = inputRef.value;
        if (!$input) {
          return;
        }
        $input.focus();
        const range = window.getSelection();
        range.selectAllChildren($input);
        !option.selectRange && range.collapseToEnd();
      });
    };
    const handleKeydown = (event) => {
      if (loading.value) {
        event.preventDefault();
        event.stopPropagation();
        return;
      }
      const { key } = event;
      const keyMap = {
        Enter: handleEnter,
        Backspace: handleBackspace,
        Delete: handleBackspace,
        ArrowLeft: handleArrow,
        ArrowRight: handleArrow,
        ArrowUp: handleArrow,
        ArrowDown: handleArrow,
      };
      if (keyMap.hasOwnProperty(key)) {
        keyMap[key](event);
      }
      ctx.emit('keydown', event);
    };
    const handleEnter = (e) => {
      // debugger;
      e.preventDefault();
      e.stopPropagation();
      shouldUpdate.value = false;
      if (highlightIndex.value !== -1) {
        const { username } = flattenedUsers.value[highlightIndex.value];
        const disabled = disabledUsers.value.includes(username);
        if (disabled) {
          return false;
        }
        if (multiple.value) {
          const lv = [...localValue.value];
          lv.splice(inputIndex.value, 0, username);
          localValue.value = lv;
          moveInput(1, { reset: true });
        } else {
          localValue.value = [username];
          reset();
          handleBlur();
        }
      } else if (inputValue.value) {
        if (!exclude.value && !localValue.value.includes(inputValue.value)) {
          if (multiple.value) {
            const lv = [...localValue.value];
            lv.splice(inputIndex.value, 0, inputValue.value);
            localValue.value = lv;
            moveInput(1, { reset: true });
          } else {
            localValue.value = [inputValue.value];
            reset();
            handleBlur();
          }
        } else {
          reset();
        }
      } else {
        reset();
        handleBlur();
      }
      hidePopover();
    };
    const handleBackspace = (_event) => {
      if (inputValue.value || !localValue.value.length || !inputIndex.value) {
        return true;
      }
      shouldUpdate.value = false;
      const lv = [...localValue.value];
      lv.splice(inputIndex.value - 1, 1);
      localValue.value = lv;
      moveInput(-1);
      search();
    };
    const handleArrow = (event) => {
      const arrow = event.key;
      if (['ArrowLeft', 'ArrowRight'].includes(arrow)) {
        if (inputValue.value || !localValue.value.length) {
          return true;
        }
        if (arrow === 'ArrowLeft' && inputIndex.value !== 0) {
          moveInput(-1);
        } else if (
          arrow === 'ArrowRight'
          && inputIndex.value !== localValue.value.length
        ) {
          moveInput(1);
        }
      } else if (flattenedUsers.value.length) {
        event.preventDefault();
        if (arrow === 'ArrowDown') {
          if (highlightIndex.value < flattenedUsers.value.length - 1) {
            highlightIndex.value += 1;
          } else if (alternateContent.value.next) {
            alternateContent.value.$refs.alternateList.scrollTop += 32;
            alternateContent.value.handleScroll();
          } else {
            highlightIndex.value = 0;
          }
        } else if (arrow === 'ArrowUp' && highlightIndex.value !== -1) {
          highlightIndex.value -= 1;
        }
      }
    };
    const handleInput = (event) => {
      if (loading.value) {
        event.preventDefault();
        event.stopPropagation();
        return;
      }
      inputValue.value = inputRef.value.textContent.trim();
    };
    const handleBlur = () => {
      if (!shouldUpdate.value) {
        return true;
      }
      isFocus.value = false;
      hidePopover();
    };
    const getMatchedUser = (nameToMatch) => {
      const user = flattenedUsers.value.find((user) => {
        const enName = user.username;
        const cnName = user.display_name;
        const isMatch = [enName, cnName].some(name => name.toLowerCase() === nameToMatch.toLowerCase());
        const isSelected = localValue.value.includes(enName);
        return isMatch && !isSelected;
      });
      return user;
    };
    const defaultPasteValidator = async (originalValues) => {
      if (api.value) {
        const users = await request.pasteValidate(api.value, originalValues);
        const validValues = users.map(user => user.username);
        if (!exclude.value) {
          return [...new Set(validValues.concat(originalValues))];
        }
        return validValues;
      }
      if (!pasteValidator.value) {
        console.warn('No paste validator has been set');
        return Promise.resolve([]);
      }
      return pasteValidator.value(values);
    };
    const handlePaste = async (event) => {
      hidePopover();
      if (loading.value) {
        event.preventDefault();
        event.stopPropagation();
      }
      try {
        loading.value = true;
        const pasteStr = event.clipboardData.getData('text').replace(/\s/g, '');
        const values = pasteStr
          .split(/,|;/)
          .map(value => pasteFormatter.value(value))
          .filter(value => value.length);
        const uniqueValues = [...new Set(values)];
        if (!uniqueValues.length) {
          return;
        }
        const validValues = await (
          pasteValidator.value || defaultPasteValidator
        )(uniqueValues);

        const newValues = validValues.filter(value => !localValue.value.includes(value));
        if (!validValues.length) {
          return;
        }
        const lv = [...localValue.value];
        lv.splice(inputIndex.value, 0, ...newValues);
        localValue.value = lv;
        if (multiple.value) {
          isFocus.value && moveInput(newValues.length, { reset: true });
        } else {
          handleBlur();
        }
      } catch (error) {
        console.error(error);
      } finally {
        loading.value = false;
      }
    };

    const getSelectedDOM = () => Array.from(containerRef.value.querySelectorAll('.user-selector-selected'));

    const moveInput = (step, option = {}) => {
      inputIndex.value = inputIndex.value + step;
      nextTick(() => {
        const selected = getSelectedDOM();
        const $referenceTarget = selected[inputIndex.value] || null;
        containerRef.value.insertBefore(inputRef.value, $referenceTarget);
        setSelection(option);
        updatePopover();
      });
    };
    const updateScroller = () => {
      if (!alternateContent.value || !isSelector.value) {
        return false;
      }
      nextTick(() => {
        const { highlightIndex } = proxy;
        const $alternateList = alternateContent.value.$refs.alternateList;
        if (!$alternateList) {
          return false;
        }
        if (highlightIndex !== -1) {
          const $alternateItem = alternateContent.value.alternateItem[highlightIndex].$el;
          const listClientHeight = $alternateList.clientHeight;
          const listScrollTop = $alternateList.scrollTop;
          const itemOffsetTop = $alternateItem.offsetTop;
          const itemOffsetHeight = $alternateItem.offsetHeight;
          if (
            itemOffsetTop >= listScrollTop
            && itemOffsetTop + itemOffsetHeight <= listScrollTop + listClientHeight
          ) {
            return false;
          }
          if (itemOffsetTop <= listScrollTop) {
            $alternateList.scrollTop = itemOffsetTop;
          } else if (itemOffsetTop + itemOffsetHeight > listScrollTop + listClientHeight) {
            $alternateList.scrollTop = itemOffsetTop + itemOffsetHeight - listClientHeight;
          }
        } else {
          $alternateList.scrollTop = 0;
        }
      });
    };
    const overflowTimer = ref(0);
    const calcOverflow = () => {
      if (!isSelector.value) {
        return false;
      }

      removeOverflowTagNode();

      if (
        !fixedHeight.value
        || isFocus.value
        || localValue.value.length < 2
      ) {
        return false;
      }
      clearOverflowTimer();
      overflowTimer.value = setTimeout(() => {
        const selectedUsers = getSelectedDOM();
        const userIndexInSecondRow = selectedUsers.findIndex((currentUser, index) => {
          if (!index) {
            return false;
          }
          const previousUser = selectedUsers[index - 1];
          return previousUser.offsetTop !== currentUser.offsetTop;
        });
        if (userIndexInSecondRow > -1) {
          overflowTagIndex.value = userIndexInSecondRow;
        } else {
          overflowTagIndex.value = null;
        }
        containerRef.value.scrollTop = 0;
        insertOverflowTag();
      }, 0);
    };
    const clearOverflowTimer = () => {
      overflowTimer.value && clearTimeout(overflowTimer.value);
    };
    const insertOverflowTag = () => {
      if (!overflowTagIndex.value) {
        return;
      }
      getOverflowTagNode();

      const selectedUser = getSelectedDOM();
      const referenceUser = selectedUser[overflowTagIndex.value];
      if (referenceUser) {
        overflowTagNode.value.textContent = `+${
          localValue.value.length - overflowTagIndex.value
        }`;
        containerRef.value.insertBefore(overflowTagNode.value, referenceUser);
      } else {
        overflowTagIndex.value = null;
        return;
      }
      setTimeout(() => {
        const previousUser = selectedUser[overflowTagIndex.value - 1];
        if (overflowTagNode.value.offsetTop !== previousUser.offsetTop) {
          overflowTagIndex.value -= 1;
          containerRef.value.insertBefore(
            overflowTagNode.value,
            overflowTagNode.value.previousSibling,
          );
          overflowTagNode.value.textContent = `+${
            localValue.value.length - overflowTagIndex.value
          }`;
        }
      }, 0);
    };
    const getOverflowTagNode = () => {
      if (overflowTagNode.value) {
        return overflowTagNode.value;
      }
      const node = document.createElement('span');
      node.className = 'user-selector-overflow-tag';
      overflowTagNode.value = node;
    };
    const removeOverflowTagNode = () => {
      if (
        overflowTagNode.value
        && overflowTagNode.value.parentNode === containerRef.value
      ) {
        containerRef.value.removeChild(overflowTagNode.value);
      }
    };
    const handleFastClear = () => {
      localValue.value = [];
      ctx.emit('clear');
    };
    const reset = () => {
      shouldUpdate.value = true;
      highlightIndex.value = -1;
      inputValue.value = '';
      inputRef.value.innerHTML = '';
    };
    // const getRefSetter = refKey => (ref) => {
    //   !ctx.root.$arrRefs && (ctx.root.$arrRefs = {});
    //   !ctx.root.$arrRefs[refKey] && (ctx.root.$arrRefs[refKey] = []);
    //   ref && ctx.root.$arrRefs[refKey].push(ref);
    // };
    watch(inputValue, (value) => {
      if (value.length) {
        highlightIndex.value = -1;
        updateScroller();
        scheduleSearch(value);
      } else if (isFocus.value) {
        search();
      }
    });
    watch(isFocus, (isFocus) => {
      if (isFocus) {
        search();
        ctx.emit('focus');
      } else {
        reset();
        ctx.emit('blur');
      }
      calcOverflow();
    });
    watch(highlightIndex, () => {
      updateScroller();
    });

    watch(localValue, (_localValue) => {
      ctx.emit('update:modelValue', _localValue);
      ctx.emit('change', _localValue);
      calcOverflow();
      getCurrentUsers();
    });

    onMounted(() => {
      calcOverflow();
      getCurrentUsers();
    });

    // onBeforeUpdate(() => {
    //   ctx.root.$arrRefs && (ctx.root.$arrRefs = {});
    // });
    return {
      selectorHeight,
      singleRowHeight,
      inputValue,
      inputIndex,
      highlightIndex,
      shouldUpdate,
      isFocus,
      overflowTagIndex,
      currentUsers,
      matchedUsers,
      flattenedUsers,
      scheduleSearch,
      popoverInstance,
      alternateContent,
      selectedTipsTimer,
      overflowTagNode,
      loading,
      isSelector,
      containerStyle,
      localValue,
      localValueUsers,
      userInfo,
      getCurrentUsers,
      search,
      getDefaultAlternateData,
      filterUsers,
      defaultFuzzySearchMethod,
      getUserTips,
      defaultExactSearchMethod,
      getPopoverInstance,
      getAlternateContent,
      getHistoryUsers,
      updateHistoryUsers,
      updatePopover,
      showPopover,
      hidePopover,
      handlePopoverHide,
      getDisplayText,
      focus,
      handleContainerScroll,
      handleSelectedMousedown,
      handleSelectedMouseup,
      handleSelectedMouseenter,
      handleSelectedMouseleave,
      handleRemoveMouseDown,
      handleRemoveSelected,
      handleUserMousedown,
      handleUserMouseup,
      handleGroupMousedown,
      handleGroupMouseup,
      setSelection,
      handleKeydown,
      handleEnter,
      handleBackspace,
      handleArrow,
      handleInput,
      handleBlur,
      getMatchedUser,
      defaultPasteValidator,
      handlePaste,
      getSelectedDOM,
      moveInput,
      updateScroller,
      calcOverflow,
      clearOverflowTimer,
      insertOverflowTag,
      getOverflowTagNode,
      removeOverflowTagNode,
      handleFastClear,
      reset,
      // getRefSetter,
      containerRef,
      inputRef,
    };
  },
});
</script>
<style lang="css">
@import './style.css';
</style>
