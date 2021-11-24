module github.com/crossplane-contrib/provider-influxdb

go 1.16

require (
	github.com/crossplane/crossplane-runtime v0.15.1
	github.com/crossplane/crossplane-tools v0.0.0-20210320162312-1baca298c527
	github.com/influxdata/influxdb-client-go/v2 v2.5.1
	github.com/pkg/errors v0.9.1
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	k8s.io/apimachinery v0.21.3
	k8s.io/client-go v0.21.3
	k8s.io/utils v0.0.0-20211116205334-6203023598ed
	sigs.k8s.io/controller-runtime v0.9.6
	sigs.k8s.io/controller-tools v0.6.2
)
