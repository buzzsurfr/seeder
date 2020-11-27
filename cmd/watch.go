// Package cmd has the main commands for seeder
/*
Copyright © 2020 Theo Salvo <buzzsurfr>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/buzzsurfr/seeder/internal/seed"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// watchCmd represents the watch command
var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: watch,
}

func init() {
	rootCmd.AddCommand(watchCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// watchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// watchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	watchCmd.Flags().DurationP("interval", "n", time.Hour, "wait between updates")
	viper.BindPFlag("watch.interval", watchCmd.Flags().Lookup("interval"))

}

func watch(cmd *cobra.Command, args []string) {
	// AWS Session
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Load seeds from config
	seeds := seed.UnmarshalSeeds(sess, "seeds")

	// Timer
	ticker := time.NewTicker(viper.GetDuration("watch.interval"))
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			for _, s := range seeds {
				// Copy seeds from sources to targets
				s.Copy()

				// Close source and target
				s.Close()
			}
		}
	}
}
