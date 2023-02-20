import { createStore } from 'vuex'
import http from "../request"

const store = createStore({
  state: {
    userInfo: {}
  },
  mutations: {
    setUserInfo (state, payload) {
      state.userInfo = payload
    }
  },
  actions: {
    getUserInfo(context) {
      return http.get('/auth/user/info').then(res => {
        console.log(res)
        context.commit('setUserInfo', res)
      })
    }
  }
})

export default store;
