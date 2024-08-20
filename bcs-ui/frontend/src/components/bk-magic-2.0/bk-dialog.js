/*
* Tencent is pleased to support the open source community by making
* 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community Edition) available.
*
* Copyright (C) 2021 THL A29 Limited, a Tencent company.  All rights reserved.
*
* 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community Edition) is licensed under the MIT License.
*
* License for 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community Edition):
*
* ---------------------------------------------------
* Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated
* documentation files (the "Software"), to deal in the Software without restriction, including without limitation
* the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and
* to permit persons to whom the Software is furnished to do so, subject to the following conditions:
*
* The above copyright notice and this permission notice shall be included in all copies or substantial portions of
* the Software.
*
* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO
* THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF
* CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
* IN THE SOFTWARE.
*/
export default {
  name: 'bk-dialog',
  props: {
    isShow: {
      type: Boolean,
      default: false,
    },
    value: {
      type: Boolean,
      default: false,
    },
    hasFooter: {
      type: Boolean,
      default: true,
    },
    confirm: {
      type: String,
      default: window.i18n.t('generic.button.confirm'),
    },
    cancel: {
      type: String,
      default: window.i18n.t('generic.button.cancel'),
    },
    maskClose: {
      type: Boolean,
      default: true,
    },
    quickClose: {
      type: Boolean,
      default: true,
    },
    headerPosition: {
      type: String,
      default: 'left',
    },
    closeIcon: {
      type: Boolean,
      default: true,
    },
    title: {
      type: String,
    },
  },

  render(h) {
    const slots = Object.keys(this.$slots)
      .reduce((arr, key) => arr.concat(this.$slots[key]), [])
      .map((vnode) => {
        vnode.context = this._self;
        return vnode;
      });

    return h('bcs-dialog', {
      on: this.$listeners, // 透传事件
      props: Object.assign({}, {
        ...this.$props,
        value: this.$props.value || this.$props.isShow,
        showFooter: this.$props.hasFooter,
        okText: this.$props.confirm,
        cancelText: this.$props.cancel,
        maskClose: this.$props.quickClose,
      }, this.$attrs), // 透传props
      scopedSlots: this.$scopedSlots, // 透传scopedSlots
      attrs: this.$attrs, // 透传属性，非props
    }, slots);
  },
};
