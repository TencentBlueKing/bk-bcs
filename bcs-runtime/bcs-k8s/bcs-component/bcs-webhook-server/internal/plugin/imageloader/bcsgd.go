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
 */

package imageloader

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	tkexv1alpha1 "github.com/Tencent/bk-bcs/bcs-scenarios/kourse/pkg/apis/tkex/v1alpha1"
	bcsclient "github.com/Tencent/bk-bcs/bcs-scenarios/kourse/pkg/client/clientset/versioned"
	bcssche "github.com/Tencent/bk-bcs/bcs-scenarios/kourse/pkg/client/clientset/versioned/scheme"
	informers "github.com/Tencent/bk-bcs/bcs-scenarios/kourse/pkg/client/informers/externalversions"
	bcsgdlister "github.com/Tencent/bk-bcs/bcs-scenarios/kourse/pkg/client/listers/tkex/v1alpha1"
	jsonpatch "github.com/evanphx/json-patch"
	"k8s.io/api/admission/v1beta1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	"k8s.io/utils/pointer"
)

const (
	// imageUpdateAnno holds container's image update info in json-patch format.
	imageUpdateAnno = "IMAGE_UPDATE"
	// imagePreloadDoneLabel represents this is a update trigger by imageloader after image loaded.
	imagePreloadDoneLabel = "IMAGE_PRELOAD_DONE"
	// jobNameLabel holds job's name for anti affinity
	jobNameLabel = "JOB_NAME"
	// jobOnNodeAnno holds nodes which job should run on
	jobOnNodeAnno = "ON_NODE"
)

var (
	bcsCodec = bcssche.Codecs.LegacyCodec(tkexv1alpha1.SchemeGroupVersion)
)

type bcsgdWorkload struct {
	name string

	client   bcsclient.Interface
	informer cache.SharedIndexInformer
	lister   bcsgdlister.GameDeploymentLister

	i *imageLoader
}

// Name returns name the the workload.
func (b *bcsgdWorkload) Name() string {
	return b.name
}

// Init inits the workload's informer.
func (b *bcsgdWorkload) Init(i *imageLoader) error {
	b.name = metav1.GroupVersionKind{
		Group:   tkexv1alpha1.GroupVersion.Group,
		Version: tkexv1alpha1.GroupVersion.Version,
		Kind:    tkexv1alpha1.KindGameDeployment,
	}.String()
	b.i = i

	var err error
	b.client, err = bcsclient.NewForConfig(i.kubeConfig)
	if err != nil {
		blog.Errorf("%v", err)
		return err
	}
	blog.Info("connect to k8s with bcsgd client success")

	informerFactory := informers.NewSharedInformerFactory(b.client, 0)
	b.informer = informerFactory.Tkex().V1alpha1().GameDeployments().Informer()
	// set gamedeployment lister
	b.lister = informerFactory.Tkex().V1alpha1().GameDeployments().Lister()
	// start informer
	informerFactory.Start(b.i.stopCh)

	return nil
}

