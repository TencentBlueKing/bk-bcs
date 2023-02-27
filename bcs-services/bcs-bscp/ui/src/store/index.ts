import { createStore } from 'vuex'
import http from "../request"

const store = createStore({
  state: {
    loginUrl: '',
    showLoginModal: false,
    userInfo: {}
  },
  mutations: {
    handleLogin (state, url) {
      state.loginUrl = url
      state.showLoginModal = true
    },
    setUserInfo (state, payload) {
      state.userInfo = payload
    }
  },
  actions: {
    getUserInfo(context) {
      return http.get('/auth/user/info').then(res => {
        context.commit('setUserInfo', res.data)
      })
    }
  }
})

export default store;
