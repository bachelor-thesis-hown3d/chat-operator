package util

import (
	chatv1alpha1 "github.com/bachelor-thesis-hown3d/chat-operator/api/chat.accso.de/v1alpha1"
)

func DefaultLabels(name string) map[string]string {
	return map[string]string{"rocketchat": name}
}

// MergeLabels merges the toMerge map into the base map.
// If keys are the same, the value of the toMerge map are used.
func MergeLabels(base, toMerge map[string]string) map[string]string {
	if base == nil {
		return toMerge
	}
	for k, v := range toMerge {
		base[k] = v
	}
	return base
}

// HasDefaultLabels checks if the current rocket instance already has the default labels applied to it
func HasDefaultLabels(rocket *chatv1alpha1.Rocket) (exists bool) {
	currentLabels := rocket.Labels
	defaultLabels := DefaultLabels(rocket.Name)
	for key, value := range defaultLabels {
		currentValue, ok := currentLabels[key]
		// if key exists in currentlabels, check if value is the same
		if ok && currentValue == value {
			exists = true
		} else {
			exists = false
		}
	}
	return
}
