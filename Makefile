GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get -v --insecure
BINARY_NAME=thresholdsMetrics
BINARY_UNIX=$(BINARY_NAME)_linux
#GPATH=/home/lpeano/banca_peano


all: deps

build:
		$(GOBUILD) -o bin/$(BINARY_UNIX) -v banca_peano

deps:   k8s.io metav1 kubernates rest yaml.v2 getopt schema

yaml.v2:
		$(GOGET)  gopkg.in/yaml.v2

getopt:
		$(GOGET)  github.com/pborman/getopt

schema:
		$(GOGET) github.com/gorilla/schema

k8s.io:

		$(GOGET) k8s.io/apimachinery/pkg/api/errors

metav1:

		$(GOGET) k8s.io/apimachinery/pkg/apis/meta/v1

kubernates:
		$(GOGET) k8s.io/client-go/kubernetes
	
rest:
		$(GOGET) k8s.io/client-go/rest
