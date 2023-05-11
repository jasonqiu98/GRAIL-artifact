#!/bin/bash

grep -c :ok *.edn | awk -F ".edn:" '{print $1, $2}' | sort -nk 1 | awk '{print $2}'