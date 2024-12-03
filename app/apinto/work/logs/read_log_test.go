package logs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

type LokiRequest struct {
	Streams []*Stream `json:"streams"`
}

type Stream struct {
	Stream map[string]string `json:"stream"`
	Values [][]interface{}   `json:"values"`
}

func TestSendLogToLoki(t *testing.T) {
	// 1. Create a new log
	// 2. Send the log to Loki
	// 3. Check if the log is in Loki

	items, err := parseLog()
	if err != nil {
		t.Fatal(err)
	}
	client := http.Client{}
	for _, item := range items {
		body, err := json.Marshal(item)
		if err != nil {
			t.Fatal(err)
		}
		req, err := http.NewRequest(http.MethodPost, "http://localhost:3100/loki/api/v1/push", bytes.NewBuffer(body))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Scope-OrgID", "tenant1")
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != 204 {
			t.Fatal(resp.StatusCode, string(respBody))
		}
		if err := resp.Body.Close(); err != nil {
			t.Fatal(err)
		}
	}
	t.Log("Send log to Loki success")

}

func parseLog() ([]*LokiRequest, error) {
	data, err := os.ReadFile("access.log")
	if err != nil {
		return nil, err
	}
	// 换行分割
	lines := strings.Split(string(data), "\n")
	reqMap := map[string]*Stream{}
	// 解析日志
	for _, l := range lines {
		if l == "" {
			continue
		}
		tmp := make(map[string]interface{})
		err = json.Unmarshal([]byte(l), &tmp)
		if err != nil {
			return nil, err
		}
		org := map[string]string{
			"cluster":    tmp["cluster"].(string),
			"node":       tmp["node"].(string),
			"service":    tmp["service"].(string),
			"api":        tmp["api"].(string),
			"src_ip":     tmp["src_ip"].(string),
			"block_name": tmp["block_name"].(string),
		}
		key := genKey(org)
		if _, ok := reqMap[key]; !ok {
			reqMap[key] = &Stream{
				Stream: org,
				Values: make([][]interface{}, 0),
			}
		}

		reqMap[key].Values = append(reqMap[key].Values, []interface{}{strconv.FormatInt(time.UnixMilli(int64(tmp["msec"].(float64))).UnixNano(), 10), l})
	}
	reqs := make([]*LokiRequest, len(reqMap)/10+1)
	num := 0
	for _, v := range reqMap {
		index := num / 10
		if reqs[index] == nil {
			reqs[index] = &LokiRequest{
				Streams: make([]*Stream, 0, 10),
			}
		}
		reqs[index].Streams = append(reqs[index].Streams, v)
		num++
	}
	return reqs, nil
}

func genKey(org map[string]string) string {
	return fmt.Sprintf("cluster_%s-node_%s-service_%s-api_%s-src_ip_%s-block_name_%s", org["cluster"], org["node"], org["service"], org["api"], org["src_ip"], org["block_name"])
}
