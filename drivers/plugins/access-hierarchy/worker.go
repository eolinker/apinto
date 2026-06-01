package access_hierarchy

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/utils/response"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	"github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/metrics"
	"net/http"
)

var (
	_ eocontext.IFilter       = (*AccessHierarchy)(nil)
	_ http_context.HttpFilter = (*AccessHierarchy)(nil)
	_ eosc.IWorker            = (*AccessHierarchy)(nil)
)

type AccessHierarchy struct {
	drivers.WorkerBase
	rules              []ruleHandler
	response           response.IResponse
}

func (w *AccessHierarchy) Start() error {
	return nil
}

func (w *AccessHierarchy) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	config, err := assert(conf)
	if err != nil {
		return err
	}
	err = Check(config, workers)
	if err != nil {
		return err
	}
	iResponse, handlers := w.parseConfig(config)
	w.response = iResponse
	w.rules = handlers
	return nil
}

func (w *AccessHierarchy) Stop() error {
	return nil
}

func (w *AccessHierarchy) Destroy() {
}

func (w *AccessHierarchy) CheckSkill(skill string) bool {
	return http_context.FilterSkillName == skill
}

func assert(v interface{}) (*Config, error) {
	cfg, ok := v.(*Config)
	if !ok {
		return nil, eosc.ErrorConfigType
	}
	return cfg, nil
}

var (
	defaultResponse = response.Parse(&response.Response{
		StatusCode:  http.StatusForbidden,
		ContentType: "text/plain",
		Charset:     "utf-8",
		Headers:     nil,
		Body:        http.StatusText(http.StatusForbidden),
	})
)

func (w *AccessHierarchy) newHandler(a, b string, hCfg *HierarchyConfig) ruleHandler {
	am := metrics.Parse(a)
	bm := metrics.Parse(b)
	if am == nil || bm == nil {
		return nil
	}

	h := &handler{
		a: am,
		b: bm,

		appNodesPrefix:           "app_groups:",
		nodeMetaPrefix:           "group_meta:",
		nodeTargetsPrefix:        "group_resources:",
		parentNodesPrefix:        "tenant_groups:",
		parentGrantedNodesPrefix: "tenant_granted_groups:",

		metaModeField:        "mode",
		metaParentNodeField: "owner_tenant_id",

		modeExplicit:     "explicit_resources",
		modeFollowParent: "follow_tenant",
	}

	if hCfg != nil {
		if hCfg.AppNodesPrefix != "" {
			h.appNodesPrefix = hCfg.AppNodesPrefix
		} else if hCfg.AppGroupsPrefix != "" {
			h.appNodesPrefix = hCfg.AppGroupsPrefix
		}

		if hCfg.NodeMetaPrefix != "" {
			h.nodeMetaPrefix = hCfg.NodeMetaPrefix
		} else if hCfg.GroupMetaPrefix != "" {
			h.nodeMetaPrefix = hCfg.GroupMetaPrefix
		}

		if hCfg.NodeTargetsPrefix != "" {
			h.nodeTargetsPrefix = hCfg.NodeTargetsPrefix
		} else if hCfg.GroupResourcesPrefix != "" {
			h.nodeTargetsPrefix = hCfg.GroupResourcesPrefix
		}

		if hCfg.ParentNodesPrefix != "" {
			h.parentNodesPrefix = hCfg.ParentNodesPrefix
		} else if hCfg.TenantGroupsPrefix != "" {
			h.parentNodesPrefix = hCfg.TenantGroupsPrefix
		}

		if hCfg.ParentGrantedNodesPrefix != "" {
			h.parentGrantedNodesPrefix = hCfg.ParentGrantedNodesPrefix
		} else if hCfg.TenantGrantedGroupsPrefix != "" {
			h.parentGrantedNodesPrefix = hCfg.TenantGrantedGroupsPrefix
		}

		if hCfg.MetaModeField != "" {
			h.metaModeField = hCfg.MetaModeField
		}

		if hCfg.MetaParentNodeField != "" {
			h.metaParentNodeField = hCfg.MetaParentNodeField
		} else if hCfg.MetaOwnerTenantField != "" {
			h.metaParentNodeField = hCfg.MetaOwnerTenantField
		}

		if hCfg.ModeExplicit != "" {
			h.modeExplicit = hCfg.ModeExplicit
		}

		if hCfg.ModeFollowParent != "" {
			h.modeFollowParent = hCfg.ModeFollowParent
		} else if hCfg.ModeFollowTenant != "" {
			h.modeFollowParent = hCfg.ModeFollowTenant
		}
	}

	return h
}

func (w *AccessHierarchy) parseConfig(config *Config) (response.IResponse, []ruleHandler) {
	responseHandler := response.Parse(config.Response)
	if responseHandler == nil {
		responseHandler = defaultResponse
	}
	rules := make([]ruleHandler, 0)
	for _, rule := range config.Rules {
		rh := w.newHandler(rule.A, rule.B, config.Hierarchy)
		if rh != nil {
			rules = append(rules, rh)
		}
	}
	return responseHandler, rules
}
