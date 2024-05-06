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

const nodeList = [];
const clickctx = '$clickoutsideCtx';
let beginClick; // 确保鼠标按下和松开时是同一个目标

document.addEventListener('mousedown', event => (beginClick = event));

document.addEventListener('mouseup', (event) => {
  for (const node of nodeList) {
    node[clickctx].clickoutsideHandler(event, beginClick);
  }
});

function checkMousePosition(event, el) {
  const { clientX, clientY } = event;

  const dom_x = Number(el.getBoundingClientRect().left);
  const dom_x_width = Number(el.getBoundingClientRect().left + el.clientWidth);

  const dom_y = Number(el.getBoundingClientRect().top);
  const dom_y_height = Number(el.getBoundingClientRect().top + el.clientHeight);

  if (clientX > dom_x && clientX < dom_x_width && clientY > dom_y && clientY < dom_y_height) {
    return true;
  }
  return false;
}

export default {
  bind(el, binding, vnode) {
    const id = nodeList.push(el) - 1;
    const clickoutsideHandler = (mouseup = {}, mousedown = {}) => {
      if (binding.modifiers?.pos && checkMousePosition(mouseup, el)) return;
      if (!vnode.context // 点击在vue实例之外的DOM上
                || !mouseup.target
                || !mousedown.target
                || el.contains(mouseup.target) // 鼠标按下时的DOM节点是当前展开的组件的子元素
                || el.contains(mousedown.target) // 鼠标松开时的DOM节点是当前展开的组件的子元素
                || el === mouseup.target // 鼠标松开时的DOM节点是当前展开的组件的根元素
                || (
                  vnode.context.popup // 当前点击元素是有弹出层的
                    && (
                      vnode.context.popup.contains(mouseup.target) // 鼠标按下时的DOM节点是当前有弹出层元素的子节点
                        || vnode.context.popup.contains(mousedown.target) // 鼠标松开时的DOM节点是当前有弹出层元素的子节点
                    )
                )
      ) return;

      if (
        binding.expression // 传入了指令绑定的表达式
                && el[clickctx].callbackName // 当前元素的clickoutside对象中有回调函数名
                && vnode.context[el[clickctx].callbackName] // vnode中存在回调函数
      ) {
        vnode.context[el[clickctx].callbackName]();
      } else {
        el[clickctx].callbackFn?.();
      }
    };

    el[clickctx] = {
      id,
      clickoutsideHandler,
      callbackName: binding.expression,
      callbackFn: binding.value,
    };
  },
  update(el, binding) {
    el[clickctx].callbackName = binding.expression;
    el[clickctx].callbackFn = binding.value;
  },
  unbind(el) {
    for (let i = 0, len = nodeList.length; i < len; i++) {
      if (nodeList[i][clickctx].id === el[clickctx].id) {
        nodeList.splice(i, 1);
        break;
      }
    }
  },
};
