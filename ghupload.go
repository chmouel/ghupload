package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/chmouel/ghupload/ghupload"

	"github.com/urfave/cli/v2"
)

func app() error {
	ctx := context.Background()
	commonFlag := []cli.Flag{
		&cli.StringFlag{
			Name:  "token",
			Usage: "github token",
			EnvVars: []string{
				"GITHUB_TOKEN",
			},
		},
		&cli.StringFlag{
			Name:  "message",
			Usage: "commit message",
		},

		&cli.StringFlag{
			Name:  "author",
			Usage: "commit author",
			EnvVars: []string{
				"GHUPLOAD_AUTHOR",
			},
		},

		&cli.StringFlag{
			Name:  "email",
			Usage: "commit email",
			EnvVars: []string{
				"GHUPLOAD_EMAIL",
			},
		},
	}
	app := &cli.App{
		EnableBashCompletion: true,
		Version:              "1.0.0",
		Commands: []*cli.Command{
			{
				Name:  "upload",
				Flags: commonFlag,
				Action: func(c *cli.Context) error {
					if c.Args().Len() < 2 {
						return fmt.Errorf("invalid number of arguments we need at least a src and a dst")
					}
					token := c.String("token")
					if token == "" {
						return cli.Exit("github token need to be set", 1)
					}
					if strings.HasPrefix(token, "pass::") {
						passkey := strings.TrimPrefix(token, "pass::")
						var err error
						if token, err = ghupload.RunCMD("pass", ".", "show", passkey); err != nil {
							return err
						}
					}

					author := c.String("author")
					if author == "" {
						author, _ = ghupload.RunCMD("git", ".", "config", "--global", "user.name")
					}

					email := c.String("email")
					if email == "" {
						email, _ = ghupload.RunCMD("git", ".", "config", "--global", "user.email")
					}

					g := ghupload.NewGHClient(token)
					g.CreateClient(ctx)
					return g.Upload(ctx, c.Args().Slice(), author, email, c.String("message"))
				},
				Usage: "upload a file",
			},
		},
	}
	return app.Run(os.Args)
}

func main() {
	if err := app(); err != nil {
		log.Fatal(err)
	}
}
