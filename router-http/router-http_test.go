package router_http

import (
	"encoding/json"
	"github.com/eolinker/eosc"
	"github.com/eolinker/goku-eosc/router"
	"github.com/ghodss/yaml"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
)

type testRequestStruct struct {
	port     int
	testName string
	request  *http.Request
	want     string
}

var tests = []testRequestStruct{}

func insertTests() {

	request, _ := http.NewRequest("GET", "http://www.eolinker.com/ab?token=123", nil)
	request.Header.Set("app", "goku")
	request.Header.Set("user", "abc")
	request.Header.Set("version", "1.0")
	tests = append(tests, testRequestStruct{7777, "test1", request, "serviceA_rule1"})

	request, _ = http.NewRequest("GET", "http://www.apishop.net/ab?token=123", nil)
	request.Header.Set("app", "gokux")
	request.Header.Set("version", "1.0")
	tests = append(tests, testRequestStruct{7777, "test2", request, "serviceD_rule1"})

	request, _ = http.NewRequest("GET", "http://www.apibee.com/ab?token=123", nil)
	request.Header.Set("user", "abc")
	tests = append(tests, testRequestStruct{7777, "test3", request, "serviceA_rule3"})

	request, _ = http.NewRequest("GET", "http://www.apishop.net/abc?token=123", nil)
	request.Header.Set("user", "abc")
	tests = append(tests, testRequestStruct{7777, "test4", request, "serviceA_rule4"})

	request, _ = http.NewRequest("GET", "http://www.apishop.net/cxz?token=123", nil)
	request.Header.Set("user", "abc")
	request.Header.Set("token", "abc")
	tests = append(tests, testRequestStruct{7777, "test5", request, "serviceA_rule2"})

	request, _ = http.NewRequest("GET", "http://www.eolinker.com/abcd?token=123", nil)
	tests = append(tests, testRequestStruct{7777, "test6", request, "serviceB_rule1"})

	request, _ = http.NewRequest("GET", "http://www.eolinker.com/cxz?token=123", nil)
	tests = append(tests, testRequestStruct{7777, "test7", request, "serviceB_rule2"})

	request, _ = http.NewRequest("GET", "http://www.apibee.com/ab?token=123&token2=321", nil)
	request.Header.Set("user", "abc")
	request.Header.Set("token", "abc")
	tests = append(tests, testRequestStruct{7777, "test8", request, "serviceC_rule3"})

	request, _ = http.NewRequest("GET", "http://www.apibee.com/ab?token=123", nil)
	request.Header.Set("user", "abc")
	request.Header.Set("token", "abc")
	tests = append(tests, testRequestStruct{7777, "test9", request, "serviceC_rule2"})

	request, _ = http.NewRequest("GET", "http://www.adasdavera.com/abc?token=123", nil)
	request.Header.Set("user", "abc")
	request.Header.Set("token", "abc")
	tests = append(tests, testRequestStruct{80, "test10", request, "serviceE_rule4"})

}

func TestRM(t *testing.T) {
	employeeArr := loadYamlEmployee()
	if employeeArr == nil {
		log.Fatalln("空employee切片")
	}
	RM, err := newRouterHttpManager(employeeArr)
	if err != nil {
		log.Fatalln(err)
	}

	//测试路由树正确性
	//insertTests()
	//NRM := RM.(*routerManager)
	//for _, test := range tests {
	//	t.Run(test.testName, func(t *testing.T) {
	//		Tree := NRM.servers[test.port].(*routerTree)
	//		target, _ := Tree.tree.Match(test.request)
	//		if target != test.want {
	//			log.Println(test.testName, "   target == ", target)
	//			t.Fail()
	//		}
	//	})
	//}

	RM.StartAllServer()
	//测试删除实例
	//RM.Delete(80, "E")

	//测试shutdown后再启动
	//RM.ShutDownServer(80)
	//RM.StartServer(80)

	//测试新增实例
	//newEmploye := loadNewEmployee()[0]
	//RM.Set(80, newEmploye)

	select {}

}

func loadYamlEmployee() []eosc.IEmployee {
	data, err := ioutil.ReadFile("test_router_manager.yaml")
	if err != nil {
		return nil
	}

	convertData, err := yaml.YAMLToJSON(data)
	if err != nil {
		return nil
	}

	var cfg map[string][]router.Config
	err = json.Unmarshal(convertData, &cfg)
	if err != nil {
		return nil
	}

	employeeArr := make([]eosc.IEmployee, 0, len(cfg["router"]))

	for _, config := range cfg["router"] {
		cdata, _ := json.Marshal(config)
		employeeArr = append(employeeArr, router.NewEmployee(config.ID, config.Name, config.Driver, string(cdata)))
	}

	return employeeArr
}

func loadNewEmployee() []eosc.IEmployee {
	data, err := ioutil.ReadFile("test_router_manager_newEmployee.yaml")
	if err != nil {
		return nil
	}

	convertData, err := yaml.YAMLToJSON(data)
	if err != nil {
		return nil
	}

	var cfg map[string][]router.Config
	err = json.Unmarshal(convertData, &cfg)
	if err != nil {
		return nil
	}

	employeeArr := make([]eosc.IEmployee, 0, len(cfg["router"]))

	for _, config := range cfg["router"] {
		cdata, _ := json.Marshal(config)
		employeeArr = append(employeeArr, router.NewEmployee(config.ID, config.Name, config.Driver, string(cdata)))
	}

	return employeeArr
}
