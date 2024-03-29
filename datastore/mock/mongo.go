// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock

import (
	"context"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-import-api/datastore"
	"github.com/ONSdigital/dp-import-api/models"
	"sync"
)

// Ensure, that DataStorerMock does implement datastore.DataStorer.
// If this is not the case, regenerate this file with moq.
var _ datastore.DataStorer = &DataStorerMock{}

// DataStorerMock is a mock implementation of datastore.DataStorer.
//
// 	func TestSomethingThatUsesDataStorer(t *testing.T) {
//
// 		// make and configure a mocked datastore.DataStorer
// 		mockedDataStorer := &DataStorerMock{
// 			AcquireInstanceLockFunc: func(ctx context.Context, jobID string) (string, error) {
// 				panic("mock out the AcquireInstanceLock method")
// 			},
// 			AddJobFunc: func(ctx context.Context, importJob *models.Job) (*models.Job, error) {
// 				panic("mock out the AddJob method")
// 			},
// 			AddUploadedFileFunc: func(ctx context.Context, jobID string, message *models.UploadedFile) error {
// 				panic("mock out the AddUploadedFile method")
// 			},
// 			CheckerFunc: func(contextMoqParam context.Context, checkState *healthcheck.CheckState) error {
// 				panic("mock out the Checker method")
// 			},
// 			CloseFunc: func(contextMoqParam context.Context) error {
// 				panic("mock out the Close method")
// 			},
// 			GetJobFunc: func(ctx context.Context, jobID string) (*models.Job, error) {
// 				panic("mock out the GetJob method")
// 			},
// 			GetJobsFunc: func(ctx context.Context, filters []string, offset int, limit int) (*models.JobResults, error) {
// 				panic("mock out the GetJobs method")
// 			},
// 			UnlockInstanceFunc: func(ctx context.Context, lockID string)  {
// 				panic("mock out the UnlockInstance method")
// 			},
// 			UpdateJobFunc: func(ctx context.Context, jobID string, update *models.Job) error {
// 				panic("mock out the UpdateJob method")
// 			},
// 			UpdateProcessedInstanceFunc: func(ctx context.Context, id string, procInstances []models.ProcessedInstances) error {
// 				panic("mock out the UpdateProcessedInstance method")
// 			},
// 		}
//
// 		// use mockedDataStorer in code that requires datastore.DataStorer
// 		// and then make assertions.
//
// 	}
type DataStorerMock struct {
	// AcquireInstanceLockFunc mocks the AcquireInstanceLock method.
	AcquireInstanceLockFunc func(ctx context.Context, jobID string) (string, error)

	// AddJobFunc mocks the AddJob method.
	AddJobFunc func(ctx context.Context, importJob *models.Job) (*models.Job, error)

	// AddUploadedFileFunc mocks the AddUploadedFile method.
	AddUploadedFileFunc func(ctx context.Context, jobID string, message *models.UploadedFile) error

	// CheckerFunc mocks the Checker method.
	CheckerFunc func(contextMoqParam context.Context, checkState *healthcheck.CheckState) error

	// CloseFunc mocks the Close method.
	CloseFunc func(contextMoqParam context.Context) error

	// GetJobFunc mocks the GetJob method.
	GetJobFunc func(ctx context.Context, jobID string) (*models.Job, error)

	// GetJobsFunc mocks the GetJobs method.
	GetJobsFunc func(ctx context.Context, filters []string, offset int, limit int) (*models.JobResults, error)

	// UnlockInstanceFunc mocks the UnlockInstance method.
	UnlockInstanceFunc func(ctx context.Context, lockID string)

	// UpdateJobFunc mocks the UpdateJob method.
	UpdateJobFunc func(ctx context.Context, jobID string, update *models.Job) error

	// UpdateProcessedInstanceFunc mocks the UpdateProcessedInstance method.
	UpdateProcessedInstanceFunc func(ctx context.Context, id string, procInstances []models.ProcessedInstances) error

	// calls tracks calls to the methods.
	calls struct {
		// AcquireInstanceLock holds details about calls to the AcquireInstanceLock method.
		AcquireInstanceLock []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// JobID is the jobID argument value.
			JobID string
		}
		// AddJob holds details about calls to the AddJob method.
		AddJob []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// ImportJob is the importJob argument value.
			ImportJob *models.Job
		}
		// AddUploadedFile holds details about calls to the AddUploadedFile method.
		AddUploadedFile []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// JobID is the jobID argument value.
			JobID string
			// Message is the message argument value.
			Message *models.UploadedFile
		}
		// Checker holds details about calls to the Checker method.
		Checker []struct {
			// ContextMoqParam is the contextMoqParam argument value.
			ContextMoqParam context.Context
			// CheckState is the checkState argument value.
			CheckState *healthcheck.CheckState
		}
		// Close holds details about calls to the Close method.
		Close []struct {
			// ContextMoqParam is the contextMoqParam argument value.
			ContextMoqParam context.Context
		}
		// GetJob holds details about calls to the GetJob method.
		GetJob []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// JobID is the jobID argument value.
			JobID string
		}
		// GetJobs holds details about calls to the GetJobs method.
		GetJobs []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Filters is the filters argument value.
			Filters []string
			// Offset is the offset argument value.
			Offset int
			// Limit is the limit argument value.
			Limit int
		}
		// UnlockInstance holds details about calls to the UnlockInstance method.
		UnlockInstance []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// LockID is the lockID argument value.
			LockID string
		}
		// UpdateJob holds details about calls to the UpdateJob method.
		UpdateJob []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// JobID is the jobID argument value.
			JobID string
			// Update is the update argument value.
			Update *models.Job
		}
		// UpdateProcessedInstance holds details about calls to the UpdateProcessedInstance method.
		UpdateProcessedInstance []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// ID is the id argument value.
			ID string
			// ProcInstances is the procInstances argument value.
			ProcInstances []models.ProcessedInstances
		}
	}
	lockAcquireInstanceLock     sync.RWMutex
	lockAddJob                  sync.RWMutex
	lockAddUploadedFile         sync.RWMutex
	lockChecker                 sync.RWMutex
	lockClose                   sync.RWMutex
	lockGetJob                  sync.RWMutex
	lockGetJobs                 sync.RWMutex
	lockUnlockInstance          sync.RWMutex
	lockUpdateJob               sync.RWMutex
	lockUpdateProcessedInstance sync.RWMutex
}

