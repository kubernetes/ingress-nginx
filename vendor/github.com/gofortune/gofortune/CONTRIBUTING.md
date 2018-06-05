# Contributing to GoFortune

Thank you for contributing! GoFortune welcomes contributions following the guidelines described in this document.

This document outlines the procedures and what to expect when contributing documentation, a bug report, 
a feature request, or new code via a pull request. I

Please read and follow these guidelines before submitting a bug report, feature request or pull request.

 - [Code of Conduct](#code-of-conduct)
 - [Question or Problem?](#got-a-question-or-problem)
 - [Issues and Bugs](#issues-and-bugs)
 - [Feature Requests](#feature-requests)
 - [Submission Guidelines](#submission-guidelines)
 - [Coding Rules](#coding-rules)
 - [Signing the CLA](#signing-the-cla)

## Code of Conduct

Make sure you help to keep GoFortune open and inclusive. Please read and follow our [Code of Conduct](CODE_OF_CONDUCT.md).

## Got a Question or Problem?

If you have questions or a problem related with GoFortune, please direct these to the issue tracker.
Although it seems clear an issue tracker is not the best communication mechanism, for the time
being there is no need for any other mechanism.

## Found an Issue?

If you find a bug in the source code or a mistake in the documentation, you can help by
submitting an issue to our [GitHub Repository](http://www.github.com/vromero/gofortune).
Even better you can submit a Pull Request with a fix.

**Please see the [Submission Guidelines](#submission-guidelines) below.**

## Want a Feature?

You can request a new feature by submitting an issue to our [GitHub Repository](http://www.github.com/vromero/gofortune).  If you
would like to implement a new feature then consider a Pull Request.

**Please see the [Submission Guidelines](#submission-guidelines) below.**

## <a name="submit"></a> Submission Guidelines

### Submitting an Issue
Before you submit your issue search the archive, maybe your question was already answered.

If your issue appears to be a bug, and hasn't been reported, open a new issue. 

### Submitting a Pull Request

Before you submit your pull request consider the following guidelines:

* Search GitHub for an open or closed Pull Request
  that relates to your submission. You don't want to duplicate effort.
* Please sign our Contributor License Agreement (CLA) before sending pull
  requests. Code cannot be accepted without this.
* Make your changes in a new git branch:

    ```shell
    git checkout -b feature/my-new-feature master
    ```

* Create your patch, **including appropriate test cases**.
* Follow our [Coding Rules](#coding-rules).
* Run the full test suite and ensure that all tests pass.
* Commit your changes using a descriptive commit message.
* Push your branch to GitHub:

In GitHub, send a pull request to `vromero/gofortune:master`.

If the PR gets too outdated it may be asked to rebase and force push to update the PR:

```shell
git rebase master -i
git push origin feature/my-new-feature -f
```

That's it! Thank you for your contribution!

#### After your pull request is merged

After your pull request is merged, you can safely delete your branch and pull the changes
from the main (upstream) repository:

* Delete the remote branch on GitHub either through the GitHub web UI or your local shell as follows:

    ```shell
    git push origin --delete feature/my-new-feature
    ```

* Check out the master branch:

    ```shell
    git checkout master -f
    ```

* Delete the local branch:

    ```shell
    git branch -D feature/my-new-feature
    ```

* Update your master with the latest upstream version:

    ```shell
    git pull --ff upstream master
    ```

## Coding Rules

To ensure consistency throughout the source code, keep these rules in mind as you are working:

* All features or bug fixes **must be tested** by one or more tests.
* All public API methods **must be documented**.
* With the exceptions listed below, we follow the rules contained in
  [Google's Go Style Guide][https://github.com/golang/go/wiki/CodeReviewComments]:

## Signing the CLA

Please sign our Contributor License Agreement (CLA) before sending pull requests. For any code
changes to be accepted, the CLA must be signed. It's a quick process, we promise!

