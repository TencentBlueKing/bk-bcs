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

package types

import (
	"errors"
	"fmt"
)

const (
	// DefaultMaxPageLimit is the default value of the max page limitation.
	DefaultMaxPageLimit = uint(1000)
)

// DefaultPageOption is the default BasePage's option.
var DefaultPageOption = &PageOption{
	EnableUnlimitedLimit: false,
	MaxLimit:             DefaultMaxPageLimit,
	DisabledSort:         false,
}

// PageOption defines the options to validate the
// BasePage's configuration.
type PageOption struct {
	// EnableUnlimitedLimit allows user to query resources with unlimited
	// limitation. if true, then the 'Limit' option will not be checked.
	EnableUnlimitedLimit bool `json:"enable_unlimited_limit"`
	// MaxLimit defines max limit value of a page.
	MaxLimit uint `json:"max_limit"`
	// DisableSort defines the sort field is not allowed to be defined by the user.
	// then system defined sort field is used.
	// Note: this option does not work when use the page to generate SQL expression,
	// which means call the method of BasePage's SQLExpr().
	DisabledSort bool `json:"disabled_sort"`
}

// Order is the direction when do sort operation.
type Order string

const (
	// Ascending sort data with ascending direction
	// this is the default sort direction.
	Ascending Order = "ASC"
	// Descending sort data with descending direction
	Descending Order = "DESC"
)

// Validate the sort direction is valid or not
func (sd Order) Validate() error {
	if len(sd) == 0 {
		return nil
	}

	switch sd {
	case Ascending:
	case Descending:
	default:
		return fmt.Errorf("unsupported sort direction: %s", sd)
	}

	return nil
}

// Order returns the sort direction, if not set, use
// ascending as the default direction.
func (sd Order) Order() Order {
	switch sd {
	case Ascending:
		return Ascending
	case Descending:
		return Descending
	default:
		// set Ascending as the default sort direction.
		return Descending
	}
}

// BasePage define the basic page limitation to query resources.
type BasePage struct {
	// Start is the start position of the queried resource's page.
	// Note:
	// 1. Start only works when the Count = false.
	// 2. Start's minimum value is 0, not 1.
	// 3. if PageOption.EnableUnlimitedLimit = true, then Start = 0
	//   and Limit = 0 means query all the resources at once.
	Start uint32 `json:"start"`
	// Limit is the total returned resources at once query.
	// Limit only works when the Count = false.
	Limit uint `json:"limit"`
	// Sort defines use which field to sort the queried resources.
	// only 'one' field is supported to do sort.
	// Sort only works when the Count = false.
	Sort string `json:"sort"`
	// Order is the direction when do sort operation.
	// it works only when the Sort is set.
	Order Order `json:"order"`
	// All defines whether query all the resources at once.
	All bool `json:"all"`
}

// Offset 偏移量
func (bp *BasePage) Offset() int {
	return int(bp.Start)
}

// LimitInt Limit int类型值
func (bp *BasePage) LimitInt() int {
	return int(bp.Limit)
}

// Validate the base page's options.
// if the page option is not set, use the default configuration.
func (bp BasePage) Validate(opt ...*PageOption) (err error) {
	if len(opt) >= 2 {
		return errors.New("at most one page options is allows")
	}

	maxLimit := DefaultMaxPageLimit
	enableUnlimited := false
	if len(opt) != 0 {
		// option is configured, validate it
		one := opt[0]
		if one.MaxLimit > 0 {
			maxLimit = one.MaxLimit
		}

		enableUnlimited = one.EnableUnlimitedLimit

		if one.DisabledSort {
			if len(bp.Sort) > 0 {
				return errors.New("page.sort is not allowed")
			}

			if len(bp.Order) > 0 {
				return errors.New("invalid page.order, page.order is not allowed")
			}
		}
	}

	// only validate when not query all the resources
	if !enableUnlimited && !bp.All {
		// not allow the unlimited query, then valid this.
		// if the user is not allowed to query with unlimited limit, then
		// 1. limit should >=1
		// 2. validate whether the limit is larger than the max limit value
		if bp.Limit == 0 {
			return errors.New("page.limit value should >= 1")
		}

		if bp.Limit > maxLimit {
			return fmt.Errorf("invalid page.limit max value: %d", maxLimit)
		}
	}

	// if direction is set, then validate it.
	if len(bp.Order) != 0 {
		if err := bp.Order.Validate(); err != nil {
			return err
		}
	}

	return nil
}
