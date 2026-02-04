package helpers

import (
	"encoding/json"
	"fmt"
	"strings"
)

func Dump(values ...interface{}) {
	for _, v := range values {
		fmt.Println(ToJSON(v, "\t"))
	}
}

func ToJSON(i interface{}, indent string) string {
	s, _ := json.MarshalIndent(i, "", indent)
	return string(s)
}

func StringReplacer(val string, replacer map[string]string) string {
	for k, v := range replacer {
		val = strings.Replace(val, fmt.Sprintf("{{%s}}", k), v, -1)
	}
	return val
}

func InArrayString(val string, haystack []string) bool {
	for _, v := range haystack {
		if val == v {
			return true
		}
	}
	return false
}
func InArrayInt(val int, haystack []int) bool {
	for _, v := range haystack {
		if val == v {
			return true
		}
	}
	return false
}

type Debug struct {
	Property   string
	Error      error
	Additional string
}

func (e Debug) String() string {
	return fmt.Sprintf("ERROR (%v): %v | %v", e.Property, e.Error, e.Additional)
}

func RemoveDuplicateStr(strSlice []string) []string {
	allKeys := make(map[string]bool)
	list := []string{}
	for _, item := range strSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

func RemoveDuplicateInt(intSlice []int) []int {
	allKeys := make(map[int]bool)
	list := []int{}
	for _, item := range intSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}