// AcquireInstanceLock calls AcquireInstanceLockFunc.
func (mock *DataStorerMock) AcquireInstanceLock(ctx context.Context, jobID string) (string, error) {
	if mock.AcquireInstanceLockFunc == nil {
		panic("DataStorerMock.AcquireInstanceLockFunc: method is nil but DataStorer.AcquireInstanceLock was just called")
	}
	callInfo := struct {
		Ctx   context.Context
		JobID string
	}{
		Ctx:   ctx,
		JobID: jobID,
	}
	mock.lockAcquireInstanceLock.Lock()
	mock.calls.AcquireInstanceLock = append(mock.calls.AcquireInstanceLock, callInfo)
	mock.lockAcquireInstanceLock.Unlock()
	return mock.AcquireInstanceLockFunc(ctx, jobID)
}

// AcquireInstanceLockCalls gets all the calls that were made to AcquireInstanceLock.
// Check the length with:
//     len(mockedDataStorer.AcquireInstanceLockCalls())
func (mock *DataStorerMock) AcquireInstanceLockCalls() []struct {
	Ctx   context.Context
	JobID string
} {
	var calls []struct {
		Ctx   context.Context
		JobID string
	}
	mock.lockAcquireInstanceLock.RLock()
	calls = mock.calls.AcquireInstanceLock
	mock.lockAcquireInstanceLock.RUnlock()
	return calls
}

