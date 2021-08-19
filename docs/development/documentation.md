# Documentation Setup

The documentation of the [onmetal-api](https://github.com/onmetal/onmetal-api) project is written primarily using Markdown. 
All documentation related content can be found in the `/docs` folder. New content  also should be added there. 
[MkDocs](https://www.mkdocs.org/) and [MkDocs Material](https://squidfunk.github.io/mkdocs-material/) are then used to  render the contents of the `/docs` folder to have a more user-friendly experience when browsing the projects' documentation.

!!!note  
    One exception to the [common contribution process](/development/contribution/#steps-to-contribute) builds  the `docs/api-reference` folder. 
    The folder contains auto-generated CRD reference documentation of the project,  no _manual_ contributions should be applied as they will be overwritten in the next generation step.
    To read more:  [Updating API Reference Documentation](#api-reference-documentation) section.

## Local Development Setup

This project contains a local Docker based runtime environment for the documentation part. Just run

```shell
make start-docs
```

in the projects root folder and access the output in your browser under <http://localhost:8000/>

The environment will hot-rebuild your documentation  so there is no need to restart it while you make your changes.
If you want to add a new chapter (basically a new file/folder to `docs` directory) you should also add it to the `nav` section in the `mkdocs.yml` file in the projects root folder.

Use this helper Makefile directive to clean up old and stopped container instances 

```shell
make clean-docs
```

## Writing Content

### Abbreviations

Abbreviations are defined centrally in the following file `/hack/docs/abbreviations.md`. So in case you introduce any 
new abbreviation to your content please make sure to also add a corresponding entry there. At the end of each Markdown
file you also need to include the statement `--8<-- "hack/docs/abbreviations.md"`. This will ensure that the abbreviation
highlighting will work correctly.

## API Reference Documentation

The [API reference documentation](/api-reference/overview/) contains auto-generated description from the CRD definition of the [onmetal-api](https://github.com/onmetal/onmetal-api) project.

We are using the [gen-crd-api-reference-docs](https://github.com/ahmetb/gen-crd-api-reference-docs) project
to generate the content. Under the hood we are using `go generate` instructions defined in each version type
`doc.go`.

The needed instructions to generate documentation for the `core/v1alpha1` types are  in the example below 

```go
//go:generate gen-crd-api-reference-docs -api-dir . -config ../../../hack/api-reference/core-config.json -template-dir ../../../hack/api-reference/template -out-file ../../../docs/api-reference/core.md
```

Together with the comments in the corresponding type files `go generate` will call the `gen-crd-api-reference-doc` command
to generate the output in the `/docs/api-reference` folder.

This project contains a `Makefile` routine to generate the reference documentation for all types. In case you change any
of the types in the `apis` folder just run

```shell
make docs
```

The generated output should be part of your pull request.

--8<-- "hack/docs/abbreviations.md"