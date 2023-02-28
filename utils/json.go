package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"math/rand"
	"strconv"
	"time"

	"github.com/robertkrimen/otto"
)

// JSObjectToJSON 将js对象转为json
func JSObjectToJSON(s string) ([]byte, error) {
	vm := otto.New()
	v, err := vm.Run(fmt.Sprintf(`
		cs = %s
		JSON.stringify(cs)
`, s))
	if err != nil {
		return nil, err
	}
	return []byte(v.String()), nil
}

// JSONUnmarshal  将json格式的s解码成v所需的json格式
func JSONUnmarshal(s, v interface{}) error {
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// JsonSchemaMockJsUnmarshal
// 解析mockJs生成的json schema 并根据规则生成随机值
// 不是mockJs生成的json 走默认规则解析
func JsonSchemaMockJsUnmarshal(valueMap interface{}) interface{} {
	rand.Seed(time.Now().UnixMicro())
	gofakeit.Seed(time.Now().UnixMicro())

	value, vOk := valueMap.(map[string]interface{})
	if vOk {
		switch valType := value["properties"].(type) {
		case []interface{}:
			resultMap := make(map[string]interface{})
			for _, v := range valType {
				if m, ok := v.(map[string]interface{}); ok {
					name, nameOk := m["name"].(string)
					if !nameOk {
						return jsonSchemaFormat
					}
					template := m["template"]
					if t, tOk := m["type"].(string); tOk {
						rule, ruleOk := m["rule"].(map[string]interface{})
						if !ruleOk || len(rule) == 0 {
							if template != nil {
								if templateStr, ok := template.(string); ok {
									switch templateStr {
									case "@cname":
										resultMap[name] = gofakeit.Username()
									case "@cfirst":
										resultMap[name] = gofakeit.FirstName()
									case "@clast":
										resultMap[name] = gofakeit.LastName()
									case "@name", "@name(true)":
										resultMap[name] = gofakeit.Name()
									case "@first":
										resultMap[name] = gofakeit.FirstName()
									case "@last":
										resultMap[name] = gofakeit.LastName()
									case "@email":
										resultMap[name] = gofakeit.Email()
									case "@ip":
										resultMap[name] = gofakeit.IPv4Address()
									case "@zip":
										resultMap[name] = gofakeit.Address().Zip
									case "@city", "@city(true)":
										resultMap[name] = gofakeit.Address().Address
									case "@url":
										resultMap[name] = gofakeit.URL()
									default:
										resultMap[name] = template
									}
									continue
								}
								resultMap[name] = template
								continue
							}
						}

						switch t {
						case "string":

							minVal, maxVal, _, _, err := getMinMaxDminDmax(rule)
							if err != nil {
								return err
							}

							randomNum := 0
							if minVal > 0 && maxVal == 0 {
								randomNum = int(minVal)
							}

							if minVal > 0 && maxVal > 0 {
								randomNum = int(RandInt64(int64(minVal), int64(maxVal)))
							}

							if template != nil {
								templateStr, sOk := template.(string)
								if !sOk {
									return jsonSchemaFormat
								}
								temp := ""
								for i := 0; i < randomNum; i++ {
									temp = temp + templateStr
								}
								resultMap[name] = temp
								continue
							}

						case "number":
							minVal, maxVal, dminVal, dmaxVal, err := getMinMaxDminDmax(rule)
							if err != nil {
								return err
							}

							randomValue := 0.0
							if minVal > 0.0 && maxVal == 0.0 {
								randomValue = minVal
								randomValue += RandFloats(0, 1)
							} else if minVal > 0.0 && maxVal > 0.0 {
								randomValue = RandFloats(minVal, maxVal)
							}

							if randomValue == 0.0 {
								resultMap[name] = template
								continue
							}

							if dminVal > 0.0 && dmaxVal == 0.0 {
								randomValue, _ = strconv.ParseFloat(strconv.FormatFloat(randomValue, 'f', int(dminVal), 64), 64)
							} else if dminVal > 0.0 && dmaxVal > 0.0 {
								floats := RandFloats(dminVal, dmaxVal)
								randomValue, _ = strconv.ParseFloat(strconv.FormatFloat(randomValue, 'f', int(floats), 64), 64)
							} else {

							}

							resultMap[name] = randomValue

						case "boolean":
							resultMap[name] = gofakeit.Bool()
						case "object":
							templateMap, templateOk := template.(map[string]interface{})
							if templateOk {
								minVal, maxVal, _, _, err := getMinMaxDminDmax(rule)
								if err != nil {
									return err
								}
								randomNum := 0
								if minVal > 0 && maxVal == 0 {
									randomNum = int(minVal)
								}

								if minVal > 0 && maxVal > 0 {
									randomNum = int(RandInt64(int64(minVal), int64(maxVal)))
								}
								tempMap := make(map[string]interface{})
								i := 1
								for k, v := range templateMap {
									tempMap[k] = v
									if i == randomNum {
										break
									}
									i++
								}

								resultMap[name] = tempMap

							}
						case "array":
							templateList, templateOk := template.([]interface{})
							if templateOk {
								minVal, maxVal, _, _, err := getMinMaxDminDmax(rule)
								if err != nil {
									return err
								}

								randomNum := 0
								if minVal > 0 && maxVal == 0 {
									randomNum = int(minVal)
								}

								if minVal > 0 && maxVal > 0 {
									randomNum = int(RandInt64(int64(minVal), int64(maxVal)))
								}
								tempList := make([]interface{}, 0)

								for i := 0; i < randomNum; i++ {
									tempList = append(tempList, templateList[rand.Intn(len(templateList))])
								}

								resultMap[name] = tempList

							}
						}
					}
				}
			}
			return resultMap
		}
	}
	return jsonSchemaUnmarshal(value)
}

var jsonSchemaFormat = errors.New("json schema format err")

func getMinMaxDminDmax(rule map[string]interface{}) (float64, float64, float64, float64, error) {
	minVal := 0.0
	min, minOk := rule["min"]
	if minOk && min != nil {
		mOk := false
		minVal, mOk = min.(float64)
		if !mOk {
			return 0, 0, 0, 0, jsonSchemaFormat
		}

	}

	maxVal := 0.0
	max, maxOk := rule["max"]
	if maxOk && max != nil {
		mOk := false
		maxVal, mOk = max.(float64)
		if !mOk {
			return 0, 0, 0, 0, jsonSchemaFormat
		}
	}
	dminVal := 0.0
	dmin, dminOk := rule["dmin"]
	if dminOk && dmin != nil {
		mOk := false
		dminVal, mOk = dmin.(float64)
		if !mOk {
			return 0, 0, 0, 0, jsonSchemaFormat
		}
	}

	dmaxVal := 0.0
	dmax, dmaxOk := rule["dmax"]
	if dmaxOk && dmax != nil {
		mOk := false
		dmaxVal, mOk = dmax.(float64)
		if !mOk {
			return 0, 0, 0, 0, jsonSchemaFormat
		}
	}
	return minVal, maxVal, dminVal, dmaxVal, nil
}

func jsonSchemaUnmarshal(properties interface{}) interface{} {
	propertiesMap, ok := properties.(map[string]interface{})
	if !ok {
		return jsonSchemaFormat
	}
	if val, ok := propertiesMap["example"]; ok {
		return val
	} else {
		if t, tOk := propertiesMap["type"].(string); tOk {
			switch t {
			case "string":
				return gofakeit.LetterN(10)
			case "number":
				return gofakeit.Float64()
			case "integer":
				return gofakeit.Int64()
			case "boolean":
				return gofakeit.Bool()
			case "object":
				propertiesMaps, pOk := propertiesMap["properties"].(map[string]interface{})
				if !pOk {
					return jsonSchemaFormat
				}
				resultMap := make(map[string]interface{})
				for key, vProperties := range propertiesMaps {
					resultMap[key] = jsonSchemaUnmarshal(vProperties)
				}
				return resultMap
			case "array":
				items, iOk := propertiesMap["items"].(map[string]interface{})
				if !iOk {
					return jsonSchemaFormat
				}
				resultList := make([]interface{}, 0)
				resultList = append(resultList, jsonSchemaUnmarshal(items))
				return resultList
			}
		}
		return jsonSchemaFormat
	}
}
