package worker

import (
	"log"

	"github.com/JulianDominic/GatheringTheBulk/internal/models"
	"github.com/JulianDominic/GatheringTheBulk/internal/store"
)

type JobRequest struct {
	Job     *models.Job
	Handler func(store store.Store, job *models.Job) (string, error)
}

type Dispatcher struct {
	store    store.Store
	jobQueue chan JobRequest
}

func NewDispatcher(s store.Store, buffer int) *Dispatcher {
	return &Dispatcher{
		store:    s,
		jobQueue: make(chan JobRequest, buffer),
	}
}

func (d *Dispatcher) Start(numWorkers int) {
	log.Printf("Starting worker pool with %d workers", numWorkers)
	for i := 0; i < numWorkers; i++ {
		go d.worker(i)
	}
}

func (d *Dispatcher) QueueJob(req JobRequest) {
	d.jobQueue <- req
}

func (d *Dispatcher) worker(id int) {
	for req := range d.jobQueue {
		log.Printf("[Worker %d] Starting Job %s (%s)", id, req.Job.ID, req.Job.Type)

		if err := d.store.UpdateJobStatus(req.Job.ID, models.JobStatusProcessing); err != nil {
			log.Printf("Failed to update status for job %s: %v", req.Job.ID, err)
			continue
		}

		summary, err := req.Handler(d.store, req.Job)

		if err != nil {
			log.Printf("[Worker %d] Job %s Failed: %v", id, req.Job.ID, err)
			d.store.FailJob(req.Job.ID, err.Error())
		} else {
			log.Printf("[Worker %d] Job %s Completed", id, req.Job.ID)
			d.store.CompleteJob(req.Job.ID, summary)
		}
	}
}