// AddJob calls AddJobFunc.
func (mock *DataStorerMock) AddJob(ctx context.Context, importJob *models.Job) (*models.Job, error) {
	if mock.AddJobFunc == nil {
		panic("DataStorerMock.AddJobFunc: method is nil but DataStorer.AddJob was just called")
	}
	callInfo := struct {
		Ctx       context.Context
		ImportJob *models.Job
	}{
		Ctx:       ctx,
		ImportJob: importJob,
	}
	mock.lockAddJob.Lock()
	mock.calls.AddJob = append(mock.calls.AddJob, callInfo)
	mock.lockAddJob.Unlock()
	return mock.AddJobFunc(ctx, importJob)
}

// AddJobCalls gets all the calls that were made to AddJob.
// Check the length with:
//     len(mockedDataStorer.AddJobCalls())
func (mock *DataStorerMock) AddJobCalls() []struct {
	Ctx       context.Context
	ImportJob *models.Job
} {
	var calls []struct {
		Ctx       context.Context
		ImportJob *models.Job
	}
	mock.lockAddJob.RLock()
	calls = mock.calls.AddJob
	mock.lockAddJob.RUnlock()
	return calls
}

// AddUploadedFile calls AddUploadedFileFunc.
func (mock *DataStorerMock) AddUploadedFile(ctx context.Context, jobID string, message *models.UploadedFile) error {
	if mock.AddUploadedFileFunc == nil {
		panic("DataStorerMock.AddUploadedFileFunc: method is nil but DataStorer.AddUploadedFile was just called")
	}
	callInfo := struct {
		Ctx     context.Context
		JobID   string
		Message *models.UploadedFile
	}{
		Ctx:     ctx,
		JobID:   jobID,
		Message: message,
	}
	mock.lockAddUploadedFile.Lock()
	mock.calls.AddUploadedFile = append(mock.calls.AddUploadedFile, callInfo)
	mock.lockAddUploadedFile.Unlock()
	return mock.AddUploadedFileFunc(ctx, jobID, message)
}

// AddUploadedFileCalls gets all the calls that were made to AddUploadedFile.
// Check the length with:
//     len(mockedDataStorer.AddUploadedFileCalls())
func (mock *DataStorerMock) AddUploadedFileCalls() []struct {
	Ctx     context.Context
	JobID   string
	Message *models.UploadedFile
} {
	var calls []struct {
		Ctx     context.Context
		JobID   string
		Message *models.UploadedFile
	}
	mock.lockAddUploadedFile.RLock()
	calls = mock.calls.AddUploadedFile
	mock.lockAddUploadedFile.RUnlock()
	return calls
}

// Checker calls CheckerFunc.
func (mock *DataStorerMock) Checker(contextMoqParam context.Context, checkState *healthcheck.CheckState) error {
	if mock.CheckerFunc == nil {
		panic("DataStorerMock.CheckerFunc: method is nil but DataStorer.Checker was just called")
	}
	callInfo := struct {
		ContextMoqParam context.Context
		CheckState      *healthcheck.CheckState
	}{
		ContextMoqParam: contextMoqParam,
		CheckState:      checkState,
	}
	mock.lockChecker.Lock()
	mock.calls.Checker = append(mock.calls.Checker, callInfo)
	mock.lockChecker.Unlock()
	return mock.CheckerFunc(contextMoqParam, checkState)
}

// CheckerCalls gets all the calls that were made to Checker.
// Check the length with:
//     len(mockedDataStorer.CheckerCalls())
func (mock *DataStorerMock) CheckerCalls() []struct {
	ContextMoqParam context.Context
	CheckState      *healthcheck.CheckState
} {
	var calls []struct {
		ContextMoqParam context.Context
		CheckState      *healthcheck.CheckState
	}
	mock.lockChecker.RLock()
	calls = mock.calls.Checker
	mock.lockChecker.RUnlock()
	return calls
}

