package tbuspp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"k8s.io/klog"
	"net"
	"net/http"
	"net/url"
)

const (
	TbusppControllerService = "tbuspp-controller.tbuspp-system"

)

func CheckCanDelete(podName string, podNameSpace string) bool {
	type Resp struct {
		Code        int    `json:"code"`
		ErrMsg      string `json:"err_msg"`
		//AccessToken string `json:"access_token"`
	}
	res := new(Resp)

	_, err := net.ResolveIPAddr("ip", TbusppControllerService)
	if err != nil {
		klog.Errorf("can not resolve %s , please check", TbusppControllerService)
		// no need check tbuspp before scale
		return false
	}else {
		TbusppControllerServiceUrl := "http://" + TbusppControllerService + ":10086/hpa-reduction/can-reduce"
		resp, err := http.PostForm(TbusppControllerServiceUrl,
			url.Values{
				"pod_name": {podName},
				"pod_namespace": {podNameSpace},
			})
		if err != nil {
			klog.Infof(" http.PostForm error %s", err.Error())
			return false
		}
		defer resp.Body.Close()
		// 验证结果回包
		if resp.StatusCode == 200 {
			respBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				klog.Infof("ReadAll error %s", err.Error())
				return false
			}

			err = json.Unmarshal(respBytes, res)
			if err != nil {
				klog.Infof("Unmarshal error %s", err.Error())
				return false
			}
			if res.Code != 0 {
				klog.Warningf("check scale success,but cannot delete now, err %s, please try later.", res.ErrMsg)
				return false
			} else {
				fmt.Printf("check scale success, delete now.")
				return true
			}
		}else {
			fmt.Printf(" check scale failed, code %d != 200. ", resp.StatusCode)
			return false
		}
	}
}