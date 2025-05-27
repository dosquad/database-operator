package helper

import (
	"errors"
	"fmt"
	"os"
	"reflect"

	dbov1 "github.com/dosquad/database-operator/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	ctrl "sigs.k8s.io/controller-runtime"
)

// loadFile is used from the mutex.Once to load the file.
func loadFile(configFile string) (*dbov1.DatabaseAccountControllerConfig, error) {
	scheme := runtime.NewScheme()
	if scheme == nil {
		return nil, errors.New("scheme not supplied to controller configuration loader")
	}

	content, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("could not read file at %s", configFile)
	}

	codecs := serializer.NewCodecFactory(scheme)

	newObj := &dbov1.DatabaseAccountControllerConfig{}

	// Regardless of if the bytes are of any external version,
	// it will be read successfully and converted into the internal version
	if err = runtime.DecodeInto(codecs.UniversalDecoder(), content, newObj); err != nil {
		return nil, errors.New("could not decode file into runtime.Object")
	}

	return newObj, nil
}

func LoadConfigFile(configFile string, options ctrl.Options) (
	ctrl.Options, *dbov1.DatabaseAccountControllerConfig, error,
) {
	var newObj *dbov1.DatabaseAccountControllerConfig
	{
		var err error
		newObj, err = loadFile(configFile)
		if err != nil {
			return options, newObj, err
		}
	}

	options = setLeaderElectionConfig(options, newObj)

	if options.Cache.SyncPeriod == nil && newObj.SyncPeriod != nil {
		options.Cache.SyncPeriod = newObj.SyncPeriod
	}

	// if len(options.Cache.DefaultNamespaces) == 0 && newObj.CacheNamespace != "" {
	// 	options.Cache.DefaultNamespaces = map[string]cache.Config{newObj.CacheNamespace: {}}
	// }

	if options.Metrics.BindAddress == "" && newObj.MetricsBindAddress != "" {
		options.Metrics.BindAddress = newObj.MetricsBindAddress
	}

	if options.HealthProbeBindAddress == "" && newObj.Health.HealthProbeBindAddress != "" {
		options.HealthProbeBindAddress = newObj.Health.HealthProbeBindAddress
	}

	if options.ReadinessEndpointName == "" && newObj.Health.ReadinessEndpointName != "" {
		options.ReadinessEndpointName = newObj.Health.ReadinessEndpointName
	}

	if options.LivenessEndpointName == "" && newObj.Health.LivenessEndpointName != "" {
		options.LivenessEndpointName = newObj.Health.LivenessEndpointName
	}

	// if options.WebhookServer == nil {
	// 	port := 0
	// 	if newObj.Webhook.Port != nil {
	// 		port = *newObj.Webhook.Port
	// 	}
	// 	options.WebhookServer = webhook.NewServer(webhook.Options{
	// 		Port:    port,
	// 		Host:    newObj.Webhook.Host,
	// 		CertDir: newObj.Webhook.CertDir,
	// 	})
	// }

	// if newObj.Controller != nil {
	// 	if options.Controller.CacheSyncTimeout == 0 && newObj.Controller.CacheSyncTimeout != nil {
	// 		options.Controller.CacheSyncTimeout = *newObj.Controller.CacheSyncTimeout
	// 	}

	// 	if len(options.Controller.GroupKindConcurrency) == 0 && len(newObj.Controller.GroupKindConcurrency) > 0 {
	// 		options.Controller.GroupKindConcurrency = newObj.Controller.GroupKindConcurrency
	// 	}
	// }

	return options, newObj, nil
}

func setLeaderElectionConfig(o ctrl.Options, obj *dbov1.DatabaseAccountControllerConfig) ctrl.Options {
	if obj.LeaderElection == nil {
		// The source does not have any configuration; noop
		return o
	}

	if !o.LeaderElection && obj.LeaderElection.LeaderElect != nil {
		o.LeaderElection = *obj.LeaderElection.LeaderElect
	}

	if o.LeaderElectionResourceLock == "" && obj.LeaderElection.ResourceLock != "" {
		o.LeaderElectionResourceLock = obj.LeaderElection.ResourceLock
	}

	if o.LeaderElectionNamespace == "" && obj.LeaderElection.ResourceNamespace != "" {
		o.LeaderElectionNamespace = obj.LeaderElection.ResourceNamespace
	}

	if o.LeaderElectionID == "" && obj.LeaderElection.ResourceName != "" {
		o.LeaderElectionID = obj.LeaderElection.ResourceName
	}

	if o.LeaseDuration == nil && !reflect.DeepEqual(obj.LeaderElection.LeaseDuration, metav1.Duration{}) {
		o.LeaseDuration = &obj.LeaderElection.LeaseDuration.Duration
	}

	if o.RenewDeadline == nil && !reflect.DeepEqual(obj.LeaderElection.RenewDeadline, metav1.Duration{}) {
		o.RenewDeadline = &obj.LeaderElection.RenewDeadline.Duration
	}

	if o.RetryPeriod == nil && !reflect.DeepEqual(obj.LeaderElection.RetryPeriod, metav1.Duration{}) {
		o.RetryPeriod = &obj.LeaderElection.RetryPeriod.Duration
	}

	return o
}
