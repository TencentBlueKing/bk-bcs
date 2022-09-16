/* eslint-disable no-unused-expressions */
import { messageInfo } from '@/common/bkmagic'
import { copyText } from '@/common/util'

export default {
    inserted (el, bind) {
        const parentNode = el.parentNode
        if (!parentNode) return

        const tools = bind.value?.tools || ['fullscreen']
        if (!tools || !tools.length) return

        el.handleExitFullScreen = (event) => {
            if (event.code === 'Escape' && el.fullscreen) {
                el.fullscreen.className = 'bcs-icon bcs-icon-full-screen'
                tools.forEach(tool => {
                    el[tool].style.position = 'absolute'
                })
                el.classList.remove('bcs-full-screen')
            }
        }
        el.addEventListener('mouseenter', () => {
            document.addEventListener('keyup', el.handleExitFullScreen)
        })
        el.addEventListener('mouseleave', () => {
            document.removeEventListener('keyup', el.handleExitFullScreen)
        })

        el.defaultConfig = {
            fullscreen: {
                icon: 'bcs-icon bcs-icon-full-screen',
                handler: (e) => {
                    const target = e.target
                    if (target.className === 'bcs-icon bcs-icon-full-screen') {
                        target.className = 'bcs-icon bcs-icon-un-full-screen'
                        tools.forEach(tool => {
                            el[tool].style.position = 'fixed'
                        })
                        el.classList.add('bcs-full-screen')
                        messageInfo(window.i18n.t('按Esc即可退出全屏模式'))
                    } else {
                        target.className = 'bcs-icon bcs-icon-full-screen'
                        tools.forEach(tool => {
                            el[tool].style.position = 'absolute'
                        })
                        el.classList.remove('bcs-full-screen')
                    }
                }
            },
            copy: {
                icon: 'bcs-icon bcs-icon-copy',
                handler: () => {
                    copyText(bind.value?.content)
                    messageInfo(window.i18n.t('复制成功'))
                }
            }
        }
        parentNode.style.position = 'relative'

        tools.forEach((tool, index) => {
            const icon = document.createElement('i')
            icon.className = el.defaultConfig[tool]?.icon
            icon.style.cssText = `position: absolute;right: ${(index + 1) * 20}px;top: 15px;cursor: pointer;z-index: 200;margin-right: ${index * 10}px`
            el[tool] = icon
            icon.addEventListener('click', el.defaultConfig[tool]?.handler)
            parentNode.append(icon)
        })
    },
    unbind (el, bind) {
        const tools = bind.value?.tools || ['fullscreen', 'copy']
        document.removeEventListener('keyup', el.handleExitFullScreen)
        tools.forEach(tool => {
            el[tool]?.removeEventListener('click', el.defaultConfig[tool]?.handler)
            el[tool]?.remove()
        })
    }
}
