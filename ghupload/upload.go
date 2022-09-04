package ghupload

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/go-github/v47/github"
	"golang.org/x/oauth2"
)

type GH struct {
	Token  string
	Client *github.Client
	dst    blob
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
		return blob{}, fmt.Errorf("invalid dst format: %s, should be of owner/repo@branch:dst/", path)
	}
	return blob{
		owner:  match[1],
		repo:   match[2],
		branch: match[4],
		path:   match[5],
	}, nil
}

func (g *GH) checkIfAlreadyUploaded(ctx context.Context, filename string) (*string, error) {
	sha, err := RunCMD("git", ".", "hash-object", filename)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("repos/%v/%v/git/blobs/%v", g.dst.owner, g.dst.repo, sha)
	req, err := g.Client.NewRequest("HEAD", u, nil)
	if err != nil {
		return nil, err
	}

	blob := new(github.Blob)
	resp, _ := g.Client.Do(ctx, req, blob)
	if resp.StatusCode == 200 {
		return github.String(sha), nil
	}
	return nil, nil
}

func (g *GH) CreateClient(ctx context.Context) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: g.Token},
	)
	tc := oauth2.NewClient(ctx, ts)
	g.Client = github.NewClient(tc)
}

func (g *GH) getFileMode(src string) (string, error) {
	fi, err := os.Stat(src)
	if err != nil {
		return "", err
	}
	// if it's a normal file then perm == 100644
	perm := "100644"
	if fi.IsDir() {
		perm = "100644"
	} else if fi.Mode()&os.ModeSymlink != 0 {
		return "", fmt.Errorf("symlink path: %s is not supported (cause i am lazy)", src)
	}
	return perm, nil
}

func (g *GH) walkTheWalk(ctx context.Context, dir string) ([]*github.TreeEntry, error) {
	// walkg over directory
	entries := []*github.TreeEntry{}
	err := filepath.Walk(dir, func(path string, info os.FileInfo, errr error) error {
		if errr != nil {
			return errr
		}

		if info.IsDir() {
			return nil
		}

		// get git hash for file
		sha, err := g.checkIfAlreadyUploaded(ctx, path)
		if err != nil {
			return err
		}

		perm, err := g.getFileMode(path)
		if err != nil {
			return err
		}
		if sha == nil {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			content, err := io.ReadAll(file)
			if err != nil {
				return err
			}
			// create a blob out of it
			blob, _, err := g.Client.Git.CreateBlob(ctx, g.dst.owner, g.dst.repo,
				&github.Blob{
					Content: github.String(string(content)),
				})
			if err != nil {
				return err
			}
			sha = blob.SHA
		}
		entries = append(entries, &github.TreeEntry{
			Path: github.String(filepath.Join(g.dst.path, path)),
			Mode: github.String(perm),
			Type: github.String("blob"),
			SHA:  sha,
		})
		return nil
	})
	return entries, err
}

func (g *GH) Upload(ctx context.Context, args []string, author, email, commitMessage string) error {
	if commitMessage == "" {
		commitMessage = "Uploaded by ghupload"
	}
	if author == "" {
		author = "ghuploader"
	}
	if email == "" {
		email = "ghuploader@localhost"
	}
	// dst is the last argument
	dst := args[len(args)-1]
	dstBlob, err := parsePath(dst)
	if err != nil {
		return err
	}
	isDirUpload := strings.HasSuffix(dstBlob.path, "/")
	g.dst = dstBlob

	// srcs are everything but the last argument
	srcs := args[:len(args)-1]

	// if dst don't finish by a / and we have multiple srcs then error out
	if !isDirUpload && len(srcs) > 1 {
		return fmt.Errorf("dst path must end with a / if you want to upload multiple files")
	}

	entries := []*github.TreeEntry{}
	for _, src := range srcs {
		newentries, err := g.walkTheWalk(ctx, src)
		if err != nil {
			return err
		}
		entries = append(entries, newentries...)
	}

	branch := dstBlob.branch
	repoInfo, _, err := g.Client.Repositories.Get(ctx, dstBlob.owner, dstBlob.repo)
	if err != nil {
		return err
	}
	if branch == "" {
		branch = repoInfo.GetDefaultBranch()
	}
	branchInfo, _, err := g.Client.Repositories.GetBranch(ctx, dstBlob.owner, dstBlob.repo, branch, true)
	if err != nil {
		return err
	}
	lastBranchCommit := branchInfo.GetCommit().GetSHA()
	tree, _, err := g.Client.Git.CreateTree(ctx, dstBlob.owner, dstBlob.repo, lastBranchCommit, entries)
	if err != nil {
		return err
	}
	// create a commit
	commit, _, err := g.Client.Git.CreateCommit(ctx, dstBlob.owner, dstBlob.repo, &github.Commit{
		Message: github.String(commitMessage),
		Tree:    tree,
		Parents: []*github.Commit{{Tree: tree, SHA: github.String(lastBranchCommit)}},
		Author: &github.CommitAuthor{
			Name:  github.String(author),
			Email: github.String(email),
		},
	})
	if err != nil {
		return err
	}
	// update the branch with the commit
	_, _, err = g.Client.Git.UpdateRef(ctx, dstBlob.owner, dstBlob.repo, &github.Reference{
		Ref:    github.String(fmt.Sprintf("refs/heads/%s", branch)),
		Object: &github.GitObject{SHA: commit.SHA},
	}, false)
	fmt.Printf("commit has been created: %s\n", commit.GetHTMLURL())
	fmt.Printf("branch %s has been updated to commit %s\n", branch, commit.GetSHA())
	return err
}

func NewGHClient(token string) *GH {
	return &GH{
		Token: token,
	}
}
