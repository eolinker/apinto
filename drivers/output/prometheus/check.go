package prometheus

import (
	"sort"
	"strings"
)

// checkScopesChange 检查scopes配置是否有改变
func checkScopesChange(oldScopes, newScopes []string) bool {
	if len(oldScopes) != len(oldScopes) {
		return true
	}

	sort.Slice(oldScopes, func(i, j int) bool {
		return oldScopes[i] < oldScopes[j]
	})
	sort.Slice(newScopes, func(i, j int) bool {
		return newScopes[i] < newScopes[j]
	})

	for i, scope := range oldScopes {
		if newScopes[i] != scope {
			return true
		}
	}

	return false
}

func checkPathChange(oldPath, newPath string) bool {
	oldPath = "/" + strings.Trim(oldPath, "/")
	newPath = "/" + strings.Trim(newPath, "/")
	if oldPath != newPath {
		return true
	}
	return false
}

func checkMetricConfigChange(oldMC, newMC *MetricConfig) bool {
	if oldMC.Collector != newMC.Collector {
		return true
	} else if oldMC.Objectives != newMC.Objectives {
		return true
	} else if oldMC.Description != newMC.Description {
		return true
	}
	if len(oldMC.Labels) != len(newMC.Labels) {
		return true
	} else {
		for i, oldLabel := range oldMC.Labels {
			if newMC.Labels[i] != oldLabel {
				return true
			}
		}
	}

	return false
}