// LoadImageBeforeUpdate prevents workload instance from updating and
// create a image-load job.
// nolint funlen
func (b *bcsgdWorkload) LoadImageBeforeUpdate(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	// get req gd
	newGD := &tkexv1alpha1.GameDeployment{}
	if err := json.Unmarshal(ar.Request.Object.Raw, newGD); err != nil {
		return toAdmissionResponse(fmt.Errorf("get new gd failed: %v", err))
	}

	// only inplace-update need image loader
	if newGD.Spec.UpdateStrategy.Type != tkexv1alpha1.InPlaceGameDeploymentUpdateStrategyType {
		return toAdmissionResponse(nil)
	}

	// get old gd
	oldGD := &tkexv1alpha1.GameDeployment{}
	if err := json.Unmarshal(ar.Request.OldObject.Raw, oldGD); err != nil {
		return toAdmissionResponse(fmt.Errorf("get old gd failed: %v", err))
	}

	blog.V(3).Infof("bcsgd %s/%s updating", newGD.Namespace, newGD.Name)

	// diff new and old images to get patch on new and jobContainers to pull images
	originalDiffPatch, originalDiffJobContainers := diff(oldGD.ObjectMeta, newGD.ObjectMeta,
		oldGD.Spec.Template.Spec.Containers, newGD.Spec.Template.Spec.Containers)
	// apply final patch
	var finalPatch string
	// craete finalJob if it is not nil
	var finalJob *batchv1.Job
	var needToDeleteJob bool
	var err error

	// update but without container change
	if len(originalDiffJobContainers) == 0 {
		blog.V(3).Infof("recevice update requests but wihout changing container, gd: %s/%s",
			newGD.Namespace, newGD.Name)
		return toAdmissionResponse(nil, finalPatch)
	}

	updatePatch, ok := oldGD.Annotations[imageUpdateAnno]
	if !ok {
		// new image update request from user, use original diff, and create job
		blog.V(3).Infof("new image update request from user of bcsgd %s/%s", newGD.Namespace, newGD.Name)
		finalPatch = originalDiffPatch
		finalJob, err = b.generateJobByDiff(newGD, originalDiffJobContainers)
	} else {
		if _, ok = newGD.Labels[imagePreloadDoneLabel]; ok {
			// image has loaded, trigger the real update, delete the job
			// remove imageUpdateAnno and imagePreloadDoneLabel and return patch
			blog.V(3).Infof("trigger real update, bcsgd %s/%s", newGD.Namespace, newGD.Name)
			patchs := []string{
				fmt.Sprintf("{\"op\":\"remove\",\"path\":\"/metadata/annotations/%s\"}", imageUpdateAnno),
				fmt.Sprintf("{\"op\":\"remove\",\"path\":\"/metadata/labels/%s\"}", imagePreloadDoneLabel),
			}
			finalPatch = fmt.Sprintf("[%s]", strings.Join(patchs, ","))
			needToDeleteJob = true
		} else {
			// user request during image loading
			// caculate diff between new and old's imageUpdateAnno
			// patch imageUpdateAnno to updateOld, then diff new and updateOld
			updateOld, applyErr := applyPatchToGD(oldGD, updatePatch)
			if applyErr != nil {
				blog.Errorf("apply patch %s to bcsgd %s/%s failed: %v", updatePatch, oldGD.Namespace, oldGD.Name, applyErr)
				return toAdmissionResponse(applyErr)
			}
			_, updateDiffJobContainers := diff(updateOld.ObjectMeta, newGD.ObjectMeta,
				updateOld.Spec.Template.Spec.Containers, newGD.Spec.Template.Spec.Containers)
			if len(updateDiffJobContainers) == 0 {
				// no diff between user request and current update
				// use original diff, create job if not exist
				finalJob, err = b.generateJobByDiff(newGD, originalDiffJobContainers)
				if err != nil {
					blog.Errorf("generate job by original diff failed: %v", err)
					return toAdmissionResponse(err)
				}
				blog.V(3).Infof("user request original update, bcsgd: %s/%s", newGD.Namespace, newGD.Name)
				finalPatch = originalDiffPatch
			} else if len(originalDiffJobContainers) != 0 {
				// real new update request, use original diff, create the job
				blog.V(3).Infof("user request new update, bcsgd %s/%s", newGD.Namespace, newGD.Name)
				finalPatch = originalDiffPatch
				finalJob, err = b.generateJobByDiff(newGD, originalDiffJobContainers)
				if err != nil {
					return toAdmissionResponse(err)
				}
			} else {
				// user revert the current update
				// delete imageUpdateAnno and corresponding job, permit the update
				blog.V(3).Infof("user revert current update, bcsgd %v", newGD)
				finalPatch = fmt.Sprintf("[{\"op\":\"remove\",\"path\":\"/metadata/annotations/%s\"}]", imageUpdateAnno)
				needToDeleteJob = true
			}
		}
	}

	if err != nil {
		return toAdmissionResponse(err)
	}

	if needToDeleteJob {
		jobName := fmt.Sprintf("%s-%s-%s", strings.ToLower(tkexv1alpha1.KindGameDeployment), newGD.Namespace, newGD.Name)
		err = b.i.deleteJob(pluginName, jobName)
		if err != nil {
			blog.Errorf("delete job %s failed: %v", jobName, err)
		}
	}

	if finalJob != nil {
		b.generateStartEvent(newGD)
		err = b.i.createJobIfNeed(finalJob)
		if err != nil {
			blog.Errorf("create job %s/%s failed: %v", finalJob.Namespace, finalJob.Name, err)
		}
	}

	blog.V(3).Infof("finalPatch for GD %s/%s: %s", newGD.Namespace, newGD.Name, finalPatch)
	return toAdmissionResponse(nil, finalPatch)
}

