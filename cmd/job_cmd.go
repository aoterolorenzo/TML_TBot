package cmd

import (
	"TML_TBot/config"
	"TML_TBot/infrastructure/connectors"
	"TML_TBot/infrastructure/workers"
	"fmt"
	"github.com/spf13/cobra"
)

var jobCmd = &cobra.Command{
	Use:   "job",
	Short: "Execute a single job",
	Run:   executeJob,
}

func executeJob(cmd *cobra.Command, args []string) {
	// Check if at least one argument is provided
	if len(args) < 1 {
		fmt.Println("Missing argument. Please provide the required argument.")
		return
	}

	// Access the argument
	jobID := args[0]
	cron := workers.NewCronWorker()
	tgService := connectors.NewTelegramService()
	processor := workers.NewProcessor(&cron, tgService)
	err := processor.RubJobById(jobID)
	if err != nil {
		config.Log.Fatal(err)
	}

}
