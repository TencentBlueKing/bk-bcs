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

package controllers

import (
	"bytes"
	"fmt"
	"text/template"

	bkcmdbv1 "github.com/Tencent/bk-bcs/bcs-resources/bk-cmdb-operator/api/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	defaultIngressDomain = "bk-cmdb.blueking.domain"
)

type CmdbConfig struct {
	MongoHost     string
	MongoUsername string
	MongoPwd      string
	MongoDatabase string
	MongoPort     int32

	RedisHost string
	RedisPwd  string
	RedisPort int32

	ZookeeperHost string
	ZookeeperPort int32

	IngressDomain string
}

type RenderedContent struct {
	ApiserverConf      string
	AuditCtrlConf      string
	CoreConf           string
	DataCollectionConf string
	EventServerConf    string
	HostConf           string
	HostCtrlConf       string
	MigrateConf        string
	ObjectCtrlConf     string
	ProcConf           string
	ProcCtrlConf       string
	TopoConf           string
	TxcConf            string
	WebserverConf      string
	TaskConf           string
	OperationConf      string
}

// reconcileConfigMap reconcile bk-cmdb configmap
func (r *BkcmdbReconciler) reconcileConfigMap(instance *bkcmdbv1.Bkcmdb) error {
	cmdbConfig := r.generateConfig(instance)

	renderedContent, err := r.generateRenderedContent(cmdbConfig)
	if err != nil {
		return err
	}

	cmdbCm := makeCmdbConfigMap(instance, renderedContent)
	if err := controllerutil.SetControllerReference(instance, cmdbCm, r.Scheme); err != nil {
		return fmt.Errorf("failed to set bkcmdb configmap owner reference: %s", err.Error())
	}
	err = r.Client.CreateOrUpdateCm(cmdbCm)
	if err != nil {
		return fmt.Errorf("failed to create or update bkcmdb configmap: %s", err.Error())
	}

	return nil
}

// makeCmdbConfigMap build bk-cmdb configmap object
func makeCmdbConfigMap(z *bkcmdbv1.Bkcmdb, rendered *RenderedContent) *v1.ConfigMap {
	return &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      z.GetName() + "-configures",
			Namespace: z.Namespace,
		},
		Data: map[string]string{
			"apiserver.conf":        rendered.ApiserverConf,
			"auditcontroller.conf":  rendered.AuditCtrlConf,
			"coreservice.conf":      rendered.CoreConf,
			"datacollection.conf":   rendered.DataCollectionConf,
			"eventserver.conf":      rendered.EventServerConf,
			"host.conf":             rendered.HostConf,
			"hostcontroller.conf":   rendered.HostCtrlConf,
			"migrate.conf":          rendered.MigrateConf,
			"objectcontroller.conf": rendered.ObjectCtrlConf,
			"proc.conf":             rendered.ProcConf,
			"proccontroller.conf":   rendered.ProcCtrlConf,
			"topo.conf":             rendered.TopoConf,
			"txc.conf":              rendered.TxcConf,
			"webserver.conf":        rendered.WebserverConf,
			"task.conf":             rendered.TaskConf,
			"operation.conf":        rendered.OperationConf,
		},
	}
}

// generateConfig generates CmdbConfig object
func (r *BkcmdbReconciler) generateConfig(instance *bkcmdbv1.Bkcmdb) *CmdbConfig {
	var mongoPort, redisPort, zkPort int32
	mongoHost := instance.GetName() + "-mongodb"
	mongoUsername := defaultMongoUsername
	mongoPwd := defaultMongoPassword
	mongoDb := defaultMongoDatabase
	mongoPort = defaultMongoPort
	// use custom external mongodb
	if instance.Spec.MongoDb != nil {
		mongoHost = instance.Spec.MongoDb.Host
		mongoUsername = instance.Spec.MongoDb.Username
		mongoPwd = instance.Spec.MongoDb.Password
		mongoDb = instance.Spec.MongoDb.Database
		mongoPort = instance.Spec.MongoDb.Port
	}

	redisHost := instance.GetName() + "-redis-master"
	redisPwd := defaultRedisPassword
	redisPort = defaultRedisPort
	// use custom external redis
	if instance.Spec.Redis != nil {
		redisHost = instance.Spec.Redis.Host
		redisPwd = instance.Spec.Redis.Password
		redisPort = instance.Spec.Redis.Port
	}

	zkHost := instance.GetName() + "-zookeeper"
	zkPort = defaultZkClientPort
	// use custom external zookeeper
	if instance.Spec.Zookeeper != nil {
		zkHost = instance.Spec.Zookeeper.Host
		zkPort = instance.Spec.Zookeeper.Port
	}
	ingressDomain := defaultIngressDomain
	if instance.Spec.IngressDomain != "" {
		ingressDomain = instance.Spec.IngressDomain
	}

	return &CmdbConfig{
		MongoHost:     mongoHost,
		MongoUsername: mongoUsername,
		MongoPwd:      mongoPwd,
		MongoDatabase: mongoDb,
		MongoPort:     mongoPort,
		RedisHost:     redisHost,
		RedisPwd:      redisPwd,
		RedisPort:     redisPort,
		ZookeeperHost: zkHost,
		ZookeeperPort: zkPort,
		IngressDomain: ingressDomain,
	}
}

