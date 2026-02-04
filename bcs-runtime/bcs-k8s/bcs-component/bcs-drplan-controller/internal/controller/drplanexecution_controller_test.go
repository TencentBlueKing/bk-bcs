/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package controller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	drv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/api/v1alpha1"
)

var _ = Describe("DRPlanExecution Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default", // TODO(user):Modify as needed
		}
		drplanexecution := &drv1alpha1.DRPlanExecution{}

		BeforeEach(func() {
			By("creating a test DRWorkflow first")
			testWorkflow := &drv1alpha1.DRWorkflow{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-workflow",
					Namespace: "default",
				},
				Spec: drv1alpha1.DRWorkflowSpec{
					Actions: []drv1alpha1.Action{
						{
							Name: "test-action",
							Type: "HTTP",
							HTTP: &drv1alpha1.HTTPAction{
								URL:    "http://example.com",
								Method: "GET",
							},
						},
					},
				},
			}
			err := k8sClient.Get(ctx, client.ObjectKey{Name: "test-workflow", Namespace: "default"}, &drv1alpha1.DRWorkflow{})
			if err != nil && errors.IsNotFound(err) {
				Expect(k8sClient.Create(ctx, testWorkflow)).To(Succeed())
			}

			By("creating a test DRPlan")
			testPlan := &drv1alpha1.DRPlan{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-plan",
					Namespace: "default",
				},
				Spec: drv1alpha1.DRPlanSpec{
					Stages: []drv1alpha1.Stage{
						{
							Name: "test-stage",
							Workflows: []drv1alpha1.WorkflowReference{
								{
									WorkflowRef: drv1alpha1.ObjectReference{
										Name:      "test-workflow",
										Namespace: "default",
									},
								},
							},
						},
					},
				},
			}
			err = k8sClient.Get(ctx, client.ObjectKey{Name: "test-plan", Namespace: "default"}, &drv1alpha1.DRPlan{})
			if err != nil && errors.IsNotFound(err) {
				Expect(k8sClient.Create(ctx, testPlan)).To(Succeed())
			}

			By("creating the custom resource for the Kind DRPlanExecution")
			err = k8sClient.Get(ctx, typeNamespacedName, drplanexecution)
			if err != nil && errors.IsNotFound(err) {
				resource := &drv1alpha1.DRPlanExecution{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: drv1alpha1.DRPlanExecutionSpec{
						PlanRef:       "test-plan",
						OperationType: "Execute",
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &drv1alpha1.DRPlanExecution{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance DRPlanExecution")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())

			By("Cleanup the test DRPlan")
			testPlan := &drv1alpha1.DRPlan{}
			err = k8sClient.Get(ctx, client.ObjectKey{Name: "test-plan", Namespace: "default"}, testPlan)
			if err == nil {
				Expect(k8sClient.Delete(ctx, testPlan)).To(Succeed())
			}

			By("Cleanup the test DRWorkflow")
			testWorkflow := &drv1alpha1.DRWorkflow{}
			err = k8sClient.Get(ctx, client.ObjectKey{Name: "test-workflow", Namespace: "default"}, testWorkflow)
			if err == nil {
				Expect(k8sClient.Delete(ctx, testWorkflow)).To(Succeed())
			}
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &DRPlanExecutionReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
			// Example: If you expect a certain status condition after reconciliation, verify it here.
		})
	})
})
