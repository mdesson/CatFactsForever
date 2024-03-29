package scheduler

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// JobFunc will be the function run the job runner
type JobFunc func(context.Context) error

// Job to be run and all of its metadata
type Job struct {
	ID          string             // Job's uid
	Cron        string             // cron string. Supports special characters *,- only
	Description string             // Short description of job
	Active      bool               // inactive jobs are not run
	running     bool               // If job is currently running
	Rerun       bool               // Determines if jobs are rerun on next cycle following an error
	cancel      context.CancelFunc // For cancellation and timeouts
	ctx         context.Context    // Context for JobFunc
	mu          *sync.Mutex        // Only one call to Run() can hold the lock
	err         error              // error from last run, otherwise nil
	Job         JobFunc            // Actual job to be run
}

// Status returns user-friendly string with job's current status
func (j Job) Status() string {
	activeStatus := "inactive"
	if j.Active {
		activeStatus = "active"
	}
	if j.err == nil {
		return fmt.Sprintf("Job %v is %s with no error", j.ID, activeStatus)
	}
	return fmt.Sprintf("Job %v is %s with error:\n%v", j.ID, activeStatus, j.err)
}

// Cancel will stop the job's current execution and reset the context
// BUG: Context is not properly reset. It will stay cancelled
func (j *Job) Cancel() {
	j.cancel()
	ctx, cancel := context.WithCancel(context.Background())
	j.ctx = ctx
	j.cancel = cancel
}

// Run the job. Returns early if job is inactive or does not need to be run
// Updates error if job returns with error
func (j *Job) run() {
	// Ensure the task is active and not running
	j.mu.Lock()
	if !j.Active || j.running {
		j.mu.Unlock()
		return
	}
	j.running = true
	j.mu.Unlock()

	// Parse CRON tab, return early for invalid cron formats
	runnable, err := canRun(j.Cron)
	if err != nil {
		j.err = err
		return
	}

	// Run if runnable or error occurred on last run
	if runnable || (j.err != nil && j.Rerun) {
		j.err = j.Job(j.ctx)
	}
	j.running = false
}

// Parses crontab format and determines if it is time to run jobFunc
// Will return true if wait time has been exceeded
func canRun(cron string) (runnable bool, err error) {
	// Get cron values for current time
	now := time.Now()
	hour, min, _ := now.Clock()
	_, month, day := now.Date()
	weekday := now.Weekday()

	current := []int{min, hour, day, int(month), int(weekday)}

	// Split cron string into fields
	fields := strings.Split(cron, " ")

	// Ensure correct number of fields in cron string
	if len(fields) != 5 {
		return false, fmt.Errorf("Invalid length of cron string")
	}

	for i, field := range fields {
		runnable, ok := cronFieldCheck(field, current[i])
		if !ok {
			return false, fmt.Errorf("Invalid cron format. This application accepts only integers and characters '*' ',' and '-'")
		}
		if !runnable {
			return false, nil
		}
	}
	return true, nil
}

// Check if cron field is runnable
func cronFieldCheck(input string, compare int) (runnable bool, ok bool) {
	// return true if input is "*"
	if input == "*" {
		return true, true
	}

	// Validate if cron field is a list of values
	if strings.Contains(input, ",") {
		for _, val := range strings.Split(input, ",") {
			intVal, err := strconv.Atoi(val)
			if err != nil {
				return false, false
			}
			if intVal == compare {
				return true, true
			}
		}
		return false, true
	}

	// Validate if cron field is a range of values
	if strings.Contains(input, "-") {
		fieldRange := strings.Split(input, "-")
		if len(fieldRange) != 2 {
			return false, false
		}
		start, err := strconv.Atoi(fieldRange[0])
		if err != nil {
			return false, false
		}
		end, err := strconv.Atoi(fieldRange[1])
		if err != nil {
			return false, false
		}
		if start > end {
			return false, false
		}
		if compare >= start && compare <= end {
			return true, true
		}
		return false, true
	}
	inputVal, err := strconv.Atoi(input)
	if err != nil {
		return false, false
	}
	if inputVal == compare {
		return true, true
	}
	return false, true
}
