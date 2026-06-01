package access_hierarchy

import (
	"strconv"
	"time"

	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/metrics"
)

var (
	_ ruleHandler = (*handler)(nil)
)

type ruleHandler interface {
	Check(ctx eosc.IEntry) bool
}

type handler struct {
	a metrics.Metrics
	b metrics.Metrics

	appNodesPrefix           string
	nodeMetaPrefix           string
	nodeTargetsPrefix        string
	parentNodesPrefix        string
	parentGrantedNodesPrefix string

	metaModeField       string
	metaParentNodeField string

	modeExplicit     string
	modeFollowParent string
}

func (h *handler) getAppNodesPrefix() string {
	if h.appNodesPrefix == "" {
		return "app_groups:"
	}
	return h.appNodesPrefix
}

func (h *handler) getNodeMetaPrefix() string {
	if h.nodeMetaPrefix == "" {
		return "group_meta:"
	}
	return h.nodeMetaPrefix
}

func (h *handler) getNodeTargetsPrefix() string {
	if h.nodeTargetsPrefix == "" {
		return "group_resources:"
	}
	return h.nodeTargetsPrefix
}

func (h *handler) getParentNodesPrefix() string {
	if h.parentNodesPrefix == "" {
		return "tenant_groups:"
	}
	return h.parentNodesPrefix
}

func (h *handler) getParentGrantedNodesPrefix() string {
	if h.parentGrantedNodesPrefix == "" {
		return "tenant_granted_groups:"
	}
	return h.parentGrantedNodesPrefix
}

func (h *handler) getMetaModeField() string {
	if h.metaModeField == "" {
		return "mode"
	}
	return h.metaModeField
}

func (h *handler) getMetaParentNodeField() string {
	if h.metaParentNodeField == "" {
		return "owner_tenant_id"
	}
	return h.metaParentNodeField
}

func (h *handler) getModeExplicit() string {
	if h.modeExplicit == "" {
		return "explicit_resources"
	}
	return h.modeExplicit
}

func (h *handler) getModeFollowParent() string {
	if h.modeFollowParent == "" {
		return "follow_tenant"
	}
	return h.modeFollowParent
}

func (h *handler) Check(entry eosc.IEntry) bool {
	appID := h.a.Metrics(entry)
	targetID := h.b.Metrics(entry)
	if appID == "" || targetID == "" {
		return false
	}

	appNodes, has := customerVar.GetAll(h.getAppNodesPrefix() + appID)
	if !has || len(appNodes) == 0 {
		return false
	}

	now := time.Now().UnixMilli()

	visitedNodes := make(map[string]bool)
	visitedParents := make(map[string]bool)

	for nodeID, v := range appNodes {
		timestamp, _ := strconv.ParseInt(v, 10, 64)
		if timestamp > 0 && timestamp <= now {
			continue
		}

		if h.checkNodeHasTarget(nodeID, targetID, now, visitedNodes, visitedParents) {
			return true
		}
	}

	return false
}

func (h *handler) checkNodeHasTarget(nodeID string, targetID string, now int64, visitedNodes map[string]bool, visitedParents map[string]bool) bool {
	if visitedNodes[nodeID] {
		return false
	}
	visitedNodes[nodeID] = true

	meta, has := customerVar.GetAll(h.getNodeMetaPrefix() + nodeID)
	if !has {
		return false
	}

	mode := meta[h.getMetaModeField()]
	parentID := meta[h.getMetaParentNodeField()]

	if mode == h.getModeExplicit() {
		targets, hasRes := customerVar.GetAll(h.getNodeTargetsPrefix() + nodeID)
		if !hasRes {
			return false
		}
		val, exists := targets[targetID]
		if exists {
			timestamp, _ := strconv.ParseInt(val, 10, 64)
			if timestamp <= 0 || timestamp > now {
				return true
			}
		}
		return false
	}

	if mode == h.getModeFollowParent() && parentID != "" {
		return h.checkParentHasTarget(parentID, targetID, now, visitedNodes, visitedParents)
	}

	return false
}

func (h *handler) checkParentHasTarget(parentID string, targetID string, now int64, visitedNodes map[string]bool, visitedParents map[string]bool) bool {
	if visitedParents[parentID] {
		return false
	}
	visitedParents[parentID] = true

	parentNodes, hasParentNodes := customerVar.GetAll(h.getParentNodesPrefix() + parentID)
	if hasParentNodes {
		for ownedNodeID, tv := range parentNodes {
			timestamp, _ := strconv.ParseInt(tv, 10, 64)
			if timestamp > 0 && timestamp <= now {
				continue
			}

			meta, hasMeta := customerVar.GetAll(h.getNodeMetaPrefix() + ownedNodeID)
			if !hasMeta {
				continue
			}
			if meta[h.getMetaModeField()] != h.getModeExplicit() {
				continue
			}

			targets, hasRes := customerVar.GetAll(h.getNodeTargetsPrefix() + ownedNodeID)
			if !hasRes {
				continue
			}

			val, exists := targets[targetID]
			if exists {
				rt, _ := strconv.ParseInt(val, 10, 64)
				if rt <= 0 || rt > now {
					return true
				}
			}
		}
	}

	grantedNodes, hasGrantedNodes := customerVar.GetAll(h.getParentGrantedNodesPrefix() + parentID)
	if hasGrantedNodes {
		for grantedNodeID, gv := range grantedNodes {
			timestamp, _ := strconv.ParseInt(gv, 10, 64)
			if timestamp > 0 && timestamp <= now {
				continue
			}

			if h.checkNodeHasTarget(grantedNodeID, targetID, now, visitedNodes, visitedParents) {
				return true
			}
		}
	}

	if _, hasMeta := customerVar.GetAll(h.getNodeMetaPrefix() + parentID); hasMeta {
		if h.checkNodeHasTarget(parentID, targetID, now, visitedNodes, visitedParents) {
			return true
		}
	}

	return false
}
