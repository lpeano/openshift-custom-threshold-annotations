package main

import(
        "net/http"
        "github.com/prometheus/client_golang/prometheus"
        "github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/klog"
	"time"
	"github.com/davecgh/go-spew/spew"
)



var SericesMetrics Services
var PodsMetrics Pods
var PromRegister *prometheus.Registry
var handler *http.Handler
var server *http.Server

//type Prom_Metrics_Collectors  
type pCollector struct {
	
	MetrDescriptor *prometheus.Desc
	labels prometheus.Labels
	variablelabels []string
	Name string
}

func  newpCollector(Name string, Description string , labels map[string]string, varlabels []string ) *pCollector {
	klog.Infof("LABELS %v",labels)
	varlabels=append(varlabels,"namespace")
	varlabels=append(varlabels,"service")
	return &pCollector{ Name: Name,MetrDescriptor: prometheus.NewDesc(Name,Description,varlabels,labels), labels: labels ,variablelabels: varlabels}

}

func MakePodsCollectors() {
	// For each Namepspace
        for n, s := range PodsMetrics {
                // For each Pod
                for sn, srv := range s {
                        // For each Metric
                        for mn , m := range srv.metrics {
                                // For each Label
				annotations := make(map[string]string)
				annotations["namespace"]= n
				annotations["pod"]= sn
				var labels []string
                                for ln, l := range m.AdditionalLabels {
                                        klog.Infof("Namespace: %s - Pod: %s -- [[ metric.Name: %s ]] %v LabelName: %s LabelValue: %s", n, sn,  m.Name , m.Value,l.Name,l.Value)
					annotations[l.Name]=l.Value
					labels=append(labels,l.Name)
                                        Use(mn,ln)
                                }
				labels=append(labels,"namespace")
				labels=append(labels,"pod")

				klog.Infof("Metrics labels %v", labels)
        			gv:=prometheus.NewGaugeVec( prometheus.GaugeOpts{
					Namespace: "Pod",
                			Name: m.Name ,
					Help: "SIA Metrics THresholds",
                		}, labels)
                                if err := PromRegister.Register(gv) ; err != nil {
                                        klog.Errorf("Error registering Collector for Pod \"%s\"", err)
					if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
        					gv = are.ExistingCollector.(*prometheus.GaugeVec)
    					} else {
						PromRegister.Unregister(gv)
						PromRegister.Register(gv)
    					}	
                                } else {
                                        klog.Info("Collector registered")
                                }

        			gv.With(annotations).Set(m.Value)
				klog.Infof("Set Collector %v", annotations)
				//spew.Dump( gv )
				/*if err := PromRegister.Register(col) ; err != nil {
					klog.Errorf("Error registering Collector \"%s\"", err)
				} else {
					klog.Info("Collector registered")
				}*/		

                        }
                }
        }
}

func MakeServicesCollectors() {
	// For each Namepspace
        for n, s := range SericesMetrics {
                // For each Service
                for sn, srv := range s {
                        // For each Metric
                        for mn , m := range srv.metrics {
                                // For each Label
				annotations := make(map[string]string)
				annotations["namespace"]= n
				annotations["service"]= sn
				var labels []string
                                for ln, l := range m.AdditionalLabels {
                                        klog.Infof("Namespace: %s - Service: %s -- [[ metric.Name: %s ]] %v LabelName: %s LabelValue: %s", n, sn,  m.Name , m.Value,l.Name,l.Value)
					annotations[l.Name]=l.Value
					labels=append(labels,l.Name)
                                        Use(mn,ln)
                                }
				labels=append(labels,"namespace")
				labels=append(labels,"service")

				klog.Infof("Metrics labels %v", labels)
        			gv:=prometheus.NewGaugeVec( prometheus.GaugeOpts{
					Namespace: "Service",
                			Name: m.Name ,
					Help: "SIA Metrics THresholds",
                		}, labels)
                                if err := PromRegister.Register(gv) ; err != nil {
                                        klog.Errorf("Error registering Collector \"%s\"", err)
					if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
        					gv = are.ExistingCollector.(*prometheus.GaugeVec)
    					} else {
						PromRegister.Unregister(gv)
						PromRegister.Register(gv)
    					}	
                                } else {
                                        klog.Info("Collector registered")
                                }

        			gv.With(annotations).Set(m.Value)
				klog.Infof("Set Collector %v", annotations)
				//spew.Dump( gv )
				/*if err := PromRegister.Register(col) ; err != nil {
					klog.Errorf("Error registering Collector \"%s\"", err)
				} else {
					klog.Info("Collector registered")
				}*/		

                        }
                }
        }
}

func inited_cache_get() (){
	SericesMetrics=nil
	klog.Infof("Waiting for cache ")
	for  {
		SericesMetrics=ServiceCacheAtomic.Load().(Services)
		if ( SericesMetrics != nil ) {
			break;
		}

	}
	if(appConf.LOGLEVEL == "DEBUG" ) {
				klog.Info("*****************DEBUG**********************") 
				spew.Dump(SericesMetrics) 
				klog.Info("*****************DEBUG**********************")
	}
}

func start_prometheus() {
        inited_cache_get()

	
	klog.Info("Starting Prometheus instance")
	PromRegister = prometheus.NewRegistry()
	x:= promhttp.HandlerFor(PromRegister, promhttp.HandlerOpts{})
        handler = &x
	MakeServicesCollectors()
        MakePodsCollectors()
	klog.Infof("Starting Promethues Channel")
	go RefreshPrometheus()
	klog.Infof("REGISTRY %v",prometheus.DefaultRegisterer )
	server = &http.Server{
	Addr:           ":7777",
	Handler:        *handler,
	ReadTimeout:    10 * time.Second,
	WriteTimeout:   10 * time.Second,
	MaxHeaderBytes: 1 << 20, }
	server.ListenAndServe()
	klog.Infof("REGISTRY %v",prometheus.DefaultRegisterer )
}

func RefreshPrometheus() {
	for {
		klog.Info("Wait on prometheus channel")
		SericesMetrics=ServiceCacheAtomic.Load().(Services)
		PodsMetrics=PodCacheAtomic.Load().(Pods)
		if(SericesMetrics == nil ){
			klog.Warning("Service Metric Cache is null")
		} else {
			klog.Info("Service Metric Cache refreshedi ... removing register")
			PromRegister = prometheus.NewRegistry()	
			x := promhttp.HandlerFor(PromRegister, promhttp.HandlerOpts{})
			handler = &x
			MakeServicesCollectors()
        		MakePodsCollectors()
			server.Handler=*handler
		}
		time.Sleep(time.Duration(appConf.CacheRefreshIntervall ) * time.Second)	
	}
}

