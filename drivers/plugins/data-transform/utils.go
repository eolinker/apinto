package data_transform

import (
	"encoding/json"
	"fmt"
	"strings"

	. "github.com/clbanning/mxj"
)

var declarations = []string{"version", "encoding", "standalone"}

func xml2json(xmlVal []byte, xmlDeclaration map[string]string) ([]byte, error) {
	m, err := NewMapXml(xmlVal)
	if err != nil {
		return nil, err
	}
	for _, key := range declarations {
		if _, ok := m[key]; !ok {
			if xmlDeclaration != nil {
				if v, has := xmlDeclaration[key]; has {
					m[key] = v
				}
			}
		}
		if v, ok := m[key]; ok {
			vv, success := v.(string)
			if success {
				m[key] = strings.Replace(vv, `"`, "", -1)
			}
		}
	}

	jsonObject, err := m.Json()
	if err != nil {
		return nil, err
	}
	return jsonObject, nil
}

func json2xml(jsonVal []byte, rootTag string, xd map[string]string) ([]byte, error) {
	var m, n map[string]interface{}
	if err := json.Unmarshal(jsonVal, &m); err != nil {
		return nil, err
	}
	n = make(map[string]interface{})
	for _, key := range declarations {
		if v, ok := m[key]; ok {
			n[key] = v
			delete(m, key)
		} else {
			if xd != nil {
				if v, ok := xd[key]; ok {
					n[key] = v
				}
			}

		}
	}
	var x []byte
	var err error
	x, err = Map(m).XmlIndent("", "  ", rootTag)
	if err != nil {
		return x, err
	}

	declaration := xmlDeclaration(n)
	retBody := []byte(fmt.Sprintf("%s\r\n%s", declaration, string(x)))
	return retBody, nil
}

func xmlDeclaration(m map[string]interface{}) string {
	declaration := ""

	for _, a := range declarations {
		if declaration == "" {
			declaration = "<?xml"
		}
		if v, ok := m[a]; ok {
			delete(m, a)
			declaration += fmt.Sprintf(` %s="%s"`, a, v)
		}
	}
	if declaration != "" {
		declaration += "?>"
	}
	return declaration
}

func encode(ent string, origin string, statusCode int) string {
	if ent == "json" {
		tmp := map[string]interface{}{
			"message":     origin,
			"status_code": statusCode,
		}
		body, _ := json.Marshal(tmp)
		return string(body)
	}
	return origin
}
