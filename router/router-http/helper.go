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

type HttpRouterHelper struct {
	index map[string]int
}

func NewHttpRouterHelper() *HttpRouterHelper {
	index := make(map[string]int)
	for i, cmd := range cmds {
		index[cmd] = i
	}
	return &HttpRouterHelper{index: index}
}
func (h *HttpRouterHelper) cmdType(cmd string) (string, string) {
	i := strings.Index(cmd, ":")
	if i < 0 {
		return cmd, ""
	}
	if i == 0 {
		return strings.ToLower(cmd[1:]), ""
	}

	return strings.ToLower(cmd[:i]), strings.ToLower(cmd[i+1:])

}

func (h *HttpRouterHelper) Less(i, j string) bool {
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
