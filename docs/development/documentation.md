# Documentation Setup

The documentation of the [ironcore](https://github.com/ironcore-dev/ironcore) project is written primarily using Markdown.
All documentation related content can be found in the `/docs` folder. New content also should be added there.
[MkDocs](https://www.mkdocs.org/) and [MkDocs Material](https://squidfunk.github.io/mkdocs-material/) are then used to render the contents of the `/docs` folder to have a more user-friendly experience when browsing the projects' documentation.

!!! note
    One exception to the [common contribution process](/development/contribution/#steps-to-contribute) builds the `docs/api-reference` folder. The folder contains auto-generated CRD reference documentation of the project, no manual contributions should be applied as they will be overwritten in the next generation step.
    To read more: [Updating API Reference Documentation](#api-reference-documentation)  section.

## Requirements:

Following tools are required to work on that package.

* Kubernetes cluster access to deploy and test the result (via minikube, kind or docker desktop locally)
* [make](https://www.gnu.org/software/make/) - to execute build goals
* [docker](https://www.docker.com) - to run the local mkdocs environment
* [git](https://git-scm.com/downloads) - to be able to commit any changes to repository
* [kubectl](https://kubernetes.io/docs/reference/kubectl/) (>= v1.23.4) - to be able to talk to the kubernetes cluster
  
!!! note
    If you don't have Docker installed on your machine please follow one of those guides:

    * [Docker Desktop for Mac](https://docs.docker.com/desktop/mac/install/)
    * [Docker Desktop for Windows](https://docs.docker.com/desktop/windows/install/)
    * [Docker Engine for Linux](https://docs.docker.com/engine/install/#server)

## Local Development Setup
This project contains a local Docker based runtime environment for the documentation part. If you have an access to the docker registry and k8s installation that you can use for development purposes, just run following command and access the output in your browser under <http://localhost:8000/>:

```shell
make start-docs
```
The environment will hot-rebuild your documentation, so there is no need to restart it while you make your changes.
If you want to add a new chapter (basically a new file/folder to `docs` directory) you should add it to the `nav` section in the `mkdocs.yml` file in the projects root folder.
Use helper Makefile directive to clean up old and stopped container instances.

```shell
make clean-docs
```

## Writing Content

### Abbreviations
Abbreviations are defined centrally in the following file `/hack/docs/abbreviations.md`. In case you introduce any new abbreviation to your content, please make sure to add a corresponding entry there.
Please include the statement `--8<-- "hack/docs/abbreviations.md"` at the end of each Markdown file. This will ensure that the abbreviation highlighting will work correctly.

## API Reference Documentation

The [API reference documentation](/api-reference/overview/) contains auto-generated description from the CRD definition of the [ironcore](https://github.com/ironcore-dev/ironcore) project.
We are using the [gen-crd-api-reference-docs](https://github.com/ahmetb/gen-crd-api-reference-docs) project to generate the content. Under the hood we are using `go generate` instructions defined in each version type `doc.go`.
The needed instructions to generate documentation for the `core/v1alpha1` types are in the example below:

```go
//go:generate gen-crd-api-reference-docs -api-dir . -config ../../../hack/api-reference/core-config.json -template-dir ../../../hack/api-reference/template -out-file ../../../docs/api-reference/core.md
```
Together with the comments in the corresponding type files `go generate` will call the `gen-crd-api-reference-doc` command to generate the output in the `/docs/api-reference` folder.
The project contains a `Makefile` routine to generate the reference documentation for all types. In case you change any of the types in the `apis` folder just run:

```shell
make docs
```

!!! note
    The generated output should be part of your pull request.

--8<-- "hack/docs/abbreviations.md"
