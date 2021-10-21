/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package lib

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
)

var opCharacterSet = []byte{'=', '!'}
var specialCharacterSet = []byte{',', '+', '(', ')'}

type Selector struct {
	Prefix      string
	SelectorStr string
	cursor      int
	conditions  []*operator.Condition
}

// TODO: error operation
// GetNextCondition get next label selector condition unit
func (s *Selector) GetNextCondition() *operator.Condition {
	if s.cursor >= len(s.SelectorStr) {
		return nil
	}
	// get key
	key := s.getWord(false)
	if key == "" {
		return nil
	}
	if s.Prefix != "" {
		key = s.Prefix + key
	}
	// get op
	op := s.getOperator()
	blog.Infof("getOperator: %s", op)
	// get value
	var cond *operator.Condition
	switch op {
	case operator.Ext:
		cond = operator.NewLeafCondition(op, operator.M{key: true})
	case operator.Eq, operator.Ne:
		// get Value
		value := s.getWord(false)
		if value == "" {
			return nil
		}
		cond = operator.NewLeafCondition(op, operator.M{key: value})
	case operator.In, operator.Nin:
		wordList := s.getWordList()
		if len(wordList) == 0 {
			return nil
		}
		cond = operator.NewLeafCondition(op, operator.M{key: wordList})
	}
	s.conditions = append(s.conditions, cond)
	blog.Infof("GetNextCondition: %+v", cond)
	// get expression end (, or end of string)
	s.getCharacter(',')
	return cond
}

// GetAllConditions get all conditions of labelSelector
func (s *Selector) GetAllConditions() []*operator.Condition {
	if s.conditions != nil {
		return s.conditions
	}
	var cond *operator.Condition
	for {
		cond = s.GetNextCondition()
		if cond == nil {
			break
		}
	}
	return s.conditions
}

// Clear reset the current cursor for labelSelector condition parsing
func (s *Selector) Clear() {
	s.cursor = 0
	s.conditions = nil
}

// getWord get an operator word or a key/value word
func (s *Selector) getWord(isOperator bool) string {
	if s.cursor >= len(s.SelectorStr) {
		return ""
	}
	var word = make([]byte, 0)
	for {
		if s.cursor >= len(s.SelectorStr) {
			break
		}
		c := s.SelectorStr[s.cursor]
		if s.cursor >= len(s.SelectorStr) ||
			s.inSpecialCharacterSet(c) ||
			(!isOperator && s.inOPCharacterSet(c)) {
			blog.Infof("break word: %v", c)
			break
		}
		word = append(word, c)
		s.cursor++
	}
	blog.Infof("getWord %s", string(word))
	return string(word)
}

// getWordList get value list enclosed by ()
func (s *Selector) getWordList() []string {
	// list start
	if !s.getCharacter('(') {
		return nil
	}

	wordList := make([]string, 0)
	for {
		word := s.getWord(false)
		if word == "" {
			return nil
		}
		wordList = append(wordList, word)
		if !s.getCharacter(',') {
			break
		}
	}

	// list end
	if !s.getCharacter(')') {
		return nil
	}
	blog.Infof("getWordList %+v", wordList)
	return wordList
}

// getOperator get operator of labelSelector
func (s *Selector) getOperator() operator.Operator {
	if s.cursor >= len(s.SelectorStr) {
		return operator.Ext
	}
	if s.getCharacter(',') {
		return operator.Ext
	}
	if s.getCharacter('=') {
		return operator.Eq
	}
	if s.getCharacter('+') {
		opstr := s.getWord(true)
		var op operator.Operator
		switch opstr {
		case "in":
			op = operator.In
		case "notin":
			op = operator.Nin
		case "=", "==":
			op = operator.Eq
		case "!=":
			op = operator.Ne
		default:
			// getOperator failed
			return operator.Tr
		}
		if s.getCharacter('+') {
			return op
		}
	}
	return operator.Tr
}

// getCharacter test whether next character is designated character
func (s *Selector) getCharacter(standard byte) bool {
	if s.cursor >= len(s.SelectorStr) {
		return false
	}
	if s.SelectorStr[s.cursor] == standard {
		s.cursor++
		return true
	}
	return false
}

// inSpecialCharacterSet test whether characted is in special character set
func (s *Selector) inSpecialCharacterSet(c byte) bool {
	for _, sc := range specialCharacterSet {
		if c == sc {
			return true
		}
	}
	return false
}

// inOPCharacterSet test whether characted is in operator character set
func (s *Selector) inOPCharacterSet(c byte) bool {
	for _, sc := range opCharacterSet {
		if c == sc {
			return true
		}
	}
	return false
}
