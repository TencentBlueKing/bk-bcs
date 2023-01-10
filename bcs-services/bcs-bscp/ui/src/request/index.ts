import axios from 'axios';
import jsCookie from 'js-cookie';
import BkMessage from 'bkui-vue/lib/message';

export const getEnvParams = () => {
  const { BK_CC_API, BK_BCS_BSCP_API, BK_APP_CODE, BK_APP_SECRET } = window as any;
  return { BK_CC_API, BK_BCS_BSCP_API, BK_APP_CODE, BK_APP_SECRET };
}

export const EnvParams = {
  ...getEnvParams()
};

export const BASE_URI: any = {
  PRO: {
    CC_API: EnvParams.BK_CC_API,
    BK_BCS_BSCP_API: EnvParams.BK_BCS_BSCP_API,
  },
  DEV: {
    CC_API: EnvParams.BK_CC_API,
    BK_BCS_BSCP_API: EnvParams.BK_BCS_BSCP_API,
  },
};

const getRequestConfig = (method: string, params: any, appendDefParam = true) => {
  if (/^put|post|delete|patch$/gi.test(method)) {
    return {
      data: {
        ...(appendDefParam ? getSharedParams() : {}),
        ...params,
      },
    };
  }

  return {
    params,
  };
};

const getSharedParams = () => {
  const curLang = jsCookie.get('blueking_language');

  return {
    // bk_app_code: EnvParams.BK_APP_CODE,
    // bk_app_secret: EnvParams.BK_APP_SECRET,
    // bk_ticket: jsCookie.get('bk_ticket'),
    // bk_username: jsCookie.get('bk_uid'),
  };
};

export interface IHttpResponse {
  result: boolean,
  code: number,
  message: string,
  data: any,
  request_id: string,
  permission: Record<string, any>
}

export interface IBkHttpResponse {
  response: IHttpResponse,
  isSuccess: (validate?: (resp: IHttpResponse) => boolean) => boolean,
  validate: (showSuccess?: boolean, showError?: boolean, validateFn?: (resp: IHttpResponse) => boolean) => Promise<any>
}

export class BkHttpResponse {
  response: IHttpResponse;
  constructor(resp: IHttpResponse) {
    this.response = resp;
  }

  isSuccess(validate?: (resp: IHttpResponse) => boolean) {
    if (validate !== undefined && validate !== null) {
      if (typeof validate === 'function') {
        return validate(this.response);
      }

      return validate;
    }

    return this.response.code === 0 || this.response.result;
  }

  validate(showSuccess = false, showError = true, validateFn?: (resp: IHttpResponse) => boolean) {
    return new Promise((resolve, reject) => {
      if (this.isSuccess(validateFn)) {
        if (showSuccess) {
          BkMessage({
            message: '请求成功',
            theme: 'success'
          });
        }
        resolve(this.response.data);

      } else {
        BkMessage({
          message: this.response.message,
          theme: 'error'
        });
        reject(this.response.message);
      }
    })
  }
}

export const BkRequest = (name: string, params: any, method: string = 'post', baseURL = '/', appendDefParam = true) =>
  axios
    .create({
      baseURL,
      withCredentials: true,
      headers: {
        'X-Requested-With': 'XMLHttpRequest',
      },
      responseType: 'json',
    })
    .request({
      method,
      url: name,
      ...getRequestConfig(method, params, appendDefParam),
    }).then(res => Promise.resolve(new BkHttpResponse(res.data)))
    .catch(err => {
      BkMessage({
        message: err.message || err,
        theme: 'error'
      });
      return Promise.reject(err);
    });

export const CC_Request = (name: string, params: any = {}, method: string = 'get', env = 'DEV') =>
  BkRequest(name, params, method, `${BASE_URI[env]?.BK_BCS_BSCP_API}/api/c/compapi/v2/cc/`);

export const Self_Request = (name: string, params: any = {}, method: string = 'get', env = 'DEV') =>
  BkRequest(name, params, method, `${BASE_URI[env]?.BK_BCS_BSCP_API}/api/v1/`);
