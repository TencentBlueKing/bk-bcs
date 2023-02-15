import axios from 'axios';
import BkMessage from 'bkui-vue/lib/message';
import { propsMixin } from 'bkui-vue/lib/modal';

const http = axios.create({
  // @ts-ignore
  baseURL: `${window.BK_BCS_BSCP_API}/api/v1`
})

// 错误处理
http.interceptors.response.use(
  (response) => {
      const { data } = response
      if (data.code === 0) {
        return data.data
      }
      BkMessage({ theme: 'error', message: data.message, ellipsisLine: 3 })
      return Promise.reject(data.message)
  },
  error => {
      const { response } = error
      if (response) {
          let message = response.statusText
          if (response.data && response.data.message) {
              message = response.data.message
          }
          BkMessage({ theme: 'error', message, ellipsisLine: 3 })
      }
      return Promise.reject(error)
  }
)

export default http;
