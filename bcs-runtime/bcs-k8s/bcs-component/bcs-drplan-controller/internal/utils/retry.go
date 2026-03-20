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

package utils

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"k8s.io/klog/v2"

	drv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/api/v1alpha1"
)

const (
	DefaultRetryLimit        = 3
	DefaultRetryInterval     = 5 * time.Second
	DefaultBackoffMultiplier = 2.0
	MaxRetryInterval         = 5 * time.Minute
)

// RetryConfig holds retry configuration
type RetryConfig struct {
	Limit             int32
	Interval          time.Duration
	BackoffMultiplier float64
}

// ParseRetryPolicy parses RetryPolicy from CRD to RetryConfig
func ParseRetryPolicy(policy *drv1alpha1.RetryPolicy) (*RetryConfig, error) {
	if policy == nil {
		return &RetryConfig{
			Limit:             DefaultRetryLimit,
			Interval:          DefaultRetryInterval,
			BackoffMultiplier: DefaultBackoffMultiplier,
		}, nil
	}

	config := &RetryConfig{
		Limit:             DefaultRetryLimit,
		Interval:          DefaultRetryInterval,
		BackoffMultiplier: DefaultBackoffMultiplier,
	}

	// Parse limit
	if policy.Limit > 0 {
		config.Limit = policy.Limit
	}

	// Parse interval
	if policy.Interval != "" {
		interval, err := time.ParseDuration(policy.Interval)
		if err != nil {
			return nil, fmt.Errorf("invalid retry interval %s: %w", policy.Interval, err)
		}
		config.Interval = interval
	}

	// Parse backoff multiplier
	if policy.BackoffMultiplier != "" {
		multiplier, err := strconv.ParseFloat(policy.BackoffMultiplier, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid backoff multiplier %s: %w", policy.BackoffMultiplier, err)
		}
		if multiplier < 1.0 {
			return nil, fmt.Errorf("backoff multiplier must be >= 1.0, got %f", multiplier)
		}
		config.BackoffMultiplier = multiplier
	}

	return config, nil
}

// RetryFunc is a function that can be retried
type RetryFunc func(ctx context.Context, attempt int32) error

// RetryWithBackoff retries a function with exponential backoff
func RetryWithBackoff(ctx context.Context, config *RetryConfig, fn RetryFunc) error {
	if config == nil {
		config = &RetryConfig{
			Limit:             DefaultRetryLimit,
			Interval:          DefaultRetryInterval,
			BackoffMultiplier: DefaultBackoffMultiplier,
		}
	}

	var lastErr error
	var attempt int32

	for attempt = 0; attempt <= config.Limit; attempt++ {
		// Execute function
		err := fn(ctx, attempt)
		if err == nil {
			if attempt > 0 {
				klog.Infof("Action succeeded after %d retries", attempt)
			}
			return nil
		}

		lastErr = err

		// Check if we've reached the limit
		if attempt >= config.Limit {
			klog.Warningf("Action failed after %d retries: %v", attempt, err)
			break
		}

		// Check context cancellation before sleeping
		select {
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled: %w", ctx.Err())
		default:
		}

		// Calculate backoff duration
		backoffDuration := calculateBackoff(config.Interval, config.BackoffMultiplier, attempt)
		klog.V(4).Infof("Action failed (attempt %d/%d), retrying in %v: %v",
			attempt+1, config.Limit+1, backoffDuration, err)

		// Wait with context cancellation support
		timer := time.NewTimer(backoffDuration)
		select {
		case <-ctx.Done():
			timer.Stop()
			return fmt.Errorf("retry cancelled during backoff: %w", ctx.Err())
		case <-timer.C:
			// Continue to next attempt
		}
	}

	return fmt.Errorf("action failed after %d retries: %w", attempt, lastErr)
}

// calculateBackoff calculates the backoff duration for a given attempt
func calculateBackoff(baseInterval time.Duration, multiplier float64, attempt int32) time.Duration {
	// Calculate: baseInterval * (multiplier ^ attempt)
	duration := float64(baseInterval)
	for i := int32(0); i < attempt; i++ {
		duration *= multiplier
	}

	backoff := time.Duration(duration)

	// Cap at max interval
	if backoff > MaxRetryInterval {
		backoff = MaxRetryInterval
	}

	return backoff
}
