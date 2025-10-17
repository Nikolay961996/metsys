package agent

import (
	"context"
	"crypto/rsa"
	"fmt"

	"github.com/Nikolay961996/metsys/models"
)

type workerJob struct {
	oneMetrics models.Metrics
}

func runReportWorker(id int, doneCtx context.Context, jobsIn <-chan workerJob, serverAddress string, keyForSigning string, publicKey *rsa.PublicKey) {
	models.Log.Info(fmt.Sprintf("Worker %d started", id))
	for {
		select {
		case job := <-jobsIn:
			err := Report(job.oneMetrics, serverAddress, keyForSigning, publicKey)
			if err != nil {
				models.Log.Error(fmt.Sprintf("%d on worker: %s", id, err.Error()))
			}
		case <-doneCtx.Done():
			models.Log.Warn(fmt.Sprintf("Worker %d stopped", id))
			return
		}
	}
}
