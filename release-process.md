---
title: Kubean Release Process
layout: page
---

# Release process

This is the process to follow to make a new release. Similar to [Semantic](https://semver.org/) Version. The project
refers to the respective components of this triple as <major>.<minor>.<patch>.

## Three types of release process

* Minor release process, that means a version which have some features , such as `v0.4.0`
* Patch release process, that means a version which includes some bug fix , such as `v0.4.1`
* RC release process, that means a preview version before the official release ,such as `v0.4.2-rc1`

## How to release a new version

### Overview

For example, A minor release is e.g. `v0.4.0`.

A minor release requires:

pre steps:

- Check pipeline success
- smoke test(optional)

core steps:

- push new tag to release
- Smoke test based on new version

post steps:

- Submit an issue of the kubean version update to the [documentation site](https://github.com/DaoCloud/DaoCloud-docs)
- Website updates，blog update
- announce Message at the NDX work WeChat(include the changelog)

### pre steps

### core steps

#### Push the new tag

```bash
new_tag=v0.4.0 ## for example
git tag $new_tag
git push origin $new_tag 
```

if a tag is pushed, the following steps will automatically run:

1. build the images with the new tag ,and then push to ghcr.io
2. build the file artifacts ,such as os-pkgs and images-amd64.tar.gz
3. push a new version helm charts to [kubean-helm-chart](https://github.com/kubean-io/kubean-helm-chart)
4. generate [a release note](https://github.com/kubean-io/kubean/releases) which includes `What's Changed` which shows
   the previous pull requests
   and `New Contributors`
5. push the kubean client library to [kubean-api](https://github.com/kubean-io/kubean-api)

#### Smoke test based on new version

1. create a new k8s cluster by minikube or kind
2. use helm charts to install the new version kubean

# How to check release notes

Go to the [site](https://kubean-io.github.io/website/zh/00Releases/) to check more documents.