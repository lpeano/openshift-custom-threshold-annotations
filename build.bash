export GOPATH=$HOME/thresholdsMetrics/
if [ "x$1" == "xall" ]
then
	dep ensure -update  -v
fi;
go build .
