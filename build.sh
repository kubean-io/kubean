#!/usr/bin/env bash

# Create a virtual environment
VENV_DIR=kubean-venv
python3 -m venv ../$VENV_DIR
source ../$VENV_DIR/bin/activate

# Install the mkdocs utility and related dependencies
pip3 install -r docs/requirements.txt

# build the docs
mkdocs build -f docs/mkdocs.yml
mkdocs build -f docs/mkdocs.zh.yml
mkdocs build -f docs/mkdocs.en.yml

# run the local server
# cd site && python -m http.server 8000
