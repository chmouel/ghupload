package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

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

func (g *GH) upload(ctx context.Context, args []string, commitMessage string) error {
	if commitMessage == "" {
		commitMessage = "Uploaded by ghupload"
	}
	// dst is the last argument
	dst := args[len(args)-1]
	dstBlob, err := parsePath(dst)
	if err != nil {
		return err
	}
	isDirUpload := strings.HasSuffix(dstBlob.path, "/")

	// srcs are everything but the last argument
	srcs := args[:len(args)-1]

	// if dst don't finish by a / and we have multiple srcs then error out
	if !isDirUpload && len(srcs) > 1 {
		return fmt.Errorf("dst path must end with a / if you want to upload multiple files")
	}
	entries := []*github.TreeEntry{}
	for _, src := range srcs {
		fi, err := os.Stat(src)
		if err != nil {
			return err
		}
		// if it's a normal file then perm == 100644
		perm := "100644"
		if fi.IsDir() {
			perm = "100644"
		} else if fi.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink is not supported (cause i am lazy)")
		}

		file, err := os.Open(src)
		if err != nil {
			return err
		}
		content, err := io.ReadAll(file)
		if err != nil {
			return err
		}
		entries = append(entries, &github.TreeEntry{
			Path:    github.String(filepath.Join(dstBlob.path, fi.Name())),
			Mode:    github.String(perm),
			Type:    github.String("commit"),
			Content: github.String(string(content)),
		})

	}

	branch := dstBlob.branch
	repoInfo, _, err := g.client.Repositories.Get(ctx, dstBlob.owner, dstBlob.repo)
	if err != nil {
		return err
	}
	if branch == "" {
		branch = repoInfo.GetDefaultBranch()
	}
	branchInfo, _, err := g.client.Repositories.GetBranch(ctx, dstBlob.owner, dstBlob.repo, branch, true)
	if err != nil {
		return err
	}
	lastBranchCommit := branchInfo.GetCommit().GetSHA()
	tree, _, err := g.client.Git.CreateTree(ctx, dstBlob.owner, dstBlob.repo, lastBranchCommit, entries)
	if err != nil {
		return err
	}
	// create a commit
	commit, _, err := g.client.Git.CreateCommit(ctx, dstBlob.owner, dstBlob.repo, &github.Commit{
		Message: github.String(commitMessage),
		Tree:    tree,
		Parents: []*github.Commit{{Tree: tree, SHA: github.String(lastBranchCommit)}},
		Author:  &github.CommitAuthor{Name: github.String("ghupload"), Email: github.String("ghupload@ghupload.com")},
	})
	if err != nil {
		return err
	}
	// update the branch with the commit
	_, _, err = g.client.Git.UpdateRef(ctx, dstBlob.owner, dstBlob.repo, &github.Reference{
		Ref:    github.String(fmt.Sprintf("refs/heads/%s", branch)),
		Object: &github.GitObject{SHA: commit.SHA},
	}, false)
	fmt.Printf("commit has been created: %s\n", commit.GetHTMLURL())
	fmt.Printf("branch %s has been updated to commit %s\n", branch, commit.GetSHA())
	return err
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
					if c.Args().Len() != 2 {
						return fmt.Errorf("invalid number of arguments we need at least a src and a dst")
					}
					token := c.String("token")
					if token == "" {
						return cli.Exit("github token need to be set", 1)
					}
					g := newGH(token)
					g.createClient(ctx)
					return g.upload(ctx, c.Args().Slice(), c.String("message"))
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
