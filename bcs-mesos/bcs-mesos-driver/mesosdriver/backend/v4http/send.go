/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package v4http

import (
	"encoding/json"
	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	bhttp "github.com/Tencent/bk-bcs/bcs-common/common/http"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	"github.com/bitly/go-simplejson"
)

func (s *Scheduler) SendMessageApplication(ns, name, taskgroupId string, body []byte) (string, error) {
	blog.Info("send message to application (%s.%s) taskgroup (%s) param(%s)", ns, name, taskgroupId, string(body))

	//encoding the parameter of sending message
	var param SendMsgOpeParam
	if err := json.Unmarshal(body, &param); err != nil {
		blog.Error("parse sending message operation parameters failed, err(%s)", err.Error())
		err = bhttp.InternalError(common.BcsErrCommJsonDecode, common.BcsErrCommJsonDecodeStr+err.Error())
		return err.Error(), err
	}

	// message data
	msgData, err := json.Marshal(param.MsgData)
	if err != nil {
		blog.Error("encode message data to json failed, err:%s", err.Error())
		err = bhttp.InternalError(common.BcsErrCommJsonEncode, common.BcsErrCommJsonEncodeStr+err.Error())
		return err.Error(), err
	}

	var sendData []byte
	var reply string
	var rpyErr error
	//deal with different message type
	switch param.MsgType {
	case "local":
		sendData, reply, rpyErr = s.makeMsg_Local(msgData)
	case "remote":
		sendData, reply, rpyErr = s.makeMsg_Remote(msgData)
	case "signal":
		sendData, reply, rpyErr = s.makeMsg_Signal(msgData)
	case "env-key":
		sendData, reply, rpyErr = s.makeMsg_EnvKey(msgData)
	default: //unkown message type
		blog.Error("unkown message type(%s) which will be sent to application(%s) under runAs(%s)", param.MsgType, param.Name, param.RunAs)
		err = bhttp.InternalError(common.BcsErrMesosDriverSendMsgUnknowType, common.BcsErrMesosDriverSendMsgUnknowTypeStr)
		return err.Error(), err
	}

	if rpyErr != nil {
		return string(reply), rpyErr
	}

	if s.GetHost() == "" {
		blog.Error("no scheduler is connected by driver")
		err := bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+"scheduler not exist")
		return err.Error(), err
	}

	url := s.GetHost() + "/v1/apps/" + ns + "/" + name + "/message"
	if "" != taskgroupId {
		url = url + "/" + taskgroupId
	}
	blog.Info("post a request to url(%s), request(%s)", url, string(sendData))

	rpyPost, rpyError := s.client.POST(url, nil, sendData)
	if rpyError != nil {
		blog.Error("post request to url(%s) failed! err(%s)", url, rpyError.Error())
		err = bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+rpyError.Error())
		return err.Error(), err
	}

	return string(rpyPost), nil
}

func (s *Scheduler) makeMsg_Local(data []byte) ([]byte, string, error) {
	js, err := simplejson.NewJson(data)
	if err != nil {
		blog.Error("parse local message failed, data(%s), err(%s)", string(data), err.Error())
		err = bhttp.InternalError(common.BcsErrCommJsonDecode, common.BcsErrCommJsonDecodeStr+err.Error())
		return nil, err.Error(), err
	}

	to, _ := js.Get("to").String()
	user, _ := js.Get("user").String()
	right, _ := js.Get("right").String()
	ctx, _ := js.Get("ctx").String()

	var msg types.BcsMessage
	msg.Type = types.Msg_LOCALFILE.Enum()
	msg.Local = &types.Msg_LocalFile{
		To:     &to,
		User:   &user,
		Right:  &right,
		Base64: &ctx,
	}

	msgData, err := json.Marshal(&msg)
	if err != nil {
		blog.Error("encode bcs local message failed, err(%s)", err.Error())
		err = bhttp.InternalError(common.BcsErrCommJsonEncode, common.BcsErrCommJsonEncodeStr+err.Error())
		return nil, err.Error(), err
	}

	return msgData, "", nil
}

