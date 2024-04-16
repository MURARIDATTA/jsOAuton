package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

func main() {
	inputJSON := `{
		"number_1": {"N": "1.50"},
		"string_1": {"S": "784498 "},
		"string_2": {"S": "2014-07-16T20:55:46Z"},
		"map_1": {
			"M": {
				"bool_1": {"BOOL": "truthy"},
				"null_1": {"NULL ": "true"},
				"list_1": {
					"L": [
						{"S": ""},
						{"N": "011"},
						{"N": "5215s"},
						{"BOOL": "f"},
						{"NULL": "0"}
					]
				}
			}
		}
	}`

	var data map[string]json.RawMessage
	err := json.Unmarshal([]byte(inputJSON), &data)
	if err != nil {
		log.Fatalf("Error parsing JSON: %s", err)
	}

	output := make(map[string]interface{})
	transformJSON(data, output)
	outputJSON, err := json.MarshalIndent([]interface{}{output}, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling JSON: %s", err)
	}

	fmt.Println(string(outputJSON))
}

func transformJSON(data map[string]json.RawMessage, output map[string]interface{}) {
	for key, raw := range data {
		if key = strings.TrimSpace(key); key == "" {
			continue
		}

		var field map[string]interface{}
		if err := json.Unmarshal(raw, &field); err != nil {
			continue
		}

		for typ, val := range field {
			typ = strings.TrimSpace(typ)
			switch typ {
			case "S":
				transformString(key, val.(string), output)
			case "N":
				transformNumber(key, val.(string), output)
			case "BOOL":
				transformBool(key, val.(string), output)
			case "NULL":
				transformNull(key, val.(string), output)
			case "L":
				transformList(key, val, output)
			case "M":
				transformMap(key, val, output)
			}
		}
	}
}

func transformString(key, value string, output map[string]interface{}) {
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		output[key] = t.Unix()
	} else {
		output[key] = value
	}
}

func transformNumber(key, value string, output map[string]interface{}) {
	value = strings.TrimSpace(value)
	if n, err := strconv.ParseFloat(value, 64); err == nil {
		output[key] = n
	}
}

func transformBool(key, value string, output map[string]interface{}) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "1" || value == "t" || value == "true" {
		output[key] = true
	} else if value == "0" || value == "f" || value == "false" {
		output[key] = false
	}
}

func transformNull(key, value string, output map[string]interface{}) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "1" || value == "t" || value == "true" {
		output[key] = nil
	}
}

func transformList(key string, value interface{}, output map[string]interface{}) {
	var items []interface{}
	if dataList, ok := value.([]interface{}); ok {
		for _, item := range dataList {
			var transformedItem interface{}
			itemMap := item.(map[string]interface{})
			for k, v := range itemMap {
				vStr := strings.TrimSpace(v.(string))
				if k == "N" && vStr != "" {
					if n, err := strconv.Atoi(strings.TrimLeft(vStr, "0")); err == nil {
						transformedItem = n
					}
				} else if k == "BOOL" && strings.ToLower(vStr) == "f" {
					transformedItem = false
				}
			}
			if transformedItem != nil {
				items = append(items, transformedItem)
			}
		}
	}
	if len(items) > 0 {
		output[key] = items
	}
}

func transformMap(key string, value interface{}, output map[string]interface{}) {
	if rawMap, ok := value.(json.RawMessage); ok {
		subMap := make(map[string]json.RawMessage)
		if err := json.Unmarshal(rawMap, &subMap); err == nil {
			resultMap := make(map[string]interface{})
			for subKey, subValue := range subMap {
				transformJSON(map[string]json.RawMessage{subKey: subValue}, resultMap)
			}
			output[key] = resultMap
		}
	}
}
