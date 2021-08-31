/*
 * Copyright (c) 2021. Lorem ipsum dolor sit amet, consectetur adipiscing elit.
 * Morbi non lorem porttitor neque feugiat blandit. Ut vitae ipsum eget quam lacinia accumsan.
 * Etiam sed turpis ac ipsum condimentum fringilla. Maecenas magna.
 * Proin dapibus sapien vel ante. Aliquam erat volutpat. Pellentesque sagittis ligula eget metus.
 * Vestibulum commodo. Ut rhoncus gravida arcu.
 */

package router_http

import "strings"

var cmds = []string{
	cmdHost,
	cmdMethod,
	cmdLocation,
	cmdHeader,
	cmdQuery,
}

//HTTPRouterHelper http路由指标类型排序helper
type HTTPRouterHelper struct {
	index map[string]int
}

//NewHTTPRouterHelper 新建一个http路由指标类型排序helper
func NewHTTPRouterHelper() *HTTPRouterHelper {
	index := make(map[string]int)
	for i, cmd := range cmds {
		index[cmd] = i
	}
	return &HTTPRouterHelper{index: index}
}

func (h *HTTPRouterHelper) cmdType(cmd string) (string, string) {
	i := strings.Index(cmd, ":")
	if i < 0 {
		return cmd, ""
	}
	if i == 0 {
		return strings.ToLower(cmd[1:]), ""
	}

	return strings.ToLower(cmd[:i]), strings.ToLower(cmd[i+1:])

}

//Less 排序指标类型的匹配顺序
func (h *HTTPRouterHelper) Less(i, j string) bool {
	cmdI, keyI := h.cmdType(i)
	cmdJ, keyJ := h.cmdType(j)
	if cmdI != cmdJ {
		ii, hasI := h.index[cmdI]
		jj, hasJ := h.index[cmdJ]
		if !hasI && !hasJ {
			return cmdI < cmdJ
		}
		if !hasJ {
			return true
		}
		if !hasI {
			return false
		}
		return ii < jj
	}
	return keyI < keyJ
}
