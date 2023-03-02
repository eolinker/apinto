package utils

import (
	"encoding/json"
	"testing"
)

// 官方格式json Schema
var test1 = `
{
    "properties":{
        "field0":{
            "example":"abcd",
            "type":"string"
        },
        "field1":{
            "example":123.12,
            "type":"number"
        },
        "field3":{
            "properties":{
                "field3_1":{
                    "type":"string"
                },
                "field3_2":{
                    "properties":{
                        "field3_2_1":{
                            "example":true,
                            "type":"boolean"
                        },
                        "field3_2_2":{
                            "items":{
                                "example":155.55,
                                "type":"integer"
                            },
                            "type":"array"
                        }
                    },
                    "type":"object"
                }
            },
            "type":"object"
        },
        "field2":{
            "items":{
                "type":"string"
            },
            "type":"array"
        }
    },
    "type":"object"
}`

/*
mockJs template

	var template = {
	  'name': '@cname', // 生成中文名字
	  'age|18-30': 20, // 生成18~30之间的随机整数
	  'gender|1': ['男', '女'], // 从数组中随机选取一个元素
	  'email': '@email' // 生成邮箱
	}

Mock.toJSONSchema(template)生成的json schema
*/
var mock1 = `{
    "template": {
        "name": "@cname",
        "age|18-30": 20,
        "gender|1": [
            "男",
            "女"
        ],
        "email": "@email"
    },
    "type": "object",
    "rule": {},
    "path": [
        "ROOT"
    ],
    "properties": [
        {
            "name": "name",
            "template": "@cname",
            "type": "string",
            "rule": {},
            "path": [
                "ROOT",
                "name"
            ]
        },
        {
            "name": "age",
            "template": 20,
            "type": "number",
            "rule": {
                "parameters": [
                    "age|18-30",
                    "age",
                    null,
                    "18-30",
                    null
                ],
                "range": [
                    "18-30",
                    "18",
                    "30"
                ],
                "min": 18,
                "max": 30,
                "count": 22
            },
            "path": [
                "ROOT",
                "age"
            ]
        },
        {
            "name": "gender",
            "template": [
                "男",
                "女"
            ],
            "type": "array",
            "rule": {
                "parameters": [
                    "gender|1",
                    "gender",
                    null,
                    "1",
                    null
                ],
                "range": [
                    "1",
                    "1",
                    null
                ],
                "min": 1,
                "count": 1
            },
            "path": [
                "ROOT",
                "gender"
            ],
            "items": [
                {
                    "name": 0,
                    "template": "男",
                    "type": "string",
                    "rule": {},
                    "path": [
                        "ROOT",
                        "gender",
                        0
                    ]
                },
                {
                    "name": 1,
                    "template": "女",
                    "type": "string",
                    "rule": {},
                    "path": [
                        "ROOT",
                        "gender",
                        1
                    ]
                }
            ]
        },
        {
            "name": "email",
            "template": "@email",
            "type": "string",
            "rule": {},
            "path": [
                "ROOT",
                "email"
            ]
        }
    ]
}`

/*
mockJs template

	var template = {
  'list|1-10': [{
        'id|+1': 1,
        'email': '@EMAIL'
    }]
}

Mock.toJSONSchema(template)生成的json schema
*/

var mock2 = `{
    "template": {
        "list|1-10": [
            {
                "id|+1": 7,
                "email": "@EMAIL"
            }
        ]
    },
    "type": "object",
    "rule": {},
    "path": [
        "ROOT"
    ],
    "properties": [
        {
            "name": "list",
            "template": [
                {
                    "id|+1": 7,
                    "email": "@EMAIL"
                }
            ],
            "type": "array",
            "rule": {
                "parameters": [
                    "list|1-10",
                    "list",
                    null,
                    "1-10",
                    null
                ],
                "range": [
                    "1-10",
                    "1",
                    "10"
                ],
                "min": 1,
                "max": 10,
                "count": 2
            },
            "path": [
                "ROOT",
                "list"
            ],
            "items": [
                {
                    "name": 0,
                    "template": {
                        "id|+1": 7,
                        "email": "@EMAIL"
                    },
                    "type": "object",
                    "rule": {},
                    "path": [
                        "ROOT",
                        "list",
                        0
                    ],
                    "properties": [
                        {
                            "name": "id",
                            "template": 7,
                            "type": "number",
                            "rule": {
                                "parameters": [
                                    "id|+1",
                                    "id",
                                    "1",
                                    null,
                                    null
                                ]
                            },
                            "path": [
                                "ROOT",
                                "list",
                                0,
                                "id"
                            ]
                        },
                        {
                            "name": "email",
                            "template": "@EMAIL",
                            "type": "string",
                            "rule": {},
                            "path": [
                                "ROOT",
                                "list",
                                0,
                                "email"
                            ]
                        }
                    ]
                }
            ]
        }
    ]
}`