// JobDoneHook is called after image-load job is done.
// This function trigger the update keepgoing or attachs failed event to the workload instance.
func (b *bcsgdWorkload) JobDoneHook(namespace, name string, event *corev1.Event) error {
	// get gd and update
	gd, err := b.lister.GameDeployments(namespace).Get(name)
	if err != nil {
		blog.Errorf("get gd %s-%s failed: %v", namespace, name, err)
		return err
	}
	blog.V(3).Infof("job done, update bcsgd %s/%s", gd.Namespace, gd.Name)

	// handle warning event
	if event != nil {
		// add event to gd and return
		// add object ref
		// cannot get gvk from lister's object, use const instead
		event.Namespace = gd.Namespace
		event.InvolvedObject = corev1.ObjectReference{
			Kind:            tkexv1alpha1.KindGameDeployment,
			Namespace:       gd.Namespace,
			Name:            gd.Name,
			UID:             gd.UID,
			APIVersion:      "tkex.tencent.com/v1alpha1",
			ResourceVersion: gd.ResourceVersion,
		}
		return nil
	}

	updatePatch, ok := gd.Annotations[imageUpdateAnno]
	if !ok {
		// no imageUpdateAnno found, finish the job
		blog.Errorf("no imageUpdateAnno of bcsgd(%s-%s) when job is done", namespace, name)
		return nil
	}
	// add imagePreloadDoneLabel
	// DOTO retry on conflict
	updatePatch = updatePatch[:len(updatePatch)-1] + "," +
		fmt.Sprintf("{\"op\":\"add\",\"path\":\"/metadata/labels/%s\",\"value\":\"1\"}",
			imagePreloadDoneLabel) + "]"
	_, err = b.client.TkexV1alpha1().GameDeployments(gd.Namespace).Patch(context.Background(),
		gd.Name,
		types.JSONPatchType,
		[]byte(updatePatch),
		metav1.PatchOptions{})
	if err != nil {
		blog.Errorf("execute update failed: %v", err)
		return err
	}
	b.generateFinishEvent(gd)
	return nil
}

func applyPatchToGD(old *tkexv1alpha1.GameDeployment, patch string) (*tkexv1alpha1.GameDeployment, error) {
	updateOld := &tkexv1alpha1.GameDeployment{}
	// transfer old object to json
	oldJS, err := runtime.Encode(bcsCodec, old)
	if err != nil {
		return nil, err
	}
	// construct json patch by update patch in annotations
	patchObj, err := jsonpatch.DecodePatch([]byte(patch))
	if err != nil {
		return nil, err
	}
	// apply patch to old object
	patchedJS, err := patchObj.Apply(oldJS)
	if err != nil {
		return nil, err
	}
	// transfer applied object to updateOld
	err = runtime.DecodeInto(bcsCodec, patchedJS, updateOld)
	if err != nil {
		return nil, err
	}
	return updateOld, nil
}

