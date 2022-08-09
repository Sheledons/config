#!/bin/bash

DIR=`cd $(dirname $0);pwd`
cd $DIR

glide install

go build -o xlconfig-sdk-go ./