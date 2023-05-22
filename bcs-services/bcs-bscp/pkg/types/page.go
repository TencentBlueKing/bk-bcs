/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package types

import (
	"errors"
	"fmt"
	"strconv"

	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/runtime/filter"
)

const (
	// DefaultMaxPageLimit is the default value of the max page limitation.
	DefaultMaxPageLimit = uint(200)
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
	// Count describe if this query only return the total request
	// count of the resources.
	// If true, then the request will only return the total count
	// without the resource's detail infos. and start, limit must
	// be 0.
	Count bool `json:"count"`
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

	if bp.Count {
		if bp.Start > 0 {
			return errors.New("count is enabled, page.start should be 0")
		}

		if bp.Limit > 0 {
			return errors.New("count is enabled, page.limit should be 0")
		}

		if len(bp.Sort) > 0 {
			return errors.New("count is enabled, page.sort should be null")
		}

		if len(bp.Order) > 0 {
			return errors.New("count is enabled, page.order should be empty")
		}

		return nil
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

	if enableUnlimited {
		// allow the unlimited query, then valid this.
		if bp.Start < 0 || bp.Limit < 0 {
			return errors.New("page.start >= 0, page.limit value should >= 0")
		}
	} else {
		// if the user is not allowed to query with unlimited limit, then
		// 1. limit should >=1
		// 2. validate whether the limit is larger than the max limit value
		if bp.Limit <= 0 {
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

// PageSQLOption defines the options to generate a sql expression
// based on the BasePage.
type PageSQLOption struct {
	// Sort defines the field to do sort.
	// Note:
	// 1. If set, then user defined Sort field will be overlapped.
	// 2. Sort field should always be an indexed field in db.
	Sort SortOption `json:"sort"`
}

// SortOption defines how to set the order column when do the BasePage.SQLExpr
// operation.
type SortOption struct {
	// Sort defines the sorted column.
	Sort string `json:"sort"`
	// IfNotPresent means if the sort column is not defined by user, then
	// use this Sort column as default.
	IfNotPresent bool `json:"if_not_present"`
	// ForceOverlap means no matter what sort column defined, use this
	// Sort column overlapped.
	// Note: ForceOverlap option have more priority than IfNotPresent
	ForceOverlap bool `json:"force_overlap"`
}

// SQLExpr return the expression of the query clause based one the page options.
// Note:
//  1. do not call this, when it's a count request.
//  2. if sort is not set, use the default resource's identity 'id' as the sort key.
//  3. if Sort is set by the system(PageSQLOption.Sort), then use its Sort value
//     according to the various options.
//
// see the test case to get more returned example and learn the supported scenarios.
func (bp BasePage) SQLExpr(ps *PageSQLOption) (where string, err error) {
	defer func() {
		if err != nil {
			err = errf.New(errf.InvalidParameter, err.Error())
		}
	}()

	if ps == nil {
		return "", errors.New("page sql option is nil")
	}

	if bp.Count {
		// this is a count query clause.
		return "", errors.New("page.count is enabled, do not support generate SQL expression")
	}

	var sort string
	if ps.Sort.ForceOverlap {
		// force overlapped user defined sort field.
		sort = ps.Sort.Sort
	} else {
		if ps.Sort.IfNotPresent && len(bp.Sort) == 0 {
			// user note defined sort, then use default sort.
			sort = ps.Sort.Sort
		} else {
			// use user defined sort column
			sort = bp.Sort
		}
	}

	if len(sort) == 0 {
		// if sort is not set, use the default resource's
		// identity id as the default sort column.
		sort = "id"
	}
	var sqlSentence []string
	sqlSentence = append(sqlSentence, " ORDER BY ", sort)
	expr := filter.SqlJoint(sqlSentence)

	if bp.Start == 0 && bp.Limit == 0 {
		// this is a special scenario, which means query all the resources at once.
		return fmt.Sprintf(" %s %s", expr, bp.Order.Order()), nil
	}

	// if Start >=1, then Limit can not be 0.
	if bp.Limit == 0 {
		return "", errors.New("page.limit value should >= 1")
	}

	// bp.Limit is > 0, already validated upper.
	var sqlSentenceLimit []string
	sqlSentenceLimit = append(sqlSentenceLimit, expr, " ", string(bp.Order.Order()), " LIMIT ", strconv.Itoa(int(bp.Limit)), " OFFSET ", strconv.Itoa(int(bp.Start)))
	expr = filter.SqlJoint(sqlSentenceLimit)
	return expr, nil
}
