// Copyright © 2021 Cisco Systems, Inc. and its affiliates.
// All rights reserved.
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

package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/urfave/cli"

	_cli "github.com/openclarity/speculator/pkg/cli"
)

func run(c *cli.Context) {
	_cli.Run(c)
}

func main() {
	viper.AutomaticEnv()

	app := cli.NewApp()
	app.Usage = ""
	app.Name = "Speculator CLI"
	app.Version = "0.1"

	runCommand := cli.Command{
		Name:   "learn",
		Usage:  "CLI to generate OAS from HTTP transaction files",
		Action: run,
		Flags: []cli.Flag{
			cli.StringSliceFlag{
				Name:  "t",
				Usage: "path to a telemetry json file (can be ran with multiple files, e.g. -t file1.json -t file2.json)",
			},
			cli.StringFlag{
				Name:  "state",
				Usage: "path to an encoded speculator state file",
			},
			cli.StringFlag{
				Name:  "save",
				Usage: "save speculator state to a given path after learning",
			},
		},
	}
	runCommand.UsageText = runCommand.Name

	app.Commands = []cli.Command{
		runCommand,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
