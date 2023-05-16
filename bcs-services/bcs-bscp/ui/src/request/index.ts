import axios from 'axios';
import BkMessage from 'bkui-vue/lib/message';
import { pinia } from '../store/index';
import { useUserStore } from '../store/user';
import { useGlobalStore } from '../store/global';


const http = axios.create({
  baseURL: `${(<any>window).BK_BCS_BSCP_API}/api/v1`,
  withCredentials: true
})

// 错误处理
http.interceptors.response.use(
  (response) => {
      // response.data 目前有一下几种情况
      // 1. data为string类型，下载配置内容
      // 2. data为object类型，{ code?, data?, message }
      const { data } = response

      if (Object.prototype.toString.call(data) === '[Object object]' && 'code' in data && data.code !== 0) {
        BkMessage({ theme: 'error', message: data.message, ellipsisLine: 3 })
        return Promise.reject(data.message)
      }
      return response.data
  },
  error => {
      const { response } = error
      if (response) {
          const userStore = useUserStore(pinia)
          const globalStore = useGlobalStore(pinia)
          let message = response.statusText

          if (response.status === 401) {
              userStore.$patch((state) => {
                state.loginUrl = response.data.error.data.login_url
                state.showLoginModal = true
              })
              return
          } else if (response.status === 403) {
            globalStore.$patch((state) => {
              state.showSpacePermApply = true
              state.applyPermUrl = response.data.error.data.apply_url
            })
            return response.data.error
          }
          if (response.data.error) {
            message = response.data.error.message
          } else if (response.data.message) {
              message = response.data.message
          }
          BkMessage({ theme: 'error', message, ellipsisLine: 3 })
      }
      return Promise.reject(error)
  }
)

export default http;
