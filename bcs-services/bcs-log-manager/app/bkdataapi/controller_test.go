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

package bkdataapi

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	moc_bkdata "github.com/Tencent/bk-bcs/bcs-common/pkg/esb/apigateway/bkdata/mock"
	bcsv1 "github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/pkg/apis/bkbcs.tencent.com/v1"
	apiextensionclientset "github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/pkg/mock/apiextension/clientset"
	apiextensionclientsetv1beta1 "github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/pkg/mock/apiextension/clientset/v1beta1"
	bkdataclientset "github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/pkg/mock/bkdataapiconfig/clientset"
	bkdataclientsetv1 "github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/pkg/mock/bkdataapiconfig/clientset/v1"
	bkdatainformerf "github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/pkg/mock/bkdataapiconfig/informer/factory"
	bkdatainformerfb "github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/pkg/mock/bkdataapiconfig/informer/factory/bkbcs"
	bkdatainformerfbv1 "github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/pkg/mock/bkdataapiconfig/informer/factory/bkbcs/v1"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/pkg/mock/informer"
)

// TestObtainDataid test obtain dataid method
func TestObtainDataid(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	// bkdataapiclient
	mockCreator := moc_bkdata.NewMockClientCreatorInterface(ctrl)
	mockClient := moc_bkdata.NewMockClientInterface(ctrl)
	errRet := fmt.Errorf("error ObtainDataID test")
	mockClient.EXPECT().ObtainDataID(gomock.Any()).Return(int64(-1), errRet)
	mockClient.EXPECT().ObtainDataID(gomock.Any()).Return(int64(21093), nil)
	mockCreator.EXPECT().NewClientFromConfig(gomock.Any()).Return(mockClient).Times(2)

	// apiextensionClientset
	mockApiextensionClientset := apiextensionclientset.NewMockInterface(ctrl)
	mockApiextensionClientsetV1beta1 := apiextensionclientsetv1beta1.NewMockApiextensionsV1beta1Interface(ctrl)
	mockApiextensionClientsetV1beta1I := apiextensionclientsetv1beta1.NewMockCustomResourceDefinitionInterface(ctrl)

	// bkdataapiconfig clientset
	mockBkDataAPIConfigClientset := bkdataclientset.NewMockInterface(ctrl)
	mockBkdataclientsetBV1 := bkdataclientsetv1.NewMockBkbcsV1Interface(ctrl)
	mockBkdataclientsetBV1I := bkdataclientsetv1.NewMockBKDataApiConfigInterface(ctrl)

	// bkdataapiconfig informer factory
	mockBkDataAPIConfigInformerFactory := bkdatainformerf.NewMockSharedInformerFactory(ctrl)
	mockBkdataInformerFB := bkdatainformerfb.NewMockInterface(ctrl)
	mockBkdataInformerFBV1 := bkdatainformerfbv1.NewMockInterface(ctrl)
	mockBkdataInformerFBV1I := bkdatainformerfbv1.NewMockBKDataApiConfigInformer(ctrl)

	// bkdataapiconfig informer
	mockBkDataAPIConfigInformer := informer.NewMockSharedIndexInformer(ctrl)

	// bkdataapiconfig lister
	// mockBkDataApiConfigLister := lister.NewMockBKDataApiConfigLister(ctrl)

	// bkdata informer initialization
	var handlerFuncs cache.ResourceEventHandlerFuncs
	mockBkDataAPIConfigInformer.EXPECT().AddEventHandler(gomock.Any()).Do(func(funcs cache.ResourceEventHandlerFuncs) {
		handlerFuncs = funcs
	}).Times(2)
	mockBkdataInformerFBV1I.EXPECT().Informer().Return(mockBkDataAPIConfigInformer).Times(2)
	// mockBkdataInformerFBV1I.EXPECT().Lister().Return(mockBkDataApiConfigLister).Times(2)
	mockBkdataInformerFBV1.EXPECT().BKDataApiConfigs().Return(mockBkdataInformerFBV1I).Times(2)
	mockBkdataInformerFB.EXPECT().V1().Return(mockBkdataInformerFBV1).Times(2)
	mockBkDataAPIConfigInformerFactory.EXPECT().Bkbcs().Return(mockBkdataInformerFB).Times(2)
	mockBkDataAPIConfigInformerFactory.EXPECT().Start(gomock.Any()).Return().Times(2)
	mockBkDataAPIConfigInformerFactory.EXPECT().WaitForCacheSync(gomock.Any()).Return(nil).Times(2)

	// apiextension clientset initialization
	// already exist
	mockApiextensionClientsetV1beta1I.EXPECT().Create(gomock.Any()).Return(nil, apierrors.NewAlreadyExists(schema.GroupResource{}, ""))
	// error
	mockApiextensionClientsetV1beta1I.EXPECT().Create(gomock.Any()).Return(nil, fmt.Errorf("error test"))
	// normal
	mockApiextensionClientsetV1beta1I.EXPECT().Create(gomock.Any()).Return(nil, nil)
	mockApiextensionClientsetV1beta1.EXPECT().CustomResourceDefinitions().Return(mockApiextensionClientsetV1beta1I).Times(3)
	mockApiextensionClientset.EXPECT().ApiextensionsV1beta1().Return(mockApiextensionClientsetV1beta1).Times(3)

	// bkdataapiconfig clientset initialization
	var cnt = 0
	mockBkdataclientsetBV1I.EXPECT().Update(gomock.Any()).DoAndReturn(func(conf *bcsv1.BKDataApiConfig) (*bcsv1.BKDataApiConfig, error) {
		switch cnt {
		// obtain 失败
		case 0:
			if conf.Spec.Response.Result {
				t.Errorf("BKDataController.v3_access_deploy_plan_post return status(%v), expect status(false)1", conf.Spec.Response.Result)
			}
		// obtain 成功
		case 1:
			if !conf.Spec.Response.Result {
				t.Errorf("BKDataController.v3_access_deploy_plan_post return status(%v), expect status(true)2", conf.Spec.Response.Result)
			}
		// api名称指定错误
		case 2:
			if conf.Spec.Response.Result {
				t.Errorf("BKDataController.v3_access_deploy_plan_post return status(%v), expect status(false)3", conf.Spec.Response.Result)
			}
		}
		cnt++
		return nil, nil
	}).Times(3)
	mockBkdataclientsetBV1.EXPECT().BKDataApiConfigs(gomock.Any()).Return(mockBkdataclientsetBV1I).Times(3)
	mockBkDataAPIConfigClientset.EXPECT().BkbcsV1().Return(mockBkdataclientsetBV1).Times(3)

	controller := &BKDataController{
		StopCh:                         make(chan struct{}),
		ClientCreator:                  mockCreator,
		ApiextensionClientset:          mockApiextensionClientset,
		BkDataAPIConfigInformerFactory: mockBkDataAPIConfigInformerFactory,
		BkDataAPIConfigClientset:       mockBkDataAPIConfigClientset,
		KubeConfig:                     "test",
		RestConfig:                     &rest.Config{},
	}
	// bkdataapiconfigs already exists
	err := controller.Start()
	if err != nil {
		t.Errorf("BKDataController.Start returns error(%+v), expect nil", err)
	}
	// bkdataapiconfigs error
	err = controller.Start()
	if err == nil {
		t.Errorf("BKDataController.Start returns error(%+v), expect error(error test)", err)
	}
	// bkdataapiconfigs custom
	err = controller.Start()
	if err != nil {
		t.Errorf("BKDataController.Start returns error(%+v), expect nil", err)
	}
	handlerFuncs.AddFunc(&bcsv1.BKDataApiConfig{
		Spec: bcsv1.BKDataApiConfigSpec{
			ApiName: "v3_access_deploy_plan_post",
		},
	})
	handlerFuncs.AddFunc(&bcsv1.BKDataApiConfig{
		Spec: bcsv1.BKDataApiConfigSpec{
			ApiName: "v3_access_deploy_plan_post",
		},
	})
	handlerFuncs.AddFunc(&bcsv1.BKDataApiConfig{})
}

