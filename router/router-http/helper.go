/*
 * Copyright (c) 2021. Lorem ipsum dolor sit amet, consectetur adipiscing elit.
 * Morbi non lorem porttitor neque feugiat blandit. Ut vitae ipsum eget quam lacinia accumsan.
 * Etiam sed turpis ac ipsum condimentum fringilla. Maecenas magna.
 * Proin dapibus sapien vel ante. Aliquam erat volutpat. Pellentesque sagittis ligula eget metus.
 * Vestibulum commodo. Ut rhoncus gravida arcu.
 */

package router

type HttpRouterHelper struct {

}

func NewHttpRouterHelper() *HttpRouterHelper {
	return &HttpRouterHelper{}
}

func (h *HttpRouterHelper) Less(i, j string) bool {
	panic("implement me")
}

