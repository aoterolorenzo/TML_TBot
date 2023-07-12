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
func (p *Processor) RunUseCase(job models.Job, useCase interfaces.UseCase) {

	res, err := useCase.Run()
	if err != nil {
		config.Log.Error()
	}

	for _, msg := range res {
		for _, target := range job.Response {
			if msg.MSG != "" || msg.Media != nil {
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
			} else {
				config.Log.Infof("Job finished successfully with no messages to send")
			}
		}

	}
}

// StartCronBot Reads the jobs in the settings and creates a cronjob entry for each one together with its execution function
func (p *Processor) StartCronBot() {
	for _, job := range config.Settings.Jobs {
		config.Log.Infof("Preparing %s job", job.ID)
		useCase := parseUseCase(job.ID)
		if useCase != nil {
			err := p.cronWorker.AddToCron(job, func() { p.RunUseCase(job, useCase) })
			if err != nil {
				config.Log.Error(err)
			}
		} else {
			config.Log.Infof("Skipping  %s job", job.ID)
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

	config.Log.Infof("Executing %s job", foundJob.ID)

	useCase := parseUseCase(foundJob.ID)

	p.RunUseCase(foundJob, useCase)
	return nil
}

// parseUseCase Parse a job.ID and returns its related use case
func parseUseCase(str string) interfaces.UseCase {
	switch str {
	case "weather":
		return &usecases.WeatherController{}
	case "lineUp":
		return usecases.NewTMLLineUpController()
	case "instagramPost":
		return usecases.NewInstagramPostsController()
	}
	config.Log.Errorf("Unparseable %s job", str)
	return nil
}
