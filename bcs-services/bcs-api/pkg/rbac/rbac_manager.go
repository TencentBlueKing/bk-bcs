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

package rbac

import (
	"fmt"
	"reflect"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/rbac/template"

	mapset "github.com/deckarep/golang-set"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

const (
	clusterRoleBindingLabelUser = "clusterrolebinding.user.bke.bcs"
	clusterRoleBindingLabelFrom = "clusterrolebinding.from.bke.bcs"

	clusterRoleBindingTypeFromAny       = "from-any-cluster"
	clusterRoleBindingTypeFromCommon    = "from-common"
	clusterRoleBindingTypeFromNamespace = "from-any-namespaces"

	roleBindingLabelUser = "rolebinding.user.bke.bcs"
)

type rbacManager struct {
	clusterId  string
	kubeClient *kubernetes.Clientset
}

func newRbacManager(cluster string, client *kubernetes.Clientset) *rbacManager {
	return &rbacManager{
		clusterId:  cluster,
		kubeClient: client,
	}
}

// ensureRoles确保待sync的clusterrole在集群中已经存在，如果不存在，则创建
func (rm *rbacManager) ensureRoles(rolesList []string) error {
	for _, role := range rolesList {
		err := rm.ensureRole(role)
		if err != nil {
			return err
		}
	}
	return nil
}

// ensureRoleBindings 从配置文件中全量同步 rbac 数据
func (rm *rbacManager) ensureClusterRoleBindings(username string, rolesList []string) error {
	// 获取该用户已创建的bindings
	label := map[string]string{clusterRoleBindingLabelUser: username}
	alreadyRoleBindings, err := rm.kubeClient.RbacV1().ClusterRoleBindings().List(metav1.ListOptions{LabelSelector: labels.Set(label).AsSelector().String()})
	if err != nil {
		return fmt.Errorf("error when list clusterrolebindings from cluster %s for user %s: %s", rm.clusterId, username, err.Error())
	}
	alreadyBinded := make(map[string]string)
	for _, existedRb := range alreadyRoleBindings.Items {
		alreadyBinded[existedRb.RoleRef.Name] = existedRb.Name
	}
	blog.Infof("already existed clusterrolebindings for user %s in cluster %s: %s", username, rm.clusterId, alreadyBinded)

	toCreate, toDelete := rm.getClusterRoles(alreadyBinded, rolesList)

	blog.Infof("clusterroles to bind for user %s in cluster %s: %s", username, rm.clusterId, toCreate)
	blog.Infof("clusterroles to delete its bind for user %s in cluster %s: %s", username, rm.clusterId, toDelete)

	for _, role := range toCreate {
		err := rm.createClusterRoleBinding(username, role, "")
		if err != nil {
			return fmt.Errorf("error when creating clusterrolebinding for user %s to bind clusterrole %s: %s", username, role, err.Error())
		}
	}

	for _, role := range toDelete {
		err := rm.deleteClusterRoleBinding(alreadyBinded[role])
		if err != nil {
			return fmt.Errorf("error when deleting clusterrolebinding %s for username %s: %s", alreadyBinded[role], username, err.Error())
		}
	}

	return nil
}

// ensureRole 确保 clusterrole 已经在集群中创建
func (rm *rbacManager) ensureRole(role string) error {
	clusterRoleName := template.ClusterRolePrefix + role
	roleTemplate, ok := template.RoleTemplateStore[clusterRoleName]
	if !ok {
		return fmt.Errorf("role not existed in roletemplate：%s", clusterRoleName)
	}

	rules := roleTemplate.Rules
	existedClusterRole, err := rm.kubeClient.RbacV1().ClusterRoles().Get(clusterRoleName, metav1.GetOptions{})
	// 如果clusterrole不存在，则创建
	if err != nil && errors.IsNotFound(err) {
		err1 := rm.createClusterRole(clusterRoleName, rules)
		if err1 != nil {
			return fmt.Errorf("error when creating clusterrole %s to cluster %s: %s", clusterRoleName, rm.clusterId, err.Error())
		}
		return nil
	} else if err != nil {
		return fmt.Errorf("error when get clusterrole %s from cluster %s: %s", clusterRoleName, rm.clusterId, err.Error())
	} else {
		// clusterrole已存在，判断rules是否一致，如果一致则跳过
		if reflect.DeepEqual(existedClusterRole.Rules, rules) {
			return nil
		}
		// 如果不一致，则更新已存在的clusterrole
		existedClusterRole = existedClusterRole.DeepCopy()
		existedClusterRole.Rules = rules
		err := rm.updateClusterRole(existedClusterRole)
		if err != nil {
			return fmt.Errorf("error when updating clusterrole %s to cluster %s: %s", clusterRoleName, rm.clusterId, err.Error())
		}
		return nil
	}
}

// createClusterRole 创建 clusterrole
func (rm *rbacManager) createClusterRole(clusterRoleName string, rules []rbacv1.PolicyRule) error {
	_, err := rm.kubeClient.RbacV1().ClusterRoles().Create(&rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterRoleName,
		},
		Rules: rules,
	})
	return err
}

