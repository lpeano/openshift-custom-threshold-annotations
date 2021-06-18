package main
import (
	"encoding/json"
	"k8s.io/klog"
        metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
        "k8s.io/api/core/v1"
	"github.com/davecgh/go-spew/spew"
        "k8s.io/client-go/kubernetes"
	"sync/atomic"

	)


type Service struct {
	Service string
	Namespace string
        resourceVersion string
	metrics Metrics 
}

// Global Service Cache 
var ServiceCache = Services{}
var ServiceCacheAtomic atomic.Value


// Map of thresholds where key is service name
type Services map[string]map[string]Service

func (s Services) Addservice(Srv Service)(bool, string) {
	var reason string
	
	if _,ok:=s[Srv.Namespace][Srv.Service]; ok {
		reason="Service allready exist"		
		return false, reason
	} else {
		reason="Added"
		klog.Infof("Adding service %s of namespace %s", Srv.Service,Srv.Namespace)
		if s[Srv.Namespace] == nil {
			klog.Infof("Namespace (%s) doesn't exixt",Srv.Namespace)
			s[Srv.Namespace]=make(map[string]Service)
		}
		s[Srv.Namespace][Srv.Service]=Srv
		return true, reason
	}
}
//QUI
// Set new service instance
func  (srv Service) SetService(ServiceName string, resourceVersion string) {
	srv.Service= ServiceName
	srv.resourceVersion= resourceVersion	
}

func (s Service) SetMetricAnnotation(Met Metrics) {
	s.metrics=Met
}

// Service Watcher
func (s Services ) Service_Watcher(clientset *kubernetes.Clientset) {
                timeout := int64(1800)
		namespace := ""
                for{
		if appConf.Namespace{
                        namespace=appConf.Namespace
                }

                ServiceWatcher, err := clientset.CoreV1().Services(namespace).Watch(metav1.ListOptions{TimeoutSeconds: &timeout, Watch: true, })
                if err != nil {
                        klog.Error("Error")
                }
                ch := ServiceWatcher.ResultChan()


                for event := range ch {
                        svcev, ok := event.Object.(*v1.Service)
                        if !ok {
                                klog.Errorf("unexpected type %v \n", svcev)
                        } else {
                                klog.Infof("Events Namespace: %s ServiceName: %s ResourveVersion: %s EventType: %s \n",svcev.GetNamespace(),svcev.GetName(),svcev.GetResourceVersion(),event.Type )
				s.ManageEvent(svcev, string(event.Type))
				
                        }

                }}

}

func (s Services) ManageEvent( Srv *v1.Service,event string) {
	switch event {
		case "MODIFIED":
			if x,found:=s[Srv.GetNamespace()][Srv.GetName()]; found {
				klog.Infof("Event : %s - Srv %s present",event,x.Service,found)
				annotations:=Srv.GetAnnotations()
				found, v:=is_annotated(annotations)
				if(!found) {
					// If annotation is missing service must be removed from chahce
					klog.Infof("Removing in namespace %s service %s from cache" ,Srv.GetNamespace(),Srv.GetName())
					delete(s[Srv.GetNamespace()],Srv.GetName())
					Use(v)
				} else {
					klog.Infof("Service in namespace %s servicename %s modified" ,Srv.GetNamespace(),Srv.GetName())
					delete(s[Srv.GetNamespace()],Srv.GetName())
					srv:=Buildservice(Srv)
                                	r, reason := s.Addservice(srv)
					klog.Infof("Service modifies return \"%s\" dueue to \"%s\"", r,reason)
					
				}		
			 	
			} else {
				klog.Warningf("Event : %s - Srv %s not present must do CACHEADD",event,x.Service,found)
				srv:=Buildservice(Srv)
				r, reason := s.Addservice(srv)
				klog.Infof("Service add return \"%s\" dueue to \"%s\"", r,reason)
				//Use(v)
				// TO HERE
			}
                        if(appConf.LOGLEVEL == "DEBUG" ) {spew.Dump(s)}
		case "ADDED":
			if x,found:=s[Srv.GetNamespace()][Srv.GetName()]; found {
				klog.Warningf("Event : %s - Srv %s allready present",event,x.Service,found)
				srv := Service{Service: Srv.GetName(), Namespace: Srv.GetNamespace(), resourceVersion: Srv.GetResourceVersion()}
				annotations:=Srv.GetAnnotations()
				found, v:=is_annotated(annotations)
				if (found) {
					r, reason := s.Addservice(srv)
					klog.Infof("Service add return \"%s\" dueue to \"%s\"", r,reason)
				} else {
					klog.Infof("Service %s.%s not annotated IGNORING",  Srv.GetName(), Srv.GetNamespace())
				}
				Use(v)
			} else {
				annotations:=Srv.GetAnnotations()
				found, v:=is_annotated(annotations)
				if ( found ) {
					//Adding new service to cache
					srv := Service{Service: Srv.GetName(), Namespace: Srv.GetNamespace(), resourceVersion: Srv.GetResourceVersion()}
                                	f, an := get_annotated(annotations)
					if( f ) {
                                        err:= json.Unmarshal( []byte(an) ,&srv.metrics)
                                        	if err != nil {
                                                	klog.Errorf("Error marshaling metrics \"%s\"", err)
                                        	} else {
                                                	srv.SetMetricAnnotation(metrics)
                                        	}
					}
					r, reason := s.Addservice(srv)
					klog.Infof("Service add return \"%s\" dueue to \"%s\"", r,reason)
				}
				Use(v,found)
			}
                        if(appConf.LOGLEVEL == "DEBUG" ) {spew.Dump(s)}
		default:
			klog.Warningf("Event (%s) not supported",event)
	}
	klog.Info("Storing atomic cache")
	ServiceCacheAtomic.Store(s)
}

