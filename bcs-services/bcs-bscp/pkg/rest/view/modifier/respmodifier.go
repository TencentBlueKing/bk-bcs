package modifier

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	pbcs "bscp.io/pkg/protocol/config-server"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// ModifyRespFunc is a function to modify response
type ModifyRespFunc func(resp []byte) ([]byte, error)

// AllModifyRespFunc is all functions for protobuf message which need to modify response
var AllModifyRespFunc = map[string]ModifyRespFunc{
	string(proto.MessageName(&pbcs.ListContentsResp{})): modifyListContentsResp,
	string(proto.MessageName(&pbcs.ListCommitsResp{})):  modifyListCommitsResp,
}

// modifyListContentsResp convert byte_size type from string to int64
// see grpc-gateway issue: https://github.com/grpc-ecosystem/grpc-gateway/issues/296
func modifyListContentsResp(resp []byte) ([]byte, error) {
	js := string(resp)
	result := gjson.Get(js, "details.#.spec.byte_size")
	if !result.Exists() {
		return nil, fmt.Errorf("can't find json path details.#.spec.byte_size in response")
	}

	destJs := js
	rs := result.Array()
	var err error
	for i, r := range rs {
		// convert byte_size type from string to int64
		destJs, err = sjson.Set(destJs, fmt.Sprintf("details.%d.spec.byte_size", i), r.Int())
		if err != nil {
			return nil, err
		}
	}
	return []byte(destJs), nil
}

// modifyListCommitsResp convert byte_size type from string to int64
// see grpc-gateway issue: https://github.com/grpc-ecosystem/grpc-gateway/issues/296
func modifyListCommitsResp(resp []byte) ([]byte, error) {
	js := string(resp)
	result := gjson.Get(js, "details.#.spec.content.byte_size")
	if !result.Exists() {
		return nil, fmt.Errorf("can't find json path details.#.spec.content.byte_size in response")
	}

	destJs := js
	rs := result.Array()
	var err error
	for i, r := range rs {
		// convert byte_size type from string to int64
		destJs, err = sjson.Set(destJs, fmt.Sprintf("details.%d.spec.content.byte_size", i), r.Int())
		if err != nil {
			return nil, err
		}
	}
	return []byte(destJs), nil
}
