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

package printer

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/tidwall/pretty"
	"github.com/ugorji/go/codec"
)

const (
	timeFormatter = "2006-01-02 15:04:05"
)

// defaultTableWriter create the tablewriter instance
func defaultTableWriter() *tablewriter.Table {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeaderLine(false)
	table.SetRowLine(false)
	table.SetColWidth(150)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorder(false)
	table.SetColumnSeparator("")
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	return table
}

const (
	outputTypeJSON = "json"
	outputTypeWide = "wide"
)

func encodeJSON(v interface{}) error {
	var data []byte
	if err := encodeJSONWithIndent(4, v, &data); err != nil {
		return errors.Wrapf(err, "encode json failed")
	}
	fmt.Println(string(pretty.Color(pretty.Pretty(data), nil)))
	return nil
}

func encodeJSONWithIndent(indent int8, v interface{}, s *[]byte) error {
	enc := codec.NewEncoderBytes(s, &codec.JsonHandle{
		MapKeyAsString: true,
		Indent:         indent,
	})
	return enc.Encode(v)
}