// renderTemplate render configmap data
func (r *BkcmdbReconciler) renderTemplate(tmplName, contentTemplate string, cmdbConfig *CmdbConfig) (string, error) {
	tmpl, err := template.New(tmplName).Parse(contentTemplate)
	if err != nil {
		return "", err
	}

	var tplOutput bytes.Buffer
	if err := tmpl.Execute(&tplOutput, cmdbConfig); err != nil {
		return "", err
	}

	return tplOutput.String(), nil
}

// generateRenderedContent generate configmap content
func (r *BkcmdbReconciler) generateRenderedContent(cmdbConfig *CmdbConfig) (*RenderedContent, error) {
	auditCtrlConf, err := r.renderTemplate("auditCtrl", auditConfContentTemplate, cmdbConfig)
	if err != nil {
		return nil, fmt.Errorf("error render auditcontroller conf: %s", err.Error())
	}
	coreConf, err := r.renderTemplate("coreservice", coreConfContenetTemplate, cmdbConfig)
	if err != nil {
		return nil, fmt.Errorf("error render coreservice conf: %s", err.Error())
	}
	dataCollectionConf, err := r.renderTemplate("datacollection", dataCnConfContentTemplate, cmdbConfig)
	if err != nil {
		return nil, fmt.Errorf("error render datacollection conf: %s", err.Error())
	}
	eventServerConf, err := r.renderTemplate("eventserver", eventServerConfContentTemplate, cmdbConfig)
	if err != nil {
		return nil, fmt.Errorf("error render eventserver conf: %s", err.Error())
	}
	hostConf, err := r.renderTemplate("host", hostConfContentTemplate, cmdbConfig)
	if err != nil {
		return nil, fmt.Errorf("error render host conf: %s", err.Error())
	}
	hostCtrlConf, err := r.renderTemplate("hostctrl", hostCtrlConfContentTemplate, cmdbConfig)
	if err != nil {
		return nil, fmt.Errorf("error render hostctrl conf: %s", err.Error())
	}
	migrateConf, err := r.renderTemplate("migrate", migrateConfContentTemplate, cmdbConfig)
	if err != nil {
		return nil, fmt.Errorf("error render migrate conf: %s", err.Error())
	}
	objectCtrlConf, err := r.renderTemplate("objectctrl", objectCtrlConfContentTemplate, cmdbConfig)
	if err != nil {
		return nil, fmt.Errorf("error render objectctrl conf: %s", err.Error())
	}
	procConf, err := r.renderTemplate("proc", procConfContentTemplate, cmdbConfig)
	if err != nil {
		return nil, fmt.Errorf("error render proc conf: %s", err.Error())
	}
	procCtrlConf, err := r.renderTemplate("procctrl", procCtrlConfContentTemplate, cmdbConfig)
	if err != nil {
		return nil, fmt.Errorf("error render procctrl conf: %s", err.Error())
	}
	topoConf, err := r.renderTemplate("topo", topoConfContentTemplate, cmdbConfig)
	if err != nil {
		return nil, fmt.Errorf("error render topo conf: %s", err.Error())
	}
	txcConf, err := r.renderTemplate("txc", txcConfContentTemplate, cmdbConfig)
	if err != nil {
		return nil, fmt.Errorf("error render txc conf: %s", err.Error())
	}
	webserverConf, err := r.renderTemplate("webserver", webserverConfContentTemplate, cmdbConfig)
	if err != nil {
		return nil, fmt.Errorf("error render webserver conf: %s", err.Error())
	}
	taskConf, err := r.renderTemplate("task", taskConfContentTemplate, cmdbConfig)
	if err != nil {
		return nil, fmt.Errorf("error render task conf: %s", err.Error())
	}
	operationConf, err := r.renderTemplate("operation", operationConfContentTemplate, cmdbConfig)
	if err != nil {
		return nil, fmt.Errorf("error render operation conf: %s", err.Error())
	}
	return &RenderedContent{
		ApiserverConf:      apiserverConfContent,
		AuditCtrlConf:      auditCtrlConf,
		CoreConf:           coreConf,
		DataCollectionConf: dataCollectionConf,
		EventServerConf:    eventServerConf,
		HostConf:           hostConf,
		HostCtrlConf:       hostCtrlConf,
		MigrateConf:        migrateConf,
		ObjectCtrlConf:     objectCtrlConf,
		ProcConf:           procConf,
		ProcCtrlConf:       procCtrlConf,
		TopoConf:           topoConf,
		TxcConf:            txcConf,
		WebserverConf:      webserverConf,
		TaskConf:           taskConf,
		OperationConf:      operationConf,
	}, nil
}
