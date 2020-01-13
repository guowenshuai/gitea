#########################################################################
# File Name: build.sh
# Author: WenShuai
# mail: guowenshuai8207@163.com
# Created Time: Thu 09 Jan 2020 10:10:29 PM CST
#########################################################################
#!/bin/bash



cd $(dirname $0)/..

make generate-swagger
git commit -am "update swagger"
make swagger-check
make docker
