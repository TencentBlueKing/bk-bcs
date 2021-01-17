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

package strategy

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/mesosproto/mesos"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	offerP "github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/sched/offer"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/store"

	"github.com/danwakefield/fnmatch"
)

// ConstraintsFit Check whether an offer matches with the constraints for an application
func ConstraintsFit(version *types.Version, offer *mesos.Offer, store store.Store, taskgroupID string) (bool, error) {
	constraints := version.Constraints
	//taints & toleration
	attribute, _ := offerP.GetOfferAttribute(offer, types.MesosAttributeNoSchedule)
	if attribute != nil {
		fit, _ := checkToleration(offer, version)
		if !fit {
			return false, nil
		}
	}

	var itemInsance int
	if constraints == nil && !isVersionRequestIp(version) {
		blog.V(3).Infof("to check constraints: version(%s.%s) not set constraints", version.RunAs, version.ID)
		return true, nil
	}
	if constraints != nil {
		itemInsance = len(constraints.IntersectionItem)
	}

	blog.V(3).Infof("to check constraints: version(%s.%s) have %d constraints", version.RunAs, version.ID, itemInsance)

	i := 0
	if itemInsance > 0 {
		for _, oneConstraint := range constraints.IntersectionItem {
			if oneConstraint == nil {
				continue
			}
			isFit, _ := constraintDataItemFit(oneConstraint, offer, version, store)
			if isFit == false {
				blog.V(3).Infof("check constraints[%d]: not fit, so this offer is not fit", i)
				return false, nil
			}
			blog.V(3).Infof("check constraints[%d]: fit, continue to check next constraints", i)
			i++
		}
	}

	isFit, err := checkRequestIP(version, offer, store, taskgroupID)
	if err != nil {
		blog.V(3).Infof("requestip constraint check error(%s)", err.Error())
		return isFit, err
	}
	if isFit == false {
		blog.V(3).Infof("requestip constraint not fit, taskgroupID(%s)", taskgroupID)
		return false, nil
	}
	blog.V(3).Infof("requestip constraint fit, taskgroupID(%s)", taskgroupID)

	blog.V(3).Infof("all constraints are fit, so this offer is fit")
	return true, nil
}

func constraintDataItemFit(constraintItem *commtypes.ConstraintDataItem, offer *mesos.Offer, version *types.Version, store store.Store) (bool, error) {

	i := 0
	for _, constraintData := range constraintItem.UnionData {
		if constraintData == nil {
			continue
		}
		isFit, _ := contraintDataFit(constraintData, offer, version, store)
		if isFit == true {
			blog.V(3).Infof("check constraintData[%d]: fit, so this ConstraintItem is fit", i)
			return true, nil
		}
		blog.V(3).Infof("check constraintData[%d]: not fit, continue to check next ConstraintData", i)

		i++
	}

	blog.V(3).Infof("all constraintDatas are not fit, so this constraintItem is not fit")

	return false, nil
}

func contraintDataFit(constraint *commtypes.ConstraintData, offer *mesos.Offer, version *types.Version, store store.Store) (bool, error) {

	name := constraint.Name
	operate := constraint.Operate

	//there's no need to judge toleration here, always true
	if constraint.Operate == commtypes.Constraint_Type_TOLERATION {
		return true, nil
	}

	valueType := constraint.Type
	blog.V(3).Infof("check constraint(%s) for (name:%s, type:%d) for offer from %s", operate, name, valueType, offer.GetHostname())

	if operate == commtypes.Constraint_Type_EXCLUDE {
		return checkExclude(constraint, offer.GetHostname(), version, store)
	}

	var attribute *mesos.Attribute
	// construct an attribute for hostname
	if name == "hostname" {
		var attr mesos.Attribute
		var attrName = "hostname"
		attr.Name = &attrName
		var attrType mesos.Value_Type = mesos.Value_TEXT
		attr.Type = &attrType
		var attrValue mesos.Value_Text
		var host string = offer.GetHostname()
		attrValue.Value = &host
		attr.Text = &attrValue
		attribute = &attr
	} else {
		attribute, _ = offerP.GetOfferAttribute(offer, name)
	}
	if attribute == nil {
		blog.V(3).Infof("offer from %s, get attribute(%s) return nil, contraint not fit", offer.GetHostname(), name)
		return false, nil
	}

	switch operate {
	case commtypes.Constraint_Type_UNIQUE:
		return checkUnique(constraint, attribute, version, store)
	case commtypes.Constraint_Type_CLUSTER:
		return checkCluster(constraint, attribute)
	case commtypes.Constraint_Type_GROUP_BY:
		return checkGroupBy(constraint, attribute, version, store)
	case commtypes.Constraint_Type_MAX_PER:
		return checkMaxPer(constraint, attribute, version, store)
	case commtypes.Constraint_Type_LIKE:
		return checkLike(constraint, attribute)
	case commtypes.Constraint_Type_UNLIKE:
		return checkUnLike(constraint, attribute)
	case commtypes.Constraint_Type_GREATER:
		return checkGreater(constraint, attribute)
	//there's no need to judge toleration here, always true
	case commtypes.Constraint_Type_TOLERATION:
		return true, nil
	default:
		blog.Warnf("constraint operate type(%s) for attr(%s) not supported", operate, name)
		return false, errors.New("constraint operate type error")
	}
}

