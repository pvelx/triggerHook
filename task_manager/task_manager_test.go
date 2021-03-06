package task_manager

import (
	"context"
	"testing"
	"time"

	"github.com/pvelx/triggerhook/contracts"
	"github.com/pvelx/triggerhook/domain"
	"github.com/pvelx/triggerhook/error_service"
	"github.com/pvelx/triggerhook/monitoring_service"
	"github.com/pvelx/triggerhook/repository"
	"github.com/pvelx/triggerhook/util"
	"github.com/stretchr/testify/assert"
)

func TestTaskManager_Delete(t *testing.T) {
	tests := []struct {
		name                        string
		inputErrorRepository        []error
		inputTask                   domain.Task
		expectedError               error
		countCallMethodOfRepository int
		expectedEvents              []string
	}{
		{
			name:                        "main flow - without error",
			inputErrorRepository:        []error{nil},
			expectedError:               contracts.TmErrorTaskNotFound,
			countCallMethodOfRepository: 1,
			expectedEvents:              []string{},
		},
		{
			name:                        "1 times retryable error",
			inputErrorRepository:        []error{contracts.RepoErrorDeadlock, nil},
			expectedError:               contracts.TmErrorTaskNotFound,
			countCallMethodOfRepository: 2,
			expectedEvents:              []string{contracts.RepoErrorDeadlock.Error()},
		},
		{
			name:                        "2 times retryable error",
			inputErrorRepository:        []error{contracts.RepoErrorDeadlock, contracts.RepoErrorDeadlock, nil},
			expectedError:               contracts.TmErrorTaskNotFound,
			countCallMethodOfRepository: 3,
			expectedEvents:              []string{contracts.RepoErrorDeadlock.Error(), contracts.RepoErrorDeadlock.Error()},
		},
		{
			name:                        "3 times  retryable error",
			inputErrorRepository:        []error{contracts.RepoErrorDeadlock, contracts.RepoErrorDeadlock, contracts.RepoErrorDeadlock, nil},
			expectedError:               contracts.TmErrorDeletingTask,
			countCallMethodOfRepository: 3,
			expectedEvents: []string{
				contracts.RepoErrorDeadlock.Error(),
				contracts.RepoErrorDeadlock.Error(),
				contracts.RepoErrorDeadlock.Error(),
				contracts.RepoErrorDeadlock.Error(),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			countCallMethodOfRepository := 0
			r := &repository.RepositoryMock{DeleteMock: func(ctx context.Context, tasks []domain.Task) (affected int64, err error) {
				err = test.inputErrorRepository[countCallMethodOfRepository]
				countCallMethodOfRepository++

				return
			}}

			countCallNewOfEventHandler := 0
			eh := &error_service.ErrorHandlerMock{NewMock: func(level contracts.Level, eventMessage string, extra map[string]interface{}) {
				assert.Equal(t, contracts.LevelError, level, "must be LevelError")
				assert.Equal(t, test.expectedEvents[countCallNewOfEventHandler], eventMessage, "must be LevelError")
				countCallNewOfEventHandler++
			}}

			tm := New(r, eh, &monitoring_service.MonitoringMock{}, nil)

			result := tm.Delete(context.Background(), util.NewId())

			assert.Equal(t, test.expectedError, result, "error from task manager is not correct")

			assert.Equal(t, test.countCallMethodOfRepository, countCallMethodOfRepository,
				"is not correct call method delete of repository")

			assert.Equal(t, len(test.expectedEvents), countCallNewOfEventHandler,
				"is not correct count of event")
		})
	}
}

func TestTaskManager_Create(t *testing.T) {
	tests := []struct {
		name                        string
		inputErrorRepository        []error
		inputTask                   domain.Task
		expectedError               error
		countCallMethodOfRepository int
		expectedEvents              []string
	}{
		{
			name:                        "main flow - without error",
			inputErrorRepository:        []error{nil},
			expectedError:               nil,
			countCallMethodOfRepository: 1,
			expectedEvents:              []string{},
		},
		{
			name:                        "3 times retryable error",
			inputErrorRepository:        []error{contracts.RepoErrorDeadlock, contracts.RepoErrorDeadlock, contracts.RepoErrorDeadlock, nil},
			expectedError:               contracts.TmErrorCreatingTasks,
			countCallMethodOfRepository: 3,
			expectedEvents: []string{
				contracts.RepoErrorDeadlock.Error(),
				contracts.RepoErrorDeadlock.Error(),
				contracts.RepoErrorDeadlock.Error(),
				contracts.RepoErrorDeadlock.Error(),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			countCallMethodOfRepository := 0
			r := &repository.RepositoryMock{CreateMock: func(ctx context.Context, task domain.Task, isTaken bool) (err error) {
				err = test.inputErrorRepository[countCallMethodOfRepository]
				countCallMethodOfRepository++

				return
			}}

			countCallNewOfEventHandler := 0
			eh := &error_service.ErrorHandlerMock{NewMock: func(level contracts.Level, eventMessage string, extra map[string]interface{}) {
				assert.Equal(t, contracts.LevelError, level, "must be LevelError")
				assert.Equal(t, test.expectedEvents[countCallNewOfEventHandler], eventMessage, "must be LevelError")
				countCallNewOfEventHandler++
			}}

			tm := New(r, eh, &monitoring_service.MonitoringMock{}, nil)

			result := tm.Create(context.Background(), &domain.Task{}, true)

			assert.Equal(t, test.expectedError, result, "error from task manager is not correct")

			assert.Equal(t, test.countCallMethodOfRepository, countCallMethodOfRepository,
				"is not correct call method delete of repository")

			assert.Equal(t, len(test.expectedEvents), countCallNewOfEventHandler,
				"is not correct count of event")
		})
	}
}

