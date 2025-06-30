package util

import (
	"fmt"
	"testing"
)

func TestGetJSONKeys(t *testing.T) {
	var jsonStr = `
	{
		"Name": "test",
		"TableName": "test",
		"TemplateID": "test",
		"TemplateInfo": "test",
		"Limit": 0
}`
	keys, err := GetJSONKeys(jsonStr)
	if err != nil {
		t.Errorf("GetJSONKeys failed: %s\n", err.Error())
		return
	}
	if len(keys) != 5 {
		t.Errorf("GetJSONKeys failed: %s\n", "keys length is not 5")
		return
	}
	if keys[0] != "Name" {
		t.Errorf("GetJSONKeys failed: %s\n", "keys[0] is not Name")

		return
	}
	if keys[1] != "TableName" {
		t.Errorf("GetJSONKeys failed: %s\n", "keys[1] is not TableName")

		return
	}
	if keys[2] != "TemplateID" {
		t.Errorf("GetJSONKeys failed: %s\n", "keys[2] is not TemplateID")

		return
	}
	if keys[3] != "TemplateInfo" {
		t.Errorf("GetJSONKeys failed: %s\n", "keys[3] is not TemplateInfo")

		return
	}
	if keys[4] != "Limit" {
		t.Errorf("GetJSONKeys failed: %s\n", "keys[4] is not Limit")

		return
	}

	fmt.Println(keys)
}
