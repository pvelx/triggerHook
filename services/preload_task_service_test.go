package services

import (
	"fmt"
	"github.com/pvelx/triggerHook/contracts"
	"github.com/pvelx/triggerHook/domain"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"sync/atomic"
	"testing"
	"time"
)

func TestTaskAdding(t *testing.T) {
	var isTakenActual bool
	taskManagerMock := &taskManagerMock{createMock: func(task domain.Task, isTaken bool) error {
		isTakenActual = isTaken
		return nil
	}}

	preloadingTaskService := NewPreloadingTaskService(taskManagerMock, nil)

	now := time.Now().Unix()
	tests := []struct {
		task            domain.Task
		isTakenExpected bool
	}{
		{domain.Task{Id: uuid.NewV4().String(), ExecTime: now - 10}, true},
		{domain.Task{Id: uuid.NewV4().String(), ExecTime: now - 1}, true},
		{domain.Task{Id: uuid.NewV4().String(), ExecTime: now + 0}, true},
		{domain.Task{Id: uuid.NewV4().String(), ExecTime: now + 1}, true},
		{domain.Task{Id: uuid.NewV4().String(), ExecTime: now + 9}, true},
		{domain.Task{Id: uuid.NewV4().String(), ExecTime: now + 10}, false},
		{domain.Task{Id: uuid.NewV4().String(), ExecTime: now + 11}, false},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			chPreloadedTask := preloadingTaskService.GetPreloadedChan()
			if err := preloadingTaskService.AddNewTask(tt.task); err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tt.isTakenExpected, isTakenActual, "The task must be taken")
			if tt.isTakenExpected {
				assert.Equal(t, 1, len(chPreloadedTask), "Len of channel is wrong")
				assert.Equal(t, tt.task, <-chPreloadedTask, "The task must be send in channel")
			} else {
				assert.Len(t, chPreloadedTask, 0, "Len of channel is wrong")
			}
		})
	}
}

func TestMainFlow(t *testing.T) {

	type collectionsType []struct {
		TaskCount int
		Duration  time.Duration
	}

	data := []struct {
		collections collectionsType
	}{
		{
			collections: collectionsType{
				{TaskCount: 1000, Duration: 50 * time.Millisecond},
				{TaskCount: 1000, Duration: 50 * time.Millisecond},
				{TaskCount: 1000, Duration: 50 * time.Millisecond},
				{TaskCount: 1000, Duration: 50 * time.Millisecond},
				{TaskCount: 1000, Duration: 50 * time.Millisecond},
				{TaskCount: 1000, Duration: 50 * time.Millisecond},
				{TaskCount: 1000, Duration: 50 * time.Millisecond},
				{TaskCount: 1, Duration: 10 * time.Millisecond},
				{TaskCount: 1, Duration: 10 * time.Millisecond},
				{TaskCount: 1, Duration: 10 * time.Millisecond},
				{TaskCount: 333, Duration: 20 * time.Millisecond},
			},
		},
		{
			collections: collectionsType{
				{TaskCount: 1000, Duration: 50 * time.Millisecond},
				{TaskCount: 1000, Duration: 50 * time.Millisecond},
				{TaskCount: 1000, Duration: 50 * time.Millisecond},
				{TaskCount: 1, Duration: 10 * time.Millisecond},
				{TaskCount: 1, Duration: 10 * time.Millisecond},
				{TaskCount: 333, Duration: 20 * time.Millisecond},
			},
		},
		{
			collections: collectionsType{},
		},
		{
			collections: collectionsType{
				{TaskCount: 1000, Duration: 50 * time.Millisecond},
			},
		},
	}

	var globalCurrentFinding int32 = 0

	taskManagerMock := &taskManagerMock{
		getTasksToCompleteMock: func(preloadingTimeRange time.Duration) (contracts.CollectionsInterface, error) {

			currentFinding := atomic.LoadInt32(&globalCurrentFinding)
			if len(data) > int(currentFinding) {
				atomic.AddInt32(&globalCurrentFinding, 1)
				collections := data[currentFinding].collections

				var globalCurrentCollection int32 = 0
				return &collectionsMock{nextMock: func() (tasks []domain.Task, isEnd bool, err error) {

					currentCollection := atomic.LoadInt32(&globalCurrentCollection)
					if len(collections) > int(currentCollection) {
						atomic.AddInt32(&globalCurrentCollection, 1)
						collection := collections[currentCollection]
						for i := 0; i < collection.TaskCount; i++ {
							tasks = append(tasks, domain.Task{
								Id:       fmt.Sprintf("%d/%d/%d", currentFinding, currentCollection, i),
								ExecTime: time.Now().Unix(),
							})
						}
						time.Sleep(collection.Duration)

						return tasks, false, nil
					}

					/*
						Collections of tasks were ended
					*/
					return nil, true, nil
				}}, nil
			}

			/*
				Tasks with a suitable time of execute not found
			*/
			return nil, contracts.NoTasksFound
		},
	}

	preloadingTaskService := NewPreloadingTaskService(taskManagerMock, nil)

	chPreloadedTask := preloadingTaskService.GetPreloadedChan()
	go preloadingTaskService.Preload()

	receivedTasks := make(map[string]domain.Task)

	time.Sleep(6 * time.Second)

	for _, item := range data {
		for _, collection := range item.collections {

			if len(chPreloadedTask) == 0 {
				t.Fatal("tasks was received not enough")
			}

			for i := 0; i < collection.TaskCount; i++ {
				taskActual := <-chPreloadedTask
				if _, exist := receivedTasks[taskActual.Id]; exist {
					t.Fatal(fmt.Sprintf("task '%s' already received", taskActual.Id))
				}
				receivedTasks[taskActual.Id] = taskActual
			}
		}
	}
	assert.Equal(t, 0, len(chPreloadedTask), "was received extra task")
}