/*
	var template = {
	  'list|1-10': {
	        'id|+1': 1,
	        'email': '@EMAIL'
	    }
	}
*/
var mock3 = `{
    "template": {
        "list|1-10": {
            "id|+1": 1,
            "email": "@EMAIL"
        }
    },
    "type": "object",
    "rule": {},
    "path": [
        "ROOT"
    ],
    "properties": [
        {
            "name": "list",
            "template": {
                "id|+1": 1,
                "email": "@EMAIL"
            },
            "type": "object",
            "rule": {
                "parameters": [
                    "list|1-10",
                    "list",
                    null,
                    "1-10",
                    null
                ],
                "range": [
                    "1-10",
                    "1",
                    "10"
                ],
                "min": 1,
                "max": 10,
                "count": 4
            },
            "path": [
                "ROOT",
                "list"
            ],
            "properties": [
                {
                    "name": "id",
                    "template": 1,
                    "type": "number",
                    "rule": {
                        "parameters": [
                            "id|+1",
                            "id",
                            "1",
                            null,
                            null
                        ]
                    },
                    "path": [
                        "ROOT",
                        "list",
                        "id"
                    ]
                },
                {
                    "name": "email",
                    "template": "@EMAIL",
                    "type": "string",
                    "rule": {},
                    "path": [
                        "ROOT",
                        "list",
                        "email"
                    ]
                }
            ]
        }
    ]
}`

/*
	var template = {
	    'key|1-10': '★'
	}
*/
var mock4 = `{
    "template": {
        "key|1-10": "★"
    },
    "type": "object",
    "rule": {},
    "path": [
        "ROOT"
    ],
    "properties": [
        {
            "name": "key",
            "template": "★",
            "type": "string",
            "rule": {
                "parameters": [
                    "key|1-10",
                    "key",
                    null,
                    "1-10",
                    null
                ],
                "range": [
                    "1-10",
                    "1",
                    "10"
                ],
                "min": 1,
                "max": 10,
                "count": 2
            },
            "path": [
                "ROOT",
                "key"
            ]
        }
    ]
}`

/*
	var template = {
    'title': 'Syntax Demo',

    'string1|1-10': '★',
    'string2|3': 'value',

    'number1|+1': 100,
    'number2|1-100': 100,
    'number3|1-100.1-10': 1,
    'number4|123.1-10': 1,
    'number5|123.3': 1,
    'number6|123.10': 1.123,

    'boolean1|1': true,
    'boolean2|1-2': true,

    'object1|2-4': {
        '110000': '北京市',
        '120000': '天津市',
        '130000': '河北省',
        '140000': '山西省'
    },
    'object2|2': {
        '310000': '上海市',
        '320000': '江苏省',
        '330000': '浙江省',
        '340000': '安徽省'
    },
    'object3|2': {
        '310000': '@name',
        '320000': '@ip',
        '330000': '@email'
    },

    'array1|1': ['AMD', 'CMD', 'KMD', 'UMD'],
    'array2|1-10': [{
    	'id':10,
      'ip':'@ip'
    }],
    'array3|3': ['Mock.js'],
    'array4|3-5': [10,20,30,40]
}
*/

