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
	bcsgslister "github.com/Tencent/bk-bcs/bcs-scenarios/kourse/pkg/client/listers/tkex/v1alpha1"
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

var (
	bcsgsCodec = bcssche.Codecs.LegacyCodec(tkexv1alpha1.SchemeGroupVersion)
)

type bcsgsWorkload struct {
	name string

	client   bcsclient.Interface
	informer cache.SharedIndexInformer
	lister   bcsgslister.GameStatefulSetLister

	i *imageLoader
}

// Name returns name the the workload.
func (b *bcsgsWorkload) Name() string {
	return b.name
}

// Init inits the workload's informer.
func (b *bcsgsWorkload) Init(i *imageLoader) error {
	b.name = metav1.GroupVersionKind{
		Group:   tkexv1alpha1.GroupVersion.Group,
		Version: tkexv1alpha1.GroupVersion.Version,
		Kind:    tkexv1alpha1.KindGameStatefulSet,
	}.String()
	b.i = i

	var err error
	b.client, err = bcsclient.NewForConfig(i.kubeConfig)
	if err != nil {
		blog.Errorf("%v", err)
		return err
	}
	blog.Info("connect to k8s with bcsgs client success")

	informerFactory := informers.NewSharedInformerFactory(b.client, 0)
	b.informer = informerFactory.Tkex().V1alpha1().GameStatefulSets().Informer()
	// set gamestatefulset lister
	b.lister = informerFactory.Tkex().V1alpha1().GameStatefulSets().Lister()
	// start informer
	informerFactory.Start(b.i.stopCh)

	return nil
}

// LoadImageBeforeUpdate prevents workload instance from updating and
// create a image-load job.
// nolint funlen
func (b *bcsgsWorkload) LoadImageBeforeUpdate(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	// get req gs
	newGS := &tkexv1alpha1.GameStatefulSet{}
	if err := json.Unmarshal(ar.Request.Object.Raw, newGS); err != nil {
		return toAdmissionResponse(fmt.Errorf("get new gs failed: %v", err))
	}

	// only inplace-update need image loader
	if newGS.Spec.UpdateStrategy.Type != tkexv1alpha1.InplaceUpdateGameStatefulSetStrategyType {
		return toAdmissionResponse(nil)
	}

	// get old gs
	oldGS := &tkexv1alpha1.GameStatefulSet{}
	if err := json.Unmarshal(ar.Request.OldObject.Raw, oldGS); err != nil {
		return toAdmissionResponse(fmt.Errorf("get old gs failed: %v", err))
	}

	blog.V(3).Infof("bcsgs %s/%s updating", newGS.Namespace, newGS.Name)

	// diff new and old images to get patch on new and jobContainers to pull images
	originalDiffPatch, originalDiffJobContainers := diff(oldGS.ObjectMeta, newGS.ObjectMeta,
		oldGS.Spec.Template.Spec.Containers, newGS.Spec.Template.Spec.Containers)
	// apply final patch
	var finalPatch string
	// craete finalJob if it is not nil
	var finalJob *batchv1.Job
	var needToDeleteJob bool
	var err error

	// update but without container change
	if len(originalDiffJobContainers) == 0 {
		blog.V(3).Infof("recevice update requests but wihout changing container, gs: %s/%s",
			newGS.Namespace, newGS.Name)
		return toAdmissionResponse(nil, finalPatch)
	}

	updatePatch, ok := oldGS.Annotations[imageUpdateAnno]
	if !ok {
		// new image change request from user
		// use original diff, and create job
		blog.V(3).Infof("new image update request from user of bcsgs %s/%s", newGS.Namespace, newGS.Name)
		finalPatch = originalDiffPatch
		finalJob, err = b.generateJobByDiff(newGS, originalDiffJobContainers)
	} else {
		// during a image loading process
		if _, ok := newGS.Labels[imagePreloadDoneLabel]; ok {
			// image loaded, trigger the real update, delete the job
			// remove imageUpdateAnno and imagePreloadDoneLabel and return patch
			blog.V(3).Infof("trigger real update, bcsgs %s/%s", newGS.Namespace, newGS.Name)
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
			updateOld, applyErr := applyPatchToGS(oldGS, updatePatch)
			if applyErr != nil {
				blog.Errorf("apply patch %s to bcsgs %s/%s failed: %v", updatePatch, oldGS.Namespace, oldGS.Name, applyErr)
				return toAdmissionResponse(applyErr)
			}

			// diff new and updateOld
			_, updateDiffJobContainers := diff(updateOld.ObjectMeta, newGS.ObjectMeta,
				updateOld.Spec.Template.Spec.Containers, newGS.Spec.Template.Spec.Containers)
			if len(updateDiffJobContainers) == 0 {
				// no diff between user request and current update
				finalJob, err = b.generateJobByDiff(newGS, originalDiffJobContainers)
				if err != nil {
					blog.Errorf("generate job by original diff failed: %v", err)
					return toAdmissionResponse(err)
				}
				blog.V(3).Infof("user request original update, bcsgs: %s/%s", newGS.Namespace, newGS.Name)
				finalPatch = originalDiffPatch
			} else if len(originalDiffJobContainers) != 0 {
				// real new update request, use original diff, create the job
				blog.V(3).Infof("user request new update, bcsgs %s/%s", newGS.Namespace, newGS.Name)
				finalPatch = originalDiffPatch
				finalJob, err = b.generateJobByDiff(newGS, originalDiffJobContainers)
				if err != nil {
					return toAdmissionResponse(err)
				}
			} else {
				// user revert the current update
				// delete imageUpdateAnno and corresponding job, permit the update
				blog.V(3).Infof("user revert old update, bcsgs %v", newGS)
				finalPatch = fmt.Sprintf("[{\"op\":\"remove\",\"path\":\"/metadata/annotations/%s\"}]", imageUpdateAnno)
				needToDeleteJob = true
			}
		}
	}

	if err != nil {
		return toAdmissionResponse(err)
	}

	if needToDeleteJob {
		jobName := fmt.Sprintf("%s-%s-%s", strings.ToLower(tkexv1alpha1.KindGameStatefulSet), newGS.Namespace, newGS.Name)
		err = b.i.deleteJob(pluginName, jobName)
		if err != nil {
			blog.Errorf("delete job %s failed: %v", jobName, err)
		}
	}

	if finalJob != nil {
		b.generateStartEvent(newGS)
		err = b.i.createJobIfNeed(finalJob)
		if err != nil {
			blog.Errorf("create job %s/%s failed: %v", finalJob.Namespace, finalJob.Name, err)
		}
	}

	blog.V(3).Infof("finalPatch for GS %s/%s: %s", newGS.Namespace, newGS.Name, finalPatch)
	return toAdmissionResponse(nil, finalPatch)
}

