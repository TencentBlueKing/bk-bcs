package xrequests

import (
	"encoding/json"
	"net/url"
	"reflect"
	"strconv"
	"time"
)

// buildQuery ...
func buildQuery(query any) (map[string]string, error) {
	switch v := reflect.ValueOf(query); v.Kind() {
	case reflect.Ptr:
		return buildQuery(v.Elem().Interface())
	case reflect.Struct:
		return fromQueryStruct(v.Interface())
	case reflect.String:
		return fromQueryString(v.String())
	case reflect.Map:
		return fromQueryMap(v.Interface())
	}
	return nil, nil
}

func fromQueryMap(query any) (map[string]string, error) {
	return fromQueryStruct(query)
}

func fromQueryStruct(query any) (map[string]string, error) {
	if queryBytes, err := json.Marshal(query); err != nil {
		return nil, err
	} else {
		var mapping map[string]any
		if err := json.Unmarshal(queryBytes, &mapping); err != nil {
			return nil, err
		}
		queryMap := make(map[string]string, 0)
		for k, v := range mapping {
			var s string
			switch t := v.(type) {
			case string:
				s = t
			case float64:
				s = strconv.FormatFloat(t, 'f', -1, 64)
			case time.Time:
				s = t.Format(time.RFC3339)
			default:
				j, err := json.Marshal(v)
				if err != nil {
					continue
				}
				s = string(j)
			}
			queryMap[k] = s
		}
		return queryMap, nil
	}
}

func fromQueryString(query string) (map[string]string, error) {
	queryMap := make(map[string]string, 0)
	if err := json.Unmarshal([]byte(query), &queryMap); err == nil {
		return queryMap, nil
	}
	mapList, err := url.ParseQuery(query)
	for k, v := range mapList {
		queryMap[k] = v[0]
	}
	return queryMap, err
}
