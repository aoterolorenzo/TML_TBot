package workers

import (
	"TML_TBot/application/interfaces"
	"TML_TBot/application/usecases"
	"TML_TBot/config"
	"TML_TBot/domain/models"
	"TML_TBot/infrastructure/connectors"
	"errors"
)

type Processor struct {
	cronWorker *CronWorker
	telegram   *connectors.TelegramService
}

func NewProcessor(cron *CronWorker, telegram *connectors.TelegramService) *Processor {
	return &Processor{cronWorker: cron, telegram: telegram}
}

// RunUseCase runs the use case of a job, preparing and executing the proper response based on job settings
func (p *Processor) RunUseCase(job models.Job) {

	useCase := parseUseCase(job.ID)
	res, err := useCase.Run()
	if err != nil {
		config.Log.Error()
	}

	for _, msg := range res {
		for _, target := range job.Response {

			switch msg.Kind {
			case models.KindMessage:
				err := p.telegram.SendMessage(msg.MSG, target.ChatID, &target.TopicID)
				if err != nil {
					config.Log.Fatal(err)
					return
				}
				break
			case models.KindAnimation:
				err := p.telegram.SendAnimation(msg.MSG, msg.Media, target.ChatID, &target.TopicID)
				if err != nil {
					config.Log.Fatal(err)
					return
				}
				break
			case models.KindMedia:
				err := p.telegram.SendMedia(msg.MSG, msg.Media, target.ChatID, &target.TopicID)
				if err != nil {
					config.Log.Fatal(err)
					return
				}
				break

			}
		}

	}
}

// StartCronBot Reads the jobs in the settings and creates a cronjob entry for each one together with its execution function
func (p *Processor) StartCronBot() {
	for _, job := range config.Settings.Jobs {
		err := p.cronWorker.AddToCron(job, func() { p.RunUseCase(job) })
		if err != nil {
			config.Log.Error(err)
		}
	}
	p.cronWorker.Cron.Start()
}

func (p *Processor) RubJobById(jobID string) error {
	var foundJob models.Job
	for _, job := range config.Settings.Jobs {
		if job.ID == jobID {
			foundJob = job
		}
	}

	if foundJob.ID == "" {
		return errors.New("job not found in config")
	}

	p.RunUseCase(foundJob)
	return nil
}

// parseUseCase Parse a job.ID and returns its related use case
func parseUseCase(str string) interfaces.UseCase {
	switch str {
	case "weather":
		return &usecases.WeatherController{}
	case "lineUp":
		return &usecases.TMLLineUpController{}
	}
	config.Log.Fatal("unparseable job")
	return nil
}