// updateClusterRole update clusterrole
func (rm *rbacManager) updateClusterRole(clusterRole *rbacv1.ClusterRole) error {
	_, err := rm.kubeClient.RbacV1().ClusterRoles().Update(clusterRole)
	return err
}

// ensureAddClusterRoleBinding 确保创建 clusterrolebinding
func (rm *rbacManager) ensureAddClusterRoleBinding(username, role, bindingType string) error {
	// 获取该用户已创建的bindings
	label := map[string]string{clusterRoleBindingLabelUser: username, clusterRoleBindingLabelFrom: bindingType}
	alreadyClusterRoleBindings, err := rm.kubeClient.RbacV1().ClusterRoleBindings().List(metav1.ListOptions{LabelSelector: labels.Set(label).AsSelector().String()})
	if err != nil {
		return fmt.Errorf("error when list clusterrolebindings from cluster %s for user %s: %s", rm.clusterId, username, err.Error())
	}

	clusterRole := template.ClusterRolePrefix + role
	for _, clusterRoleBinding := range alreadyClusterRoleBindings.Items {
		if clusterRoleBinding.RoleRef.Name == clusterRole {
			blog.Info("clusterrolebinding %s already existed, binded to clusterrole %s for user %s", clusterRoleBinding.Name, role, username)
			return nil
		}
	}

	err = rm.createClusterRoleBinding(username, clusterRole, bindingType)
	if err != nil {
		return fmt.Errorf("error when creating clusterrolebinding for user %s to bind clusterrole %s in cluster %s: %s", username, role, rm.clusterId, err.Error())
	}
	return nil
}

// ensureDeleteClusterRoleBinding 确保删除 clusterrolebinding
func (rm *rbacManager) ensureDeleteClusterRoleBinding(username, role, bindingType string) error {
	// 获取该用户已创建的bindings
	label := map[string]string{clusterRoleBindingLabelUser: username, clusterRoleBindingLabelFrom: bindingType}
	alreadyClusterRoleBindings, err := rm.kubeClient.RbacV1().ClusterRoleBindings().List(metav1.ListOptions{LabelSelector: labels.Set(label).AsSelector().String()})
	if err != nil {
		return fmt.Errorf("error when list clusterrolebindings from cluster %s for user %s: %s", rm.clusterId, username, err.Error())
	}

	clusterRole := template.ClusterRolePrefix + role
	for _, clusterRoleBinding := range alreadyClusterRoleBindings.Items {
		if clusterRoleBinding.RoleRef.Name == clusterRole {
			return rm.deleteClusterRoleBinding(clusterRoleBinding.Name)
		}
	}
	return nil
}

// createClusterRoleBinding 调用 k8s api 创建 clusterrolebinding
func (rm *rbacManager) createClusterRoleBinding(username, clusterRole, bindingType string) error {
	// 给每个用户创建的clusterrolebinding都打上一个特有的label
	label := map[string]string{clusterRoleBindingLabelUser: username, clusterRoleBindingLabelFrom: bindingType}
	objectMeta := metav1.ObjectMeta{
		GenerateName: "clusterrolebinding-",
		Labels:       label,
	}

	subject := rbacv1.Subject{
		Kind: "User",
		Name: username,
	}

	roleRef := rbacv1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "ClusterRole",
		Name:     clusterRole,
	}

	_, err := rm.kubeClient.RbacV1().ClusterRoleBindings().Create(&rbacv1.ClusterRoleBinding{
		ObjectMeta: objectMeta,
		Subjects:   []rbacv1.Subject{subject},
		RoleRef:    roleRef,
	})
	return err
}

// deleteClusterRoleBinding 调用 k8s api 删除  clusterrolebinding
func (rm *rbacManager) deleteClusterRoleBinding(clusterRoleBindingName string) error {
	err := rm.kubeClient.RbacV1().ClusterRoleBindings().Delete(clusterRoleBindingName, &metav1.DeleteOptions{})
	return err
}

