// KIProtect Go-Helpers - Golang Utility Functions
// Copyright (C) 2019-2021  KIProtect GmbH (HRB 208395B) - Germany
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the 3-Clause BSD License.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// license for more details.
//
// You should have received a copy of the 3-Clause BSD License
// along with this program.  If not, see <https://opensource.org/licenses/BSD-3-Clause>.

package errors

import (
	"encoding/json"
	"reflect"
)

type StructuredError struct {
	Message string      `json:"message"`
	Code    string      `json:"code"`
	Data    interface{} `json:"data,omitempty"`
}

type ErrorScope int

const (
	ExternalError ErrorScope = 1
	InternalError ErrorScope = 2
)

type StructuredErrorWithTraceback struct {
	StructuredError
	Traceback []StructuredError `json:"traceback"`
}

type ChainableError interface {
	Error() string
	Message() string
	Scope() ErrorScope
	Data() interface{}
	Code() string
	Parent() error
}

type BaseChainableError struct {
	data    interface{}
	code    string
	scope   ErrorScope
	message string
	parent  error
}

func MakeExternalError(message, code string, data interface{}, parent error) *BaseChainableError {
	return MakeError(ExternalError, message, code, data, parent)
}

func MakeInternalError(message, code string, data interface{}, parent error) *BaseChainableError {
	return MakeError(InternalError, message, code, data, parent)
}

func MakeError(scope ErrorScope, message, code string, data interface{}, parent error) *BaseChainableError {
	return &BaseChainableError{
		message: message,
		code:    code,
		scope:   scope,
		data:    data,
		parent:  parent,
	}
}

func (c *BaseChainableError) MarshalJSON() ([]byte, error) {
	return json.Marshal(MakeStructuredErrorWithTraceback(c, ExternalError))
}

func (c *BaseChainableError) Scope() ErrorScope {
	return c.scope
}

func (c *BaseChainableError) Message() string {
	return c.message
}

func (c *BaseChainableError) Data() interface{} {
	return c.data
}

func (c *BaseChainableError) Code() string {
	return c.code
}

func (c *BaseChainableError) Error() string {
	message := c.message
	parent := c.parent
	for {
		if parent == nil {
			break
		}
		chainableParent, ok := parent.(ChainableError)
		if ok {
			message += ": " + chainableParent.Message()
			parent = chainableParent.Parent()
		} else {
			message += ": " + parent.Error()
			parent = nil
		}
	}
	return message
}

func (c *BaseChainableError) Parent() error {
	return c.parent
}

func MakeStructuredError(err ChainableError, scope ErrorScope) StructuredError {
	if err.Scope() > scope {
		return StructuredError{
			Message: "undisclosed error",
			Code:    err.Code(),
		}
	}
	return StructuredError{
		Message: err.Message(),
		Code:    err.Code(),
		Data:    err.Data(),
	}
}

func MakeStructuredErrorWithTraceback(err ChainableError, scope ErrorScope) StructuredErrorWithTraceback {
	traceback := make([]StructuredError, 0)
	parent := err.Parent()
	for {
		if parent == nil {
			break
		}
		cParent, ok := parent.(ChainableError)
		if !ok {
			// if we show internal errors we include the error message
			// as well as its type to improve debugability.
			if scope >= InternalError {
				t := reflect.TypeOf(parent)
				traceback = append(traceback, StructuredError{
					Message: parent.Error(),
					Code:    "GO-ERROR",
					Data: map[string]interface{}{
						"type": t.Name(),
					},
				})
			}
			break
		}
		traceback = append(traceback, MakeStructuredError(cParent, scope))
		parent = cParent.Parent()
	}
	return StructuredErrorWithTraceback{
		StructuredError: MakeStructuredError(err, scope),
		Traceback:       traceback,
	}
}