// Close calls CloseFunc.
func (mock *DataStorerMock) Close(contextMoqParam context.Context) error {
	if mock.CloseFunc == nil {
		panic("DataStorerMock.CloseFunc: method is nil but DataStorer.Close was just called")
	}
	callInfo := struct {
		ContextMoqParam context.Context
	}{
		ContextMoqParam: contextMoqParam,
	}
	mock.lockClose.Lock()
	mock.calls.Close = append(mock.calls.Close, callInfo)
	mock.lockClose.Unlock()
	return mock.CloseFunc(contextMoqParam)
}

// CloseCalls gets all the calls that were made to Close.
// Check the length with:
//     len(mockedDataStorer.CloseCalls())
func (mock *DataStorerMock) CloseCalls() []struct {
	ContextMoqParam context.Context
} {
	var calls []struct {
		ContextMoqParam context.Context
	}
	mock.lockClose.RLock()
	calls = mock.calls.Close
	mock.lockClose.RUnlock()
	return calls
}

// GetJob calls GetJobFunc.
func (mock *DataStorerMock) GetJob(ctx context.Context, jobID string) (*models.Job, error) {
	if mock.GetJobFunc == nil {
		panic("DataStorerMock.GetJobFunc: method is nil but DataStorer.GetJob was just called")
	}
	callInfo := struct {
		Ctx   context.Context
		JobID string
	}{
		Ctx:   ctx,
		JobID: jobID,
	}
	mock.lockGetJob.Lock()
	mock.calls.GetJob = append(mock.calls.GetJob, callInfo)
	mock.lockGetJob.Unlock()
	return mock.GetJobFunc(ctx, jobID)
}

// GetJobCalls gets all the calls that were made to GetJob.
// Check the length with:
//     len(mockedDataStorer.GetJobCalls())
func (mock *DataStorerMock) GetJobCalls() []struct {
	Ctx   context.Context
	JobID string
} {
	var calls []struct {
		Ctx   context.Context
		JobID string
	}
	mock.lockGetJob.RLock()
	calls = mock.calls.GetJob
	mock.lockGetJob.RUnlock()
	return calls
}

// GetJobs calls GetJobsFunc.
func (mock *DataStorerMock) GetJobs(ctx context.Context, filters []string, offset int, limit int) (*models.JobResults, error) {
	if mock.GetJobsFunc == nil {
		panic("DataStorerMock.GetJobsFunc: method is nil but DataStorer.GetJobs was just called")
	}
	callInfo := struct {
		Ctx     context.Context
		Filters []string
		Offset  int
		Limit   int
	}{
		Ctx:     ctx,
		Filters: filters,
		Offset:  offset,
		Limit:   limit,
	}
	mock.lockGetJobs.Lock()
	mock.calls.GetJobs = append(mock.calls.GetJobs, callInfo)
	mock.lockGetJobs.Unlock()
	return mock.GetJobsFunc(ctx, filters, offset, limit)
}

// GetJobsCalls gets all the calls that were made to GetJobs.
// Check the length with:
//     len(mockedDataStorer.GetJobsCalls())
func (mock *DataStorerMock) GetJobsCalls() []struct {
	Ctx     context.Context
	Filters []string
	Offset  int
	Limit   int
} {
	var calls []struct {
		Ctx     context.Context
		Filters []string
		Offset  int
		Limit   int
	}
	mock.lockGetJobs.RLock()
	calls = mock.calls.GetJobs
	mock.lockGetJobs.RUnlock()
	return calls
}

