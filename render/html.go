// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package render

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

func parseTag(b []byte) (html.Token, error) {
	tk := html.NewTokenizer(bytes.NewReader(b))
	tt := tk.Next()
	if tt == html.ErrorToken {
		return html.Token{}, fmt.Errorf("failed to parse %q", b)
	}
	return tk.Token(), nil
}

func unmarshalAttrs(attrs []html.Attribute, dst interface{}) error {
	am := make(map[string]string, len(attrs))
	for _, a := range attrs {
		if _, ok := am[a.Key]; ok {
			return fmt.Errorf("duplicate %q attribute", a.Key)
		}
		am[a.Key] = a.Val
	}

	dv := reflect.ValueOf(dst).Elem()
	dt := reflect.TypeOf(dst).Elem()
	for i := 0; i < dv.NumField(); i++ {
		tag := dt.Field(i).Tag.Get("html")
		if tag == "" {
			continue
		}
		sv, ok := am[tag]
		if !ok {
			continue
		}

		fn := dt.Field(i).Name
		fld := dv.Field(i)
		switch fld.Kind() {
		case reflect.Bool:
			fld.SetBool(true)
		case reflect.Int:
			v, err := strconv.Atoi(sv)
			if err != nil {
				return fmt.Errorf("unable to parse field %v value %q as int", fn, sv)
			}
			fld.SetInt(int64(v))
		case reflect.String:
			// Replace newlines with spaces to handle wrapped values.
			fld.SetString(strings.ReplaceAll(sv, "\n", " "))
		default:
			return fmt.Errorf("field %v has unsupported type %v", fn, fld.Kind())
		}
	}

	return nil
}
