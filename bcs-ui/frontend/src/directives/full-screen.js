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
import { messageInfo } from '@/common/bkmagic';
import { copyText } from '@/common/util';

export default {
  inserted(el, bind) {
    const tools = bind.value?.tools || ['fullscreen'];
    if (!tools?.length) return;

    el.handleExitFullScreen = (event) => {
      if (event.code === 'Escape' && el.fullscreen) {
        el.fullscreen.className = 'bcs-icon bcs-icon-full-screen';
        tools.forEach((tool) => {
          el[tool].style.position = 'absolute';
        });
        el.classList.remove('bcs-full-screen');
      }
    };
    el.addEventListener('mouseenter', () => {
      document.addEventListener('keyup', el.handleExitFullScreen);
    });
    el.addEventListener('mouseleave', () => {
      document.removeEventListener('keyup', el.handleExitFullScreen);
    });

    el.defaultConfig = {
      fullscreen: {
        icon: 'bcs-icon bcs-icon-full-screen',
        handler: (e) => {
          const { target } = e;
          if (target.className === 'bcs-icon bcs-icon-full-screen') {
            target.className = 'bcs-icon bcs-icon-un-full-screen';
            tools.forEach((tool) => {
              el[tool].style.position = 'fixed';
            });
            el.classList.add('bcs-full-screen');
            messageInfo(window.i18n.t('generic.button.fullScreen.msg'));
          } else {
            target.className = 'bcs-icon bcs-icon-full-screen';
            tools.forEach((tool) => {
              el[tool].style.position = 'absolute';
            });
            el.classList.remove('bcs-full-screen');
          }
        },
      },
      copy: {
        icon: 'bcs-icon bcs-icon-copy',
        handler: () => {
          copyText(bind.value?.content);
          messageInfo(window.i18n.t('generic.msg.success.copy'));
        },
      },
    };
    el.style.position = 'relative';

    const css = bind.value?.css || '';
    tools.forEach((tool, index) => {
      const icon = document.createElement('i');
      icon.className = el.defaultConfig[tool]?.icon;
      icon.style.cssText = `position: absolute;right: ${(index + 1) * 20}px;top: 15px;cursor: pointer;z-index: 200;margin-right: ${index * 10}px;color: #fff;${css}`;
      el[tool] = icon;
      icon.addEventListener('click', el.defaultConfig[tool]?.handler);
      el.append(icon);
    });
  },
  unbind(el, bind) {
    const tools = bind.value?.tools || ['fullscreen', 'copy'];
    document.removeEventListener('keyup', el.handleExitFullScreen);
    tools.forEach((tool) => {
      el[tool]?.removeEventListener('click', el.defaultConfig[tool]?.handler);
      el[tool]?.remove();
    });
  },
};
