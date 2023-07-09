package workers

import (
	"TML_TBot/domain/models"
	"github.com/robfig/cron/v3"
)

type CronWorker struct {
	Cron cron.Cron
}

func NewCronWorker() CronWorker {
	return CronWorker{
		Cron: *cron.New(),
	}
}

func (cw *CronWorker) AddToCron(job models.Job, usecase func()) error {
	_, err := cw.Cron.AddFunc(job.CronString, usecase)
	if err != nil {
		return err
	}
	return nil
}