// ensureAddRoleBinding 确保创建 rolebinding
func (rm *rbacManager) ensureAddRoleBinding(username, role, namespace string) error {
	// 获取该用户已创建的bindings
	label := map[string]string{roleBindingLabelUser: username}
	alreadyRoleBindings, err := rm.kubeClient.RbacV1().RoleBindings(namespace).List(metav1.ListOptions{LabelSelector: labels.Set(label).AsSelector().String()})
	if err != nil {
		return fmt.Errorf("error when list rolebindings from cluster %s namespace %s for user %s: %s", rm.clusterId, namespace, username, err.Error())
	}

	clusterRole := template.ClusterRolePrefix + role
	for _, roleBinding := range alreadyRoleBindings.Items {
		if roleBinding.RoleRef.Name == clusterRole {
			blog.Infof("rolebinding %s already existed, binded to role %s in namespace %s for user %s", roleBinding.Name, role, namespace, username)
			return nil
		}
	}

	err = rm.createRoleBinding(username, clusterRole, namespace)
	if err != nil {
		return fmt.Errorf("error when creating rolebinding for user %s to bind role %s in cluster %s namespace %s: %s", username, role, rm.clusterId, namespace, err.Error())
	}
	return nil
}

// ensureDeleteRoleBinding 确保删除 rolebinding
func (rm *rbacManager) ensureDeleteRoleBinding(username, role, namespace string) error {
	// 获取该用户已创建的bindings
	label := map[string]string{roleBindingLabelUser: username}
	alreadyRoleBindings, err := rm.kubeClient.RbacV1().RoleBindings(namespace).List(metav1.ListOptions{LabelSelector: labels.Set(label).AsSelector().String()})
	if err != nil {
		return fmt.Errorf("error when list rolebindings from cluster %s namespace %s for user %s: %s", rm.clusterId, namespace, username, err.Error())
	}

	clusterRole := template.ClusterRolePrefix + role
	for _, roleBinding := range alreadyRoleBindings.Items {
		if roleBinding.RoleRef.Name == clusterRole {
			return rm.deleteRoleBinding(roleBinding.Name, namespace)
		}
	}
	return nil
}

// createRoleBinding 调用 k8s api 创建 rolebinding
func (rm *rbacManager) createRoleBinding(username, clusterRole, namespace string) error {
	// 给每个用户创建的rolebinding都打上一个特有的label
	label := map[string]string{roleBindingLabelUser: username}
	objectMeta := metav1.ObjectMeta{
		GenerateName: "rolebinding-",
		Namespace:    namespace,
		Labels:       label,
	}

	subject := rbacv1.Subject{
		Kind: "User",
		Name: username,
	}

	roleRef := rbacv1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "ClusterRole",
		Name:     clusterRole,
	}

	_, err := rm.kubeClient.RbacV1().RoleBindings(namespace).Create(&rbacv1.RoleBinding{
		ObjectMeta: objectMeta,
		Subjects:   []rbacv1.Subject{subject},
		RoleRef:    roleRef,
	})
	return err
}

// deleteRoleBinding 调用 k8s api 删除 rolebinding
func (rm *rbacManager) deleteRoleBinding(roleBindingName, namespace string) error {
	err := rm.kubeClient.RbacV1().RoleBindings(namespace).Delete(roleBindingName, &metav1.DeleteOptions{})
	return err
}

// getClusterRoles 通过对比，得到待创建和待删除的 clusterrolebinding
func (rm *rbacManager) getClusterRoles(alreadyBinded map[string]string, rolesList []string) ([]string, []string) {
	alreadySet := mapset.NewSet()
	for role := range alreadyBinded {
		alreadySet.Add(role)
	}
	newSet := mapset.NewSet()
	for _, role := range rolesList {
		newSet.Add(template.ClusterRolePrefix + role)
	}

	toCreateSet := newSet.Difference(alreadySet)
	toDeleteSet := alreadySet.Difference(newSet)
	toCreateIt := toCreateSet.Iterator()
	toDeleteIt := toDeleteSet.Iterator()
	var toCreateArray, toDeleteArray []string
	for elem := range toCreateIt.C {
		toCreateArray = append(toCreateArray, elem.(string))
	}
	for elem := range toDeleteIt.C {
		toDeleteArray = append(toDeleteArray, elem.(string))
	}

	return toCreateArray, toDeleteArray
}