func checkRequestIP(version *types.Version, offer *mesos.Offer, store store.Store, taskgroupID string) (bool, error) {
	runAs := version.RunAs
	appID := version.ID
	if isVersionRequestIp(version) == false {
		blog.V(3).Infof("check requestip: version(%s.%s) no requestip", runAs, appID)
		return true, nil
	}

	blog.V(3).Infof("check requestip: version(%s.%s) taskgroup(%s)", runAs, appID, taskgroupID)

	store.LockApplication(runAs + "." + appID)
	defer store.UnLockApplication(runAs + "." + appID)

	app, _ := store.FetchApplication(runAs, appID)
	if app == nil {
		blog.Errorf("check requestip: fetch application(%s.%s) return nil", runAs, appID)
		return false, errors.New("application not exist")
	}
	index := app.Instances
	if taskgroupID != "" {
		taskgroup, _ := store.FetchTaskGroup(taskgroupID)
		if taskgroup == nil {
			blog.Errorf("check requestip: fetch taskgroup(%s) return nil", taskgroupID)
			return false, errors.New("taskgroup not exist")
		}
		index = taskgroup.InstanceID
	}
	var v string
	var ok bool
	k := "io.tencent.bcs.netsvc.requestip." + strconv.Itoa(int(index))
	v, ok = version.Labels[k]
	if !ok {
		v, ok = version.ObjectMeta.Annotations[k]
		if !ok {
			blog.Error("check requestip: version(%s.%s) label,annotation(%s) not exist", runAs, appID, k)
			return false, errors.New("requestip not exist")
		}
	}

	splitV := strings.Split(v, "|")
	if len(splitV) < 2 {
		blog.Infof("check requestip: version(%s.%s) label(%s:%s) has no constraints, pass", runAs, appID, k, v)
		return true, nil
	}

	splitConstraint := strings.Split(splitV[1], "=")
	if len(splitConstraint) != 2 {
		blog.Warnf("check requestip: version(%s.%s) label(%s:%s) constraint information(%s) not correct",
			runAs, appID, splitV[1])
		return false, nil
	}
	constraintName := splitConstraint[0]
	constraintValues := splitConstraint[1]

	attributeValue := ""
	if constraintName == "hostname" {
		attributeValue = offer.GetHostname()
	} else {
		attribute, _ := offerP.GetOfferAttribute(offer, constraintName)
		if attribute == nil {
			blog.Warn("check requestip: version(%s.%s), offer from %s, get attribute(%s) return nil",
				runAs, appID, offer.GetHostname(), constraintName)
			return false, nil
		}
		if attribute.GetText() != nil {
			attributeValue = attribute.GetText().GetValue()
		}
		if attributeValue == "" {
			blog.Warn("check requestip: version(%s.%s), offer from %s, get empty attribute(%s)",
				runAs, appID, offer.GetHostname(), constraintName)
			return false, nil
		}
	}

	constraints := strings.Split(constraintValues, ";")
	for _, oneItem := range constraints {
		isMatch := fnmatch.Match(oneItem, attributeValue, 0)
		if isMatch == true {
			blog.Infof("check requestip: version(%s.%s %d), attribute(%s: %s) fit constraint(%s) by fnmatch",
				runAs, appID, index, constraintName, attributeValue, oneItem)
			return true, nil
		}

		r, _ := regexp.Compile(oneItem)
		if r != nil {
			isMatch = r.MatchString(attributeValue)
		}
		if isMatch == true {
			blog.Infof("check requestip: version(%s.%s %d), attribute(%s: %s) fit constraint(%s) by regexp",
				runAs, appID, index, constraintName, attributeValue, oneItem)
			return true, nil
		}
	}

	blog.V(3).Infof("check requestip: version(%s.%s %d), attribute(%s: %s) not fit constraints(%s)",
		runAs, appID, index, constraintName, attributeValue, constraintValues)

	return false, nil
}

