#!/bin/sh
(cd cmd/png2svg; go build -mod=vendor -v)
cmd/png2svg/png2svg -v -l -o img/rainforest4096.svg img/rainforest.png
cmd/png2svg/png2svg -v -l -n 4 -o img/rainforest_4c.svg img/rainforest.png
cmd/png2svg/png2svg -v -l -n 8 -o img/rainforest_8c.svg img/rainforest.png
cmd/png2svg/png2svg -v -l -n 16 -o img/rainforest_16c.svg img/rainforest.png
cmd/png2svg/png2svg -v -l -n 32 -o img/rainforest_32c.svg img/rainforest.png
cmd/png2svg/png2svg -v -l -n 64 -o img/rainforest_64c.svg img/rainforest.png
cmd/png2svg/png2svg -v -l -n 96 -o img/rainforest_96c.svg img/rainforest.png
