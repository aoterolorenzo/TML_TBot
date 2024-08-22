package workers

import (
	"TML_TBot/application/interfaces"
	"TML_TBot/application/usecases"
	"TML_TBot/config"
	"TML_TBot/domain/models"
	"TML_TBot/infrastructure/connectors"
	"errors"
	"fmt"
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

	config.Log.Infof(fmt.Sprintf("Executing %s job", job.ID))

	for _, msg := range res {
		for _, target := range job.Response {
			if msg.MSG != "" || msg.Media != nil {
				config.Log.Infof("Sending response to telegram")

				if msg.UnpinAll {
					p.telegram.UnPinAll(target.ChatID, &target.TopicID)
				}

				switch msg.Kind {
				case models.KindMessage:
					err := p.telegram.SendMessage(msg.MSG, target.ChatID, &target.TopicID, msg.Pin)
					if err != nil {
						config.Log.Fatal(err)
						return
					}
					break
				case models.KindAnimation:
					err := p.telegram.SendAnimation(msg.MSG, msg.Media, target.ChatID, &target.TopicID, msg.Pin)
					if err != nil {
						config.Log.Fatal(err)
						return
					}
					break
				case models.KindMedia:
					err := p.telegram.SendMedia(msg.MSG, msg.Media, target.ChatID, &target.TopicID, msg.Pin)
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
	config.Log.Infof(fmt.Sprintf("Execution finished for %s job", job.ID))
}

// StartCronBot Reads the jobs in the settings and creates a cronjob entry for each one together with its execution function
func (p *Processor) StartCronBot() {
	for _, job := range config.Settings.Jobs {
		useCase := parseUseCase(job)
		if useCase != nil && job.CronString != "loop" {
			j := job

			err := p.cronWorker.AddToCron(j, func() { p.RunUseCase(j, useCase) })
			config.Log.Infof("%s job added to Cron: %s", job.ID, job.CronString)
			if err != nil {
				config.Log.Error(err)
			}
		} else if job.CronString == "loop" {
			go func() {
				j := job
				p.RunUseCase(j, useCase)
			}()
		} else {
			config.Log.Infof("Skipping  %s job", job.ID)
		}
	}
	p.cronWorker.Cron.Start()
	sleepForever()
}

func (p *Processor) RubJobById(jobID string) error {
	var foundJob models.Job
	for _, job := range config.Settings.Jobs {
		if job.ID == jobID {
			foundJob = job
		}
	}

	if foundJob.ID == "" {
		return errors.New(fmt.Sprintf("job %s not found in config", jobID))
	}

	config.Log.Infof("Executing %s job", foundJob.ID)

	useCase := parseUseCase(foundJob)

	p.RunUseCase(foundJob, useCase)
	return nil
}

func (p *Processor) RunAllJobs() error {
	for _, job := range config.Settings.Jobs {
		if job.CronString != "loop" {
			config.Log.Infof("Executing %s job", job.ID)
			useCase := parseUseCase(job)
			p.RunUseCase(job, useCase)
		}
	}
	return nil
}

// parseUseCase Parse a job.ID and returns its related use case
func parseUseCase(job models.Job) interfaces.UseCase {
	switch job.ID {
	case "weather":
		return &usecases.WeatherController{}
	case "lineUp":
		return usecases.NewTMLLineUpController()
	case "instagramPost":
		return usecases.NewInstagramPostsController()
	case "sumarize":
		return usecases.NewTMLSumarizerController(job)
	case "antiSpoilers":
		return usecases.NewTMLAntiSpoilersController(job)
	case "rain":
		return usecases.RainController{}
	}
	config.Log.Errorf("Unparseable %s job", job.ID)
	return nil
}

func sleepForever() {
	wait := make(chan struct{})
	for {
		<-wait
	}
}
