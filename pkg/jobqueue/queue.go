package jobqueue

import (
	"sync"
	"time"

	"github.com/davidsonmarra/receitas-app/pkg/log"
)

// JobStatus representa o status de um job
type JobStatus string

const (
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
)

// JobResult representa o resultado de um job assíncrono
type JobResult struct {
	JobID   string      `json:"job_id"`
	Status  JobStatus   `json:"status"`
	Result  interface{} `json:"result,omitempty"`
	Error   string      `json:"error,omitempty"`
	Created time.Time   `json:"created_at"`
}

// JobQueue gerencia jobs assíncronos em memória
type JobQueue struct {
	jobs  map[string]*JobResult
	mutex sync.RWMutex
}

// GlobalQueue é a instância global da fila de jobs
var GlobalQueue *JobQueue

func init() {
	GlobalQueue = NewJobQueue()
	go GlobalQueue.CleanupOldJobs()
}

// NewJobQueue cria uma nova fila de jobs
func NewJobQueue() *JobQueue {
	return &JobQueue{
		jobs: make(map[string]*JobResult),
	}
}

// CreateJob cria um novo job com status "processing"
func (q *JobQueue) CreateJob(jobID string) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	q.jobs[jobID] = &JobResult{
		JobID:   jobID,
		Status:  JobStatusProcessing,
		Created: time.Now(),
	}

	log.Info("job criado", "job_id", jobID)
}

// GetJob busca um job por ID (thread-safe)
func (q *JobQueue) GetJob(jobID string) (*JobResult, bool) {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	job, exists := q.jobs[jobID]
	return job, exists
}

// CompleteJob marca um job como completado e salva o resultado
func (q *JobQueue) CompleteJob(jobID string, result interface{}) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if job, exists := q.jobs[jobID]; exists {
		job.Status = JobStatusCompleted
		job.Result = result
		log.Info("job completado", "job_id", jobID)
	}
}

// FailJob marca um job como falhou e salva a mensagem de erro
func (q *JobQueue) FailJob(jobID string, errorMsg string) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if job, exists := q.jobs[jobID]; exists {
		job.Status = JobStatusFailed
		job.Error = errorMsg
		log.Warn("job falhou", "job_id", jobID, "error", errorMsg)
	}
}

// CleanupOldJobs remove jobs com mais de 30 minutos (loop infinito)
func (q *JobQueue) CleanupOldJobs() {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		q.mutex.Lock()
		now := time.Now()
		removedCount := 0

		for jobID, job := range q.jobs {
			if now.Sub(job.Created) > 30*time.Minute {
				delete(q.jobs, jobID)
				removedCount++
			}
		}

		q.mutex.Unlock()

		if removedCount > 0 {
			log.Info("jobs antigos removidos", "count", removedCount)
		}
	}
}

// GetJobCount retorna o número de jobs na fila (útil para debug/testes)
func (q *JobQueue) GetJobCount() int {
	q.mutex.RLock()
	defer q.mutex.RUnlock()
	return len(q.jobs)
}

