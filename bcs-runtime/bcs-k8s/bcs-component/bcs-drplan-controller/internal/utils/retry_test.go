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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	drv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/api/v1alpha1"
)

var _ = Describe("ParseRetryPolicy", func() {
	Context("with nil policy", func() {
		It("should return defaults", func() {
			config, err := ParseRetryPolicy(nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(config.Limit).To(Equal(int32(DefaultRetryLimit)))
			Expect(config.Interval).To(Equal(DefaultRetryInterval))
			Expect(config.BackoffMultiplier).To(Equal(DefaultBackoffMultiplier))
		})
	})

	Context("with empty policy (zero values)", func() {
		It("should use defaults for unset fields", func() {
			policy := &drv1alpha1.RetryPolicy{}
			config, err := ParseRetryPolicy(policy)
			Expect(err).NotTo(HaveOccurred())
			Expect(config.Limit).To(Equal(int32(DefaultRetryLimit)))
			Expect(config.Interval).To(Equal(DefaultRetryInterval))
			Expect(config.BackoffMultiplier).To(Equal(DefaultBackoffMultiplier))
		})
	})

	Context("with custom values", func() {
		It("should parse all custom fields", func() {
			policy := &drv1alpha1.RetryPolicy{
				Limit:             5,
				Interval:          "10s",
				BackoffMultiplier: "3.0",
			}
			config, err := ParseRetryPolicy(policy)
			Expect(err).NotTo(HaveOccurred())
			Expect(config.Limit).To(Equal(int32(5)))
			Expect(config.Interval).To(Equal(10 * time.Second))
			Expect(config.BackoffMultiplier).To(Equal(3.0))
		})

		It("should parse minute-based intervals", func() {
			policy := &drv1alpha1.RetryPolicy{
				Interval: "2m",
			}
			config, err := ParseRetryPolicy(policy)
			Expect(err).NotTo(HaveOccurred())
			Expect(config.Interval).To(Equal(2 * time.Minute))
		})
	})

	Context("with invalid values", func() {
		It("should return error for invalid interval format", func() {
			policy := &drv1alpha1.RetryPolicy{Interval: "not-a-duration"}
			_, err := ParseRetryPolicy(policy)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid retry interval"))
		})

		It("should return error for non-numeric backoff multiplier", func() {
			policy := &drv1alpha1.RetryPolicy{BackoffMultiplier: "abc"}
			_, err := ParseRetryPolicy(policy)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid backoff multiplier"))
		})

		It("should return error for backoff multiplier < 1.0", func() {
			policy := &drv1alpha1.RetryPolicy{BackoffMultiplier: "0.5"}
			_, err := ParseRetryPolicy(policy)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("must be >= 1.0"))
		})
	})
})

var _ = Describe("RetryWithBackoff", func() {
	Context("when function succeeds immediately", func() {
		It("should return nil without retries", func() {
			callCount := 0
			err := RetryWithBackoff(context.Background(), &RetryConfig{
				Limit:             3,
				Interval:          10 * time.Millisecond,
				BackoffMultiplier: 1.0,
			}, func(_ context.Context, _ int32) error {
				callCount++
				return nil
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(callCount).To(Equal(1))
		})
	})

	Context("when function fails then succeeds", func() {
		It("should retry and eventually succeed", func() {
			callCount := 0
			err := RetryWithBackoff(context.Background(), &RetryConfig{
				Limit:             3,
				Interval:          10 * time.Millisecond,
				BackoffMultiplier: 1.0,
			}, func(_ context.Context, _ int32) error {
				callCount++
				if callCount < 3 {
					return fmt.Errorf("temporary error")
				}
				return nil
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(callCount).To(Equal(3))
		})
	})

	Context("when function always fails", func() {
		It("should exhaust retries and return last error", func() {
			callCount := 0
			err := RetryWithBackoff(context.Background(), &RetryConfig{
				Limit:             2,
				Interval:          10 * time.Millisecond,
				BackoffMultiplier: 1.0,
			}, func(_ context.Context, _ int32) error {
				callCount++
				return fmt.Errorf("persistent error %d", callCount)
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("persistent error"))
			Expect(callCount).To(Equal(3)) // attempt 0, 1, 2
		})
	})

	Context("with context cancellation", func() {
		It("should stop retrying when context is canceled", func() {
			ctx, cancel := context.WithCancel(context.Background())
			callCount := 0
			err := RetryWithBackoff(ctx, &RetryConfig{
				Limit:             10,
				Interval:          50 * time.Millisecond,
				BackoffMultiplier: 1.0,
			}, func(_ context.Context, _ int32) error {
				callCount++
				if callCount == 2 {
					cancel()
				}
				return fmt.Errorf("keep failing")
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("canceled"))
		})
	})

	Context("with nil config", func() {
		It("should use default configuration", func() {
			callCount := 0
			err := RetryWithBackoff(context.Background(), nil, func(_ context.Context, _ int32) error {
				callCount++
				return nil
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(callCount).To(Equal(1))
		})
	})
})

var _ = Describe("calculateBackoff", func() {
	It("should return base interval for attempt 0", func() {
		d := calculateBackoff(time.Second, 2.0, 0)
		Expect(d).To(Equal(time.Second))
	})

	It("should apply multiplier for each attempt", func() {
		d := calculateBackoff(time.Second, 2.0, 1)
		Expect(d).To(Equal(2 * time.Second))

		d = calculateBackoff(time.Second, 2.0, 2)
		Expect(d).To(Equal(4 * time.Second))

		d = calculateBackoff(time.Second, 2.0, 3)
		Expect(d).To(Equal(8 * time.Second))
	})

	It("should cap at MaxRetryInterval", func() {
		d := calculateBackoff(time.Minute, 10.0, 5)
		Expect(d).To(Equal(MaxRetryInterval))
	})

	It("should handle multiplier of 1.0 (no backoff)", func() {
		d := calculateBackoff(5*time.Second, 1.0, 3)
		Expect(d).To(Equal(5 * time.Second))
	})
})
