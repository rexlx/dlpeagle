#!/bin/bash

source .venv/bin/activate
python3 -m pikepdf --version
python3 add.py "$@"
echo "Python script executed successfully."