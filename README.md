# ghupload - GitHub repository uploader

## What is it?

ghupload is a tool to upload files to a GitHub repository. It is a command line
tool that can be used in scripts or in a CI environment. It allows you to upload
directly via the GitHub API without having to `git clone` && `git commmit` &&
`git push` for a simple file upload.

## How to use it?

You will need a GitHub token. You can create one in your GitHub account settings.

Then you can use it like this:

```shell
    ghupload upload --token <token> file1 dir/ dir2/ owner/repo@branch:dir/
```

* dirs are uploaded recursively
* it doesn't handle synchronization, so deletion need to be manual.

The token can be specified in the environment variable `GitHub_TOKEN`. You can
specify a [pass](https://www.passwordstore.org/) entry to get the token from
there if you prefix with `pass::` :

```shell
    # this will grab the value from GitHub/token `pass` entry
    export GitHub_TOKEN=pass::GitHub/token
```

* You can omit to specify a branch, it will grab the default branch from your
  repository (i.e: `master`, `main`)
* You can specify a commit message with the `--message` option. If you don't
  specify one, a default one will be used.
* You can specify a commit author with the `--author` option. If you don't
  specify one, it will try to get the value from your git config.
* You can specify an author email with the `--email` option. If you don't
  specify one, it will try to get the value from your git config.

### Installation

### Release

Go to the [release](https://GitHub.com/chmouel/ghupload/releases) page and
choose your archive or package for your platform.

### HomeBrew

```shell
brew tap chmouel/ghupload https://github.com/chmouel/ghupload
brew install ghupload
```

### GO install

```shell
go install github.com/chmouel/ghupload@latest
```

## Copyright

[Apache-2.0](./LICENSE)

## Authors

Chmouel Boudjnah <[@chmouel](https://twitter.com/chmouel)>
