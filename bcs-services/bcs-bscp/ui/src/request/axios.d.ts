import * as axios from 'axios';

declare module 'axios' {
  type AxiosInstance = (config: AxiosRequestConfig) => Promise<any>
}
