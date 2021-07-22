package router_http

import (
	"flag"
	"log"
	"net/http"
	"testing"
)

func BenchmarkRouterMatch(b *testing.B) {
	flag.Parse()
	employeeArr := loadYamlEmployee()
	if employeeArr == nil {
		log.Fatalln("空employee切片")
	}
	RM, err := newRouterHttpManager(employeeArr)
	if err != nil {
		log.Fatalln(err)
	}

	RM.StartAllServer()

	insertBenchMarkTests()

	client := &http.Client{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		//for _, test := range benchMarkTests {
		client.Do(benchMarkTests[0].request)
		//}
	}

}

var benchMarkTests = []testRequestStruct{}

func insertBenchMarkTests() {

	request, _ := http.NewRequest("GET", "http://127.0.0.1:80/abc?token=123", nil)
	request.Header.Set("user", "abc")
	request.Header.Set("token", "abc")
	benchMarkTests = append(benchMarkTests, testRequestStruct{80, "test10", request, "serviceE_rule4"})

}
