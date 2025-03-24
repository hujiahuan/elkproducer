package elkproducer

import (
	"fmt"
	"testing"
	"time"
)

type TestStruct struct {
	Aaa  string
	Bbb  int
	Date time.Time
}

func TestAddLog(t *testing.T) {
	elkConf := Config{
		Addresses: []string{
			"http://150.138.84.21:9200",
		},
		Username: "elastic",
		Password: "sursen@admin",
	}
	configelk := ESConfig{
		ESConf: elkConf,
		Index:  "test2",
		//IndexType: "log",
		//Url:       "http://101.237.34.55:6010",
		DebugMode: true,
		//From:      0,
		//Size:      10000,
	}
	es, _ := NewClient(configelk)
	type LogData struct {
		E      string   `json:"E"`
		B      string   `json:"B"`
		D      []string `json:"D"`
		Folder string   `json:"folder"`
		Pass   string   `json:"pass"`
		A      int      `json:"A"`
		C      string   `json:"C"`
	}
	//t1 := LogData{
	//	A: 1,
	//	B: "b",
	//	C: "c",
	//	D: []string{"a1", "b2", "c3", "d4"},
	//	E: "e",
	//}
	//t2 := LogData{
	//	A: 2,
	//	B: "b",
	//	C: "c",
	//	D: []string{"a1", "b2", "c3", "d4"},
	//	E: "e",
	//}
	//t3 := LogData{
	//	A: 3,
	//	B: "b",
	//	C: "c",
	//	D: []string{"a1", "b2", "c3", "d4"},
	//	E: "e",
	//}
	//for i := 0; i < 100000; i++ {
	//	es.AddLog(t1)
	//	es.AddLog(t2)
	//	es.AddLog(t3)
	//}
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"wildcard": map[string]interface{}{
				"book.keyword": "*中和之美——普遍艺术和谐观与特定艺术风*",
			},
		},
	}
	fmt.Println("llllll", es.GetData(query)["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"])
	//fmt.Println("llllll", es.GetData(query)["hits"].(map[string]interface{})["hits"].([]interface{})[0].(map[string]interface{})["_source"])
	fmt.Println("llllll", es.GetData(query)["hits"].(map[string]interface{})["hits"])
}
