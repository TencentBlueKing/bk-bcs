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

// Package orm NOTES
//
//nolint:all // This legacy package is complex and not be used now, so don't lint it
package orm

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"unicode"

	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
)

var (
	// EndBracketsReg ....
	EndBracketsReg = regexp.MustCompile(`\([^()]*\)\s*$`)
)

var allowedBindRunes = []*unicode.RangeTable{unicode.Letter, unicode.Digit}

// bindArray NOTES
func bindArray(bindType int, query string, arg interface{}, m *reflectx.Mapper) (string, []interface{}, error) {
	bound, names, err := compileNamedQuery([]byte(query), bindType)
	if err != nil {
		return "", []interface{}{}, err
	}
	arrayValue := reflect.ValueOf(arg)
	arrayLen := arrayValue.Len()
	if arrayLen == 0 {
		return "", []interface{}{}, fmt.Errorf("length of array is 0: %#v", arg)
	}

	var arglist []interface{}
	for i := 0; i < arrayLen; i++ {
		elemArglist, err := bindArgs(names, arrayValue.Index(i).Interface(), m)
		if err != nil {
			return "", []interface{}{}, err
		}
		arglist = append(arglist, elemArglist...)
	}
	if arrayLen > 1 {
		bound = fixBound(bound, arrayLen)
	}
	return bound, arglist, nil
}

// fixBound NOTES
func fixBound(bound string, loop int) string {
	endBrackets := EndBracketsReg.FindString(bound)
	if endBrackets == "" {
		return bound
	}
	var buffer bytes.Buffer
	buffer.WriteString(bound)
	for i := 0; i < loop-1; i++ {
		buffer.WriteString(",")
		buffer.WriteString(endBrackets)
	}
	return buffer.String()
}

// compileNamedQuery NOTES
func compileNamedQuery(qs []byte, bindType int) (query string, names []string, err error) {
	names = make([]string, 0, 10)
	rebound := make([]byte, 0, len(qs))

	inName := false
	last := len(qs) - 1
	currentVar := 1
	name := make([]byte, 0, 10)

	for i, b := range qs {
		// a ':' while we're in a name is an error
		if b == ':' {
			// if this is the second ':' in a '::' escape sequence, append a ':'
			if inName && i > 0 && qs[i-1] == ':' {
				rebound = append(rebound, ':')
				inName = false
				continue
			} else if inName {
				err = errors.New("unexpected `:` while reading named param at " + strconv.Itoa(i))
				return query, names, err
			}
			inName = true
			name = []byte{}
		} else if inName && i > 0 && b == '=' {
			rebound = append(rebound, ':', '=')
			inName = false
			continue
			// if we're in a name, and this is an allowed character, continue
		} else if inName && (unicode.IsOneOf(allowedBindRunes, rune(b)) || b == '_' || b == '.') && i != last {
			// append the byte to the name if we are in a name and not on the last byte
			name = append(name, b)
			// if we're in a name and it's not an allowed character, the name is done
		} else if inName {
			inName = false
			// if this is the final byte of the string and it is part of the name, then
			// make sure to add it to the name
			if i == last && unicode.IsOneOf(allowedBindRunes, rune(b)) {
				name = append(name, b)
			}
			// add the string representation to the names list
			names = append(names, string(name))
			// add a proper bindvar for the bindType
			switch bindType {
			// oracle only supports named type bind vars even for positional
			case sqlx.NAMED:
				rebound = append(rebound, ':')
				rebound = append(rebound, name...)
			case sqlx.QUESTION, sqlx.UNKNOWN:
				rebound = append(rebound, '?')
			case sqlx.DOLLAR:
				rebound = append(rebound, '$')
				for _, b := range strconv.Itoa(currentVar) {
					rebound = append(rebound, byte(b))
				}
				currentVar++
			case sqlx.AT:
				rebound = append(rebound, '@', 'p')
				for _, b := range strconv.Itoa(currentVar) {
					rebound = append(rebound, byte(b))
				}
				currentVar++
			}
			// add this byte to string unless it was not part of the name
			if i != last {
				rebound = append(rebound, b)
			} else if !unicode.IsOneOf(allowedBindRunes, rune(b)) {
				rebound = append(rebound, b)
			}
		} else {
			// this is a normal byte and should just go onto the rebound query
			rebound = append(rebound, b)
		}
	}

	return string(rebound), names, err
}

// bindArgs NOTES
func bindArgs(names []string, arg interface{}, m *reflectx.Mapper) ([]interface{}, error) {
	arglist := make([]interface{}, 0, len(names))

	// grab the indirected value of arg
	v := reflect.ValueOf(arg)
	for v = reflect.ValueOf(arg); v.Kind() == reflect.Ptr; {
		v = v.Elem()
	}

	err := m.TraversalsByNameFunc(v.Type(), names, func(i int, t []int) error {
		if len(t) == 0 {
			return fmt.Errorf("could not find name %s in %#v", names[i], arg)
		}

		val := reflectx.FieldByIndexesReadOnly(v, t)
		arglist = append(arglist, val.Interface())

		return nil
	})

	return arglist, err
}
