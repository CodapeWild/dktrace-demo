/*
 *   Copyright (c) 2023 CodapeWild
 *   All rights reserved.

 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at

 *   http://www.apache.org/licenses/LICENSE-2.0

 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package main

import (
	"context"
	"log"
	"os"

	"github.com/CodapeWild/dktrace-data-benchmark/agent"
)

func main() {
	Execute()

	log.SetFlags(log.Lshortfile | log.LstdFlags)
	log.SetOutput(os.Stdout)
	if gBenchConf.DisableLog {
		log.Println("log disabled")
		log.SetOutput(nil)
	}

	if gBenchConf == nil || len(gBenchConf.Tasks) == 0 {
		log.Println("dktrace-data-benchmark not configurated properly")

		return
	}

	runTaskThread()
}

var (
	gTaskChan = make(chan *taskConfig, 20)
	gCloser   = make(chan struct{})
)

func runTaskThread() {
	for {
		select {
		case <-gCloser:
			return
		case task := <-gTaskChan:
			var (
				canceler context.CancelFunc
				finish   chan struct{}
				err      error
			)
			switch task.Tracer {
			case dd:
				canceler, finish, err = benchDDTraceCollector(task)
			case jg:
			case otel:
			case pp:
			case sky:
			case zpk:
			default:
				log.Printf("unrecognized task, Name: %s Tracer %s\n", task.Name, task.Tracer)
			}
			if err != nil {
				canceler()
				log.Println(err.Error())
				continue
			}
			// waiting for the current task to complete and then start the next one multiple
			// threads benchmark task will seriously affect local host performance
			if finish != nil {
				<-finish
			}
		}
	}
}

func benchDDTraceCollector(taskConf *taskConfig) (canceler context.CancelFunc, finish chan struct{}, err error) {
	var r route
	if r, err = newRouteFromJSONFile(taskConf.RouteConfig); err != nil {
		return
	}

	tr := r.createTree(&ddtracerwrapper{})
	agentAddress := agent.NewRandomPortWithLocalHost()
	canceler, finish, err = agent.BuildDDAgentForWork(agentAddress, taskConf.CollectorIP, taskConf.CollectorPort, taskConf.CollectorPath, tr.count(), taskConf.SendThreads, taskConf.SendTimesPerThread)
	tr.spawn(context.TODO(), agentAddress)

	return
}
