package access_hierarchy

import "github.com/eolinker/apinto/utils/response"

type Rule struct {
	A string `yaml:"a" json:"a,omitempty" label:"App Key" description:"App key metrics syntax" require:"true"`
	B string `yaml:"b" json:"b,omitempty" label:"Resource Key" description:"Resource key metrics syntax" require:"true"`
}

type HierarchyConfig struct {
	AppNodesPrefix           string `yaml:"app_nodes_prefix" json:"app_nodes_prefix" label:"App Nodes Prefix"`
	NodeMetaPrefix           string `yaml:"node_meta_prefix" json:"node_meta_prefix" label:"Node Meta Prefix"`
	NodeTargetsPrefix        string `yaml:"node_targets_prefix" json:"node_targets_prefix" label:"Node Targets Prefix"`
	ParentNodesPrefix        string `yaml:"parent_nodes_prefix" json:"parent_nodes_prefix" label:"Parent Nodes Prefix"`
	ParentGrantedNodesPrefix string `yaml:"parent_granted_nodes_prefix" json:"parent_granted_nodes_prefix" label:"Parent Granted Nodes Prefix"`
	MetaModeField            string `yaml:"meta_mode_field" json:"meta_mode_field" label:"Meta Mode Field"`
	MetaParentNodeField      string `yaml:"meta_parent_node_field" json:"meta_parent_node_field" label:"Meta Parent Node Field"`
	ModeExplicit             string `yaml:"mode_explicit" json:"mode_explicit" label:"Mode Explicit"`
	ModeFollowParent         string `yaml:"mode_follow_parent" json:"mode_follow_parent" label:"Mode Follow Parent"`

	AppGroupsPrefix           string `yaml:"app_groups_prefix" json:"app_groups_prefix" label:"App Groups Prefix"`
	GroupMetaPrefix           string `yaml:"group_meta_prefix" json:"group_meta_prefix" label:"Group Meta Prefix"`
	GroupResourcesPrefix      string `yaml:"group_resources_prefix" json:"group_resources_prefix" label:"Group Resources Prefix"`
	TenantGroupsPrefix        string `yaml:"tenant_groups_prefix" json:"tenant_groups_prefix" label:"Tenant Groups Prefix"`
	TenantGrantedGroupsPrefix string `yaml:"tenant_granted_groups_prefix" json:"tenant_granted_groups_prefix" label:"Tenant Granted Groups Prefix"`
	MetaOwnerTenantField      string `yaml:"meta_owner_tenant_field" json:"meta_owner_tenant_field" label:"Meta Owner Tenant Field"`
	ModeFollowTenant          string `yaml:"mode_follow_tenant" json:"mode_follow_tenant" label:"Mode Follow Tenant"`
}

type Config struct {
	Rules     []*Rule            `yaml:"rules" json:"rules" label:"Rules"`
	Response  *response.Response `yaml:"response" json:"response" label:"Response"`
	Hierarchy *HierarchyConfig   `yaml:"hierarchy" json:"hierarchy" label:"Hierarchy Config"`
}
