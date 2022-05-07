/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package filewatcher

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/informers"
	k8scorecliset "k8s.io/client-go/kubernetes"
	k8slistcorev1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

// Lister interface for listing file names
type Lister interface {
	List() (map[string]string, error)
}

// FileLister lists filenames in one directory
type FileLister struct {
	Dir string
}

// NewFileLister create new file lister
func NewFileLister(dir string) *FileLister {
	return &FileLister{
		Dir: dir,
	}
}

// GetDir return directory name
func (fl *FileLister) GetDir() string {
	return fl.Dir
}

// List implements Lister
func (fl *FileLister) List() (map[string]string, error) {
	files, err := ioutil.ReadDir(fl.Dir)
	if err != nil {
		return nil, fmt.Errorf("read dir %s failed, err %s", fl.Dir, err.Error())
	}
	retMap := make(map[string]string)
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		fPath := filepath.Join(fl.Dir, file.Name())
		tmpF, errOpen := os.Open(fPath)
		if errOpen != nil {
			return nil, fmt.Errorf("open file %s failed, err %s", fPath, err.Error())
		}
		dataBytes, errRead := ioutil.ReadAll(tmpF)
		if errRead != nil {
			return nil, fmt.Errorf("read file %s failed, err %s", fPath, err.Error())
		}
		retMap[file.Name()] = string(dataBytes)
	}
	return retMap, nil
}

// SecretLister list kubeconfig files from k8s secret
type SecretLister struct {
	Kubeconfig      string
	SecretName      string
	SecretNamespace string

	secretLister k8slistcorev1.SecretLister
}

// NewSecretLister create secret lister by kubeconfig, secret name and secret namespace
func NewSecretLister(kubeconfig, secretName, secretNamespace string) (*SecretLister, error) {
	var restConfig *rest.Config
	var err error
	if len(kubeconfig) == 0 {
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("use incluster config to list secret failed, err %s", err.Error())
		}
	} else {
		//parse configuration
		restConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
	}
	//initialize k8s client
	cliset, err := k8scorecliset.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	factory := informers.NewSharedInformerFactory(cliset, 0)
	// informer and lister for k8s service
	secretInformer := factory.Core().V1().Secrets().Informer()
	secretLister := factory.Core().V1().Secrets().Lister()

	stopCh := make(chan struct{})
	go secretInformer.Run(stopCh)
	if !cache.WaitForCacheSync(stopCh, secretInformer.HasSynced) {
		return nil, fmt.Errorf("timeout for waiting secret informer synced")
	}
	return &SecretLister{
		Kubeconfig:      kubeconfig,
		SecretName:      secretName,
		SecretNamespace: secretNamespace,
		secretLister:    secretLister,
	}, nil
}

// List implements Lister
func (sl *SecretLister) List() (map[string]string, error) {
	retMap := make(map[string]string)
	secret, err := sl.secretLister.Secrets(sl.SecretNamespace).Get(sl.SecretName)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			zap.L().Warn("secret not found", zap.String("name", sl.SecretName), zap.String("ns", sl.SecretNamespace))
			return retMap, nil
		}
		return nil, fmt.Errorf("get secret %s/%s failed, err %s", sl.SecretName, sl.SecretNamespace, err.Error())
	}
	for key, data := range secret.Data {
		retMap[key] = string(data)
	}
	return retMap, nil
}

// MockLister lists filenames for unit test
type MockLister struct {
	FileMd5Map map[string]string
}

// List implements Lister
func (ml *MockLister) List() (map[string]string, error) {
	return ml.FileMd5Map, nil
}
