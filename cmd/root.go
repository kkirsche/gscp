// Copyright © 2019 Kevin Kirsche <kev.kirsche@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"

	"golang.org/x/crypto/ssh"

	"github.com/kkirsche/gscp/scp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	preserve bool
	verbose  bool

	cfgFile string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gscp",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		logrus.SetLevel(logrus.DebugLevel)
		agent, err := scp.SSHAgent()
		if err != nil {
			logrus.WithError(err).Fatalln("failed to connect to ssh agent")
		}

		config := &ssh.ClientConfig{
			User:            "root",
			Auth:            []ssh.AuthMethod{agent},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}

		conn, err := ssh.Dial("tcp", "deceiveyour.team:22", config)
		if err != nil {
			logrus.WithError(err).Fatalln("failed to connect using key")
		}
		defer conn.Close()

		session, err := conn.NewSession()
		if err != nil {
			logrus.WithError(err).Fatalln("failed to create new session")
		}
		defer session.Close()

		stdin, err := session.StdinPipe()
		if err != nil {
			logrus.WithError(err).Fatalln("failed to setup stdin for session")
		}

		stdout, err := session.StdoutPipe()
		if err != nil {
			logrus.WithError(err).Fatalln("failed to setup stdout for session")
		}

		go func() {
			cmd := "scp"
			if verbose {
				cmd = fmt.Sprintf("%s -v", cmd)
			}

			if preserve {
				cmd = fmt.Sprintf("%s -p", cmd)
			}
			cmd = fmt.Sprintf("%s -t", cmd)

			cmd = fmt.Sprintf("%s /tmp", cmd)

			fmt.Printf("Sending command %s\n", cmd)
			session.Run(cmd)
		}()

		fmt.Printf("Copying %s\n", args[0])
		err = scp.UploadFile(stdin, stdout, args[0], preserve)
		if err != nil {
			logrus.WithError(err).Fatalln("failed to transfer file")
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&preserve, "preserve", "p", false, "preserve access / modification times")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose logging")
}
