package types

import (
	"reflect"
	"strings"
	"unicode"
)

const (
	tagName         = "auditdb"
	indexIdentifier = "index"
)

func getIndexes(event interface{}) (map[string]string, map[string]interface{}) {
	v := reflect.ValueOf(event)
	t := reflect.TypeOf(event)

	if t == nil {
		return nil, nil
	}

	indexes := map[string]string{}
	rest := map[string]interface{}{}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if !v.Field(i).CanInterface() {
			continue
		}

		if field.Type.Name() != "string" {
			if field.Tag.Get("json") != "-" {
				rest[lowerInitial(field.Name)] = v.Field(i).Interface()
			}
			continue
		}

		tag := field.Tag.Get(tagName)
		if tag != indexIdentifier {
			rest[lowerInitial(field.Name)] = v.Field(i).Interface()
			continue
		}

		name := ""
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" {
			jsonTags := strings.Split(jsonTag, ",")

			if len(jsonTags) > 0 {
				name = jsonTags[0]
			}
		} else {
			name = lowerInitial(field.Name)
		}

		indexes[name] = v.Field(i).Interface().(string)
	}

	return indexes, rest
}

func lowerInitial(str string) string {
	for i, v := range str {
		return string(unicode.ToLower(v)) + str[i+1:]
	}
	return ""
}
