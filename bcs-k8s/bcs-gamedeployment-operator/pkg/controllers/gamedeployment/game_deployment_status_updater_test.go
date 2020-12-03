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

package gamedeployment

/*func TestGameDeploymentUpdaterUpdateStatus(t *testing.T) {
	pod1 := test.NewPod()
	pod1.Status.Phase = v1.PodRunning
	pod1.Status.Conditions = append(pod1.Status.Conditions, v1.PodCondition{Type: v1.PodReady, Status: v1.ConditionTrue})
	pod1.Labels = make(map[string]string)
	pod1.Labels[apps.ControllerRevisionHashLabelKey] = "foo-1"

	pod2 := test.NewPod()
	pod2.Status.Phase = v1.PodRunning
	pod2.Status.Conditions = append(pod1.Status.Conditions, v1.PodCondition{Type: v1.PodReady, Status: v1.ConditionTrue})
	pod2.Labels = make(map[string]string)
	pod2.Labels[apps.ControllerRevisionHashLabelKey] = "foo-2"

	pod3 := test.NewPod()
	pod3.Status.Phase = v1.PodFailed
	pod3.Status.Conditions = append(pod1.Status.Conditions, v1.PodCondition{Type: v1.PodReady, Status: v1.ConditionTrue})
	pod3.Labels = make(map[string]string)
	pod3.Labels[apps.ControllerRevisionHashLabelKey] = "foo-2"

	var pods []*v1.Pod
	pods = append(pods, pod1, pod2, pod3)

	deploy := test.NewGameDeployment(3)
	status := tkexv1alpha1.GameDeploymentStatus{ObservedGeneration: 1, UpdateRevision: "foo-2"}
	fakeClient := &fake.Clientset{}
	updater := NewRealGameDeploymentStatusUpdater(fakeClient, nil)
	fakeClient.AddReactor("update", "gamedeployments", func(action core.Action) (bool, runtime.Object, error) {
		update := action.(core.UpdateAction)
		return true, update.GetObject(), nil
	})
	if err := updater.UpdateGameDeploymentStatus(deploy, &status, pods); err != nil {
		t.Errorf("Error returned on successful status update: %s", err)
	}
	if deploy.Status.Replicas != 3 {
		t.Errorf("UpdateGameDeploymentStatus mutated the replicas %d", deploy.Status.Replicas)
	}

	if deploy.Status.ReadyReplicas != 2 {
		t.Errorf("UpdateGameDeploymentStatus mutated the ready replicas %d", deploy.Status.ReadyReplicas)
	}

	if deploy.Status.UpdatedReplicas != 2 {
		t.Errorf("UpdateGameDeploymentStatus mutated the updated replicas %d", deploy.Status.UpdatedReplicas)
	}

	if deploy.Status.UpdatedReadyReplicas != 1 {
		t.Errorf("UpdateGameDeploymentStatus mutated the updated ready replicas %d", deploy.Status.UpdatedReadyReplicas)
	}
}*/
