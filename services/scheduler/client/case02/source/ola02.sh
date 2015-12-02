#!/bin/bash
#get images from docker hub.

ret=0
docker pull busybox
if [ $? -ne 0 ];then
	echo "FAILED"
	ret=1
else
	echo "PASS"
fi

docker images | grep busybox
if [ $? -ne 0 ];then
	echo "FAILED"
	ret=2
else
	echo "PASS"
fi

exit $ret