func isVersionRequestIp(version *types.Version) bool {
	for k := range version.Labels {
		splitK := strings.Split(k, ".")
		if len(splitK) == 6 && splitK[3] == "netsvc" && splitK[4] == "requestip" {
			return true
		}
	}
	for k := range version.ObjectMeta.Annotations {
		splitK := strings.Split(k, ".")
		if len(splitK) == 6 && splitK[3] == "netsvc" && splitK[4] == "requestip" {
			return true
		}
	}
	return false
}

func checkUnique(constraint *commtypes.ConstraintData, attribute *mesos.Attribute, version *types.Version, store store.Store) (bool, error) {

	blog.V(3).Infof("constraint UNIQUE for attribute(name: %s)", attribute.GetName())

	runAs := version.RunAs
	appID := version.ID

	store.LockApplication(runAs + "." + appID)
	taskGroups, err := store.ListTaskGroups(runAs, appID)
	store.UnLockApplication(runAs + "." + appID)
	if err != nil {
		blog.Error("constraint UNIQUE: list taskgroup(%s %s) err:%s", runAs, appID, err.Error())
		return false, errors.New("constraint UNIQUE: list taskgroup err")
	}

	if taskGroups != nil {
		for _, taskGroup := range taskGroups {
			for _, taskGroupAttr := range taskGroup.Attributes {
				isSame, _ := compareAttribute(taskGroupAttr, attribute)
				if isSame == true && taskGroup.Status != types.TASKGROUP_STATUS_FINISH && taskGroup.Status != types.TASKGROUP_STATUS_FAIL {
					blog.V(3).Infof("constraint UNIQUE: taskgroup(%s) attribute(%s) is the same, so UNIQUE not fit", taskGroup.ID, attribute.GetName())
					return false, nil
				}
			}
		}
	}

	return true, nil
}

func checkGreater(constraint *commtypes.ConstraintData, attribute *mesos.Attribute) (bool, error) {
	blog.V(3).Infof("constraint Greater by attribute(name:%s)", constraint.Name)

	switch constraint.Type {
	case commtypes.ConstValueType_Scalar:
		scalar := constraint.Scalar.Value
		if attribute.GetType() != mesos.Value_SCALAR {
			blog.Errorf("constraint %s type ConstValueType_Scalar, but attribute type %s", constraint.Name, attribute.GetType().String())
			return false, nil
		}

		attrScalar := attribute.GetScalar().GetValue()

		if attrScalar > scalar {
			return true, nil
		}
		return false, nil

	default:
		blog.Errorf("constraint %s type %d is invalid", constraint.Name, constraint.Type)
	}

	return false, nil
}

func checkToleration(offer *mesos.Offer, version *types.Version) (bool, error) {
	attribute, _ := offerP.GetOfferAttribute(offer, types.MesosAttributeNoSchedule)
	if version.Constraints == nil {
		blog.V(3).Infof("version(%s:%s) don't toleration offer %s taint(%v)",
			version.RunAs, version.ID, offer.GetHostname(), attribute.Set.Item)
		return false, nil
	}

	for _, union := range version.Constraints.IntersectionItem {
		if union == nil {
			continue
		}

		for _, item := range union.UnionData {
			if item.Operate != commtypes.Constraint_Type_TOLERATION {
				continue
			}

			if item.Type != commtypes.ConstValueType_Text {
				blog.V(3).Infof("version(%s:%s) constraint(%s:%s) type %d is invalid",
					version.RunAs, version.ID, item.Name, item.Operate, item.Type)
				continue
			}

			for _, kv := range attribute.Set.Item {
				kvs := strings.Split(kv, "=")
				if item.Name == kvs[0] && item.Text.Value == kvs[1] {
					blog.V(3).Infof("version(%s:%s) toleration offer %s taint(%s:%s)",
						version.RunAs, version.ID, offer.GetHostname(), kvs[0], kvs[1])
					return true, nil
				}
			}
		}
	}

	blog.V(3).Infof("version(%s:%s) don't toleration offer %s taint(%v)",
		version.RunAs, version.ID, offer.GetHostname(), attribute.Set.Item)
	return false, nil
}

