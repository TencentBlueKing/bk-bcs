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
function path2tree(arr, conf) {
  const tree = {
    name: '/',
    title: '/',
    expanded: true,
    children: [],
  };

  if (!conf) {
    conf = {};
  }

  function addNode(obj, index) {
    const splitpath = obj.name.replace(/^\/|\/$/g, '').split('/');
    let ptr = tree;
    let iChild;

    for (let i = 0; i < splitpath.length; i++) {
      const node = {
        openedIcon: 'icon-folder-open',
        closedIcon: 'icon-folder',
        icon: 'icon-folder',
        name: splitpath[i],
        title: splitpath[i],
        expanded: false,
        children: [],
      };

      // 将第一个目录节点下的所有层级展开
      if (index === 0) {
        node.expanded = true;
      }

      // 找到最后一个节点，设置为叶子，并带上value属性
      if (i === splitpath.length - 1) {
        delete node.children;

        if (obj.value) {
          node.value = obj.value;
          node.icon = 'icon-file';
          delete node.openedIcon;
          delete node.closedIcon;
        }

        // default 选中第一个文件
        if (conf.useFirstDefault && index === 0) {
          node.selected = true;
        }
      }

      if (ptr.children) {
        const childrenCounts = ptr.children.length;
        for (iChild = 0; iChild < childrenCounts; iChild++) {
          const child = ptr.children[iChild];
          if (child.name === node.name) {
            if (i === splitpath.length - 1) {
              delete node.children;

              if (obj.value) {
                node.value = obj.value;
                node.selected = true;
                node.icon = 'icon-file';
                delete node.openedIcon;
                delete node.closedIcon;
              }
            } else {
              ptr = child;
            }
            break;
          }
        }

        // 循环结束后还没有找到name匹配的children，说明node不在children列表中，将node加入列表
        if (iChild >= childrenCounts) {
          ptr.children.push(node);
          ptr = node;
        }
      }
    }
  }

  arr.forEach((item, index) => {
    addNode(item, index);
  });

  return tree;
}

export default path2tree;
