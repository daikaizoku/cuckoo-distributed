package main

import (
	"log"
	"time"
)

const TASK_HIGH_COUNT = 100
const TASK_MEDIUM_COUNT = 50
const TASK_LOW_COUNT = 25

var node_status_count map[string]int = make(map[string]int)

func check_population(pending int) string {
	if pending >= 0 && pending < TASK_LOW_COUNT {
		return "low"
	} else if pending > TASK_LOW_COUNT && pending < TASK_MEDIUM_COUNT {
		return "medium"
	} else {
		return "high"
	}
}

func (a *App) population_monitoring() {
	for {
		nodes, err := getNodes(a.DB)
		if err != nil {
			log.Fatal(err.Error())
		}
		for _, node := range nodes {
			n_struct := new(CuckooStruct)
			JSONGet("http://"+node.Host+":8090/cuckoo/status", n_struct)
			node_status_count[n_struct.Hostname] = n_struct.Tasks.Pending
		}
		time.Sleep(time.Second * 30)
	}
}