func TestTaskManager_ConfirmExecution(t *testing.T) {
	tests := []struct {
		name                        string
		inputErrorRepository        []error
		inputTask                   domain.Task
		expectedError               error
		countCallMethodOfRepository int
		expectedEvents              []string
	}{
		{
			name:                        "main flow - without error",
			inputErrorRepository:        []error{nil},
			expectedError:               nil,
			countCallMethodOfRepository: 1,
			expectedEvents:              []string{},
		},
		{
			name:                        "3 times retryable error",
			inputErrorRepository:        []error{contracts.RepoErrorDeadlock, contracts.RepoErrorDeadlock, contracts.RepoErrorDeadlock, nil},
			expectedError:               contracts.TmErrorConfirmationTasks,
			countCallMethodOfRepository: 3,
			expectedEvents: []string{
				contracts.RepoErrorDeadlock.Error(),
				contracts.RepoErrorDeadlock.Error(),
				contracts.RepoErrorDeadlock.Error(),
				contracts.RepoErrorDeadlock.Error(),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			countCallMethodOfRepository := 0
			r := &repository.RepositoryMock{DeleteMock: func(ctx context.Context, tasks []domain.Task) (affected int64, err error) {
				err = test.inputErrorRepository[countCallMethodOfRepository]
				countCallMethodOfRepository++

				return
			}}

			countCallNewOfEventHandler := 0
			eh := &error_service.ErrorHandlerMock{NewMock: func(level contracts.Level, eventMessage string, extra map[string]interface{}) {
				assert.Equal(t, contracts.LevelError, level, "must be LevelError")
				assert.Equal(t, test.expectedEvents[countCallNewOfEventHandler], eventMessage, "must be LevelError")
				countCallNewOfEventHandler++
			}}

			tm := New(r, eh, &monitoring_service.MonitoringMock{}, nil)

			result := tm.ConfirmExecution(context.Background(), []domain.Task{{}, {}, {}})

			assert.Equal(t, test.expectedError, result, "error from task manager is not correct")

			assert.Equal(t, test.countCallMethodOfRepository, countCallMethodOfRepository,
				"is not correct call method delete of repository")

			assert.Equal(t, len(test.expectedEvents), countCallNewOfEventHandler,
				"is not correct count of event")
		})
	}
}

func TestTaskManagerMock_GetTasksToComplete(t *testing.T) {
	tests := []struct {
		name             string
		repositoryResult []struct {
			error      error
			collection contracts.CollectionsInterface
		}
		inputTask                   domain.Task
		expectedError               error
		expectedResult              contracts.CollectionsInterface
		countCallMethodOfRepository int
		expectedEvents              []string
	}{
		{
			name: "main flow - without error",
			repositoryResult: []struct {
				error      error
				collection contracts.CollectionsInterface
			}{
				{nil, &repository.Collections{}},
			},
			expectedResult:              &repository.Collections{},
			expectedError:               nil,
			countCallMethodOfRepository: 1,
			expectedEvents:              []string{},
		},
		{
			name: "3 times retryable error",
			repositoryResult: []struct {
				error      error
				collection contracts.CollectionsInterface
			}{
				{contracts.RepoErrorDeadlock, nil},
				{contracts.RepoErrorDeadlock, nil},
				{contracts.RepoErrorDeadlock, nil},
				{nil, &repository.Collections{}},
			},
			expectedError:               contracts.TmErrorGetTasks,
			expectedResult:              nil,
			countCallMethodOfRepository: 3,
			expectedEvents: []string{
				contracts.RepoErrorDeadlock.Error(),
				contracts.RepoErrorDeadlock.Error(),
				contracts.RepoErrorDeadlock.Error(),
				contracts.RepoErrorDeadlock.Error(),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			countCallMethodOfRepository := 0
			r := &repository.RepositoryMock{FindBySecToExecTimeMock: func(ctx context.Context, preloadingTimeRange time.Duration) (
				collection contracts.CollectionsInterface,
				err error,
			) {
				err = test.repositoryResult[countCallMethodOfRepository].error
				collection = test.repositoryResult[countCallMethodOfRepository].collection
				countCallMethodOfRepository++

				return
			}}

			countCallNewOfEventHandler := 0
			eh := &error_service.ErrorHandlerMock{NewMock: func(level contracts.Level, eventMessage string, extra map[string]interface{}) {
				assert.Equal(t, contracts.LevelError, level, "must be LevelError")
				assert.Equal(t, test.expectedEvents[countCallNewOfEventHandler], eventMessage, "must be LevelError")
				countCallNewOfEventHandler++
			}}

			tm := New(r, eh, &monitoring_service.MonitoringMock{}, nil)

			result, err := tm.GetTasksToComplete(context.Background(), time.Second)

			assert.Equal(t, test.expectedResult, result, "result from task manager is not correct")
			assert.Equal(t, test.expectedError, err, "error from task manager is not correct")

			assert.Equal(t, test.countCallMethodOfRepository, countCallMethodOfRepository,
				"is not correct call method delete of repository")

			assert.Equal(t, len(test.expectedEvents), countCallNewOfEventHandler,
				"is not correct count of event")
		})
	}
}
