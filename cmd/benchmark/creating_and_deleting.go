package main

import (
	"fmt"
	"github.com/cheggaaa/pb/v3"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/pvelx/triggerHook"
	"github.com/pvelx/triggerHook/connection"
	"github.com/pvelx/triggerHook/contracts"
	"github.com/pvelx/triggerHook/domain"
)

func creatingAndDeleting(taskCount int) [][]string {
	var durationDeleting time.Duration

	triggerHookService := triggerHook.Build(triggerHook.Config{
		Connection: connection.Options{
			User:     "root",
			Password: "secret",
			Host:     "127.0.0.1:3306",
			DbName:   "test_db",
		},
	})

	go func() {
		if err := triggerHookService.Run(); err != nil {
			log.Fatal(err)
		}
	}()

	point := time.Now()

	taskToDelete := createTasks(triggerHookService, taskCount, 300)
	duration := time.Since(point)

	point2 := time.Now()
	deleteTasks(taskToDelete, triggerHookService)
	durationDeleting = time.Since(point2)

	return [][]string{
		{
			"Creating task",
			fmt.Sprintf("%v", duration),
			fmt.Sprintf("%f", float64(taskCount)/duration.Seconds()),
		},
		{
			"Deleting task",
			fmt.Sprintf("%v", durationDeleting),
			fmt.Sprintf("%f", float64(taskCount)/durationDeleting.Seconds()),
		},
	}
}

func deleteTasks(tasks <-chan *domain.Task, triggerHookService contracts.TasksDeferredInterface) {
	fmt.Println("\nDeleting task")
	preparingBar := pb.StartNew(len(tasks))

	wg := sync.WaitGroup{}
	for w := 0; w < 10; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range tasks {
				preparingBar.Add(1)
				if err := triggerHookService.Delete(task.Id); err != nil {
					log.Fatal(err)
				}
			}
		}()
	}

	wg.Wait()
	preparingBar.Finish()
}

func createTasks(
	triggerHookService contracts.TasksDeferredInterface,
	numberOfTask int,
	dispersion int,
) <-chan *domain.Task {
	createdTask := make(chan *domain.Task, numberOfTask)
	rand.Seed(time.Now().UnixNano())
	wg := sync.WaitGroup{}
	workers := 8
	numberOfTaskForWorker := numberOfTask / workers

	fmt.Println("\nCreating task")
	preparingBar := pb.StartNew(numberOfTask)

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < numberOfTaskForWorker; i++ {
				preparingBar.Add(1)
				task := &domain.Task{
					ExecTime: time.Now().Add(time.Hour + time.Duration(rand.Intn(dispersion))*time.Second).Unix(),
				}
				if err := triggerHookService.Create(task); err != nil {
					fmt.Println(err)
				}

				createdTask <- task
			}
		}()
	}

	wg.Wait()
	preparingBar.Finish()
	close(createdTask)

	return createdTask
}