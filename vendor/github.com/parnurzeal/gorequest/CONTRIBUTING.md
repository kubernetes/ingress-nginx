# Contributing to GoRequest

Thanks for taking the time to contribute!!

GoRequest welcomes any kind of contributions including documentation, bug reports,
issues, feature requests, feature implementations, pull requests, helping to manage and answer issues, etc.

### Code Guidelines

To make the contribution process as seamless as possible, we ask for the following:

* Go ahead and fork the project and make your changes.  We encourage pull requests to allow for review and discussion of code changes.
* When you’re ready to create a pull request, be sure to:
    * Have test cases for the new code.
    * Follow [GoDoc](https://blog.golang.org/godoc-documenting-go-code) guideline and always add documentation for new function/variable definitions.
    * Run `go fmt`.
    * Additonally, add documentation to README.md if you are adding new features or changing functionality.
    * Squash your commits into a single commit. `git rebase -i`. It’s okay to force update your pull request with `git push -f`.
    * Make sure `go test ./...` passes, and `go build` completes.
    * Follow the **Git Commit Message Guidelines** below.

### Writing Commit Message

Follow this [blog article](http://chris.beams.io/posts/git-commit/). It is a good resource for learning how to write good commit messages,
the most important part being that each commit message should have a title/subject in imperative mood starting with a capital letter and no trailing period:
*"Return error when sending incorrect JSON format"*, **NOT** *"returning some error."*
Also, if your commit references one or more GitHub issues, always end your commit message body with *See #1234* or *Fixes #1234*.
Replace *1234* with the GitHub issue ID. The last example will close the issue when the commit is merged into *master*.

### Sending a Pull Request

Due to the way Go handles package imports, the best approach for working on a
fork is to use Git Remotes.  You can follow the instructions below:

1. Get the latest sources:

    ```
    go get -u -t github.com/parnurzeal/gorequest/...
    ```

1. Change to the GoRequest source directory:

    ```
    cd $GOPATH/src/github.com/parnurzeal/gorequest
    ```

1. Create a new branch for your changes (the branch name is arbitrary):

    ```
    git checkout -b issue_1234
    ```

1. After making your changes, commit them to your new branch:

    ```
    git commit -a -v
    ```

1. Fork GoRequest in Github.

1. Add your fork as a new remote (the remote name, "fork" in this example, is arbitrary):

    ```
    git remote add fork git://github.com/USERNAME/gorequest.git
    ```

1. Push the changes to your new remote:

    ```
    git push --set-upstream fork issue_1234
    ```

1. You're now ready to submit a PR based upon the new branch in your forked repository.