// JobDoneHook is called after image-load job is done.
// This function trigger the update keepgoing or attachs failed event to the workload instance.
func (b *bcsgsWorkload) JobDoneHook(namespace, name string, event *corev1.Event) error {
	// get gs and update
	gs, err := b.lister.GameStatefulSets(namespace).Get(name)
	if err != nil {
		blog.Errorf("get gs %s-%s failed: %v", namespace, name, err)
		return err
	}
	blog.V(3).Infof("job done, update bcsgs %s/%s", namespace, name)

	// handle event
	if event != nil {
		// add event to gs and return
		// add object ref
		// finish the job
		// event.Name = gs.Name + "-imageloadfailed"
		event.Namespace = gs.Namespace
		event.InvolvedObject = corev1.ObjectReference{
			Kind:            tkexv1alpha1.KindGameStatefulSet,
			Namespace:       gs.Namespace,
			Name:            gs.Name,
			UID:             gs.UID,
			APIVersion:      "tkex.tencent.com/v1alpha1",
			ResourceVersion: gs.ResourceVersion,
		}
		return nil
	}

	updatePatch, ok := gs.Annotations[imageUpdateAnno]
	if !ok {
		// no imageUpdateAnno found, finish the job
		blog.Errorf("no imageUpdateAnno of bcsgs(%s-%s) when job is done", namespace, name)
		return nil
	}
	// add imagePreloadDoneLabel
	// DOTO retry on conflict
	updatePatch = updatePatch[:len(updatePatch)-1] + "," +
		fmt.Sprintf("{\"op\":\"add\",\"path\":\"/metadata/labels/%s\",\"value\":\"1\"}",
			imagePreloadDoneLabel) + "]"
	_, err = b.client.TkexV1alpha1().GameStatefulSets(gs.Namespace).Patch(
		context.Background(),
		gs.Name,
		types.JSONPatchType,
		[]byte(updatePatch),
		metav1.PatchOptions{})
	if err != nil {
		blog.Errorf("execute update failed: %v", err)
		return err
	}
	b.generateFinishEvent(gs)
	return nil
}

func applyPatchToGS(old *tkexv1alpha1.GameStatefulSet, patch string) (*tkexv1alpha1.GameStatefulSet, error) {
	updateOld := &tkexv1alpha1.GameStatefulSet{}
	// transfer old object to json
	oldJS, err := runtime.Encode(bcsgsCodec, old)
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
	err = runtime.DecodeInto(bcsgsCodec, patchedJS, updateOld)
	if err != nil {
		return nil, err
	}
	return updateOld, nil
}