var mock5 = `{
    "template": {
        "title": "Syntax Demo",
        "string1|1-10": "★",
        "string2|3": "value",
        "number1|+1": 101,
        "number2|1-100": 100,
        "number3|1-100.1-10": 1,
        "number4|123.1-10": 1,
        "number5|123.3": 1,
        "number6|123.10": 1.123,
        "boolean1|1": true,
        "boolean2|1-2": true,
        "object1|2-4": {
            "110000": "北京市",
            "120000": "天津市",
            "130000": "河北省",
            "140000": "山西省"
        },
        "object2|2": {
            "310000": "上海市",
            "320000": "江苏省",
            "330000": "浙江省",
            "340000": "安徽省"
        },
        "object3|2": {
            "310000": "@name",
            "320000": "@ip",
            "330000": "@email"
        },
        "array1|1": [
            "AMD",
            "CMD",
            "KMD",
            "UMD"
        ],
        "array2|1-10": [
            {
                "id": 10,
                "ip": "@ip"
            }
        ],
        "array3|3": [
            "Mock.js"
        ],
        "array4|3-5": [
            10,
            20,
            30,
            40
        ]
    },
    "type": "object",
    "rule": {},
    "path": [
        "ROOT"
    ],
    "properties": [
        {
            "name": "title",
            "template": "Syntax Demo",
            "type": "string",
            "rule": {},
            "path": [
                "ROOT",
                "title"
            ]
        },
        {
            "name": "string1",
            "template": "★",
            "type": "string",
            "rule": {
                "parameters": [
                    "string1|1-10",
                    "string1",
                    null,
                    "1-10",
                    null
                ],
                "range": [
                    "1-10",
                    "1",
                    "10"
                ],
                "min": 1,
                "max": 10,
                "count": 5
            },
            "path": [
                "ROOT",
                "string1"
            ]
        },
        {
            "name": "string2",
            "template": "value",
            "type": "string",
            "rule": {
                "parameters": [
                    "string2|3",
                    "string2",
                    null,
                    "3",
                    null
                ],
                "range": [
                    "3",
                    "3",
                    null
                ],
                "min": 3,
                "count": 3
            },
            "path": [
                "ROOT",
                "string2"
            ]
        },
        {
            "name": "number1",
            "template": 101,
            "type": "number",
            "rule": {
                "parameters": [
                    "number1|+1",
                    "number1",
                    "1",
                    null,
                    null
                ]
            },
            "path": [
                "ROOT",
                "number1"
            ]
        },
        {
            "name": "number2",
            "template": 100,
            "type": "number",
            "rule": {
                "parameters": [
                    "number2|1-100",
                    "number2",
                    null,
                    "1-100",
                    null
                ],
                "range": [
                    "1-100",
                    "1",
                    "100"
                ],
                "min": 1,
                "max": 100,
                "count": 61
            },
            "path": [
                "ROOT",
                "number2"
            ]
        },
        {
            "name": "number3",
            "template": 1,
            "type": "number",
            "rule": {
                "parameters": [
                    "number3|1-100.1-10",
                    "number3",
                    null,
                    "1-100",
                    "1-10"
                ],
                "range": [
                    "1-100",
                    "1",
                    "100"
                ],
                "min": 1,
                "max": 100,
                "count": 70,
                "decimal": [
                    "1-10",
                    "1",
                    "10"
                ],
                "dmin": 1,
                "dmax": 10,
                "dcount": 6
            },
            "path": [
                "ROOT",
                "number3"
            ]
        },
        {
            "name": "number4",
            "template": 1,
            "type": "number",
            "rule": {
                "parameters": [
                    "number4|123.1-10",
                    "number4",
                    null,
                    "123",
                    "1-10"
                ],
                "range": [
                    "123",
                    "123",
                    null
                ],
                "min": 123,
                "count": 123,
                "decimal": [
                    "1-10",
                    "1",
                    "10"
                ],
                "dmin": 1,
                "dmax": 10,
                "dcount": 7
            },
            "path": [
                "ROOT",
                "number4"
            ]
        },
        {
            "name": "number5",
            "template": 1,
            "type": "number",
            "rule": {
                "parameters": [
                    "number5|123.3",
                    "number5",
                    null,
                    "123",
                    "3"
                ],
                "range": [
                    "123",
                    "123",
                    null
                ],
                "min": 123,
                "count": 123,
                "decimal": [
                    "3",
                    "3",
                    null
                ],
                "dmin": 3,
                "dmax": null,
                "dcount": 3
            },
            "path": [
                "ROOT",
                "number5"
            ]
        },
        {
            "name": "number6",
            "template": 1.123,
            "type": "number",
            "rule": {
                "parameters": [
                    "number6|123.10",
                    "number6",
                    null,
                    "123",
                    "10"
                ],
                "range": [
                    "123",
                    "123",
                    null
                ],
                "min": 123,
                "count": 123,
                "decimal": [
                    "10",
                    "10",
                    null
                ],
                "dmin": 10,
                "dmax": null,
                "dcount": 10
            },
            "path": [
                "ROOT",
                "number6"
            ]
        },
        {
            "name": "boolean1",
            "template": true,
            "type": "boolean",
            "rule": {
                "parameters": [
                    "boolean1|1",
                    "boolean1",
                    null,
                    "1",
                    null
                ],
                "range": [
                    "1",
                    "1",
                    null
                ],
                "min": 1,
                "count": 1
            },
            "path": [
                "ROOT",
                "boolean1"
            ]
        },
        {
            "name": "boolean2",
            "template": true,
            "type": "boolean",
            "rule": {
                "parameters": [
                    "boolean2|1-2",
                    "boolean2",
                    null,
                    "1-2",
                    null
                ],
                "range": [
                    "1-2",
                    "1",
                    "2"
                ],
                "min": 1,
                "max": 2,
                "count": 2
            },
            "path": [
                "ROOT",
                "boolean2"
            ]
        },
        {
            "name": "object1",
            "template": {
                "110000": "北京市",
                "120000": "天津市",
                "130000": "河北省",
                "140000": "山西省"
            },
            "type": "object",
            "rule": {
                "parameters": [
                    "object1|2-4",
                    "object1",
                    null,
                    "2-4",
                    null
                ],
                "range": [
                    "2-4",
                    "2",
                    "4"
                ],
                "min": 2,
                "max": 4,
                "count": 4
            },
            "path": [
                "ROOT",
                "object1"
            ],
            "properties": [
                {
                    "name": "110000",
                    "template": "北京市",
                    "type": "string",
                    "rule": {},
                    "path": [
                        "ROOT",
                        "object1",
                        "110000"
                    ]
                },
                {
                    "name": "120000",
                    "template": "天津市",
                    "type": "string",
                    "rule": {},
                    "path": [
                        "ROOT",
                        "object1",
                        "120000"
                    ]
                },
                {
                    "name": "130000",
                    "template": "河北省",
                    "type": "string",
                    "rule": {},
                    "path": [
                        "ROOT",
                        "object1",
                        "130000"
                    ]
                },
                {
                    "name": "140000",
                    "template": "山西省",
                    "type": "string",
                    "rule": {},
                    "path": [
                        "ROOT",
                        "object1",
                        "140000"
                    ]
                }
            ]
        },
        {
            "name": "object2",
            "template": {
                "310000": "上海市",
                "320000": "江苏省",
                "330000": "浙江省",
                "340000": "安徽省"
            },
            "type": "object",
            "rule": {
                "parameters": [
                    "object2|2",
                    "object2",
                    null,
                    "2",
                    null
                ],
                "range": [
                    "2",
                    "2",
                    null
                ],
                "min": 2,
                "count": 2
            },
            "path": [
                "ROOT",
                "object2"
            ],
            "properties": [
                {
                    "name": "310000",
                    "template": "上海市",
                    "type": "string",
                    "rule": {},
                    "path": [
                        "ROOT",
                        "object2",
                        "310000"
                    ]
                },
                {
                    "name": "320000",
                    "template": "江苏省",
                    "type": "string",
                    "rule": {},
                    "path": [
                        "ROOT",
                        "object2",
                        "320000"
                    ]
                },
                {
                    "name": "330000",
                    "template": "浙江省",
                    "type": "string",
                    "rule": {},
                    "path": [
                        "ROOT",
                        "object2",
                        "330000"
                    ]
                },
                {
                    "name": "340000",
                    "template": "安徽省",
                    "type": "string",
                    "rule": {},
                    "path": [
                        "ROOT",
                        "object2",
                        "340000"
                    ]
                }
            ]
        },
        {
            "name": "object3",
            "template": {
                "310000": "@name",
                "320000": "@ip",
                "330000": "@email"
            },
            "type": "object",
            "rule": {
                "parameters": [
                    "object3|2",
                    "object3",
                    null,
                    "2",
                    null
                ],
                "range": [
                    "2",
                    "2",
                    null
                ],
                "min": 2,
                "count": 2
            },
            "path": [
                "ROOT",
                "object3"
            ],
            "properties": [
                {
                    "name": "310000",
                    "template": "@name",
                    "type": "string",
                    "rule": {},
                    "path": [
                        "ROOT",
                        "object3",
                        "310000"
                    ]
                },
                {
                    "name": "320000",
                    "template": "@ip",
                    "type": "string",
                    "rule": {},
                    "path": [
                        "ROOT",
                        "object3",
                        "320000"
                    ]
                },
                {
                    "name": "330000",
                    "template": "@email",
                    "type": "string",
                    "rule": {},
                    "path": [
                        "ROOT",
                        "object3",
                        "330000"
                    ]
                }
            ]
        },
        {
            "name": "array1",
            "template": [
                "AMD",
                "CMD",
                "KMD",
                "UMD"
            ],
            "type": "array",
            "rule": {
                "parameters": [
                    "array1|1",
                    "array1",
                    null,
                    "1",
                    null
                ],
                "range": [
                    "1",
                    "1",
                    null
                ],
                "min": 1,
                "count": 1
            },
            "path": [
                "ROOT",
                "array1"
            ],
            "items": [
                {
                    "name": 0,
                    "template": "AMD",
                    "type": "string",
                    "rule": {},
                    "path": [
                        "ROOT",
                        "array1",
                        0
                    ]
                },
                {
                    "name": 1,
                    "template": "CMD",
                    "type": "string",
                    "rule": {},
                    "path": [
                        "ROOT",
                        "array1",
                        1
                    ]
                },
                {
                    "name": 2,
                    "template": "KMD",
                    "type": "string",
                    "rule": {},
                    "path": [
                        "ROOT",
                        "array1",
                        2
                    ]
                },
                {
                    "name": 3,
                    "template": "UMD",
                    "type": "string",
                    "rule": {},
                    "path": [
                        "ROOT",
                        "array1",
                        3
                    ]
                }
            ]
        },
        {
            "name": "array2",
            "template": [
                {
                    "id": 10,
                    "ip": "@ip"
                }
            ],
            "type": "array",
            "rule": {
                "parameters": [
                    "array2|1-10",
                    "array2",
                    null,
                    "1-10",
                    null
                ],
                "range": [
                    "1-10",
                    "1",
                    "10"
                ],
                "min": 1,
                "max": 10,
                "count": 1
            },
            "path": [
                "ROOT",
                "array2"
            ],
            "items": [
                {
                    "name": 0,
                    "template": {
                        "id": 10,
                        "ip": "@ip"
                    },
                    "type": "object",
                    "rule": {},
                    "path": [
                        "ROOT",
                        "array2",
                        0
                    ],
                    "properties": [
                        {
                            "name": "id",
                            "template": 10,
                            "type": "number",
                            "rule": {},
                            "path": [
                                "ROOT",
                                "array2",
                                0,
                                "id"
                            ]
                        },
                        {
                            "name": "ip",
                            "template": "@ip",
                            "type": "string",
                            "rule": {},
                            "path": [
                                "ROOT",
                                "array2",
                                0,
                                "ip"
                            ]
                        }
                    ]
                }
            ]
        },
        {
            "name": "array3",
            "template": [
                "Mock.js"
            ],
            "type": "array",
            "rule": {
                "parameters": [
                    "array3|3",
                    "array3",
                    null,
                    "3",
                    null
                ],
                "range": [
                    "3",
                    "3",
                    null
                ],
                "min": 3,
                "count": 3
            },
            "path": [
                "ROOT",
                "array3"
            ],
            "items": [
                {
                    "name": 0,
                    "template": "Mock.js",
                    "type": "string",
                    "rule": {},
                    "path": [
                        "ROOT",
                        "array3",
                        0
                    ]
                }
            ]
        },
        {
            "name": "array4",
            "template": [
                10,
                20,
                30,
                40
            ],
            "type": "array",
            "rule": {
                "parameters": [
                    "array4|3-5",
                    "array4",
                    null,
                    "3-5",
                    null
                ],
                "range": [
                    "3-5",
                    "3",
                    "5"
                ],
                "min": 3,
                "max": 5,
                "count": 5
            },
            "path": [
                "ROOT",
                "array4"
            ],
            "items": [
                {
                    "name": 0,
                    "template": 10,
                    "type": "number",
                    "rule": {},
                    "path": [
                        "ROOT",
                        "array4",
                        0
                    ]
                },
                {
                    "name": 1,
                    "template": 20,
                    "type": "number",
                    "rule": {},
                    "path": [
                        "ROOT",
                        "array4",
                        1
                    ]
                },
                {
                    "name": 2,
                    "template": 30,
                    "type": "number",
                    "rule": {},
                    "path": [
                        "ROOT",
                        "array4",
                        2
                    ]
                },
                {
                    "name": 3,
                    "template": 40,
                    "type": "number",
                    "rule": {},
                    "path": [
                        "ROOT",
                        "array4",
                        3
                    ]
                }
            ]
        }
    ]
}`

func TestJsonSchemaMockJsUnmarshal(t *testing.T) {
	type args struct {
		valueMap interface{}
	}

	valueMap := make(map[string]interface{})
	json.Unmarshal([]byte(mock5), &valueMap)
	tests := []struct {
		name string
		args args
		want interface{}
	}{
		{
			name: "",
			args: args{
				valueMap: valueMap,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := JsonSchemaMockJsUnmarshal(tt.args.valueMap)
			bytes, _ := json.Marshal(got)
			t.Logf(string(bytes))
		})
	}
}
