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

type Pod struct {
	Pod string
	Namespace string
        resourceVersion string
	metrics Metrics 
}



// Global Pod Cache 
var PodCache = Pods{}
var PodCacheAtomic atomic.Value


// Map of thresholds where key is pod name
type Pods map[string]map[string]Pod

func (p Pods) Addpod(Pd Pod)(bool, string) {
	var reason string
	
	if _,ok:=p[Pd.Namespace][Pd.Pod]; ok {
		reason="Pod allready exist"		
		return false, reason
	} else {
		reason="Added"
		klog.Infof("Adding pod %s of namespace %s", Pd.Pod,Pd.Namespace)
		if p[Pd.Namespace] == nil {
			klog.Infof("Namespace (%s) doesn't exixt",Pd.Namespace)
			p[Pd.Namespace]=make(map[string]Pod)
		}
		p[Pd.Namespace][Pd.Pod]=Pd
		return true, reason
	}
}

func (p Pod) SetMetricAnnotation(Met Metrics) {
	p.metrics=Met
}

// Pod Watcher
func (s Pods ) Pod_Watcher(clientset *kubernetes.Clientset) {
                timeout := int64(1800)
		namespace := ""
                for{
		if appConf.Namespace{
			namespace=appConf.Namespace
		}
               	PodWatcher, err := clientset.CoreV1().Pods(namespace).Watch(metav1.ListOptions{TimeoutSeconds: &timeout, Watch: true, })
                if err != nil {
                        klog.Error("Error")
                }
                ch := PodWatcher.ResultChan()


                for event := range ch {
                        svcev, ok := event.Object.(*v1.Pod)
                        if !ok {
                                klog.Errorf("unexpected type %v \n", svcev)
                        } else {
                                klog.Infof("Events Namespace: %s PodName: %s ResourveVersion: %s EventType: %s \n",svcev.GetNamespace(),svcev.GetName(),svcev.GetResourceVersion(),event.Type )
				s.ManageEvent(svcev, string(event.Type))
				
                        }

                }}

}

func (s Pods) ManageEvent( Pd *v1.Pod,event string) {
	switch event {
		case "MODIFIED":
			if x,found:=s[Pd.GetNamespace()][Pd.GetName()]; found {
				klog.Infof("Event : %s - Pd %s present",event,x.Pod,found)
				annotations:=Pd.GetAnnotations()
				found, v:=is_annotated(annotations)
				if(!found) {
					// If annotation is missing pod must be removed from chahce
					klog.Infof("Removing in namespace %s pod %s from cache" ,Pd.GetNamespace(),Pd.GetName())
					delete(s[Pd.GetNamespace()],Pd.GetName())
					Use(v)
				} else {
					klog.Infof("Pod in namespace %s podname %s modified" ,Pd.GetNamespace(),Pd.GetName())
					delete(s[Pd.GetNamespace()],Pd.GetName())
					pd:=Buildpod(Pd)
                                	r, reason := s.Addpod(pd)
					klog.Infof("Pod modifies return \"%s\" dueue to \"%s\"", r,reason)
					
				}		
			 	
			} else {
				klog.Warningf("Event : %s - Pd %s not present must do CACHEADD",event,x.Pod,found)
				pd:=Buildpod(Pd)
				r, reason := s.Addpod(pd)
				klog.Infof("Pod add return \"%s\" dueue to \"%s\"", r,reason)
				//Use(v)
				// TO HERE
			}
                        if(appConf.LOGLEVEL == "DEBUG" ) {spew.Dump(s)}
		case "ADDED":
			if x,found:=s[Pd.GetNamespace()][Pd.GetName()]; found {
				klog.Warningf("Event : %s - Pd %s allready present",event,x.Pod,found)
				pd := Pod{Pod: Pd.GetName(), Namespace: Pd.GetNamespace(), resourceVersion: Pd.GetResourceVersion()}
				annotations:=Pd.GetAnnotations()
				found, v:=is_annotated(annotations)
				if (found) {
					r, reason := s.Addpod(pd)
					klog.Infof("Pod add return \"%s\" dueue to \"%s\"", r,reason)
				} else {
					klog.Infof("Pod %s.%s not annotated IGNORING",  Pd.GetName(), Pd.GetNamespace())
				}
				Use(v)
			} else {
				annotations:=Pd.GetAnnotations()
				found, v:=is_annotated(annotations)
				if ( found ) {
					//Adding new pod to cache
					pd := Pod{Pod: Pd.GetName(), Namespace: Pd.GetNamespace(), resourceVersion: Pd.GetResourceVersion()}
                                	f, an := get_annotated(annotations)
					if( f ) {
                                        err:= json.Unmarshal( []byte(an) ,&pd.metrics)
                                        	if err != nil {
                                                	klog.Errorf("Error marshaling metrics \"%s\"", err)
                                        	} else {
                                                	pd.SetMetricAnnotation(metrics)
                                        	}
					}
					r, reason := s.Addpod(pd)
					klog.Infof("Pod add return \"%s\" dueue to \"%s\"", r,reason)
				}
				Use(v,found)
			}
                        if(appConf.LOGLEVEL == "DEBUG" ) {spew.Dump(s)}

                case "DELETED":
                         if x,found:=s[Pd.GetNamespace()][Pd.GetName()]; found {
                                klog.Infof("Event : %s - Pod %s present",event,x.Pod,found)
                                klog.Infof("Pod in namespace %s podname %s deleted" ,Pd.GetNamespace(),Pd.GetName())
                                delete(s[Pd.GetNamespace()],Pd.GetName())
                        }
                        if(appConf.LOGLEVEL == "DEBUG" ) {spew.Dump(s)}

		default:
			klog.Warningf("Event (%s) not supported",event)
	}
	klog.Info("Storing atomic cache")
	PodCacheAtomic.Store(s)
}