func checkLike(constraint *commtypes.ConstraintData, attribute *mesos.Attribute) (bool, error) {

	blog.V(3).Infof("constraint LIKE by attribute(name:%s)", constraint.Name)

	attrValueType := attribute.GetType()

	switch attrValueType {
	case mesos.Value_TEXT:
		attrValue := attribute.GetText().GetValue()
		if constraint.Type == commtypes.ConstValueType_Text {

			isMatch := fnmatch.Match(constraint.Text.Value, attrValue, 0)
			if isMatch == true {
				blog.Info("constraint LIKE: attrValue(%s) Like constraint(%s) by fnmatch", attrValue, constraint.Text.Value)
			} else {
				r, _ := regexp.Compile(constraint.Text.Value)
				if r != nil {
					isMatch = r.MatchString(attrValue)
				}
				if isMatch == true {
					blog.Info("constraint LIKE: attrValue(%s) Like constraint(%s) by regexp", attrValue, constraint.Text.Value)
				}
			}

			if isMatch == true {
				blog.Infof("constraint LIKE: attrValue(%s) Like constraint(%s)", attrValue, constraint.Text.Value)
				return true, nil
			}
			blog.V(3).Infof("constraint LIKE: attrValue(%s) not Like constraint(%s)", attrValue, constraint.Text.Value)
			return false, nil
		}

		if constraint.Type == commtypes.ConstValueType_Set {
			for _, oneItem := range constraint.Set.Item {

				isMatch := fnmatch.Match(oneItem, attrValue, 0)
				if isMatch == true {
					blog.Info("constraint LIKE: attrValue(%s) Like constraint(%s) by fnmatch", attrValue, oneItem)
				} else {
					r, _ := regexp.Compile(oneItem)
					if r != nil {
						isMatch = r.MatchString(attrValue)
					}
					if isMatch == true {
						blog.Info("constraint LIKE: attrValue(%s) Like constraint(%s) by regexp", attrValue, oneItem)
					}
				}

				if isMatch == true {
					blog.Infof("constraint LIKE: attrValue(%s) Like constraint(%s)", attrValue, oneItem)
					return true, nil
				}
				blog.V(3).Infof("constraint LIKE: attrValue(%s) not Like constraint(%s)", attrValue, oneItem)
			}
			return false, nil
		}

		blog.Error("unprocessed constraint(Like) for attribute value type(%d) and constrain value type(%d)", attrValueType, constraint.Type)
		return true, nil
	default:
		blog.Error("unprocessed constraint(Like) for attribute value type(%s)", attrValueType)
		return true, nil
	}
}

func checkUnLike(constraint *commtypes.ConstraintData, attribute *mesos.Attribute) (bool, error) {

	blog.V(3).Infof("constraint UNLIKE by attribute(name:%s)", constraint.Name)

	attrValueType := attribute.GetType()

	switch attrValueType {
	case mesos.Value_TEXT:
		attrValue := attribute.GetText().GetValue()
		if constraint.Type == commtypes.ConstValueType_Text {

			isMatch := fnmatch.Match(constraint.Text.Value, attrValue, 0)
			if isMatch == true {
				blog.Info("constraint UNLIKE: attrValue(%s) Like constraint(%s) by fnmatch", attrValue, constraint.Text.Value)
			} else {
				r, _ := regexp.Compile(constraint.Text.Value)
				if r != nil {
					isMatch = r.MatchString(attrValue)
				}
				if isMatch == true {
					blog.Info("constraint UNLIKE: attrValue(%s) Like constraint(%s) by regexp", attrValue, constraint.Text.Value)
				}
			}

			if isMatch == true {
				blog.V(3).Infof("constraint UNLIKE: attrValue(%s) Like constraint(%s)", attrValue, constraint.Text.Value)
				return false, nil
			}
			blog.Infof("constraint UNLIKE: attrValue(%s) not Like constraint(%s)", attrValue, constraint.Text.Value)
			return true, nil
		}

		if constraint.Type == commtypes.ConstValueType_Set {
			for _, oneItem := range constraint.Set.Item {

				isMatch := fnmatch.Match(oneItem, attrValue, 0)
				if isMatch == true {
					blog.Info("constraint UNLIKE: attrValue(%s) Like constraint(%s) by fnmatch", attrValue, oneItem)
				} else {
					r, _ := regexp.Compile(oneItem)
					if r != nil {
						isMatch = r.MatchString(attrValue)
					}
					if isMatch == true {
						blog.Info("constraint UNLIKE: attrValue(%s) Like constraint(%s) by regexp", attrValue, oneItem)
					}
				}

				if isMatch == true {
					blog.V(3).Infof("constraint UNLIKE: attrValue(%s) Like constraint(%s)", attrValue, oneItem)
					return false, nil
				}
				blog.Infof("constraint UNLIKE: attrValue(%s) not Like constraint(%s)", attrValue, oneItem)
			}

			return true, nil
		}

		blog.Error("unprocessed constraint(UnLike) for attribute value type(%d) and constrain value type(%d)", attrValueType, constraint.Type)
		return true, nil
	default:
		blog.Error("unprocessed constraint(UnLike) for attribute value type(%s)", attrValueType)
		return true, nil
	}
}

