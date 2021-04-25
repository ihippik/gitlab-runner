package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/ihippik/gitlab-runner/runner"
)

// GITVersion contains the hash of the commit - set on build.
var GITVersion = "local"

const gitlabAPI = "api/v4"

func main() {
	var (
		srv    *runner.Service
		logger *logrus.Entry
	)

	app := &cli.App{
		Name:    "GitLab Runner",
		Usage:   "a GitLab Runner",
		Version: GITVersion,
		Authors: []*cli.Author{
			{Name: "ihippik", Email: "hippik80@gmail.com"},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "config", Aliases: []string{"c"}, Value: "config.yml"},
		},
		Before: func(c *cli.Context) error {
			cfg, err := initConfig(c.String("c"))
			if err != nil {
				return fmt.Errorf("init config: %w", err)
			}

			logger = initLogger(cfg.Logger, GITVersion, cfg.Runner.Executor)
			api := runner.NewGitlabAPI(http.DefaultClient, cfg.Runner.URL+gitlabAPI)
			srv = runner.NewService(logger, cfg, api, executorFactory(logger, cfg.Runner.Executor))

			return nil
		},
		Commands: []*cli.Command{
			{
				Name:    "register",
				Aliases: []string{"r"},
				Usage:   "register gitlab-runner",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "token", Aliases: []string{"t"}, Required: true},
				},
				Action: func(c *cli.Context) error {
					regToken := c.String("token")
					token, err := srv.Registration(regToken)
					if err != nil {
						return fmt.Errorf("registration: %w", err)
					}

					logger.WithField("token", token).Infoln("add token in config")

					return nil
				},
			},
		},
		Action: func(c *cli.Context) error {
			if err := srv.Start(); err != nil {
				return fmt.Errorf("start: %w", err)
			}
			return nil
		},
	}
	err := app.Run(os.Args)

	if err != nil {
		logrus.Fatalln(err)
	}
}
