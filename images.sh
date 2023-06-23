#!/bin/sh

(cd cmd/png2svg; go build -mod=vendor -v)

for x in 128 96 64 32 16 8 6 4 2; do
  cmd/png2svg/png2svg -n ${x} -v -l -o img/rainforest_${x}c.svg img/rainforest.png
  echo svgo
  svgo --multipass img/rainforest_${x}c.svg -o img/rainforest_${x}c_opt.svg
done

cmd/png2svg/png2svg -v -l -o img/rainforest4096.svg img/rainforest.png
echo svgo
svgo --multipass img/rainforest4096.svg -o img/rainforest_opt.svg

cmd/png2svg/png2svg -v -l -o img/spaceships4096.svg img/spaceships.png
echo svgo
svgo --multipass img/spaceships4096.svg -o img/spaceships_opt.svg
