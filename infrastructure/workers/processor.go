package workers

import (
	"TML_TBot/application/interfaces"
	"TML_TBot/domain/models"
)

type Processor struct {
}

func (p Processor) RunUseCase(job models.Job, useCase interfaces.UseCase) {
	// TODO: Runs UseCase.Run() with job specifications
}

func (p Processor) SetUpBot() {
	// TODO: Bot entrypoint logic
	// Reads the jobs in the settings and creates the cron entry for each one
	// and starts its process. It will be running forever.
}