func checkCluster(constraint *commtypes.ConstraintData, attribute *mesos.Attribute) (bool, error) {

	blog.V(3).Infof("constraint CLUSTER by attribute(name:%s)", constraint.Name)

	attrValueType := attribute.GetType()

	switch attrValueType {
	case mesos.Value_TEXT:
		attrValue := attribute.GetText().GetValue()
		if constraint.Type == commtypes.ConstValueType_Text {
			if constraint.Text.Value == attrValue {
				blog.V(3).Infof("constraint CLUSTER: attrValue(%s) fit Cluster constraint(%s)", attrValue, constraint.Text.Value)
				return true, nil
			}
			return false, nil
		}
		if constraint.Type == commtypes.ConstValueType_Set {
			for _, oneItem := range constraint.Set.Item {
				if oneItem == attrValue {
					blog.V(3).Infof("constraint CLUSTER: attrValue(%s)fit Cluster constraint(%s)", attrValue, oneItem)
					return true, nil
				}
			}
			return false, nil
		}
		blog.Error("constraint CLUSTER: unprocessed constraint(Cluster) for attribute value type(%d) and constrain value type(%d)", attrValueType, constraint.Type)

	default:
		blog.Error("constraint CLUSTER: unprocessed constraint(Cluster) for attribute value type(%s)", attrValueType)
	}

	return true, nil

}