// TestSetCleanStrategy test create data clean stategy method
func TestSetCleanStrategy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	// bkdataapiclient
	mockCreator := moc_bkdata.NewMockClientCreatorInterface(ctrl)
	mockClient := moc_bkdata.NewMockClientInterface(ctrl)
	errRet := fmt.Errorf("error SetCleanStrategy test")
	mockClient.EXPECT().SetCleanStrategy(gomock.Any()).Return(errRet)
	mockClient.EXPECT().SetCleanStrategy(gomock.Any()).Return(nil)
	mockCreator.EXPECT().NewClientFromConfig(gomock.Any()).Return(mockClient).Times(2)

	// apiextensionClientset
	mockApiextensionClientset := apiextensionclientset.NewMockInterface(ctrl)
	mockApiextensionClientsetV1beta1 := apiextensionclientsetv1beta1.NewMockApiextensionsV1beta1Interface(ctrl)
	mockApiextensionClientsetV1beta1I := apiextensionclientsetv1beta1.NewMockCustomResourceDefinitionInterface(ctrl)

	// bkdataapiconfig clientset
	mockBkDataAPIConfigClientset := bkdataclientset.NewMockInterface(ctrl)
	mockBkdataclientsetBV1 := bkdataclientsetv1.NewMockBkbcsV1Interface(ctrl)
	mockBkdataclientsetBV1I := bkdataclientsetv1.NewMockBKDataApiConfigInterface(ctrl)

	// bkdataapiconfig informer factory
	mockBkDataAPIConfigInformerFactory := bkdatainformerf.NewMockSharedInformerFactory(ctrl)
	mockBkdataInformerFB := bkdatainformerfb.NewMockInterface(ctrl)
	mockBkdataInformerFBV1 := bkdatainformerfbv1.NewMockInterface(ctrl)
	mockBkdataInformerFBV1I := bkdatainformerfbv1.NewMockBKDataApiConfigInformer(ctrl)

	// bkdataapiconfig informer
	mockBkDataAPIConfigInformer := informer.NewMockSharedIndexInformer(ctrl)

	// bkdataapiconfig lister
	// mockBkDataApiConfigLister := lister.NewMockBKDataApiConfigLister(ctrl)

	// bkdata informer initialization
	var handlerFuncs cache.ResourceEventHandlerFuncs
	mockBkDataAPIConfigInformer.EXPECT().AddEventHandler(gomock.Any()).Do(func(funcs cache.ResourceEventHandlerFuncs) {
		handlerFuncs = funcs
	}).Times(2)
	mockBkdataInformerFBV1I.EXPECT().Informer().Return(mockBkDataAPIConfigInformer).Times(2)
	// mockBkdataInformerFBV1I.EXPECT().Lister().Return(mockBkDataApiConfigLister).Times(2)
	mockBkdataInformerFBV1.EXPECT().BKDataApiConfigs().Return(mockBkdataInformerFBV1I).Times(2)
	mockBkdataInformerFB.EXPECT().V1().Return(mockBkdataInformerFBV1).Times(2)
	mockBkDataAPIConfigInformerFactory.EXPECT().Bkbcs().Return(mockBkdataInformerFB).Times(2)
	mockBkDataAPIConfigInformerFactory.EXPECT().Start(gomock.Any()).Return().Times(2)
	mockBkDataAPIConfigInformerFactory.EXPECT().WaitForCacheSync(gomock.Any()).Return(nil).Times(2)

	// apiextension clientset initialization
	// already exist
	mockApiextensionClientsetV1beta1I.EXPECT().Create(gomock.Any()).Return(nil, apierrors.NewAlreadyExists(schema.GroupResource{}, ""))
	// error
	mockApiextensionClientsetV1beta1I.EXPECT().Create(gomock.Any()).Return(nil, fmt.Errorf("error test"))
	// normal
	mockApiextensionClientsetV1beta1I.EXPECT().Create(gomock.Any()).Return(nil, nil)
	mockApiextensionClientsetV1beta1.EXPECT().CustomResourceDefinitions().Return(mockApiextensionClientsetV1beta1I).Times(3)
	mockApiextensionClientset.EXPECT().ApiextensionsV1beta1().Return(mockApiextensionClientsetV1beta1).Times(3)

	// bkdataapiconfig clientset initialization
	var cnt = 0
	mockBkdataclientsetBV1I.EXPECT().Update(gomock.Any()).DoAndReturn(func(conf *bcsv1.BKDataApiConfig) (*bcsv1.BKDataApiConfig, error) {
		switch cnt {
		// set 失败
		case 0:
			if conf.Spec.Response.Result {
				t.Errorf("BKDataController.v3_databus_cleans_post return status(%v), expect status(false)1", conf.Spec.Response.Result)
			}
		// set 成功
		case 1:
			if !conf.Spec.Response.Result {
				t.Errorf("BKDataController.v3_databus_cleans_post return status(%v), expect status(true)2", conf.Spec.Response.Result)
			}
		// api名称指定错误
		case 2:
			if conf.Spec.Response.Result {
				t.Errorf("BKDataController.v3_databus_cleans_post return status(%v), expect status(false)3", conf.Spec.Response.Result)
			}
		}
		cnt++
		return nil, nil
	}).Times(3)
	mockBkdataclientsetBV1.EXPECT().BKDataApiConfigs(gomock.Any()).Return(mockBkdataclientsetBV1I).Times(3)
	mockBkDataAPIConfigClientset.EXPECT().BkbcsV1().Return(mockBkdataclientsetBV1).Times(3)

	controller := &BKDataController{
		StopCh:                         make(chan struct{}),
		ClientCreator:                  mockCreator,
		ApiextensionClientset:          mockApiextensionClientset,
		BkDataAPIConfigInformerFactory: mockBkDataAPIConfigInformerFactory,
		BkDataAPIConfigClientset:       mockBkDataAPIConfigClientset,
		KubeConfig:                     "test",
		RestConfig:                     &rest.Config{},
	}
	// bkdataapiconfigs already exists
	err := controller.Start()
	if err != nil {
		t.Errorf("BKDataController.Start returns error(%+v), expect nil", err)
	}
	// bkdataapiconfigs error
	err = controller.Start()
	if err == nil {
		t.Errorf("BKDataController.Start returns error(%+v), expect error(error test)", err)
	}
	// bkdataapiconfigs custom
	err = controller.Start()
	if err != nil {
		t.Errorf("BKDataController.Start returns error(%+v), expect nil", err)
	}
	handlerFuncs.AddFunc(&bcsv1.BKDataApiConfig{
		Spec: bcsv1.BKDataApiConfigSpec{
			ApiName: "v3_databus_cleans_post",
		},
	})
	handlerFuncs.AddFunc(&bcsv1.BKDataApiConfig{
		Spec: bcsv1.BKDataApiConfigSpec{
			ApiName: "v3_databus_cleans_post",
		},
	})
	handlerFuncs.AddFunc(&bcsv1.BKDataApiConfig{})
}
