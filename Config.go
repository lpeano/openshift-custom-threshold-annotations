package main

import (
    "gopkg.in/yaml.v2"
    "io/ioutil"
    "k8s.io/klog"
)


const problemYaml = `
name: GO_BUILDER
`

type appConfig struct {
	AnnotationFlag string `yaml:"AnnotationFlag"` 
	AnnotationNameThreshold string `yaml:"AnnotationNameThreshold"` 
        NameSpace string `yaml:"a,omitempty"`
	LOGLEVEL string  `yaml:"LOGLEVEL"`
        CacheRefreshIntervall int `yaml:"CacheRefreshIntervall"`
}

func (c *appConfig) GetConf(file string) (* appConfig) {

    yamlFile, err := ioutil.ReadFile(file)
    if err != nil {
        klog.Errorf("yamlFile.Get err   #%v ", err)
	panic(err)
    }
    err = yaml.Unmarshal(yamlFile, c)
    if err != nil {
        klog.Errorf("Unmarshal: %v", err)
	panic(err)
    }
    
    return c
}


