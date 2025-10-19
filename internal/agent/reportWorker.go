package agent

import (
	"crypto/rsa"
	"fmt"

	"github.com/Nikolay961996/metsys/models"
)

type workerJob struct {
	oneMetrics models.Metrics
}

func runReportWorker(id int, jobsIn <-chan workerJob, serverAddress string, keyForSigning string, publicKey *rsa.PublicKey) {
	models.Log.Info(fmt.Sprintf("Worker %d started", id))
	for job := range jobsIn {
		err := Report(job.oneMetrics, serverAddress, keyForSigning, publicKey)
		if err != nil {
			models.Log.Error(fmt.Sprintf("%d on worker: %s", id, err.Error()))
		}
	}
	models.Log.Warn(fmt.Sprintf("Worker %d stopped", id))
}
