package cmd

import (
	"TML_TBot/infrastructure/connectors"
	"TML_TBot/infrastructure/workers"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(jobCmd)
}

var rootCmd = &cobra.Command{
	Use:   "TML Bot",
	Short: "Tomorrowland 2023 ES Bot",
	Long:  "Tomorrowland 2023 ES Bot",
	Run:   executeRoot,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func executeRoot(cmd *cobra.Command, args []string) {
	cron := workers.NewCronWorker()
	tgService := connectors.NewTelegramService()
	processor := workers.NewProcessor(&cron, tgService)
	processor.StartCronBot()
}
