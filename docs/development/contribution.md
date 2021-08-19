# Contributors Guide

## Contributing 

The Gardener on Metal project uses Github to manage reviews of pull requests.

1. If you are looking to make your first contribution, follow [Steps to Contribute](#steps-to-contribute)

2. If you have a trivial fix or improvement, go ahead and create a pull request and
address (with @...) a suitable maintainer of this repository 
(see [CODEOWNERS](https://raw.githubusercontent.com/onmetal/onmetal-api/main/CODEOWNERS) 
of this repository) in the description of the pull request.

3. If you plan to do something more involved, first discuss your ideas by creating an 
[issue](https://github.com/onmetal/onmetal-api/issues) for this repository. This will avoid unnecessary work and surely give you 
and us a good deal of inspiration.

> Note: please follow these style guidelines to have your contribution considered by the maintainers:
Coding style guidelines Go Code Review Comments (https://github.com/golang/go/wiki/CodeReviewComments)
Formatting and style section of Peter Bourgon’s Go: Go: Best Practices for Production Environments. (http://peter.bourgon.org/go-in-production/#formatting-and-style).

## Steps to Contribute

Do you want to work on an issue?  You are welcome to claim an existing one by commenting on it in GitHub. 
>Note: perform a cursory search to see if the issue has already been taken by someone else. 
This will prevent misunderstanding and duplication of  effort from contributors on the same issue.

If you have questions about one of the issues please comment on them and one of the 
maintainers will clarify it.

We kindly ask you to follow the [Pull Request Checklist](#pull-request-checklist) to ensure reviews can happen accordingly.

## Contributing Code

You are welcome to contribute code to the Gardener on Metal project in order to fix a bug or to implement a new feature.

The following rules govern code contributions:

* Contributions must be licensed under the [Apache 2.0 License](http://www.apache.org/licenses/LICENSE-2.0)
* You need to sign the [Developer Certificate of Origin](#developer-certificate-of-origin).

## Contributing Documentation

You are welcome to contribute documentation to the Gardener on Metal project.

The following rules govern documentation contributions:

* Contributions must be licensed under the [Creative Commons Attribution 4.0 International License](https://creativecommons.org/licenses/by/4.0/legalcode)
* You need to sign the [Developer Certificate of Origin](#developer-certificate-of-origin).

## Developer Certificate of Origin

Due to legal reasons, contributors will be asked to accept a Developer Certificate of Origin (DCO) before they submit 
the first pull request to the Gardener on Metal project, this happens in an automated fashion during the submission 
process. We use [the standard DCO text of the Linux Foundation](https://developercertificate.org/).

## Pull Request Checklist

* Fork and clone the repository to you local machine.

```shell
git clone git@github.com:YOUR_GITHUB_USER/onmetal-api.git
cd onmetal-api
# add an upstream remote to fetch new changes
git remote add upstream git@github.com:onmetal/onmetal-api.git
```

* Create a feature branch from the `main` branch and, if needed, rebase to the current `main` branch before submitting 
your pull request. If it doesn't merge cleanly with `main` you may be asked to rebase your changes.

```shell
git checkout -b my_feature
# rebase if necessary
git fetch upstream main
git rebase upstream/main
```

* Commits should be as small as possible, while ensuring that each commit is correct independently 
(i.e. each commit should compile and pass tests).

* Test your changes as thoroughly as possible before you commit them. Preferably, automate your test by unit / integration tests. 
If tested manually, provide information about the test scope in the PR description. Now you can commit
your changes to your feature branch and push it to your fork.

```shell
git add .
git commit -m "Something meaningful"
git push origin my_feature
# alternatively you can amend your commit before pushing if you forgot something
git commit --amend
```

* Create _Work In Progress [WIP]_ pull requests only if you need a clarification or an explicit review before you can 
continue your work item.

* If your patch is not getting reviewed, or you need a specific person to review it, you can @-reply a reviewer asking 
for a review in the pull request or a comment.

* Post review:
    * If a reviewer requires you to change your commit(s), please test the changes again.
    * Amend the affected commit(s) and force push onto your branch.
    * Set respective comments in your GitHub review as resolved.
    * Create a general PR comment to notify the reviewers that your amendments are ready for another round of review.

## Issues and Planning

We use GitHub issues to track bugs and enhancement requests. Please provide as much context as possible when you open 
an issue. The information you provide must be comprehensive enough to reproduce that issue for the assignee. 
Therefore, contributors may use but aren't restricted to the issue template provided by the Gardener on Metal maintainers.

Issues and pull requests are tracked in the [backlog](https://github.com/onmetal/onmetal-api/projects/1) for this project.
