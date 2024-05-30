import { defineStore } from 'pinia';
import { ref } from 'vue';
import http from '../request';

interface IUserInfo {
  avatar_url: string;
  username: string;
}

export default defineStore('user', () => {
  const userInfo = ref<IUserInfo>({
    avatar_url: '',
    username: '',
  });

  const getUserInfo = () =>
    http.get('/auth/user/info').then((res) => {
      userInfo.value = res.data as IUserInfo;
    });

  return { userInfo, getUserInfo };
});
