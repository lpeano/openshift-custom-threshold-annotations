# Parametric Prometheus thresholds based on application configurations
Openshift plugin that allows you to parameterize the Alert Manager Prometheus thresholds based on annotations on Pods and Services.
# How does it work
It allows you to use the following annotations to set thresholds on existing prometheus metrics.

	oc annotate svc/<service-name> -n <namespace-name> sia.io/thresholds_config=true
	oc annotate svc/<service-name> -n <namespace-name> sia.io/thresholds='[{"Name":"<metricnam>","Value":<threshold-value>},...]'

First annotation activate metric thresholds injection, second establish metric name and value on witch thresholds are injected.

# Example
In the following example whe have 
