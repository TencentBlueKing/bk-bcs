/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"k8s.io/klog"
)

// autoScaling is the interface represents a specific aspect of the auto-scaling service provided by AWS SDK for use in CA
type autoScaling interface {
	DescribeAutoScalingGroupsPages(input *autoscaling.DescribeAutoScalingGroupsInput, fn func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool) error
	DescribeLaunchConfigurations(*autoscaling.DescribeLaunchConfigurationsInput) (*autoscaling.DescribeLaunchConfigurationsOutput, error)
	DescribeTagsPages(input *autoscaling.DescribeTagsInput, fn func(*autoscaling.DescribeTagsOutput, bool) bool) error
	SetDesiredCapacity(input *autoscaling.SetDesiredCapacityInput) (*autoscaling.SetDesiredCapacityOutput, error)
	TerminateInstanceInAutoScalingGroup(input *autoscaling.TerminateInstanceInAutoScalingGroupInput) (*autoscaling.TerminateInstanceInAutoScalingGroupOutput, error)
}

// autoScalingWrapper provides several utility methods over the auto-scaling service provided by AWS SDK
type autoScalingWrapper struct {
	autoScaling
	launchConfigurationInstanceTypeCache map[string]string
}

func (m autoScalingWrapper) getInstanceTypeByLCName(name string) (string, error) {
	if instanceType, found := m.launchConfigurationInstanceTypeCache[name]; found {
		return instanceType, nil
	}

	params := &autoscaling.DescribeLaunchConfigurationsInput{
		LaunchConfigurationNames: []*string{aws.String(name)},
		MaxRecords:               aws.Int64(1),
	}
	launchConfigurations, err := m.DescribeLaunchConfigurations(params)
	if err != nil {
		klog.V(4).Infof("Failed LaunchConfiguration info request for %s: %v", name, err)
		return "", err
	}
	if len(launchConfigurations.LaunchConfigurations) < 1 {
		return "", fmt.Errorf("unable to get first LaunchConfiguration for %s", name)
	}

	instanceType := *launchConfigurations.LaunchConfigurations[0].InstanceType
	m.launchConfigurationInstanceTypeCache[name] = instanceType
	return instanceType, nil
}

func (m *autoScalingWrapper) getAutoscalingGroupsByNames(names []string) ([]*autoscaling.Group, error) {
	if len(names) == 0 {
		return nil, nil
	}

	asgs := make([]*autoscaling.Group, 0)

	// AWS only accepts up to 50 ASG names as input, describe them in batches
	for i := 0; i < len(names); i += maxAsgNamesPerDescribe {
		end := i + maxAsgNamesPerDescribe

		if end > len(names) {
			end = len(names)
		}

		input := &autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: aws.StringSlice(names[i:end]),
			MaxRecords:            aws.Int64(maxRecordsReturnedByAPI),
		}
		if err := m.DescribeAutoScalingGroupsPages(input, func(output *autoscaling.DescribeAutoScalingGroupsOutput, _ bool) bool {
			asgs = append(asgs, output.AutoScalingGroups...)
			// We return true while we want to be called with the next page of
			// results, if any.
			return true
		}); err != nil {
			return nil, err
		}
	}

	return asgs, nil
}

func (m *autoScalingWrapper) getAutoscalingGroupNamesByTags(kvs map[string]string) ([]string, error) {
	// DescribeTags does an OR query when multiple filters on different tags are
	// specified. In other words, DescribeTags returns [asg1, asg1] for keys
	// [t1, t2] when there's only one asg tagged both t1 and t2.
	filters := []*autoscaling.Filter{}
	for key, value := range kvs {
		filter := &autoscaling.Filter{
			Name:   aws.String("key"),
			Values: []*string{aws.String(key)},
		}
		filters = append(filters, filter)
		if value != "" {
			filters = append(filters, &autoscaling.Filter{
				Name:   aws.String("value"),
				Values: []*string{aws.String(value)},
			})
		}
	}

	tags := []*autoscaling.TagDescription{}
	input := &autoscaling.DescribeTagsInput{
		Filters:    filters,
		MaxRecords: aws.Int64(maxRecordsReturnedByAPI),
	}
	if err := m.DescribeTagsPages(input, func(out *autoscaling.DescribeTagsOutput, _ bool) bool {
		tags = append(tags, out.Tags...)
		// We return true while we want to be called with the next page of
		// results, if any.
		return true
	}); err != nil {
		return nil, err
	}

	// According to how DescribeTags API works, the result contains ASGs which
	// not all but only subset of tags are associated. Explicitly select ASGs to
	// which all the tags are associated so that we won't end up calling
	// DescribeAutoScalingGroups API multiple times on an ASG.
	asgNames := []string{}
	asgNameOccurrences := make(map[string]int)
	for _, t := range tags {
		asgName := aws.StringValue(t.ResourceId)
		occurrences := asgNameOccurrences[asgName] + 1
		if occurrences >= len(kvs) {
			asgNames = append(asgNames, asgName)
		}
		asgNameOccurrences[asgName] = occurrences
	}

	return asgNames, nil
}