func Buildservice (Srv *v1.Service) (Service){
	// FROM HERE	
	srv := Service{Service: Srv.GetName(), Namespace: Srv.GetNamespace(), resourceVersion: Srv.GetResourceVersion()}
	annotations:=Srv.GetAnnotations()
	found, v:=is_annotated(annotations)
	if (found) {
                f, an := get_annotated(annotations)
                if( f ) {
                        err:= json.Unmarshal( []byte(an) ,&srv.metrics)
                               if err != nil {
                                  klog.Errorf("Error marshaling metrics \"%s\"", err)
                               } else {
                                  srv.SetMetricAnnotation(metrics)
                               }
                        }
		} else {
			klog.Infof("Service %s.%s not annotated IGNORING",  Srv.GetName(), Srv.GetNamespace())
		}
	Use(v)
	return srv
	// TO HERE
}

// Get List Of Service Configs
func (sc Services)Get_services(clientset *kubernetes.Clientset) {

                services, err := clientset.CoreV1().Services("").List(metav1.ListOptions{})
                if err != nil {
                        panic(err.Error())
                }
                klog.Infof("There are %d services in the cluster\n", len(services.Items))

                for index, srv := range  services.Items {
                        annotations:=srv.GetAnnotations()
                        found, v:=is_annotated(annotations)
		        klog.Infof("ANNOTATED: %v ", found)		
                        if( found ) {
                        	klog.Infof("Service %d %s ...",index,srv.GetName() )
                                f, an := get_annotated(annotations)
                                if( ! f ) {
                                        klog.Warningf("No metric annotation present for service %s", srv.GetName())
                                } else {
                                        klog.Infof("Annotation in %s service \n\t [ %s ]", srv.GetName(), an)
                                        srv:= Service{Service:srv.GetName(), resourceVersion: srv.GetResourceVersion(), Namespace: srv.GetNamespace()}
                                        err:= json.Unmarshal( []byte(an) ,&srv.metrics)
                                        if err != nil {
                                                klog.Errorf("Error marshaling metrics \"%s\"", err)
                                        } else {
                                                srv.SetMetricAnnotation(metrics)
                                        }
                                        //spew.Dump(srv)
                                        sc.Addservice(srv)
                                        if(appConf.LOGLEVEL == "DEBUG" ) {spew.Dump(sc)}
                                }
                        } else {
				klog.Infof(" SERVICE %s.%s IGNORED",srv.GetName(),srv.GetNamespace())
			}
                        Use(v)

                }
		ServiceCacheAtomic.Store( sc );		
		if(appConf.LOGLEVEL == "ATOMICCACHEDEBUG" ) {spew.Dump(ServiceCacheAtomic)}
}

