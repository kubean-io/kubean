# Local Run Documentation service

``` bash

# Go to the kubean repository directory
$ cd kubean/

# Install the mkdocs utility and related dependencies
$ pip3 install -r docs/requirements.txt 

# Run Chinese documentation locally
$ mkdocs serve -f docs/mkdocs.zh.yml

# Run English documentation locally
$ mkdocs serve -f docs/mkdocs.en.yml

```

> For more details on the development of the documentation site, please see: [mkdocs-material](https://squidfunk.github.io/mkdocs-material/)