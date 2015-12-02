#!/bin/bash
#use ubuntu image to create a docker container,and print "hello, world" in the container.

ret=0
var=`docker run --rm ubuntu bash -c "echo hello, world"`
if [ "$var"x != "hello, world"x ];then
	echo "FAILD"
	ret=1
else
        echo "PASS"
fi
		
exit $ret
