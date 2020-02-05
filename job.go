package main

import (
	"fmt"
	"time"

	"github.com/ekas-data-portal/models"
	"github.com/hibiken/asynq"
)

// both in client.go and workers.go
var rediss = &asynq.RedisClientOpt{
	Addr: "db-redis-cluster-do-user-4666162-0.db.ondigitalocean.com:25061",
	// Omit if no password is required
	Password: "wdbsxehbizfl5kbu",
	// Use a dedicated db number for asynq.
	// By default, Redis offers 16 databases (0..15)
	DB: 2,
}

// JobSchedule ...
func JobSchedule(clientJobs chan models.ClientJob, asynqClient *asynq.Client) {

	for {
		// Wait for the next job to come off the queue.
		clientJob := <-clientJobs

		// Do something thats keeps the CPU busy for a whole second.
		// for start := time.Now(); time.Now().Sub(start) < time.Second; {
		LogToRedis(clientJob.DeviceData)
		// go SaveAllData(clientJob.DeviceData)

		t := asynq.NewTask(
			"save_to_mysql",
			map[string]interface{}{"data": clientJob.DeviceData})

		// Schedule the task t to be processed a minute from now.
		if err := asynqClient.Schedule(t, time.Now().Add(time.Second)); err != nil {
			fmt.Println("Failed to save to redis")
		}

	}
}
