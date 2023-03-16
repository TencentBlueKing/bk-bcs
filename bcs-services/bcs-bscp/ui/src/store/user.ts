import { defineStore } from 'pinia'
import { ref } from 'vue'
import http from '../request'

interface IUserInfo {
  avatar_url: string;
  username: string
}

export const useUserStore = defineStore('user', () => {
  const loginUrl = ref('')
  const showLoginModal = ref(false)
  const userInfo = ref<IUserInfo>({
    avatar_url: '',
    username: ''
  })

  const getUserInfo = () => {
    return http.get('/auth/user/info').then(res => {
      userInfo.value = <IUserInfo>res.data
    })
  }

  return { loginUrl, showLoginModal, userInfo, getUserInfo }
})