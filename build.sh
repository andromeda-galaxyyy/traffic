#!/bin/bash

#alias go="go1.15.4"
#env GOOS=linux
(cd ./gen &&go  build -o ../bin/gogen && cp ../bin/gogen /tmp/gogen)
(cd ./golisten &&go  build -o ../bin/golisten && cp ../bin/golisten /tmp/golisten)
(cd ./api &&go  build -o ../bin/api && cp ../bin/api /tmp/api)
(cd ./collector &&go  build -o ../bin/collector && cp ../bin/collector /tmp/collector)
(cd ./classifier && go build -o ../bin/classifier && cp ../bin/classifier /tmp/classifier)
(cd ./guiserver && go build -o ../bin/guiserver && cp ../bin/guiserver /tmp/guiserver)
(cd ./guiserver && GOOS=windows GOARCH=amd64 go build -o ../bin/winguiserver.exe && cp ../bin/winguiserver.exe /tmp/winserver.exe)

(cd ./bin && cp ./* /home/stack/code/graduate/sim/system/traffic/gogen/bin/)
