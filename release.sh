#!/bin/sh
#
# Create release tarballs/zip for 64-bit linux, BSD and Plan9 + 64-bit ARM + raspberry pi 2/3
#
name=png2svg
version=$(./version.sh | cut -d' ' -f5 | cut -d, -f1)
echo "Version $version"

echo 'Compiling...'
cd cmd/png2svg
export GOARCH=amd64

echo '* Linux'
GOOS=linux go build -mod=vendor -o $name.linux
echo '* Plan9'
GOOS=plan9 go build -mod=vendor -o $name.plan9
echo '* macOS'
GOOS=darwin go build -mod=vendor -o $name.macos
echo '* FreeBSD'
GOOS=freebsd go build -mod=vendor -o $name.freebsd
echo '* NetBSD'
GOOS=netbsd go build -mod=vendor -o $name.netbsd
echo '* OpenBSD'
GOOS=openbsd go build -mod=vendor -o $name.openbsd
echo '* Linux ARM64'
GOOS=linux GOARCH=arm64 go build -mod=vendor -o $name.linux_arm64
echo '* RPI 2/3/4'
GOOS=linux GOARCH=arm GOARM=7 go build -mod=vendor -o $name.rpi
echo '* Linux static w/ upx'
CGO_ENABLED=0 GOOS=linux go build -mod=vendor -v -trimpath -ldflags "-s" -a -o $name.linux_static && upx $name.linux_static

# Compress the Linux releases with xz
for p in linux linux_arm64 rpi linux_static; do
  echo "Compressing $name-$version.$p.tar.xz"
  mkdir "$name-$version-$p"
  cp $name.$p "$name-$version-$p/$name"
  cp ../../LICENSE "$name-$version-$p/"
  tar Jcf "../../$name-$version-$p.tar.xz" "$name-$version-$p/"
  rm -r "$name-$version-$p"
  rm $name.$p
done

# Compress the other tarballs with gz
for p in macos freebsd netbsd openbsd plan9; do
  echo "Compressing $name-$version.$p.tar.gz"
  mkdir "$name-$version-$p"
  cp $name.$p "$name-$version-$p/$name"
  cp ../../LICENSE "$name-$version-$p/"
  tar zcf "../../$name-$version-$p.tar.gz" "$name-$version-$p/"
  rm -r "$name-$version-$p"
  rm $name.$p
done
