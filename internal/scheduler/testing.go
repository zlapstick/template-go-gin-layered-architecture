package scheduler

import "log"

type TestingJob struct {
}

func NewTestingJob() *TestingJob {
	return &TestingJob{}
}

func (j *TestingJob) Run() {
	log.Printf("testing")
}
