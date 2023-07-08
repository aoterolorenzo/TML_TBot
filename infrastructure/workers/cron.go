package workers

import (
	"TML_TBot/application/interfaces"
	"TML_TBot/application/usecases"
	"TML_TBot/domain/models"
	"github.com/robfig/cron/v3"
)

type CronWorker struct {
	cron      cron.Cron
	processor Processor
}

func NewCronWorker() *CronWorker {
	return &CronWorker{
		cron:      *cron.New(),
		processor: Processor{},
	}
}

func (cw *CronWorker) AddToCron(job models.Job) error {
	useCase := parseUseCase(job.ID)
	_, err := cw.cron.AddFunc(job.CronString, func() { cw.processor.RunUseCase(job, useCase) })
	if err != nil {
		return err
	}
	return nil
}

func parseUseCase(str string) interfaces.UseCase {
	switch str {
	case "weather":
		return &usecases.WeatherController{}
	}
	return nil
}