// UnlockInstance calls UnlockInstanceFunc.
func (mock *DataStorerMock) UnlockInstance(ctx context.Context, lockID string) {
	if mock.UnlockInstanceFunc == nil {
		panic("DataStorerMock.UnlockInstanceFunc: method is nil but DataStorer.UnlockInstance was just called")
	}
	callInfo := struct {
		Ctx    context.Context
		LockID string
	}{
		Ctx:    ctx,
		LockID: lockID,
	}
	mock.lockUnlockInstance.Lock()
	mock.calls.UnlockInstance = append(mock.calls.UnlockInstance, callInfo)
	mock.lockUnlockInstance.Unlock()
	mock.UnlockInstanceFunc(ctx, lockID)
}

// UnlockInstanceCalls gets all the calls that were made to UnlockInstance.
// Check the length with:
//     len(mockedDataStorer.UnlockInstanceCalls())
func (mock *DataStorerMock) UnlockInstanceCalls() []struct {
	Ctx    context.Context
	LockID string
} {
	var calls []struct {
		Ctx    context.Context
		LockID string
	}
	mock.lockUnlockInstance.RLock()
	calls = mock.calls.UnlockInstance
	mock.lockUnlockInstance.RUnlock()
	return calls
}

// UpdateJob calls UpdateJobFunc.
func (mock *DataStorerMock) UpdateJob(ctx context.Context, jobID string, update *models.Job) error {
	if mock.UpdateJobFunc == nil {
		panic("DataStorerMock.UpdateJobFunc: method is nil but DataStorer.UpdateJob was just called")
	}
	callInfo := struct {
		Ctx    context.Context
		JobID  string
		Update *models.Job
	}{
		Ctx:    ctx,
		JobID:  jobID,
		Update: update,
	}
	mock.lockUpdateJob.Lock()
	mock.calls.UpdateJob = append(mock.calls.UpdateJob, callInfo)
	mock.lockUpdateJob.Unlock()
	return mock.UpdateJobFunc(ctx, jobID, update)
}

// UpdateJobCalls gets all the calls that were made to UpdateJob.
// Check the length with:
//     len(mockedDataStorer.UpdateJobCalls())
func (mock *DataStorerMock) UpdateJobCalls() []struct {
	Ctx    context.Context
	JobID  string
	Update *models.Job
} {
	var calls []struct {
		Ctx    context.Context
		JobID  string
		Update *models.Job
	}
	mock.lockUpdateJob.RLock()
	calls = mock.calls.UpdateJob
	mock.lockUpdateJob.RUnlock()
	return calls
}

// UpdateProcessedInstance calls UpdateProcessedInstanceFunc.
func (mock *DataStorerMock) UpdateProcessedInstance(ctx context.Context, id string, procInstances []models.ProcessedInstances) error {
	if mock.UpdateProcessedInstanceFunc == nil {
		panic("DataStorerMock.UpdateProcessedInstanceFunc: method is nil but DataStorer.UpdateProcessedInstance was just called")
	}
	callInfo := struct {
		Ctx           context.Context
		ID            string
		ProcInstances []models.ProcessedInstances
	}{
		Ctx:           ctx,
		ID:            id,
		ProcInstances: procInstances,
	}
	mock.lockUpdateProcessedInstance.Lock()
	mock.calls.UpdateProcessedInstance = append(mock.calls.UpdateProcessedInstance, callInfo)
	mock.lockUpdateProcessedInstance.Unlock()
	return mock.UpdateProcessedInstanceFunc(ctx, id, procInstances)
}

// UpdateProcessedInstanceCalls gets all the calls that were made to UpdateProcessedInstance.
// Check the length with:
//     len(mockedDataStorer.UpdateProcessedInstanceCalls())
func (mock *DataStorerMock) UpdateProcessedInstanceCalls() []struct {
	Ctx           context.Context
	ID            string
	ProcInstances []models.ProcessedInstances
} {
	var calls []struct {
		Ctx           context.Context
		ID            string
		ProcInstances []models.ProcessedInstances
	}
	mock.lockUpdateProcessedInstance.RLock()
	calls = mock.calls.UpdateProcessedInstance
	mock.lockUpdateProcessedInstance.RUnlock()
	return calls
}
