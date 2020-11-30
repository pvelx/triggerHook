package services

import (
	"fmt"
	"github.com/VladislavPav/trigger-hook/domain"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func taskBunch(execTime int64, count int) []domain.Task {
	var taskBunch []domain.Task
	for i := 0; i < count; i++ {
		taskBunch = append(taskBunch, domain.Task{Id: int64(i), ExecTime: execTime})
	}
	return taskBunch
}

func Test(t *testing.T) {
	chPreloadedTask := make(chan domain.Task, 10000)
	chTasksReadyToSend := make(chan domain.Task, 10000)
	waitingTaskService := NewWaitingTaskService(chPreloadedTask, chTasksReadyToSend)
	countOfTasksOnSameTime := 1000
	countOfSeconds := 10
	var pauseSec float32 = 0.5

	go func() {
		for i := 0; i < countOfSeconds; i++ {
			execTime := time.Now().Unix() + int64(i)
			for _, task := range taskBunch(execTime, countOfTasksOnSameTime) {
				chPreloadedTask <- task
			}
			time.Sleep(time.Duration(pauseSec*1000) * time.Millisecond)
		}
	}()

	actualCountOfTasks := 0
	go func() {
		for {
			select {
			case task := <-chTasksReadyToSend:
				//fmt.Println(".", actualCountOfTasks)
				now := time.Now().Unix()
				assert.True(
					t,
					now == task.ExecTime,
					fmt.Sprintf("Time execution in not equal to current time. "+
						"Expected: %d, actual: %d", now, task.ExecTime),
				)
				actualCountOfTasks++
			}
		}
	}()

	go waitingTaskService.WaitUntilExecTime()

	sleepTime := countOfSeconds + int(float32(countOfSeconds)*pauseSec)
	time.Sleep(time.Duration(sleepTime) * time.Second)

	expectedCountOfTasks := countOfSeconds * countOfTasksOnSameTime
	assert.True(
		t,
		actualCountOfTasks == countOfSeconds*countOfTasksOnSameTime,
		fmt.Sprintf("Handled count of the task was not equal to expected. "+
			"Actual: %d, expected: %d", actualCountOfTasks, expectedCountOfTasks),
	)
}