func checkGroupBy(constraint *commtypes.ConstraintData, attribute *mesos.Attribute, version *types.Version, store store.Store) (bool, error) {

	blog.V(3).Infof("constraint GROUPBY by attribute(name:%s)", constraint.Name)

	runAs := version.RunAs
	appID := version.ID

	store.LockApplication(runAs + "." + appID)
	taskGroups, err := store.ListTaskGroups(runAs, appID)
	store.UnLockApplication(runAs + "." + appID)
	if err != nil {
		blog.Error("constraint GROUPBY: list taskgroup(%s %s) err:%s", runAs, appID, err.Error())
		return false, errors.New("list taskgroup err")
	}

	if constraint.Type != commtypes.ConstValueType_Set {
		blog.Error("constraint GROUPBY, type must be ConstValueType_Set(%d)", commtypes.ConstValueType_Set)
		return false, errors.New("constraint GROUPBY, type must be ConstValueType_Set")
	}

	if constraint.Set == nil {
		blog.Error("constraint GROUPBY: constraint.Set is nil")
		return false, errors.New("constraint GROUPBY: constraint.Set is nil")
	}

	// confirm that offer attribute value is fit constraint
	attrValueType := attribute.GetType()
	valueFit := false
	switch attrValueType {
	case mesos.Value_TEXT:
		attrValue := attribute.GetText().GetValue()
		if constraint.Type == commtypes.ConstValueType_Set {
			for _, oneItem := range constraint.Set.Item {
				if oneItem == attrValue {
					blog.V(3).Infof("constraint GROUPBY: attrValue(%s) fit Cluster(%s)", attrValue, oneItem)
					valueFit = true
					break
				}
			}
		} else {
			blog.Error("constraint GROUPBY for constraint value type(%d) not supported", constraint.Type)
			return false, errors.New("constraint GROUPBY: constraint value type error")
		}
	default:
		blog.Error("constraint GROUPBY for for attribute value type(%s) not supported", attrValueType)
		return false, errors.New("constraint GROUPBY: attribute value type error")
	}
	if valueFit == false {
		blog.V(3).Infof("constraint GROUPBY: attribute not in constraint's set, not fit")
		return false, nil
	}

	// confirm that offer current limit is fit constraint
	setLen := len(constraint.Set.Item)
	if setLen <= 0 {
		blog.Error("constraint GROUPBY: constraint.Set Len(%d) <= 0", setLen)
		return false, errors.New("constraint GROUPBY: constraint.Set len <= 0")
	}
	taskGroupNum := len(taskGroups) + 1
	limit := taskGroupNum / setLen
	if taskGroupNum%setLen != 0 {
		limit++
	}
	blog.V(3).Infof("constraint GROUPBY: taskGroupNum(%d), setLen(%d), curr limit(%d)", taskGroupNum, setLen, limit)
	num := 0
	if taskGroups != nil {
		for _, taskGroup := range taskGroups {
			for _, taskGroupAttr := range taskGroup.Attributes {
				isSame, _ := compareAttribute(taskGroupAttr, attribute)
				if isSame == true && taskGroup.Status != types.TASKGROUP_STATUS_FINISH && taskGroup.Status != types.TASKGROUP_STATUS_FAIL {
					blog.V(3).Infof("constraint GROUPBY: taskgroup(%s) attribute(%s) is the same, num++", taskGroup.ID, attribute.GetName())
					num++
					break
				}
			}
		}
	}
	if num >= limit {
		blog.V(3).Infof("constraint GROUPBY: num(%d) >= curr limit(%d), not fit", num, limit)
		return false, nil
	}
	blog.V(3).Infof("constraint GROUPBY: num(%d) < curr limit(%d), fit", num, limit)

	return true, nil
}

func checkMaxPer(constraint *commtypes.ConstraintData, attribute *mesos.Attribute, version *types.Version, store store.Store) (bool, error) {

	blog.V(3).Infof("constraint MAXPER: for attribute(name: %s)", attribute.GetName())

	runAs := version.RunAs
	appID := version.ID

	store.LockApplication(runAs + "." + appID)
	taskGroups, err := store.ListTaskGroups(runAs, appID)
	store.UnLockApplication(runAs + "." + appID)
	if err != nil {
		blog.Error("constraint MAXPER: list taskgroup(%s %s) err:%s", runAs, appID, err.Error())
		return false, errors.New("constraint MAXPER: list taskgroup err")
	}

	if constraint.Type != commtypes.ConstValueType_Text {
		blog.Error("constraint MAXPER, type must be ConstValueType_Text(%d)", commtypes.ConstValueType_Text)
		return false, errors.New("constraint MAXPER, type must be ConstValueType_Text")
	}

	if constraint.Text == nil {
		blog.Error("constraint MAXPER: constraint.Text is nil")
		return false, errors.New("constraint MAXPER: constraint.Text is nil")
	}

	limit, err := strconv.Atoi(constraint.Text.Value)
	if err != nil {
		return false, err
	}

	blog.V(3).Infof("constraint MAXPER, limit count(%d)", limit)

	num := 0
	if taskGroups != nil {
		for _, taskGroup := range taskGroups {
			for _, taskGroupAttr := range taskGroup.Attributes {
				isSame, _ := compareAttribute(taskGroupAttr, attribute)
				if isSame == true && taskGroup.Status != types.TASKGROUP_STATUS_FINISH && taskGroup.Status != types.TASKGROUP_STATUS_FAIL {
					blog.V(3).Infof("constraint MAXPER: taskgroup(%s) attribute(%s) is the same, num++", taskGroup.ID, attribute.GetName())
					num++
					break
				}
			}
		}
	}

	if num >= limit {
		blog.V(3).Infof("constraint MAXPER, num(%d) >= maxper(%d), not fit", num, limit)
		return false, nil
	}

	blog.V(3).Infof("constraint MAXPER, num(%d) < maxper(%d), fit", num, limit)
	return true, nil
}

