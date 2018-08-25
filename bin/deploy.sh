#!/bin/bash
BASE_DIR=$(cd $(dirname $BASH_SOURCE)/../; pwd)
rm $BASE_DIR/* -rf
cd $BASE_DIR
wget http://cp01-shifen-001.cp01.baidu.com:8001/goproxy.tar.gz
tar zxvf goproxy.tar.gz