func (b *bcsgsWorkload) generateJobByDiff(
	gs *tkexv1alpha1.GameStatefulSet, diffContainers []corev1.Container) (*batchv1.Job, error) {
	job := newJob(diffContainers)
	job.Name = fmt.Sprintf("%s-%s-%s", strings.ToLower(tkexv1alpha1.KindGameStatefulSet), gs.Namespace, gs.Name)
	// add fields to set anti affinity
	job.Labels[jobNameLabel] = job.Name
	job.Labels[workloadInsNameLabel] = gs.Name
	job.Labels[workloadInsNamespaceLabel] = gs.Namespace
	job.Annotations[workloadNameAnno] = b.Name()
	nodes := b.nodesOfGS(gs)
	if len(nodes) == 0 {
		return nil, fmt.Errorf("get nodes of job failed in bcsgs %s-%s", gs.Namespace, gs.Name)
	}
	// add fields to check image on nodes
	job.Annotations[jobOnNodeAnno] = strings.Join(nodes, ",")

	// add fields to set pod number of the job
	var podNumber = int32(len(nodes))
	job.Spec.Parallelism = &podNumber
	job.Spec.Completions = &podNumber

	// add fields to select pod and pod affinity
	job.Spec.Template.Labels[jobNameLabel] = job.Name

	// add affinity to execute job with pod of gs
	// add anti affinity to make sure no two pods of a job execute at same node
	job.Spec.Template.Spec.Affinity = &corev1.Affinity{
		PodAffinity: &corev1.PodAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
				{
					Namespaces:    []string{gs.Namespace},
					LabelSelector: gs.Spec.Selector,
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
	return job, injectImagePullSecrets(b.i.secretLister, b.i.k8sClient, gs.Namespace, job, gs.Spec.Template)
}

func (b *bcsgsWorkload) nodesOfGS(gs *tkexv1alpha1.GameStatefulSet) []string {
	ret := []string{}
	// get all pods of the gs
	set := labels.Set(gs.Spec.Selector.MatchLabels)
	listOptions := metav1.ListOptions{LabelSelector: set.AsSelector().String()}
	pods, err := b.i.k8sClient.CoreV1().Pods(gs.Namespace).List(context.Background(), listOptions)
	if err != nil {
		blog.Errorf("get pods of bcsgs(%s-%s) failed: %v", gs.Namespace, gs.Name, err)
		return ret
	}
	for _, pod := range pods.Items {
		ret = append(ret, pod.Spec.NodeName)
	}
	return ret
}

// WaitForCacheSync waits the workload informer to be synced.
func (b *bcsgsWorkload) WaitForCacheSync(stopCh chan struct{}) bool {
	return cache.WaitForCacheSync(stopCh, b.informer.HasSynced)
}

func (b *bcsgsWorkload) generateStartEvent(gs *tkexv1alpha1.GameStatefulSet) {
	event := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: pluginName + "-",
		},
		Reason:  "ImageLoading",
		Message: "Start to preload images",
		Type:    corev1.EventTypeNormal,
		InvolvedObject: corev1.ObjectReference{
			Kind:            gs.Kind,
			Namespace:       gs.Namespace,
			Name:            gs.Name,
			UID:             gs.UID,
			APIVersion:      gs.APIVersion,
			ResourceVersion: gs.ResourceVersion,
		},
		FirstTimestamp: metav1.Now(),
		LastTimestamp:  metav1.Now(),
	}
	_, err := b.i.k8sClient.CoreV1().Events(gs.Namespace).Create(context.Background(), event, metav1.CreateOptions{})
	if err != nil {
		blog.Errorf("create event %v failed: %v", event, err)
	}
}

func (b *bcsgsWorkload) generateFinishEvent(gs *tkexv1alpha1.GameStatefulSet) {
	event := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: pluginName + "-",
		},
		InvolvedObject: corev1.ObjectReference{
			Kind:            tkexv1alpha1.KindGameStatefulSet,
			Namespace:       gs.Namespace,
			Name:            gs.Name,
			UID:             gs.UID,
			APIVersion:      "tkex.tencent.com/v1alpha1",
			ResourceVersion: gs.ResourceVersion,
		},
		Reason:         "ImageLoaded",
		Message:        "Successfully preload image",
		Type:           corev1.EventTypeNormal,
		FirstTimestamp: metav1.Now(),
		LastTimestamp:  metav1.Now(),
	}
	_, err := b.i.k8sClient.CoreV1().Events(gs.Namespace).Create(context.Background(), event, metav1.CreateOptions{})
	if err != nil {
		blog.Errorf("create event %v failed: %v", event, err)
	}
}
