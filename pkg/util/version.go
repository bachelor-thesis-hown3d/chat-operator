package util

import (
	"fmt"

	controllerruntime "sigs.k8s.io/controller-runtime"
)

// Version is the version of the controller, set on build time
var Version string

// PrintVersion prints the current version
func PrintVersion() {
	controllerruntime.Log.Info(fmt.Sprintf("Running on Version %v", Version))
}