func (s *Scheduler) makeMsg_Remote(data []byte) ([]byte, string, error) {
	js, err := simplejson.NewJson(data)
	if err != nil {
		blog.Error("parse remote message failed, data(%s), err(%s)", string(data), err.Error())
		err = bhttp.InternalError(common.BcsErrCommJsonDecode, common.BcsErrCommJsonDecodeStr+err.Error())
		return nil, err.Error(), err
	}

	to, _ := js.Get("to").String()
	user, _ := js.Get("user").String()
	right, _ := js.Get("right").String()
	remote, _ := js.Get("remote").String()

	var msg types.BcsMessage
	msg.Type = types.Msg_REMOTE.Enum()
	msg.Remote = &types.Msg_Remote{
		To:    &to,
		User:  &user,
		Right: &right,
		From:  &remote,
	}

	msgData, err := json.Marshal(&msg)
	if err != nil {
		blog.Error("encode bcs remote message failed, err(%s)", err.Error())
		err = bhttp.InternalError(common.BcsErrCommJsonEncode, common.BcsErrCommJsonEncodeStr+err.Error())
		return nil, err.Error(), err
	}

	return msgData, "", nil
}

func (s *Scheduler) makeMsg_Signal(data []byte) ([]byte, string, error) {
	js, err := simplejson.NewJson(data)
	if err != nil {
		blog.Error("parse signal message failed, data(%s), err(%s)", string(data), err.Error())
		err = bhttp.InternalError(common.BcsErrCommJsonDecode, common.BcsErrCommJsonDecodeStr+err.Error())
		return nil, err.Error(), err
	}

	sig, err := js.Get("signal").Int()
	if err != nil {
		blog.Error("get signal from message failed, data(%s), err(%s)", string(data), err.Error())
		err = bhttp.InternalError(common.BcsErrCommJsonDecode, common.BcsErrCommJsonDecodeStr+err.Error())
		return nil, err.Error(), err
	}

	processname, _ := js.Get("processname").String()

	signal := uint32(sig)
	var msg types.BcsMessage
	msg.Type = types.Msg_SIGNAL.Enum()
	msg.Sig = &types.Msg_Signal{
		ProcessName: &processname,
		Signal:      &signal,
	}

	msgData, err := json.Marshal(&msg)
	if err != nil {
		blog.Error("encode bcs signal message failed, err(%s)", err.Error())
		err = bhttp.InternalError(common.BcsErrCommJsonEncode, common.BcsErrCommJsonEncodeStr+err.Error())
		return nil, err.Error(), err
	}

	return msgData, "", nil
}

func (s *Scheduler) makeMsg_EnvKey(data []byte) ([]byte, string, error) {
	js, err := simplejson.NewJson(data)
	if err != nil {
		blog.Error("parse env-key message failed, data(%s), err(%s)", string(data), err.Error())
		err = bhttp.InternalError(common.BcsErrCommJsonDecode, common.BcsErrCommJsonDecodeStr)
		return nil, err.Error(), err
	}

	envKey, _ := js.Get("env-key").String()
	envValue, _ := js.Get("env-value").String()
	//ifRep, _ := js.Get("append").Bool()

	var msg types.BcsMessage
	msg.Type = types.Msg_ENV.Enum()
	msg.Env = &types.Msg_Env{
		Name:  &envKey,
		Value: &envValue,
		//Rep:   ifRep,
	}

	msgData, err := json.Marshal(&msg)
	if err != nil {
		blog.Error("encode bcs env-key message failed, err(%s)", err.Error())
		err = bhttp.InternalError(common.BcsErrCommJsonEncode, common.BcsErrCommJsonEncodeStr+err.Error())
		return nil, err.Error(), err
	}

	return msgData, "", nil
}
