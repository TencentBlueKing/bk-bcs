/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
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
	bcsgsapi "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubebkbcs/apis/tkex/v1alpha1"
	bcsgstkexv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubebkbcs/apis/tkex/v1alpha1"
	bcsclient "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubebkbcs/generated/clientset/versioned"
	bcssche "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubebkbcs/generated/clientset/versioned/scheme"
	informers "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubebkbcs/generated/informers/externalversions"
	bcsgslister "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubebkbcs/generated/listers/tkex/v1alpha1"

	jsonpatch "github.com/evanphx/json-patch"
	"k8s.io/api/admission/v1beta1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
)

var (
	bcsgsCodec = bcssche.Codecs.LegacyCodec(bcsgstkexv1alpha1.SchemeGroupVersion)
)

type bcsgsWorkload struct {
	name string

	client   *bcsclient.Clientset
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
		Group:   bcsgsapi.GroupVersion.Group,
		Version: bcsgsapi.GroupVersion.Version,
		Kind:    "GameStatefulSet",
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
func (b *bcsgsWorkload) LoadImageBeforeUpdate(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	// get req gs
	newGS := &bcsgstkexv1alpha1.GameStatefulSet{}
	raw := ar.Request.Object.Raw
	err := json.Unmarshal(raw, newGS)
	if err != nil {
		blog.Errorf("get new gs failed: %v", err)
		return toAdmissionResponse(err)
	}

	// only inplace-update need image loader
	if newGS.Spec.UpdateStrategy.Type != bcsgstkexv1alpha1.InplaceUpdateGameStatefulSetStrategyType {
		return toAdmissionResponse(nil)
	}

	// get old gs
	oldGS := &bcsgstkexv1alpha1.GameStatefulSet{}
	raw = ar.Request.OldObject.Raw
	err = json.Unmarshal(raw, oldGS)
	if err != nil {
		blog.Errorf("get old gs failed: %v", err)
		return toAdmissionResponse(err)
	}

	blog.V(3).Infof("bcsgs %v updating", newGS)

	// diff new and old images to get patch on new and jobContainers to pull images
	originalDiffPatch, originalDiffJobContainers := b.imageDiff(oldGS, newGS)
	// apply final patch
	var finalPatch string
	// craete finalJob if it is not nil
	var finalJob *batchv1.Job
	// delete current job created if this is true
	deleteCurrentJob := false

	if updatePatch, ok := oldGS.Annotations[imageUpdateAnno]; !ok {
		// new image change request from user
		// use original diff, and create job
		blog.V(3).Infof("new image change request from user of bcsgs %v", newGS)
		finalPatch = originalDiffPatch
		finalJob, err = b.generateJobByDiff(newGS, originalDiffJobContainers)
		if err != nil {
			blog.Errorf("generate job by original diff failed: %v", err)
			return toAdmissionResponse(err)
		}
	} else {
		// during a image loading process
		if _, ok := newGS.Labels[imagePreloadDoneLabel]; ok {
			// image loaded, trigger the real update, delete the job
			// remove imageUpdateAnno and imagePreloadDoneLabel and return patch
			blog.V(3).Infof("trigger real update, bcsgs %v", newGS)
			patchs := []string{
				fmt.Sprintf("{\"op\":\"remove\",\"path\":\"/metadata/annotations/%s\"}", imageUpdateAnno),
				fmt.Sprintf("{\"op\":\"remove\",\"path\":\"/metadata/labels/%s\"}", imagePreloadDoneLabel),
			}
			finalPatch = fmt.Sprintf("[%s]", strings.Join(patchs, ","))
			deleteCurrentJob = true
		} else {
			// user request during image loading
			// caculate diff between new and old's imageUpdateAnno
			// patch imageUpdateAnno to updateOld, then diff new and updateOld
			updateOld, err := applyPatchToGS(oldGS, updatePatch)
			if err != nil {
				blog.Errorf("apply patch %s to bcsgs failed: %v", updatePatch, oldGS)
				return toAdmissionResponse(err)
			}

			// diff new and updateOld
			_, updateDiffJobContainers := b.imageDiff(updateOld, newGS)
			if len(updateDiffJobContainers) == 0 {
				// no diff between user request and current update
				// use original diff, TODO create job if not exist
				blog.V(3).Infof("user request original update, bcsgs: %v", newGS)
				finalPatch = originalDiffPatch
			} else {
				// new user request differ from current update
				if len(originalDiffJobContainers) != 0 {
					// real new update request, use original diff, create the job
					blog.V(3).Infof("user request new update, bcsgs %v", newGS)
					// TODO delete old job to avoid repeat name
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
					deleteCurrentJob = true
				}
			}
		}
	}

	if deleteCurrentJob {
		// TODO find the job and delete
	}
	if finalJob != nil {
		go b.i.createJob(finalJob)
	}

	return toAdmissionResponse(nil, finalPatch)
}

// imageDiff diffs old and new gs, and return patch string and containers for job to load images.
func (b *bcsgsWorkload) imageDiff(o, n *bcsgstkexv1alpha1.GameStatefulSet) (string, []corev1.Container) {
	oldContainers := o.Spec.Template.Spec.Containers
	newContainers := n.Spec.Template.Spec.Containers
	// for quick index
	oldImageMap := make(map[string]string)
	for _, c := range oldContainers {
		oldImageMap[c.Name] = c.Image
	}
	// container image update and update patch
	revertPatch := make([]string, len(newContainers)+1)
	imageChangeCount := 0
	// patch to annotations, used for trigger real update
	updatePatch := make([]string, len(newContainers))
	retContainers := make([]corev1.Container, 0)
	for i, c := range newContainers {
		if oi, ok := oldImageMap[c.Name]; ok && oi != c.Image {
			// TODO do not create the job if the image is already on the node
			// this is an image update
			// generate mutate patch
			revertPatch[imageChangeCount] = fmt.Sprintf("{\"op\":\"replace\",\"path\":\"/spec/template/spec/containers/%d/image\",\"value\":\"%s\"}", i, oi)
			updatePatch[imageChangeCount] = fmt.Sprintf("{\"op\":\"replace\",\"path\":\"/spec/template/spec/containers/%d/image\",\"value\":\"%s\"}", i, c.Image)
			imageChangeCount++
			// add a image loader container
			retContainers = append(retContainers,
				corev1.Container{
					Name:            c.Name,
					Image:           c.Image,
					ImagePullPolicy: corev1.PullIfNotPresent,
					Command:         []string{"echo", "pull " + c.Image}})
		}
	}
	// set image patch to label and append label patch
	updatePatchStr := fmt.Sprintf("[%s]", strings.Join(updatePatch[:imageChangeCount], ","))
	revertPatch[imageChangeCount] = fmt.Sprintf("{\"op\":\"add\",\"path\":\"/metadata/annotations/%s\",\"value\":\"%s\"}",
		imageUpdateAnno, strings.ReplaceAll(updatePatchStr, "\"", "\\\""))
	// combine patch string
	patchStr := fmt.Sprintf("[%s]", strings.Join(revertPatch[:imageChangeCount+1], ","))

	return patchStr, retContainers
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
	blog.V(3).Infof("job done, update bcsgs %v", gs)

	// handle event
	if event != nil {
		// add event to gs and return
		// add object ref
		// finish the job
		//event.Name = gs.Name + "-imageloadfailed"
		event.Namespace = gs.Namespace
		event.InvolvedObject = corev1.ObjectReference{
			Kind:            "GameStatefulSet",
			Namespace:       gs.Namespace,
			Name:            gs.Name,
			UID:             gs.UID,
			APIVersion:      gs.APIVersion,
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
	// TODO retry on conflict
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
	return nil
}

func applyPatchToGS(old *bcsgstkexv1alpha1.GameStatefulSet, patch string) (*bcsgstkexv1alpha1.GameStatefulSet, error) {
	updateOld := &bcsgstkexv1alpha1.GameStatefulSet{}
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
	gs *bcsgstkexv1alpha1.GameStatefulSet, diffContainers []corev1.Container) (*batchv1.Job, error) {
	job := newJob(diffContainers)
	job.Name = fmt.Sprintf("%s-%s-%s", strings.ToLower("GameStatefulSet"), gs.Namespace, gs.Name)
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
					Namespaces: []string{gs.Namespace},
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

	return job, nil
}

func (b *bcsgsWorkload) nodesOfGS(gs *bcsgstkexv1alpha1.GameStatefulSet) []string {
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
