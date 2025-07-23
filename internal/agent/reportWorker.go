package agent

import (
	"fmt"
	"github.com/Nikolay961996/metsys/models"
)

type workerJob struct {
	oneMetrics models.Metrics
}

func runReportWorkers(doneChan chan any, workersCount int, jobsIn <-chan workerJob, serverAddress string, keyForSigning string) {
	for i := 0; i < workersCount; i++ {
		go func(id int) {
			models.Log.Info(fmt.Sprintf("Worker %d started", id))
			for {
				select {
				case job := <-jobsIn:
					err := Report(job.oneMetrics, serverAddress, keyForSigning)
					if err != nil {
						models.Log.Error(err.Error())
					}
				case <-doneChan:
					models.Log.Warn(fmt.Sprintf("Worker %d stopped", id))
					return
				}
			}
		}(i)
	}
}
