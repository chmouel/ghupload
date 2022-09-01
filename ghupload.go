package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"

	"github.com/google/go-github/v47/github"
	"github.com/urfave/cli/v2"
	"golang.org/x/oauth2"
)

type GH struct {
	token  string
	client *github.Client
}

type blob struct {
	owner, repo, path, branch string
}

// parsePath takes a path that look like this owner/repo@branch:path/to/file
// and returns the owner, repo, branch and path, @branch can be optional and
// will default to the default branch as set on the repo
func parsePath(path string) (blob, error) {
	regexp := regexp.MustCompile(`^(?P<owner>[^/]+)\/(?P<repo>[^@]+)(@(?P<branch>[^:]+))?:(?P<src>.+)$`)
	match := regexp.FindStringSubmatch(path)
	if match == nil {
		return blob{}, fmt.Errorf("invalid path format: %s", path)
	}
	return blob{
		owner:  match[1],
		repo:   match[2],
		branch: match[4],
		path:   match[5],
	}, nil
}

func (g *GH) createClient(ctx context.Context) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: g.token},
	)
	tc := oauth2.NewClient(ctx, ts)
	g.client = github.NewClient(tc)
}

func (g *GH) upload(ctx context.Context, src, dst, commitMessage string) error {
	blob, err := parsePath(dst)
	if err != nil {
		return err
	}
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	// read file
	content, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	branch := blob.branch
	if branch == "" {
		branchInfo, _, err := g.client.Repositories.Get(ctx, blob.owner, blob.repo)
		if err != nil {
			return err
		}
		branch = branchInfo.GetDefaultBranch()
	}
	fc, dc, _, _ := g.client.Repositories.GetContents(ctx, blob.owner, blob.repo, blob.path, &github.RepositoryContentGetOptions{})
	if dc != nil {
		return fmt.Errorf("destination is a directory: %s", dst)
	}
	// if file hasn't been uploaded yet, create it
	if fc == nil {
		if commitMessage == "" {
			commitMessage = "created by ghupload"
		}
		cntResp, _, err := g.client.Repositories.CreateFile(ctx, blob.owner, blob.repo, blob.path, &github.RepositoryContentFileOptions{
			Message: github.String(commitMessage),
			Content: content,
			Branch:  github.String(branch),
		})
		if err != nil {
			return err
		}
		fmt.Printf("File has been uploaded to: %s\n", cntResp.GetContent().GetHTMLURL())
	} else { // else update the file
		if commitMessage == "" {
			commitMessage = "updated by ghupload"
		}
		ud, _, err := g.client.Repositories.UpdateFile(ctx, blob.owner, blob.repo, blob.path, &github.RepositoryContentFileOptions{
			Message: github.String(commitMessage),
			SHA:     fc.SHA,
			Content: content,
		})
		if err != nil {
			return err
		}
		fmt.Printf("File has been updated on: %s\n", ud.GetContent().GetHTMLURL())
	}
	return nil
}

func newGH(token string) *GH {
	return &GH{
		token: token,
	}
}

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
	}
	app := &cli.App{
		EnableBashCompletion: true,
		Version:              "1.0.0",
		Commands: []*cli.Command{
			{
				Name:  "upload",
				Flags: commonFlag,
				Action: func(c *cli.Context) error {
					src := c.Args().Get(0)
					dst := c.Args().Get(1)
					token := c.String("token")
					if token == "" {
						return cli.Exit("github token need to be set", 1)
					}
					g := newGH(token)
					g.createClient(ctx)
					return g.upload(ctx, src, dst, c.String("message"))
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
