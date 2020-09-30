mkdir -p bkdataapiconfig/clientset/v1
mkdir -p bkdataapiconfig/informer/factory/bkbcs/v1
mkdir -p informer
mkdir -p apiextension/clientset/v1beta1
mkdir -p manager

# bkdataapiconfig clientset
mockgen -destination ./bkdataapiconfig/clientset/bkdataapiconfigclientset_mock.go -package clientset -copyright_file ../copyright/copyright github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/pkg/generated/clientset/versioned Interface
mockgen -destination ./bkdataapiconfig/clientset/v1/bkdataapiconfigv1_mock.go -package v1 -copyright_file ../copyright/copyright github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/pkg/generated/clientset/versioned/typed/bkbcs.tencent.com/v1 BkbcsV1Interface,BKDataApiConfigInterface

# informer factory
mockgen -destination ./bkdataapiconfig/informer/factory/bkdataapiinformerfactory_mock.go -package factory -copyright_file ../copyright/copyright github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/pkg/generated/informers/externalversions SharedInformerFactory
mockgen -destination ./bkdataapiconfig/informer/factory/bkbcs/bkdataapiinformerfactorybkbcs_mock.go -package bkbcs -copyright_file ../copyright/copyright github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/pkg/generated/informers/externalversions/bkbcs.tencent.com Interface
mockgen -destination ./bkdataapiconfig/informer/factory/bkbcs/v1/bkdataapiinformerfactorybkbcsv1_mock.go -package v1 -copyright_file ../copyright/copyright github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/pkg/generated/informers/externalversions/bkbcs.tencent.com/v1 Interface,BKDataApiConfigInformer

# informer
mockgen -destination ./informer/informer_mock.go -package informer -copyright_file ../copyright/copyright k8s.io/client-go/tools/cache SharedIndexInformer

# bkdataapiconfig lister
mockgen -destination ./bkdataapiconfig/lister/bkdataapiconfiglister_mock.go -package lister -copyright_file ../copyright/copyright github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/pkg/generated/listers/bkbcs.tencent.com/v1 BKDataApiConfigLister

# crd clientset
mockgen -destination ./apiextension/clientset/apiextensionclientset_mock.go -package clientset -copyright_file ../copyright/copyright k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset Interface
mockgen -destination ./apiextension/clientset/v1beta1/apiextensionv1beta_mock.go -package v1beta1 -copyright_file ../copyright/copyright k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1 ApiextensionsV1beta1Interface,CustomResourceDefinitionInterface

# log manager
mockgen -destination ./manager/manager_mock.go -package k8s -copyright_file ../copyright/copyright github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/app/k8s LogManagerInterface