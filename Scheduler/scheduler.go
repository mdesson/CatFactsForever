package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"
)

var jobs = make(map[string]Job)
var stop = make(chan bool)

// Start begins running cron jobs
// Recommended to run as a goroutine in main with a deferred Stop()
func Start() {
	timer := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-stop:
			return
		case <-timer.C:
			for _, job := range jobs {
				go job.run()
			}
		}

	}
}

// Stop will halt the job runner
func Stop() {
	stop <- true
}

// Clear will empty the job store
func Clear() {
	jobs = make(map[string]Job)
}

// AddJob generates a new Job struct and adds it to the key-value store of jobs
func AddJob(id, cron, desc string, active, rerun bool, jobFunc JobFunc) error {
	if _, ok := jobs[id]; ok {
		return fmt.Errorf("Job store already contains job with key %v", id)
	}
	ctx, cancel := context.WithCancel(context.Background())
	mu := sync.Mutex{}
	job := Job{
		ID:          id,
		Cron:        cron,
		Description: desc,
		Active:      active,
		running:     false,
		Rerun:       rerun,
		cancel:      cancel,
		ctx:         ctx,
		mu:          &mu,
		err:         nil,
		Job:         jobFunc,
	}

	jobs[id] = job

	return nil
}

// RemoveJob removes a job from the job map
func RemoveJob(id string) bool {
	if _, ok := jobs[id]; ok {
		delete(jobs, id)
		return true
	}
	return false
}

// FindJob returns a pointer to the job if ok
func FindJob(id string) (*Job, bool) {
	if job, ok := jobs[id]; ok {
		return &job, true
	}
	return nil, false
}

// ListJobs returns a slice all job ids in the job store
func ListJobs() []string {
	jobIDs := make([]string, 0)
	for id := range jobs {
		jobIDs = append(jobIDs, id)
	}
	return jobIDs
}
