#!/bin/bash
repo_path=$(cd "$(dirname "$0")" && pwd)
rsc_path=${repo_path}/docker_files
output_path=${rsc_path}/resources
source ${rsc_path}/sources.version
mkdir -p $output_path

function is_extracted() {
	if [ -f "$1" ]; then
		echo -ne "\033[33m"
		echo -ne "done"
		echo -e "\033[0m"
	else
		echo -ne "\033[31m"
		echo -ne "failed; $1; No such file exists"
		echo -e "\033[0m"
		exit 404
	fi
	return
}

echo "Downloading Kapacitor binaries..."
cd ${output_path}
wget -O kapacitor.tar.gz https://dl.influxdata.com/kapacitor/releases/kapacitor-${KAPACITOR_VERSION}_linux_armhf.tar.gz
tar -xzf kapacitor.tar.gz
rm kapacitor.tar.gz
rm kapacitor
ln -s "kapacitor-${KAPACITOR_VERSION}"*/ kapacitor
is_extracted ${output_path}/kapacitor/usr/bin/kapacitord

echo "Setting necessary configurations..."
export GOPATH=${output_path}
go get -v -u go.uber.org/zap
go get -v -u gopkg.in/mgo.v2
cp ${rsc_path}/kapacitor.conf ${output_path}/
is_extracted ${output_path}/kapacitor.conf
is_extracted ${output_path}/src/go.uber.org/zap/Makefile

echo -ne "Setting neccessary files..."
cp ${rsc_path}/setldd.sh ${output_path}/
is_extracted ${output_path}/setldd.sh

cp /usr/bin/qemu-arm-static ${output_path}/
