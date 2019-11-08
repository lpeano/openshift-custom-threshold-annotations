/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Note: the example only works with the code within the same release/branch.
package main

import (
	"flag"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)


// Log Level 
// default:= 2 
var appConf appConfig

func initialize() {
	appConf.annotation_name="sia.io/thresholds_config"
	appConf.annotation_name_threshold="sia.io/thresholds"
	appConf.LOGLEVEL=os.Getenv("LOGLEVEL")
}

// Null Function 
func Use(vals ...interface{}) {
    for _, val := range vals {
        _ = val
    }
}

func init_nocluster() (*kubernetes.Clientset , error) {
	klog.Info("Connetting to cluster")
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	return clientset, err
}


func main() {
		// Init Envirorment parameters
		initialize()
		clientset, err := init_nocluster()	
		if err != nil {
			panic(err.Error())
		} 
		ServiceCache.Get_services( clientset )
		// Start Watching Service 
		go ServiceCache.Service_Watcher(clientset)
		start_prometheus()
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
}
	return os.Getenv("USERPROFILE") // windows
}
