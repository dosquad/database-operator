package controller

import (
	"strings"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func GetLogConstructor(mgr manager.Manager, obj runtime.Object) (func(*reconcile.Request) logr.Logger, error) {
	// Retrieve the GVK from the object we're reconciling
	// to prepopulate logger information, and to optionally generate a default name.
	gvk, err := apiutil.GVKForObject(obj, mgr.GetScheme())
	if err != nil {
		return nil, err
	}

	log := mgr.GetLogger().WithValues(
		"controller", strings.ToLower(gvk.Kind),
		"controllerGroup", gvk.Group,
		"controllerKind", gvk.Kind,
	)

	lowerCamelCaseKind := strings.ToLower(gvk.Kind[:1]) + gvk.Kind[1:]

	return func(req *reconcile.Request) logr.Logger {
		log := log
		if req != nil {
			log = log.WithValues(
				lowerCamelCaseKind, klog.KRef(req.Namespace, req.Name),
				"namespace", req.Namespace, "name", req.Name,
			)
		}
		return log
	}, nil
}