func (b *bcsgdWorkload) generateJobByDiff(
	gd *tkexv1alpha1.GameDeployment, diffContainers []corev1.Container) (*batchv1.Job, error) {
	job := newJob(diffContainers)
	job.Name = fmt.Sprintf("%s-%s-%s", strings.ToLower(tkexv1alpha1.KindGameDeployment), gd.Namespace, gd.Name)
	// add fields to set anti affinity
	job.Labels[jobNameLabel] = job.Name
	job.Labels[workloadInsNameLabel] = gd.Name
	job.Labels[workloadInsNamespaceLabel] = gd.Namespace
	job.Annotations[workloadNameAnno] = b.Name()
	nodes := b.nodesOfGD(gd)
	if len(nodes) == 0 {
		return nil, fmt.Errorf("get nodes of job failed in bcsgd %s-%s", gd.Namespace, gd.Name)
	}
	// add fields to check image on nodes
	job.Annotations[jobOnNodeAnno] = strings.Join(nodes, ",")

	// add fields to set pod number of the job
	var podNumber = int32(len(nodes))
	job.Spec.Parallelism = &podNumber
	job.Spec.Completions = &podNumber

	// add fields to select pod and pod affinity
	job.Spec.Template.Labels[jobNameLabel] = job.Name

	// add affinity to execute job with pod of gd
	// add anti affinity to make sure no two pods of a job execute at same node
	job.Spec.Template.Spec.Affinity = &corev1.Affinity{
		PodAffinity: &corev1.PodAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
				{
					Namespaces:    []string{gd.Namespace},
					LabelSelector: gd.Spec.Selector,
					TopologyKey:   corev1.LabelHostname,
				},
			},
		},
		PodAntiAffinity: &corev1.PodAntiAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
				{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							jobNameLabel: job.Name,
						},
					},
					TopologyKey: corev1.LabelHostname,
				},
			},
		},
	}

	job.Spec.ActiveDeadlineSeconds = pointer.Int64(b.i.config.JobTimeoutSeconds)
	return job, injectImagePullSecrets(b.i.secretLister, b.i.k8sClient, gd.Namespace, job, gd.Spec.Template)
}

func (b *bcsgdWorkload) nodesOfGD(gd *tkexv1alpha1.GameDeployment) []string {
	ret := []string{}
	// get all pods of the gd
	set := labels.Set(gd.Spec.Selector.MatchLabels)
	listOptions := metav1.ListOptions{LabelSelector: set.AsSelector().String()}
	pods, err := b.i.k8sClient.CoreV1().Pods(gd.Namespace).List(context.Background(), listOptions)
	if err != nil {
		blog.Errorf("get pods of bcsgd(%s-%s) failed: %v", gd.Namespace, gd.Name, err)
		return ret
	}
	for _, pod := range pods.Items {
		ret = append(ret, pod.Spec.NodeName)
	}
	return ret
}

// WaitForCacheSync waits the workload informer to be synced.
func (b *bcsgdWorkload) WaitForCacheSync(stopCh chan struct{}) bool {
	return cache.WaitForCacheSync(stopCh, b.informer.HasSynced)
}

func (b *bcsgdWorkload) generateStartEvent(gd *tkexv1alpha1.GameDeployment) {
	event := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: pluginName + "-",
		},
		Reason:  "ImageLoading",
		Message: "Start to preload images",
		Type:    corev1.EventTypeNormal,
		InvolvedObject: corev1.ObjectReference{
			Kind:            gd.Kind,
			Namespace:       gd.Namespace,
			Name:            gd.Name,
			UID:             gd.UID,
			APIVersion:      gd.APIVersion,
			ResourceVersion: gd.ResourceVersion,
		},
		FirstTimestamp: metav1.Now(),
		LastTimestamp:  metav1.Now(),
	}
	_, err := b.i.k8sClient.CoreV1().Events(gd.Namespace).Create(context.Background(), event, metav1.CreateOptions{})
	if err != nil {
		blog.Errorf("create event %v failed: %v", event, err)
	}
}

func (b *bcsgdWorkload) generateFinishEvent(gd *tkexv1alpha1.GameDeployment) {
	event := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: pluginName + "-",
		},
		InvolvedObject: corev1.ObjectReference{
			Kind:            tkexv1alpha1.KindGameDeployment,
			Namespace:       gd.Namespace,
			Name:            gd.Name,
			UID:             gd.UID,
			APIVersion:      "tkex.tencent.com/v1alpha1",
			ResourceVersion: gd.ResourceVersion,
		},
		Reason:         "ImageLoaded",
		Message:        "Successfully preload image",
		Type:           corev1.EventTypeNormal,
		FirstTimestamp: metav1.Now(),
		LastTimestamp:  metav1.Now(),
	}
	_, err := b.i.k8sClient.CoreV1().Events(gd.Namespace).Create(context.Background(), event, metav1.CreateOptions{})
	if err != nil {
		blog.Errorf("create event %v failed: %v", event, err)
	}
}