func checkExclude(constraint *commtypes.ConstraintData, hostname string, appVersion *types.Version, store store.Store) (bool, error) {

	blog.V(3).Infof("constraint EXCLUDE on host(%s)", hostname)

	if constraint.Type != commtypes.ConstValueType_Set {
		blog.Error("constraint EXCLUDE, type must be ConstValueType_Set(%d)", commtypes.ConstValueType_Set)
		return false, errors.New("constraint EXCLUDE, type must be ConstValueType_Set")
	}
	if constraint.Set == nil {
		blog.Error("constraint EXCLUDE: constraint.Set is nil")
		return false, errors.New("constraint EXCLUDE: constraint.Set is nil")
	}

	runAses, err := store.ListRunAs()
	if err != nil {
		blog.Error("constraint EXCLUDE: fail to list runAses, err:%s", err.Error())
		return false, err
	}

	for _, runAs := range runAses {
		blog.Info("constraint EXCLUDE: to check runAs(%s)", runAs)
		appIDs, err := store.ListApplicationNodes(runAs)
		if err != nil {
			blog.Error("constraint EXCLUDE: fail to list %s, err:%s", runAs, err.Error())
			return false, err
		}

		for _, appID := range appIDs {
			blog.Info("constraint EXCLUDE: to check application:%s.%s ", runAs, appID)

			version, _ := store.GetVersion(runAs, appID)
			if version == nil {
				blog.Error("constraint EXCLUDE: cannot get version for application(%s %s)", runAs, appID)
				return false, err
			}

			if appVersion.ObjectMeta.Name == version.ObjectMeta.Name && appVersion.ObjectMeta.NameSpace == version.ObjectMeta.NameSpace {
				blog.V(3).Infof("constraint EXCLUDE: app(%s.%s) is the same as current application, pass", runAs, appID)
				continue
			}

			labelMatch := false
			for k, v := range version.Labels {
				labelStr := k + ":" + v
				for _, oneItem := range constraint.Set.Item {
					if oneItem == labelStr {
						blog.V(3).Infof("constraint EXCLUDE: label(%s:%s) match, to check taskgroups", k, v)
						labelMatch = true
						break
					}
				}
				if labelMatch {
					break
				}
			}

			if labelMatch == false {
				blog.V(3).Infof("constraint EXCLUDE: label not match constraint, pass")
				continue
			}

			//check taskgroups
			store.LockApplication(runAs + "." + appID)
			taskGroups, err := store.ListTaskGroups(runAs, appID)
			store.UnLockApplication(runAs + "." + appID)
			if err != nil {
				blog.Error("constraint EXCLUDE: list taskgroup(%s %s) err:%s", runAs, appID, err.Error())
				return false, errors.New("constraint EXCLUDE: list taskgroup err")
			}

			for _, taskGroup := range taskGroups {
				if taskGroup.HostName == hostname {
					blog.Info("constraint EXCLUDE: taskgroup(%s) is on offered host(%s), constraint not pass", taskGroup.ID, taskGroup.HostName)
					return false, nil
				}
			}
		}
	}

	blog.V(3).Infof("constraint EXCLUDE: constraint pass")

	return true, nil
}

func compareAttribute(one *mesos.Attribute, other *mesos.Attribute) (bool, error) {

	if one.GetName() != other.GetName() {
		blog.V(3).Infof("attribute name different(%s, %s)", one.GetName(), other.GetName())
		return false, nil
	}

	if one.GetType() == mesos.Value_SCALAR && other.GetType() == mesos.Value_SCALAR {
		// float compare ?
		if one.GetScalar().GetValue() == other.GetScalar().GetValue() {
			blog.V(3).Infof("scalar value (%f == %f)", one.GetScalar().GetValue(), other.GetScalar().GetValue())
			return true, nil
		}

		blog.V(3).Infof("scalar value (%f != %f)", one.GetScalar().GetValue(), other.GetScalar().GetValue())
		return false, nil
	}

	if one.GetType() == mesos.Value_TEXT && other.GetType() == mesos.Value_TEXT {
		if one.GetText().GetValue() == other.GetText().GetValue() {
			blog.V(3).Infof("scalar value (%s == %s)", one.GetText().GetValue(), other.GetText().GetValue())
			return true, nil
		}
		blog.V(3).Infof("scalar value (%s != %s)", one.GetText().GetValue(), other.GetText().GetValue())
		return false, nil
	}

	blog.Error("cannot compare attribute for value type(%d) and type(%d)", one.GetType(), other.GetType())
	return false, nil
}
