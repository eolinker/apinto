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

// ruleHandler 定义校验规则接口
type ruleHandler interface {
	Check(ctx eosc.IEntry) bool
}

// handler 校验处理器，持有提取 AppID 和 ResourceID 的指标提取器
type handler struct {
	a metrics.Metrics // a 对应 AppID 的提取表达式
	b metrics.Metrics // b 对应 ResourceID 的提取表达式
}

// Check 实现核心拦截逻辑，判断当前请求提取出的 AppID 对提取出的 ResourceID 是否具有访问权限
func (h *handler) Check(entry eosc.IEntry) bool {
	// 1. 从当前 HTTP 请求中解析出 AppID 与 ResourceID
	appID := h.a.Metrics(entry)
	resourceID := h.b.Metrics(entry)
	if appID == "" || resourceID == "" {
		return false
	}

	// 2. 通过 ICustomerVar 获取当前 App 绑定的所有初始资源组 (app_groups:appID)
	appGroups, has := customerVar.GetAll("app_groups:" + appID)
	if !has || len(appGroups) == 0 {
		return false
	}

	now := time.Now().UnixMilli()

	// 初始化 DFS 递归中用于防环 (Cycle Detection) 的访问记录集合
	visitedGroups := make(map[string]bool)
	visitedTenants := make(map[string]bool)

	// 3. 遍历 App 绑定的每个资源组，只要有一个组满足资源访问权限，即放行
	for groupID, v := range appGroups {
		// 校验该组的绑定关系是否已过期
		timestamp, _ := strconv.ParseInt(v, 10, 64)
		if timestamp > 0 && timestamp <= now {
			continue
		}

		// 深入判断该资源组中是否包含了目标资源 ResourceID（开始级联递归深度遍历）
		if h.checkGroupHasResource(groupID, resourceID, now, visitedGroups, visitedTenants) {
			return true
		}
	}

	return false
}

// checkGroupHasResource 递归检查指定的资源组 groupID 是否含有目标资源 resourceID
func (h *handler) checkGroupHasResource(groupID string, resourceID string, now int64, visitedGroups map[string]bool, visitedTenants map[string]bool) bool {
	// 防止资源组依赖成环，导致死循环
	if visitedGroups[groupID] {
		return false
	}
	visitedGroups[groupID] = true
	// defer func() {
	// 	visitedGroups[groupID] = false // 递归回溯，恢复状态
	// }()

	// 1. 获取资源组的元数据 (group_meta:groupID)
	meta, has := customerVar.GetAll("group_meta:" + groupID)
	if !has {
		return false
	}

	mode := meta["mode"]
	ownerTenantID := meta["owner_tenant_id"]

	// 2. 如果是 显式绑定模式 (explicit_resources)
	if mode == "explicit_resources" {
		// 直接查询该组下绑定的资源集 (group_resources:groupID)
		resources, hasRes := customerVar.GetAll("group_resources:" + groupID)
		if !hasRes {
			return false
		}
		val, exists := resources[resourceID]
		if exists {
			timestamp, _ := strconv.ParseInt(val, 10, 64)
			if timestamp <= 0 || timestamp > now {
				return true // 资源存在且未过期
			}
		}
		return false
	}

	// 3. 如果是 跟随租户模式 (follow_tenant)
	// 此时该组包含的所有资源，等价于它的所有者租户 (ownerTenantID) 的所有资源
	if mode == "follow_tenant" && ownerTenantID != "" {
		// 递归向上：查询该租户是否拥有的此资源（考虑了多级继承关系）
		return h.checkTenantHasResource(ownerTenantID, resourceID, now, visitedGroups, visitedTenants)
	}

	return false
}

// checkTenantHasResource 递归检查指定的租户 tenantID 是否拥有目标资源 resourceID
func (h *handler) checkTenantHasResource(tenantID string, resourceID string, now int64, visitedGroups map[string]bool, visitedTenants map[string]bool) bool {
	// 防止租户级联依赖成环，导致死循环
	if visitedTenants[tenantID] {
		return false
	}
	visitedTenants[tenantID] = true
	// defer func() {
	// 	visitedTenants[tenantID] = false // 递归回溯，恢复状态
	// }()

	// 一个租户拥有的全部资源包含两个来源：

	// 来源一：自己直接拥有的所有实体组所显式绑定的资源
	tenantGroups, hasTenantGroups := customerVar.GetAll("tenant_groups:" + tenantID)
	if hasTenantGroups {
		for ownedGroupID, tv := range tenantGroups {
			timestamp, _ := strconv.ParseInt(tv, 10, 64)
			if timestamp > 0 && timestamp <= now {
				continue // 关系过期
			}

			// 【核心修复】：由于我们在获取本租户资源时，防止子组又是跟随租户模式造成 (租户 -> 组 -> 租户) 无限循环，
			// 我们需要先读取子组的元数据，过滤掉所有跟随模式的组，只获取显式绑定模式的资源。
			meta, hasMeta := customerVar.GetAll("group_meta:" + ownedGroupID)
			if !hasMeta {
				continue
			}
			if meta["mode"] != "explicit_resources" {
				continue // 过滤掉跟随该租户模式的子组，防止无限递归
			}

			// 获取该组绑定的资源，判断是否包含目标资源
			resources, hasRes := customerVar.GetAll("group_resources:" + ownedGroupID)
			if !hasRes {
				continue
			}

			val, exists := resources[resourceID]
			if exists {
				rt, _ := strconv.ParseInt(val, 10, 64)
				if rt <= 0 || rt > now {
					return true // 找到资源，允许通过
				}
			}
		}
	}

	// 来源二：租户继承上级分配/授权给该租户的资源组资源（渠道授权组，来自上一级租户）
	// 例如上一级租户授权了某个资源组（可以是显式，也可以是跟随上级的组）给该租户使用，
	// 该租户即可递归继承该授权资源组下的所有可用资源。
	grantedGroups, hasGrantedGroups := customerVar.GetAll("tenant_granted_groups:" + tenantID)
	if hasGrantedGroups {
		for grantedGroupID, gv := range grantedGroups {
			timestamp, _ := strconv.ParseInt(gv, 10, 64)
			if timestamp > 0 && timestamp <= now {
				continue // 授权关系过期
			}

			// 【级联继承】：递归进入授权组，解析授权组是显式绑定还是跟随上级的上级
			if h.checkGroupHasResource(grantedGroupID, resourceID, now, visitedGroups, visitedTenants) {
				return true
			}
		}
	}

	return false
}
