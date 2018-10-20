#!/usr/bin/env bash

git clone http://gogs.ngtech.co.il/NgTech-LTD/golang-build-software-binaries

cd tcp-to-sctp
cp -iv ../golang-build-software-binaries/* ./
make update
make all
cd ..

cd sctp-to-tcp
cp -iv ../golang-build-software-binaries/* ./
make update
make all
cd ..
