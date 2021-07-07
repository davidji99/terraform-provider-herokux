package scheduler

import (
	"fmt"
	"github.com/davidji99/simpleresty"
	"time"
)

// Jobs represents all jobs on an application.
type Jobs struct {
	Data []*Job `json:"data,omitempty"`
}

// JobModifyResponse represents the response returned from creating or updating a Job.
type JobModifyResponse struct {
	Data *Job
}

// Job represents an individual job.
type Job struct {
	Type       *string        `json:"type,omitempty"`
	ID         *string        `json:"id,omitempty"`
	Attributes *JobAttributes `json:"attributes,omitempty"`
}

// JobAttributes represents the attributes for a job.
type JobAttributes struct {
	Every       *int       `json:"every,omitempty"`
	At          *int       `json:"at,omitempty"`
	DynoSize    *string    `json:"dyno-size,omitempty"`
	Command     *string    `json:"command,omitempty"`
	RanAt       *time.Time `json:"ran-at,omitempty"`
	Pause       *time.Time `json:"schemaVersion,omitempty"`
	ExpireAfter *time.Time `json:"expire-after,omitempty"`
	CreatedAt   *time.Time `json:"created-at,omitempty"`
}

// JobRequest represents a request to create/update a job.
type JobRequest struct {
	Command  string `json:"command,omitempty"`
	DynoSize string `json:"dyno-size,omitempty"`
	Every    int    `json:"every,omitempty"`
	At       int    `json:"at,omitempty"`
}

type JobDataRequest struct {
	Attributes *JobAttributeRequest `json:"attributes"`
}

type JobAttributeRequest struct {
	Command  string `json:"command,omitempty"`
	DynoSize string `json:"dyno-size,omitempty"`
	Every    int    `json:"every,omitempty"`
	At       int    `json:"at,omitempty"`
}

type JobRequestNew struct {
	Data *JobDataRequest `json:"data"`
}

// List returns all jobs.
func (s *Scheduler) List(appID string) (*Jobs, *simpleresty.Response, error) {
	var result Jobs
	urlStr := s.http.RequestURL("/apps/%s/jobs", appID)

	// Execute the request
	response, getErr := s.http.Get(urlStr, &result, nil)

	return &result, response, getErr
}

// FindByID retrieves a job by its ID.
//
// Unfortunately, there is no `GET` endpoint to retrieve one job.
func (s *Scheduler) FindByID(appID, jobID string) (*Job, *simpleresty.Response, error) {
	jobs, listResponse, listErr := s.List(appID)
	if listErr != nil {
		return nil, listResponse, listErr
	}

	if !jobs.HasData() {
		return nil, nil, fmt.Errorf("no jobs exist for app %s", appID)
	}

	for _, j := range jobs.Data {
		if j.GetID() == jobID {
			return j, nil, nil
		}
	}

	return nil, nil, fmt.Errorf("job %s not found on app %s", appID, jobID)
}

// Create a job.
func (s *Scheduler) Create(appID string, opts *JobRequest) (*JobModifyResponse, *simpleresty.Response, error) {
	var result JobModifyResponse
	urlStr := s.http.RequestURL("/apps/%s/jobs", appID)

	body := struct {
		Data struct {
			Attributes *JobRequest `json:"attributes"`
		} `json:"data"`
	}{
		Data: struct {
			Attributes *JobRequest `json:"attributes"`
		}(struct{ Attributes *JobRequest }{Attributes: opts}),
	}

	// Execute the request
	response, createErr := s.http.Post(urlStr, &result, &body)

	return &result, response, createErr
}

// Update a job.
func (s *Scheduler) Update(appID, jobID string, opts *JobRequest) (*JobModifyResponse, *simpleresty.Response, error) {
	var result JobModifyResponse
	urlStr := s.http.RequestURL("/apps/%s/jobs/%s", appID, jobID)

	body := struct {
		Data struct {
			Attributes *JobRequest `json:"attributes"`
		} `json:"data"`
	}{
		Data: struct {
			Attributes *JobRequest `json:"attributes"`
		}(struct{ Attributes *JobRequest }{Attributes: opts}),
	}

	// Execute the request
	response, updateErr := s.http.Patch(urlStr, &result, &body)

	return &result, response, updateErr
}

// Delete a job.
func (s *Scheduler) Delete(appID, jobID string) (*simpleresty.Response, error) {
	urlStr := s.http.RequestURL("/apps/%s/jobs/%s", appID, jobID)

	// Execute the request
	response, deleteErr := s.http.Delete(urlStr, nil, nil)

	return response, deleteErr
}
