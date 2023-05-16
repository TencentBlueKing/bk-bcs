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
  name: 'bk-input',
  props: {
    value: {
      type: [String, Number],
      default: '',
    },
    // type 为 number时是否限制为整数
    int: {
      type: Boolean,
      default: false,
    },
  },
  render(h) {
    const slots = Object.keys(this.$slots)
      .reduce((arr, key) => arr.concat(this.$slots[key]), [])
      .map((vnode) => {
        vnode.context = this._self;
        return vnode;
      });

    // eslint-disable-next-line @typescript-eslint/naming-convention, @typescript-eslint/no-this-alias
    const _self = this;

    if (_self.$attrs.type === 'number') {
      this.$listeners.input = (value, event) => {
        _self.$emit('input', isNaN(value) ? value : Number(value), event);
      };
      this.$listeners.blur = (value, event) => {
        if (isNaN(value)) {
          value = _self.$attrs.min === Number.MIN_SAFE_INTEGER ? 0 : _self.$attrs.min;
        }
        _self.$emit('input', this.int ? Math.ceil(value) : Number(value), event);
      };
    }

    return h('bcs-input', {
      ref: 'input',
      on: this.$listeners, // 透传事件
      props: Object.assign({}, this.$props, this.$attrs), // 透传props
      scopedSlots: this.$scopedSlots, // 透传scopedSlots
      attrs: this.$attrs, // 透传属性，非props
    }, slots);
  },
  methods: {
    focus() {
      this.$refs.input.focus();
    },
  },
};
