// Copyright 2018 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package e2e

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestIssue90(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	webAddr, graphiteAddr := fmt.Sprintf("127.0.0.1:%d", 9118), fmt.Sprintf("127.0.0.1:%d", 9119)
	exporter := exec.Command(
		filepath.Join(cwd, "..", "graphite_exporter"),
		"--graphite.mapping-strict-match",
		"--web.listen-address", webAddr,
		"--graphite.listen-address", graphiteAddr,
		"--graphite.mapping-config", filepath.Join(cwd, "fixtures", "issue90.yml"),
	)
	err = exporter.Start()
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	defer exporter.Process.Kill()

	for i := 0; i < 10; i++ {
		if i > 0 {
			time.Sleep(1 * time.Second)
		}
		resp, err := http.Get("http://" + webAddr)
		if err != nil {
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			break
		}
	}

	testInputs := `flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.CPU.Load 0.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.CPU.Time 16550000000.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.GarbageCollector.G1-Old-Generation.Count 1.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.GarbageCollector.G1-Young-Generation.Time 11.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.ClassLoader.ClassesUnloaded 1.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.GarbageCollector.G1-Old-Generation.Time 18.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.GarbageCollector.G1-Young-Generation.Count 2.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.Memory.Direct.MemoryUsed 107940111.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.Memory.Heap.Used 73232704.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.Memory.Direct.Count 3291.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.Memory.Direct.TotalCapacity 107940110.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.Memory.Heap.Committed 966787072.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.Memory.Heap.Max 966787072.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.Memory.Mapped.TotalCapacity 0.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.Memory.Mapped.Count 0.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.Network.TotalMemorySegments 3278.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.Memory.Mapped.MemoryUsed 0.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.ClassLoader.ClassesLoaded 4911.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.Memory.NonHeap.Committed 46530560.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.Memory.NonHeap.Used 44416856.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.Threads.Count 32.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.Network.AvailableMemorySegments 3278.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.Memory.NonHeap.Max -1.000000 NOW
flink.jm.test.Status.JVM.CPU.Load 0.000000 NOW
flink.jm.test.Status.JVM.Memory.Direct.MemoryUsed 560784.000000 NOW
flink.jm.test.Status.JVM.CPU.Time 49650000000.000000 NOW
flink.jmj.test.lastCheckpointSize -1.000000 NOW
flink.jm.test.Status.JVM.Memory.Direct.TotalCapacity 560783.000000 NOW
flink.jm.test.Status.JVM.Memory.Heap.Committed 1029177344.000000 NOW
flink.jm.test.Status.JVM.Memory.Heap.Max 1029177344.000000 NOW
flink.jm.test.Status.JVM.Memory.Heap.Used 114631120.000000 NOW
flink.jm.test.Status.JVM.Memory.Mapped.Count 0.000000 NOW
flink.jm.test.Status.JVM.Memory.Mapped.MemoryUsed 0.000000 NOW
flink.jm.test.Status.JVM.Memory.Mapped.TotalCapacity 0.000000 NOW
flink.jm.test.Status.JVM.Memory.NonHeap.Committed 68550656.000000 NOW
flink.jm.test.Status.JVM.Memory.NonHeap.Max -1.000000 NOW
flink.jm.test.Status.JVM.Memory.NonHeap.Used 65371512.000000 NOW
flink.jm.test.Status.JVM.Threads.Count 50.000000 NOW
flink.jm.test.numRegisteredTaskManagers 2.000000 NOW
flink.jm.test.numRunningJobs 1.000000 NOW
flink.jm.test.taskSlotsAvailable 1.000000 NOW
flink.jm.test.taskSlotsTotal 2.000000 NOW
flink.jmj.test.downtime 0.000000 NOW
flink.jmj.test.fullRestarts 0.000000 NOW
flink.jmj.test.lastCheckpointAlignmentBuffered -1.000000 NOW
flink.jmj.test.lastCheckpointDuration -1.000000 NOW
flink.jmj.test.lastCheckpointRestoreTimestamp -1.000000 NOW
flink.jmj.test.uptime 1527925.000000 NOW
flink.jmj.test.numberOfCompletedCheckpoints 0.000000 NOW
flink.jmj.test.numberOfFailedCheckpoints 0.000000 NOW
flink.jmj.test.numberOfInProgressCheckpoints 0.000000 NOW
flink.jmj.test.restartingTime 0.000000 NOW
flink.jmj.test.totalNumberOfCheckpoints 0.000000 NOW
flink.jm.test.Status.JVM.ClassLoader.ClassesLoaded 6718.000000 NOW
flink.jm.test.Status.JVM.ClassLoader.ClassesUnloaded 0.000000 NOW
flink.jm.test.Status.JVM.GarbageCollector.PS-MarkSweep.Count 2.000000 NOW
flink.jm.test.Status.JVM.GarbageCollector.PS-Scavenge.Time 33.000000 NOW
flink.jm.test.Status.JVM.GarbageCollector.PS-MarkSweep.Time 52.000000 NOW
flink.jm.test.Status.JVM.Memory.Direct.Count 16.000000 NOW
flink.jm.test.Status.JVM.GarbageCollector.PS-Scavenge.Count 3.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunctio.0.currentOutputWatermark -9223372036854775808.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Flat-Map.0.currentInputWatermark -9223372036854775808.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Flat-Map.0.currentOutputWatermark -9223372036854775808.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Source:-Socket-Stream.0.currentOutputWatermark -9223372036854775808.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Sink:-Print-to-Std--Out.0.currentInputWatermark -9223372036854775808.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersInLocalPerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.buffers.inPoolUsage 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersInRemotePerSecond.m15_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersInRemotePerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numRecordsInPerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.buffers.inputQueueLength 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersOutPerSecond.count 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.buffers.outPoolUsage 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersInLocalPerSecond.m5_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersOutPerSecond.m1_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.buffers.outputQueueLength 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numRecordsOutPerSecond.count 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersOutPerSecond.m5_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.checkpointAlignmentTime 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numRecordsOutPerSecond.m1_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersOutPerSecond.m15_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.currentInputWatermark -9223372036854775808.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numRecordsOutPerSecond.m5_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersOutPerSecond.mean_rate 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Flat-Map.0.numRecordsIn.count 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numRecordsOutPerSecond.m15_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesInLocalPerSecond.count 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Flat-Map.0.numRecordsOut.count 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numRecordsOutPerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesInLocalPerSecond.m1_rate 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Source:-Socket-Stream.0.numRecordsIn.count 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersInLocalPerSecond.count 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesInLocalPerSecond.m5_rate 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Source:-Socket-Stream.0.numRecordsOut.count 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Sink:-Print-to-Std--Out.0.numRecordsIn.count 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesInLocalPerSecond.m15_rate 0.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.GarbageCollector.G1-Old-Generation.Count 1.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Sink:-Print-to-Std--Out.0.numRecordsOut.count 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesInLocalPerSecond.mean_rate 0.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.CPU.Load 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunctio.0.numLateRecordsDropped.count 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesInRemotePerSecond.count 0.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.CPU.Time 41440000000.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunctio.0.numRecordsIn.count 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesInRemotePerSecond.m1_rate 0.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.ClassLoader.ClassesLoaded 5699.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunctio.0.numRecordsOut.count 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesInRemotePerSecond.m5_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesInRemotePerSecond.m15_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersInLocal.count 0.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.ClassLoader.ClassesUnloaded 2.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersInRemote.count 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesInRemotePerSecond.mean_rate 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunctio.0.currentInputWatermark -9223372036854775808.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesOutPerSecond.count 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersOut.count 0.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.GarbageCollector.G1-Young-Generation.Time 68.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesInLocal.count 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesOutPerSecond.m1_rate 0.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.GarbageCollector.G1-Old-Generation.Time 30.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesOutPerSecond.m5_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesInRemote.count 0.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.GarbageCollector.G1-Young-Generation.Count 5.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesOut.count 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesOutPerSecond.m15_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numRecordsIn.count 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesOutPerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numRecordsOut.count 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numRecordsInPerSecond.count 0.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.Memory.Direct.Count 3290.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersInLocal.count 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numRecordsInPerSecond.m1_rate 0.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.Memory.Direct.MemoryUsed 107940287.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersInRemote.count 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numRecordsInPerSecond.m5_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesInRemotePerSecond.count 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersOut.count 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numRecordsInPerSecond.m15_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersInLocalPerSecond.m15_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesInLocal.count 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersInLocalPerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesInLocalPerSecond.m5_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesInRemote.count 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersInRemotePerSecond.count 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesInLocalPerSecond.m15_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesOut.count 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesOutPerSecond.mean_rate 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Source:-Socket-Stream.0.numRecordsOutPerSecond.m1_rate 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Source:-Socket-Stream.0.numRecordsOutPerSecond.m5_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesInRemotePerSecond.m1_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersInRemotePerSecond.m1_rate 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Source:-Socket-Stream.0.numRecordsInPerSecond.m5_rate 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunctio.0.numRecordsOutPerSecond.m15_rate 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Source:-Socket-Stream.0.numRecordsInPerSecond.m15_rate 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Source:-Socket-Stream.0.numRecordsInPerSecond.mean_rate 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunctio.0.numRecordsOutPerSecond.mean_rate 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Source:-Socket-Stream.0.numRecordsOutPerSecond.m15_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersInLocalPerSecond.count 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesOutPerSecond.m15_rate 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Source:-Socket-Stream.0.numRecordsOutPerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersInLocalPerSecond.m1_rate 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Sink:-Print-to-Std--Out.0.numRecordsInPerSecond.count 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersInLocalPerSecond.m5_rate 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Sink:-Print-to-Std--Out.0.numRecordsInPerSecond.m1_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersInLocalPerSecond.m15_rate 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Sink:-Print-to-Std--Out.0.numRecordsInPerSecond.m5_rate 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Sink:-Print-to-Std--Out.0.currentOutputWatermark -9223372036854775808.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Sink:-Print-to-Std--Out.0.numRecordsInPerSecond.m15_rate 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Sink:-Print-to-Std--Out.0.numRecordsInPerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersInLocalPerSecond.m1_rate 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Sink:-Print-to-Std--Out.0.numRecordsOutPerSecond.count 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersInRemotePerSecond.count 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Sink:-Print-to-Std--Out.0.numRecordsOutPerSecond.m1_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersInRemotePerSecond.m1_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numRecordsInPerSecond.count 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersInRemotePerSecond.m5_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numRecordsInPerSecond.m1_rate 0.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.Memory.Heap.Used 279919544.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.Memory.Mapped.Count 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numRecordsInPerSecond.m5_rate 0.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.Memory.Mapped.MemoryUsed 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numRecordsInPerSecond.m15_rate 0.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.Memory.Mapped.TotalCapacity 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numRecordsOutPerSecond.m1_rate 0.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.Memory.NonHeap.Committed 56492032.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numRecordsOutPerSecond.count 0.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.Memory.NonHeap.Max -1.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.Memory.NonHeap.Used 53679864.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numRecordsOutPerSecond.m5_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numRecordsOutPerSecond.m15_rate 0.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.Threads.Count 38.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.Network.AvailableMemorySegments 3278.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.Memory.Direct.TotalCapacity 107940286.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.Network.TotalMemorySegments 3278.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numRecordsOutPerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.buffers.inPoolUsage 0.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.Memory.Heap.Committed 966787072.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.buffers.inputQueueLength 0.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.Memory.Heap.Max 966787072.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesInLocalPerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.buffers.outPoolUsage 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numRecordsOut.count 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.buffers.outputQueueLength 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Source:-Socket-Stream.0.numRecordsOutPerSecond.count 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Flat-Map.0.numRecordsInPerSecond.count 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Flat-Map.0.numRecordsInPerSecond.m1_rate 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Flat-Map.0.numRecordsInPerSecond.m5_rate 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Sink:-Print-to-Std--Out.0.numRecordsOutPerSecond.m5_rate 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Flat-Map.0.numRecordsInPerSecond.m15_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesInRemotePerSecond.m5_rate 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunctio.0.numRecordsInPerSecond.m15_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesInRemotePerSecond.mean_rate 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Source:-Socket-Stream.0.numRecordsInPerSecond.count 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunctio.0.numRecordsOutPerSecond.m5_rate 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Flat-Map.0.numRecordsInPerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesOutPerSecond.m5_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numRecordsInPerSecond.mean_rate 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunctio.0.numRecordsInPerSecond.m5_rate 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Flat-Map.0.numRecordsOutPerSecond.count 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersOutPerSecond.m5_rate 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Sink:-Print-to-Std--Out.0.numRecordsOutPerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersInRemotePerSecond.m5_rate 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunctio.0.numRecordsInPerSecond.count 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersInRemotePerSecond.m15_rate 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Flat-Map.0.numRecordsOutPerSecond.m1_rate 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Flat-Map.0.numRecordsOutPerSecond.m5_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersInRemotePerSecond.mean_rate 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunctio.0.numRecordsInPerSecond.m1_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersOutPerSecond.count 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Flat-Map.0.numRecordsOutPerSecond.m15_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersOutPerSecond.m1_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesInRemotePerSecond.m15_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesInLocalPerSecond.count 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Flat-Map.0.numRecordsOutPerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersOutPerSecond.m15_rate 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunctio.0.numRecordsInPerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersOutPerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesOutPerSecond.count 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesInLocalPerSecond.m1_rate 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunctio.0.numRecordsOutPerSecond.count 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunctio.0.numRecordsOutPerSecond.m1_rate 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Sink:-Print-to-Std--Out.0.numRecordsOutPerSecond.m15_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesOutPerSecond.m1_rate 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Source:-Socket-Stream.0.numRecordsInPerSecond.m1_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numRecordsIn.count 0.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.CPU.Load 0.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.ClassLoader.ClassesLoaded 4911.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.Memory.Mapped.MemoryUsed 0.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.ClassLoader.ClassesUnloaded 1.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.GarbageCollector.G1-Old-Generation.Count 1.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.GarbageCollector.G1-Old-Generation.Time 18.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.GarbageCollector.G1-Young-Generation.Count 2.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.Memory.Heap.Max 966787072.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.Network.TotalMemorySegments 3278.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.Memory.NonHeap.Max -1.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.GarbageCollector.G1-Young-Generation.Time 11.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.Memory.NonHeap.Committed 46530560.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.Memory.Direct.Count 3291.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.Memory.Heap.Committed 966787072.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.Memory.Direct.MemoryUsed 107940111.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.Memory.Direct.TotalCapacity 107940110.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.Memory.Mapped.Count 0.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.CPU.Time 16630000000.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.Threads.Count 32.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.Memory.Heap.Used 73232704.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.Network.AvailableMemorySegments 3278.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.Memory.NonHeap.Used 44416856.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.Memory.Mapped.TotalCapacity 0.000000 NOW
flink.jm.test.Status.JVM.ClassLoader.ClassesLoaded 6718.000000 NOW
flink.jm.test.Status.JVM.Memory.Direct.TotalCapacity 560783.000000 NOW
flink.jm.test.Status.JVM.GarbageCollector.PS-Scavenge.Count 3.000000 NOW
flink.jm.test.Status.JVM.GarbageCollector.PS-MarkSweep.Time 52.000000 NOW
flink.jm.test.Status.JVM.CPU.Load 0.000000 NOW
flink.jm.test.Status.JVM.Memory.Heap.Committed 1029177344.000000 NOW
flink.jm.test.Status.JVM.Memory.NonHeap.Committed 68550656.000000 NOW
flink.jm.test.Status.JVM.GarbageCollector.PS-Scavenge.Time 33.000000 NOW
flink.jm.test.Status.JVM.Memory.Direct.Count 16.000000 NOW
flink.jm.test.Status.JVM.Memory.Mapped.MemoryUsed 0.000000 NOW
flink.jm.test.Status.JVM.Memory.Direct.MemoryUsed 560784.000000 NOW
flink.jm.test.Status.JVM.CPU.Time 49750000000.000000 NOW
flink.jm.test.Status.JVM.ClassLoader.ClassesUnloaded 0.000000 NOW
flink.jm.test.Status.JVM.Memory.Heap.Used 115417904.000000 NOW
flink.jm.test.Status.JVM.Memory.Heap.Max 1029177344.000000 NOW
flink.jm.test.Status.JVM.Memory.Mapped.TotalCapacity 0.000000 NOW
flink.jm.test.Status.JVM.GarbageCollector.PS-MarkSweep.Count 2.000000 NOW
flink.jm.test.Status.JVM.Memory.Mapped.Count 0.000000 NOW
flink.jm.test.numRegisteredTaskManagers 2.000000 NOW
flink.jm.test.Status.JVM.Memory.NonHeap.Max -1.000000 NOW
flink.jm.test.taskSlotsTotal 2.000000 NOW
flink.jm.test.Status.JVM.Memory.NonHeap.Used 65371512.000000 NOW
flink.jm.test.numRunningJobs 1.000000 NOW
flink.jm.test.Status.JVM.Threads.Count 50.000000 NOW
flink.jm.test.taskSlotsAvailable 1.000000 NOW
flink.jmj.test.lastCheckpointAlignmentBuffered -1.000000 NOW
flink.jmj.test.lastCheckpointRestoreTimestamp -1.000000 NOW
flink.jmj.test.downtime 0.000000 NOW
flink.jmj.test.lastCheckpointDuration -1.000000 NOW
flink.jmj.test.fullRestarts 0.000000 NOW
flink.jmj.test.numberOfFailedCheckpoints 0.000000 NOW
flink.jmj.test.lastCheckpointSize -1.000000 NOW
flink.jmj.test.numberOfCompletedCheckpoints 0.000000 NOW
flink.jmj.test.restartingTime 0.000000 NOW
flink.jmj.test.numberOfInProgressCheckpoints 0.000000 NOW
flink.jmj.test.uptime 1537926.000000 NOW
flink.jmj.test.totalNumberOfCheckpoints 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Flat-Map.0.currentInputWatermark -9223372036854775808.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Flat-Map.0.currentOutputWatermark -9223372036854775808.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Sink:-Print-to-Std--Out.0.currentInputWatermark -9223372036854775808.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunctio.0.currentOutputWatermark -9223372036854775808.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Source:-Socket-Stream.0.currentOutputWatermark -9223372036854775808.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Sink:-Print-to-Std--Out.0.currentOutputWatermark -9223372036854775808.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunctio.0.currentInputWatermark -9223372036854775808.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.GarbageCollector.G1-Old-Generation.Count 1.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.CPU.Load 0.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.CPU.Time 41520000000.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.ClassLoader.ClassesLoaded 5699.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.ClassLoader.ClassesUnloaded 2.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.GarbageCollector.G1-Young-Generation.Count 5.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.GarbageCollector.G1-Old-Generation.Time 30.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.GarbageCollector.G1-Young-Generation.Time 68.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.Memory.Direct.Count 3290.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.Memory.Direct.MemoryUsed 107940287.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.Memory.Direct.TotalCapacity 107940286.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.Memory.Heap.Committed 966787072.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.Memory.Heap.Used 280968120.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.Memory.Mapped.Count 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.buffers.inPoolUsage 0.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.Memory.Heap.Max 966787072.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.Memory.Mapped.MemoryUsed 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.checkpointAlignmentTime 0.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.Network.AvailableMemorySegments 3278.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.Memory.Mapped.TotalCapacity 0.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.Memory.NonHeap.Committed 56492032.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.Memory.NonHeap.Max -1.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.Memory.NonHeap.Used 53679864.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.Threads.Count 38.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.buffers.inPoolUsage 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.buffers.outputQueueLength 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.buffers.outPoolUsage 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesOut.count 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersOut.count 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.buffers.inputQueueLength 0.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.Network.TotalMemorySegments 3278.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.currentInputWatermark -9223372036854775808.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numRecordsIn.count 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesInRemote.count 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesInLocal.count 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numRecordsOut.count 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Flat-Map.0.numRecordsInPerSecond.m5_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesOut.count 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Flat-Map.0.numRecordsOut.count 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.buffers.outPoolUsage 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Source:-Socket-Stream.0.numRecordsIn.count 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Source:-Socket-Stream.0.numRecordsOut.count 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Sink:-Print-to-Std--Out.0.numRecordsIn.count 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Sink:-Print-to-Std--Out.0.numRecordsOut.count 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.buffers.inputQueueLength 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunctio.0.numLateRecordsDropped.count 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Flat-Map.0.numRecordsInPerSecond.mean_rate 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunctio.0.numRecordsIn.count 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Flat-Map.0.numRecordsInPerSecond.count 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunctio.0.numRecordsOut.count 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Flat-Map.0.numRecordsInPerSecond.m1_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.buffers.outputQueueLength 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Flat-Map.0.numRecordsIn.count 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersInLocal.count 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersInRemote.count 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Flat-Map.0.numRecordsInPerSecond.m15_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numRecordsIn.count 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Source:-Socket-Stream.0.numRecordsInPerSecond.count 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersOut.count 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numRecordsOut.count 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Flat-Map.0.numRecordsOutPerSecond.count 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersInLocal.count 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Flat-Map.0.numRecordsOutPerSecond.m1_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersInRemote.count 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesInLocal.count 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Flat-Map.0.numRecordsOutPerSecond.m5_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesInRemote.count 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Flat-Map.0.numRecordsOutPerSecond.m15_rate 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Sink:-Print-to-Std--Out.0.numRecordsOutPerSecond.count 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Flat-Map.0.numRecordsOutPerSecond.mean_rate 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunctio.0.numRecordsInPerSecond.m5_rate 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Sink:-Print-to-Std--Out.0.numRecordsOutPerSecond.m1_rate 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunctio.0.numRecordsOutPerSecond.m5_rate 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Source:-Socket-Stream.0.numRecordsInPerSecond.m15_rate 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Sink:-Print-to-Std--Out.0.numRecordsOutPerSecond.m5_rate 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Source:-Socket-Stream.0.numRecordsInPerSecond.mean_rate 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Source:-Socket-Stream.0.numRecordsOutPerSecond.count 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Source:-Socket-Stream.0.numRecordsOutPerSecond.m1_rate 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Source:-Socket-Stream.0.numRecordsInPerSecond.m5_rate 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Source:-Socket-Stream.0.numRecordsOutPerSecond.m5_rate 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunctio.0.numRecordsInPerSecond.count 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Source:-Socket-Stream.0.numRecordsOutPerSecond.m15_rate 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunctio.0.numRecordsInPerSecond.m1_rate 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Source:-Socket-Stream.0.numRecordsOutPerSecond.mean_rate 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunctio.0.numRecordsOutPerSecond.m1_rate 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Sink:-Print-to-Std--Out.0.numRecordsInPerSecond.count 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Sink:-Print-to-Std--Out.0.numRecordsOutPerSecond.m15_rate 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Source:-Socket-Stream.0.numRecordsInPerSecond.m1_rate 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunctio.0.numRecordsOutPerSecond.count 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunctio.0.numRecordsInPerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersInLocalPerSecond.mean_rate 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunctio.0.numRecordsOutPerSecond.m15_rate 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunctio.0.numRecordsOutPerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersInLocalPerSecond.count 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersInLocalPerSecond.m1_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersInLocalPerSecond.m5_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersInLocalPerSecond.m15_rate 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Sink:-Print-to-Std--Out.0.numRecordsOutPerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersOutPerSecond.m15_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersInRemotePerSecond.count 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersInRemotePerSecond.m1_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersInRemotePerSecond.m5_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersInRemotePerSecond.m15_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersInRemotePerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersOutPerSecond.count 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersOutPerSecond.m1_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersOutPerSecond.m5_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesInRemotePerSecond.m15_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersOutPerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesInLocalPerSecond.count 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesInLocalPerSecond.mean_rate 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Sink:-Print-to-Std--Out.0.numRecordsInPerSecond.m1_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesInRemotePerSecond.m1_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numRecordsInPerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numRecordsOutPerSecond.m5_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numRecordsOutPerSecond.m15_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesInRemotePerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersInLocalPerSecond.m1_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesInLocalPerSecond.m1_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesOutPerSecond.count 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numRecordsOutPerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesOutPerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesInRemotePerSecond.m1_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesInRemotePerSecond.count 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesInLocalPerSecond.m5_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numRecordsInPerSecond.count 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersInLocalPerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numRecordsInPerSecond.m1_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersOutPerSecond.m5_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesInLocalPerSecond.m15_rate 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Sink:-Print-to-Std--Out.0.numRecordsInPerSecond.m5_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersOutPerSecond.m15_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesInLocalPerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesInRemotePerSecond.m5_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersOutPerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesInRemotePerSecond.count 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Sink:-Print-to-Std--Out.0.numRecordsInPerSecond.m15_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numRecordsInPerSecond.m5_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numRecordsOutPerSecond.count 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersOutPerSecond.m1_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersInLocalPerSecond.m5_rate 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Sink:-Print-to-Std--Out.0.numRecordsInPerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersInRemotePerSecond.count 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesInLocalPerSecond.m1_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numRecordsOutPerSecond.m1_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersInRemotePerSecond.m1_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersInRemotePerSecond.m5_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numRecordsInPerSecond.m15_rate 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunctio.0.numRecordsInPerSecond.m15_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersInRemotePerSecond.m15_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersInRemotePerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesOutPerSecond.m15_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesInLocalPerSecond.m5_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesOutPerSecond.m5_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersOutPerSecond.count 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesInRemotePerSecond.m5_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesInLocalPerSecond.count 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesInLocalPerSecond.m15_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersInLocalPerSecond.m15_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesOutPerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersInLocalPerSecond.count 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesInRemotePerSecond.m15_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesInRemotePerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesOutPerSecond.count 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numRecordsOutPerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesOutPerSecond.m1_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numRecordsInPerSecond.count 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesOutPerSecond.m5_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numRecordsInPerSecond.m1_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesOutPerSecond.m15_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numRecordsInPerSecond.m5_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numRecordsOutPerSecond.count 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numRecordsOutPerSecond.m1_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numRecordsInPerSecond.m15_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numRecordsOutPerSecond.m5_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numRecordsInPerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesOutPerSecond.m1_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numRecordsOutPerSecond.m15_rate 0.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.CPU.Load 0.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.CPU.Time 16690000000.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.GarbageCollector.G1-Young-Generation.Time 11.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.Memory.Direct.MemoryUsed 107940111.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.ClassLoader.ClassesLoaded 4911.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.Memory.Direct.Count 3291.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.ClassLoader.ClassesUnloaded 1.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.GarbageCollector.G1-Old-Generation.Count 1.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.GarbageCollector.G1-Old-Generation.Time 18.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.GarbageCollector.G1-Young-Generation.Count 2.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.Memory.Mapped.MemoryUsed 0.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.Memory.Direct.TotalCapacity 107940110.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.Memory.Heap.Committed 966787072.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.Memory.NonHeap.Max -1.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.Memory.Heap.Max 966787072.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.Memory.Mapped.TotalCapacity 0.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.Memory.Heap.Used 74281280.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.Memory.NonHeap.Committed 46530560.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.Memory.Mapped.Count 0.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.Memory.NonHeap.Used 44416856.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.Network.TotalMemorySegments 3278.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.JVM.Threads.Count 32.000000 NOW
flink.tm.test.6afca7033b297cb69f5e2176d5e2c61e.Status.Network.AvailableMemorySegments 3278.000000 NOW
flink.jm.test.Status.JVM.CPU.Load 0.000000 NOW
flink.jm.test.Status.JVM.ClassLoader.ClassesLoaded 6718.000000 NOW
flink.jm.test.Status.JVM.GarbageCollector.PS-MarkSweep.Time 52.000000 NOW
flink.jm.test.Status.JVM.ClassLoader.ClassesUnloaded 0.000000 NOW
flink.jm.test.Status.JVM.CPU.Time 49860000000.000000 NOW
flink.jm.test.Status.JVM.Memory.Heap.Committed 1029177344.000000 NOW
flink.jm.test.Status.JVM.GarbageCollector.PS-Scavenge.Count 3.000000 NOW
flink.jm.test.Status.JVM.GarbageCollector.PS-Scavenge.Time 33.000000 NOW
flink.jm.test.Status.JVM.Memory.Direct.Count 16.000000 NOW
flink.jm.test.Status.JVM.Memory.Direct.MemoryUsed 560784.000000 NOW
flink.jm.test.Status.JVM.Memory.Direct.TotalCapacity 560783.000000 NOW
flink.jm.test.Status.JVM.Memory.NonHeap.Used 65381304.000000 NOW
flink.jm.test.Status.JVM.Threads.Count 50.000000 NOW
flink.jm.test.numRegisteredTaskManagers 2.000000 NOW
flink.jmj.test.uptime 1547928.000000 NOW
flink.jm.test.numRunningJobs 1.000000 NOW
flink.jm.test.Status.JVM.Memory.Heap.Max 1029177344.000000 NOW
flink.jm.test.Status.JVM.Memory.Mapped.Count 0.000000 NOW
flink.jm.test.Status.JVM.Memory.Heap.Used 116005456.000000 NOW
flink.jm.test.Status.JVM.Memory.Mapped.MemoryUsed 0.000000 NOW
flink.jm.test.Status.JVM.Memory.NonHeap.Max -1.000000 NOW
flink.jm.test.Status.JVM.Memory.Mapped.TotalCapacity 0.000000 NOW
flink.jm.test.Status.JVM.Memory.NonHeap.Committed 68550656.000000 NOW
flink.jm.test.Status.JVM.GarbageCollector.PS-MarkSweep.Count 2.000000 NOW
flink.jmj.test.lastCheckpointRestoreTimestamp -1.000000 NOW
flink.jmj.test.lastCheckpointSize -1.000000 NOW
flink.jmj.test.numberOfCompletedCheckpoints 0.000000 NOW
flink.jmj.test.numberOfFailedCheckpoints 0.000000 NOW
flink.jmj.test.numberOfInProgressCheckpoints 0.000000 NOW
flink.jmj.test.restartingTime 0.000000 NOW
flink.jmj.test.totalNumberOfCheckpoints 0.000000 NOW
flink.jmj.test.downtime 0.000000 NOW
flink.jm.test.taskSlotsAvailable 1.000000 NOW
flink.jm.test.taskSlotsTotal 2.000000 NOW
flink.jmj.test.lastCheckpointAlignmentBuffered -1.000000 NOW
flink.jmj.test.fullRestarts 0.000000 NOW
flink.jmj.test.lastCheckpointDuration -1.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Flat-Map.0.currentInputWatermark -9223372036854775808.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunctio.0.currentOutputWatermark -9223372036854775808.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Flat-Map.0.currentOutputWatermark -9223372036854775808.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Source:-Socket-Stream.0.currentOutputWatermark -9223372036854775808.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Sink:-Print-to-Std--Out.0.currentInputWatermark -9223372036854775808.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Sink:-Print-to-Std--Out.0.currentOutputWatermark -9223372036854775808.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunctio.0.currentInputWatermark -9223372036854775808.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.Memory.Direct.Count 3290.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.CPU.Load 0.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.CPU.Time 41600000000.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.ClassLoader.ClassesLoaded 5699.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.ClassLoader.ClassesUnloaded 2.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.GarbageCollector.G1-Old-Generation.Count 1.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.GarbageCollector.G1-Old-Generation.Time 30.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.GarbageCollector.G1-Young-Generation.Count 5.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.GarbageCollector.G1-Young-Generation.Time 68.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.Memory.Direct.MemoryUsed 107940287.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunctio.0.numRecordsInPerSecond.mean_rate 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Flat-Map.0.numRecordsOutPerSecond.m1_rate 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunctio.0.numRecordsOutPerSecond.mean_rate 0.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.Memory.Direct.TotalCapacity 107940286.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersInRemotePerSecond.mean_rate 0.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.Memory.Heap.Committed 966787072.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesInLocalPerSecond.m15_rate 0.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.Memory.Heap.Max 966787072.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesInRemotePerSecond.m15_rate 0.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.Memory.Heap.Used 283065272.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesOutPerSecond.m5_rate 0.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.Memory.Mapped.Count 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numRecordsOutPerSecond.m5_rate 0.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.Memory.Mapped.MemoryUsed 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesOutPerSecond.count 0.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.Memory.Mapped.TotalCapacity 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numRecordsInPerSecond.m15_rate 0.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.Memory.NonHeap.Committed 56492032.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Sink:-Print-to-Std--Out.0.numRecordsIn.count 0.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.Memory.NonHeap.Max -1.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Source:-Socket-Stream.0.numRecordsIn.count 0.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.Memory.NonHeap.Used 53681720.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Source:-Socket-Stream.0.numRecordsOut.count 0.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.JVM.Threads.Count 38.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Flat-Map.0.numRecordsOutPerSecond.m5_rate 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Sink:-Print-to-Std--Out.0.numRecordsOut.count 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.buffers.inPoolUsage 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesOut.count 0.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.Network.TotalMemorySegments 3278.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesInLocal.count 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.buffers.inPoolUsage 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.buffers.inputQueueLength 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Flat-Map.0.numRecordsInPerSecond.m1_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.buffers.outPoolUsage 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.buffers.outputQueueLength 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Source:-Socket-Stream.0.numRecordsInPerSecond.count 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.buffers.outputQueueLength 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.buffers.inputQueueLength 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.buffers.outPoolUsage 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.currentInputWatermark -9223372036854775808.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.checkpointAlignmentTime 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Flat-Map.0.numRecordsIn.count 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Flat-Map.0.numRecordsOut.count 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunctio.0.numRecordsOutPerSecond.m5_rate 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunctio.0.numLateRecordsDropped.count 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunctio.0.numRecordsIn.count 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Flat-Map.0.numRecordsOutPerSecond.m15_rate 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunctio.0.numRecordsOut.count 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Flat-Map.0.numRecordsOutPerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersInLocal.count 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersInRemote.count 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Source:-Socket-Stream.0.numRecordsInPerSecond.m1_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersOut.count 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Source:-Socket-Stream.0.numRecordsInPerSecond.m5_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesInLocal.count 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Source:-Socket-Stream.0.numRecordsInPerSecond.m15_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesInRemote.count 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Source:-Socket-Stream.0.numRecordsInPerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesOut.count 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Source:-Socket-Stream.0.numRecordsOutPerSecond.count 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Source:-Socket-Stream.0.numRecordsOutPerSecond.m1_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numRecordsIn.count 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Source:-Socket-Stream.0.numRecordsOutPerSecond.m5_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersInLocal.count 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Source:-Socket-Stream.0.numRecordsOutPerSecond.m15_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersInRemote.count 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Source:-Socket-Stream.0.numRecordsOutPerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersOut.count 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Sink:-Print-to-Std--Out.0.numRecordsInPerSecond.count 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Sink:-Print-to-Std--Out.0.numRecordsInPerSecond.m1_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesInRemote.count 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Sink:-Print-to-Std--Out.0.numRecordsInPerSecond.m5_rate 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Sink:-Print-to-Std--Out.0.numRecordsInPerSecond.m15_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numRecordsIn.count 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Sink:-Print-to-Std--Out.0.numRecordsInPerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numRecordsOut.count 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Sink:-Print-to-Std--Out.0.numRecordsOutPerSecond.count 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Flat-Map.0.numRecordsInPerSecond.count 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Sink:-Print-to-Std--Out.0.numRecordsOutPerSecond.m1_rate 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Sink:-Print-to-Std--Out.0.numRecordsOutPerSecond.m5_rate 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Flat-Map.0.numRecordsInPerSecond.m5_rate 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Sink:-Print-to-Std--Out.0.numRecordsOutPerSecond.m15_rate 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Flat-Map.0.numRecordsInPerSecond.m15_rate 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Sink:-Print-to-Std--Out.0.numRecordsOutPerSecond.mean_rate 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Flat-Map.0.numRecordsInPerSecond.mean_rate 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunctio.0.numRecordsInPerSecond.count 0.000000 NOW
flink.op.test.Source:-Socket-Stream-->-Flat-Map.Flat-Map.0.numRecordsOutPerSecond.count 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunctio.0.numRecordsInPerSecond.m1_rate 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunctio.0.numRecordsInPerSecond.m5_rate 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunctio.0.numRecordsOutPerSecond.count 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunctio.0.numRecordsInPerSecond.m15_rate 0.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunctio.0.numRecordsOutPerSecond.m1_rate 0.000000 NOW
flink.tm.test.bee4c3815f3eb72a99bb1115fbae8368.Status.Network.AvailableMemorySegments 3278.000000 NOW
flink.op.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunctio.0.numRecordsOutPerSecond.m15_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersInLocalPerSecond.count 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersInLocalPerSecond.m1_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersInLocalPerSecond.m15_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersInLocalPerSecond.m5_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersInLocalPerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersInRemotePerSecond.count 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersInRemotePerSecond.m5_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersInRemotePerSecond.m1_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersOutPerSecond.count 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersInRemotePerSecond.m15_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersOutPerSecond.m1_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersOutPerSecond.m5_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersOutPerSecond.m15_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBuffersOutPerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesInLocalPerSecond.m5_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesInLocalPerSecond.count 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesInLocalPerSecond.m1_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesInLocalPerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesInRemotePerSecond.m5_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesInRemotePerSecond.count 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesInRemotePerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesInRemotePerSecond.m1_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesOutPerSecond.m15_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesOutPerSecond.count 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesOutPerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numBytesOutPerSecond.m1_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numRecordsInPerSecond.m1_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numRecordsInPerSecond.count 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numRecordsInPerSecond.m5_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numRecordsInPerSecond.m15_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numRecordsInPerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numRecordsOutPerSecond.m1_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numRecordsOutPerSecond.count 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numRecordsOutPerSecond.m15_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numRecordsOutPerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersInLocalPerSecond.m1_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersInLocalPerSecond.m5_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersInLocalPerSecond.m15_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersInLocalPerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersInLocalPerSecond.count 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersInRemotePerSecond.count 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersInRemotePerSecond.m5_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersInRemotePerSecond.m1_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersInRemotePerSecond.m15_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersInRemotePerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersOutPerSecond.count 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersOutPerSecond.m1_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersOutPerSecond.m5_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersOutPerSecond.m15_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBuffersOutPerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesInLocalPerSecond.count 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesInLocalPerSecond.m1_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesInLocalPerSecond.m5_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesInLocalPerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesInLocalPerSecond.m15_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesInRemotePerSecond.count 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesInRemotePerSecond.m1_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesInRemotePerSecond.m5_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesInRemotePerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesInRemotePerSecond.m15_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesOutPerSecond.m5_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesOutPerSecond.m1_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numRecordsInPerSecond.count 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesOutPerSecond.m15_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numBytesOutPerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numRecordsInPerSecond.m5_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numRecordsInPerSecond.m1_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numRecordsOutPerSecond.count 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numRecordsInPerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numRecordsOutPerSecond.m1_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numRecordsOutPerSecond.m5_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numRecordsOutPerSecond.mean_rate 0.000000 NOW
flink.tsk.test.Window(TumblingProcessingTimeWindows(5000),-ProcessingTimeTrigger,-ReduceFunction$1,-PassThroughWindowFunction)-->-Sink:-Print-to-Std--Out.0.numRecordsOutPerSecond.m15_rate 0.000000 NOW
flink.tsk.test.Source:-Socket-Stream-->-Flat-Map.0.numRecordsOut.count 0.000000 NOW
`
	lines := strings.Split(testInputs, "\n")
	currSec := time.Now().Unix() - 2
	for _, input := range lines {
		conn, err := net.Dial("udp", graphiteAddr)
		updateInput := strings.ReplaceAll(input, "NOW", strconv.FormatInt(currSec, 10))

		if err != nil {
			t.Fatalf("connection error: %v", err)
		}
		_, err = conn.Write([]byte(updateInput))
		if err != nil {
			t.Fatalf("write error: %v", err)
		}
		conn.Close()
	}

	resp, err := http.Get("http://" + path.Join(webAddr, "metrics"))
	if err != nil {
		t.Fatalf("get error: %v", err)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}

	responseString := string(b)
	print(responseString)

}
