package main
import (
	"k8s.io/klog"
	"strconv"
)
type metric_annotation struct {
        Description string `json:",omitempty"`
        Name  string
        Value float64
        AdditionalLabels []OptionalLabel `json:",omitempty"`
}

type Metrics []metric_annotation


var metrics []metric_annotation

type OptionalLabel struct {
        Name string
        Value string
}

func is_annotated(annotations map[string]string) (bool , string) {

        v , found := annotations[appConf.AnnotationFlag]
        x, err :=  strconv.ParseBool(v)
        Use(err)
        if ( found && x==true ) {
                klog.V(0).Infof( "Has Annotation ... with %s",v)
                tconfig, found := annotations[appConf.AnnotationFlag]
                if (found ) {
                        klog.V(0).Infof( "thresholds_config %s", tconfig)
                }
                return true, v
        } else {
                return false, v
        }


}

func get_annotated(annotations map[string]string) (bool , string) {

        v , found := annotations[appConf.AnnotationNameThreshold]
        if ( found &&  v=="true" ) {
                klog.V(0).Infof( "level2 Has Annotation ... with %s",v)
                tconfig, found := annotations[appConf.AnnotationFlag]
                if (found ) {
                        klog.V(0).Infof( "thresholds_config %s", tconfig)
                }
        }

        return found, v

}

