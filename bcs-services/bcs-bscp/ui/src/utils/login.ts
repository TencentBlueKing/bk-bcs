// @ts-nocheck
import isCrossOriginIFrame from './is-cross-origin-iframe'

const isCrossOrigin = isCrossOriginIFrame()
const topWindow: Window = isCrossOrigin ? window : window.top
const topDocument = topWindow.document

const openLoginDialog = (src, width = 460, height = 490, method = 'get') => {
    if (!src) return
    const isWraperExit = topDocument.querySelector('#bk-gloabal-login-iframe')
    if (isWraperExit) return
    const closeIcon = topDocument.createElement('div')
    closeIcon.style.cssText = 'transform: rotate(45deg);position: absolute;right: 0;cursor: pointer;color: #979ba5;width: 26px;height: 26px;'
    const closeIconTop = topDocument.createElement('span')
    closeIconTop.style.cssText = 'width:20px;display:inline-block;border:1px solid #979ba5;position:absolute;top: 50%;'
    const closeIconBottom = topDocument.createElement('span')
    closeIconBottom.style.cssText = 'width:20px;display:inline-block;border:1px solid #979ba5;transform: rotate(90deg);position: absolute;top: 13px;'
    closeIcon.id = 'bk-gloabal-login-close'
    closeIcon.appendChild(closeIconTop)
    closeIcon.appendChild(closeIconBottom)
    topDocument.addEventListener('click', topWindow.BLUEKING.corefunc.close_login_dialog)

    const frame = topDocument.createElement('iframe')
    frame.setAttribute('src', src)
    frame.style.cssText = `border: 0;outline: 0;width:${width}px;height:${height}px;`

    const dialogDiv = topDocument.createElement('div')
    dialogDiv.style.cssText = 'position: absolute;left: 50%;top: 20%;transform: translateX(-50%);background: #ffffff;'
    dialogDiv.appendChild(closeIcon)
    dialogDiv.appendChild(frame)

    const wraper = topDocument.createElement('div')
    wraper.id = 'bk-gloabal-login-iframe'
    wraper.style.cssText = 'position: fixed;top: 0;bottom: 0;left: 0;right: 0;background-color: rgba(0,0,0,.6);height: 100%;z-index: 5000;'
    wraper.appendChild(dialogDiv)
    topDocument.body.appendChild(wraper)
}
const closeLoginDialog = (e) => {
    try {
        e.stopPropagation()
        const el = e.target
        const closeIcon = topDocument.querySelector('#bk-gloabal-login-close')
        if (closeIcon !== el) return
        topDocument.removeEventListener('click', topWindow.BLUEKING.corefunc.close_login_dialog)
        // if (el) {
        //     el.removeEventListener('click', topWindow.BLUEKING.corefunc.close_login_dialog)
        // }
        topDocument.body.removeChild(el.parentElement.parentElement)
    } catch (_) {
        topDocument.removeEventListener('click', topWindow.BLUEKING.corefunc.close_login_dialog)
        const wraper = topDocument.querySelector('#bk-gloabal-login-iframe')
        if (wraper) {
            topDocument.body.removeChild(wraper)
        }
    }
    window.location.reload()
}

try {
    window.top.BLUEKING.corefunc.open_login_dialog = openLoginDialog
    window.top.BLUEKING.corefunc.close_login_dialog = closeLoginDialog
} catch (_) {
    topWindow.BLUEKING = {
        corefunc: {
            open_login_dialog: openLoginDialog,
            close_login_dialog: closeLoginDialog
        }
    }
    // 兼容接口返回的登录成功 html
    window.open_login_dialog = openLoginDialog
    window.close_login_dialog = closeLoginDialog
}
