package bcs_storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Response struct {
	Result   bool            `json:"result"`
	Code     int             `json:"code"`           //operation code
	Message  string          `json:"message"`        //response message
	Data     json.RawMessage `json:"data,omitempty"` //response data
	Total    int32           `json:"total"`
	PageSize int32           `json:"pageSize"`
	Offset   int32           `json:"offset"`
}

type ResponseDataList []ResponseData

type ResponseData struct {
	Data         json.RawMessage `json:"data,omitempty"`
	UpdateTime   string          `json:"updateTime"`
	Id           string          `json:"_id"`
	ClusterId    string          `json:"clusterId""`
	Namespace    string          `json:"namespace"`
	ResourceName string          `json:"resourceName"`
	ResourceType string          `json:"resourceType"`
	CreateTime   string          `json:"createTime"`
}

func DecodeResp(response http.Response) ([]ResponseData, error) {

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		fmt.Printf("http storage get failed, code: %d, message: %s\n", response.StatusCode, response.Status)
		return nil, fmt.Errorf("remote err, code: %d, status: %s", response.StatusCode, response.Status)
	}
	rawData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("http storage get http status success, but read response body failed, %s\n", err)
		return nil, err
	}

	//format http response
	standarResponse := &Response{}
	if err := json.Unmarshal(rawData, standarResponse); err != nil {
		fmt.Printf("http storage decode GET %s http response failed, %s\n", "standarResponse", err)
		return nil, err
	}
	if standarResponse.Code != 0 {
		fmt.Printf("http storage GET failed, %s\n", standarResponse.Message)
		return nil, fmt.Errorf("remote err: %s", standarResponse.Message)
	}
	if len(standarResponse.Data) == 0 {
		fmt.Println("http storage GET success, but got no data")
		return nil, errors.New("Previous data err. ")
	}

	var responseData []ResponseData
	if err := json.Unmarshal(standarResponse.Data, &responseData); err != nil {
		fmt.Printf("http storage decode data object %s failed, %s\n", "responsedata", err)
		return nil, fmt.Errorf("json decode: %s", err)
	}
	return responseData, err
}
