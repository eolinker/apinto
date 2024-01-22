// Copyright 2012 Gary Burd
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package utils

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

type Error string

func (err Error) Error() string { return string(err) }

type Scanner interface {
	// RedisScan assigns a value from a Redis value. The argument src is one of
	// the reply types listed in the section `Executing Commands`.
	//
	// An error should be returned if the value cannot be stored without
	// loss of information.
	RedisScan(src interface{}) error
}

func ensureLen(d reflect.Value, n int) {
	if n > d.Cap() {
		d.Set(reflect.MakeSlice(d.Type(), n, n))
	} else {
		d.SetLen(n)
	}
}

func cannotConvert(d reflect.Value, s interface{}) error {
	var name string
	switch s.(type) {
	case string:
		name = "Redis simple string"
	case Error:
		name = "Redis error"
	case int64:
		name = "Redis integer"
	case []byte:
		name = "Redis bulk string"
	case []interface{}:
		name = "Redis array"
	default:
		name = reflect.TypeOf(s).String()
	}
	return fmt.Errorf("cannot convert from %s to %s", name, d.Type())
}

func convertAssignBulkString(d reflect.Value, s []byte) (err error) {
	switch d.Type().Kind() {
	case reflect.Float32, reflect.Float64:
		var x float64
		x, err = strconv.ParseFloat(string(s), d.Type().Bits())
		d.SetFloat(x)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var x int64
		x, err = strconv.ParseInt(string(s), 10, d.Type().Bits())
		d.SetInt(x)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var x uint64
		x, err = strconv.ParseUint(string(s), 10, d.Type().Bits())
		d.SetUint(x)
	case reflect.Bool:
		var x bool
		x, err = strconv.ParseBool(string(s))
		d.SetBool(x)
	case reflect.String:
		d.SetString(string(s))
	case reflect.Slice:
		if d.Type().Elem().Kind() != reflect.Uint8 {
			err = cannotConvert(d, s)
		} else {
			d.SetBytes(s)
		}
	default:
		err = cannotConvert(d, s)
	}
	return
}

func convertAssignInt(d reflect.Value, s int64) (err error) {
	switch d.Type().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		d.SetInt(s)
		if d.Int() != s {
			err = strconv.ErrRange
			d.SetInt(0)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if s < 0 {
			err = strconv.ErrRange
		} else {
			x := uint64(s)
			d.SetUint(x)
			if d.Uint() != x {
				err = strconv.ErrRange
				d.SetUint(0)
			}
		}
	case reflect.Bool:
		d.SetBool(s != 0)
	default:
		err = cannotConvert(d, s)
	}
	return
}

func convertAssignValue(d reflect.Value, s interface{}) (err error) {
	if d.Kind() != reflect.Ptr {
		if d.CanAddr() {
			d2 := d.Addr()
			if d2.CanInterface() {
				if scanner, ok := d2.Interface().(Scanner); ok {
					return scanner.RedisScan(s)
				}
			}
		}
	} else if d.CanInterface() {
		// Already a reflect.Ptr
		if d.IsNil() {
			d.Set(reflect.New(d.Type().Elem()))
		}
		if scanner, ok := d.Interface().(Scanner); ok {
			return scanner.RedisScan(s)
		}
	}

	switch s := s.(type) {
	case []byte:
		err = convertAssignBulkString(d, s)
	case int64:
		err = convertAssignInt(d, s)
	default:
		err = cannotConvert(d, s)
	}
	return err
}

func convertAssignArray(d reflect.Value, s []interface{}) error {
	if d.Type().Kind() != reflect.Slice {
		return cannotConvert(d, s)
	}
	ensureLen(d, len(s))
	for i := 0; i < len(s); i++ {
		if err := convertAssignValue(d.Index(i), s[i]); err != nil {
			return err
		}
	}
	return nil
}

func convertAssign(d interface{}, s interface{}) (err error) {
	if scanner, ok := d.(Scanner); ok {
		return scanner.RedisScan(s)
	}

	// Handle the most common destination types using type switches and
	// fall back to reflection for all other types.
	switch s := s.(type) {
	case nil:
		// ignore
	case []byte:
		switch d := d.(type) {
		case *string:
			*d = string(s)
		case *int:
			*d, err = strconv.Atoi(string(s))
		case *bool:
			*d, err = strconv.ParseBool(string(s))
		case *[]byte:
			*d = s
		case *interface{}:
			*d = s
		case nil:
			// skip value
		default:
			if d := reflect.ValueOf(d); d.Type().Kind() != reflect.Ptr {
				err = cannotConvert(d, s)
			} else {
				err = convertAssignBulkString(d.Elem(), s)
			}
		}
	case int64:
		switch d := d.(type) {
		case *int:
			x := int(s)
			if int64(x) != s {
				err = strconv.ErrRange
				x = 0
			}
			*d = x
		case *bool:
			*d = s != 0
		case *interface{}:
			*d = s
		case nil:
			// skip value
		default:
			if d := reflect.ValueOf(d); d.Type().Kind() != reflect.Ptr {
				err = cannotConvert(d, s)
			} else {
				err = convertAssignInt(d.Elem(), s)
			}
		}
	case string:
		switch d := d.(type) {
		case *string:
			*d = s
		case *interface{}:
			*d = s
		case nil:
			// skip value
		default:
			err = cannotConvert(reflect.ValueOf(d), s)
		}
	case []interface{}:
		switch d := d.(type) {
		case *[]interface{}:
			*d = s
		case *interface{}:
			*d = s
		case nil:
			// skip value
		default:
			if d := reflect.ValueOf(d); d.Type().Kind() != reflect.Ptr {
				err = cannotConvert(d, s)
			} else {
				err = convertAssignArray(d.Elem(), s)
			}
		}
	case Error:
		err = s
	default:
		err = cannotConvert(reflect.ValueOf(d), s)
	}
	return
}

// Scan copies from src to the values pointed at by dest.
//
// Scan uses RedisScan if available otherwise:
//
// The values pointed at by dest must be an integer, float, boolean, string,
// []byte, interface{} or slices of these types. Scan uses the standard strconv
// package to convert bulk strings to numeric and boolean types.
//
// If a dest value is nil, then the corresponding src value is skipped.
//
// If a src element is nil, then the corresponding dest value is not modified.
//
// To enable easy use of Scan in a loop, Scan returns the slice of src
// following the copied values.
func Scan(src []interface{}, dest ...interface{}) ([]interface{}, error) {
	if len(src) < len(dest) {
		return nil, errors.New("redigo.Scan: array short")
	}
	var err error
	for i, d := range dest {
		err = convertAssign(d, src[i])
		if err != nil {
			err = fmt.Errorf("redigo.Scan: cannot assign to dest %d: %v", i, err)
			break
		}
	}
	return src[len(dest):], err
}