func Buildpod (Srv *v1.Pod) (Pod){
	// FROM HERE	
	pd := Pod{Pod: Srv.GetName(), Namespace: Srv.GetNamespace(), resourceVersion: Srv.GetResourceVersion()}
	annotations:=Srv.GetAnnotations()
	found, v:=is_annotated(annotations)
	if (found) {
                f, an := get_annotated(annotations)
                if( f ) {
                        err:= json.Unmarshal( []byte(an) ,&pd.metrics)
                               if err != nil {
                                  klog.Errorf("Error marshaling metrics \"%s\"", err)
                               } else {
                                  pd.SetMetricAnnotation(metrics)
                               }
                        }
		} else {
			klog.Infof("Pod %s.%s not annotated IGNORING",  Srv.GetName(), Srv.GetNamespace())
		}
	Use(v)
	return pd
	// TO HERE
}

// Get List Of Pod Configs
func (sc Pods)Get_pods(clientset *kubernetes.Clientset) {

                pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
                if err != nil {
                        panic(err.Error())
                }
                klog.Infof("There are %d pods in the cluster\n", len(pods.Items))

                for index, pd := range  pods.Items {
                        annotations:=pd.GetAnnotations()
                        found, v:=is_annotated(annotations)
		        klog.Infof("ANNOTATED: %v ", found)		
                        if( found ) {
                        	klog.Infof("Pod %d %s ...",index,pd.GetName() )
                                f, an := get_annotated(annotations)
                                if( ! f ) {
                                        klog.Warningf("No metric annotation present for pod %s", pd.GetName())
                                } else {
                                        klog.Infof("Annotation in %s pod \n\t [ %s ]", pd.GetName(), an)
                                        pd:= Pod{Pod:pd.GetName(), resourceVersion: pd.GetResourceVersion(), Namespace: pd.GetNamespace()}
                                        err:= json.Unmarshal( []byte(an) ,&pd.metrics)
                                        if err != nil {
                                                klog.Errorf("Error marshaling metrics \"%s\"", err)
                                        } else {
                                                pd.SetMetricAnnotation(metrics)
                                        }
                                        //spew.Dump(pd)
                                        sc.Addpod(pd)
                                        if(appConf.LOGLEVEL == "DEBUG" ) {spew.Dump(sc)}
                                }
                        } else {
				klog.Infof(" SERVICE %s.%s IGNORED",pd.GetName(),pd.GetNamespace())
			}
                        Use(v)

                }
		PodCacheAtomic.Store( sc );		
		if(appConf.LOGLEVEL == "ATOMICCACHEDEBUG" ) {spew.Dump(PodCacheAtomic)}
}
