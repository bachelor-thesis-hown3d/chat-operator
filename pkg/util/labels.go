package util

func DefaultLabels(name string) map[string]string {
	return map[string]string{"rocketchat": name}
}

func MergeLabels(base map[string]string, toMerge map[string]string) map[string]string {
	for k, v := range toMerge {
		base[k] = v
	}
	return base
}
