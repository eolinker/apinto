/*
 * Copyright (c) 2021. Lorem ipsum dolor sit amet, consectetur adipiscing elit.
 * Morbi non lorem porttitor neque feugiat blandit. Ut vitae ipsum eget quam lacinia accumsan.
 * Etiam sed turpis ac ipsum condimentum fringilla. Maecenas magna.
 * Proin dapibus sapien vel ante. Aliquam erat volutpat. Pellentesque sagittis ligula eget metus.
 * Vestibulum commodo. Ut rhoncus gravida arcu.
 */

package router_http

import (
	"fmt"
	"net/textproto"
	"strings"
)

const (
	cmdLocation = "LOCATION"
	cmdHeader   = "HEADER"
	cmdQuery    = "QUERY"
	cmdHost     = "HOST"
	cmdMethod   = "METHOD"
)

func toMethod() string {
	return cmdMethod
}

func toLocation() string {
	return cmdLocation
}

func toHeader(key string) string {
	return fmt.Sprint(cmdHeader, ":", textproto.CanonicalMIMEHeaderKey(key))
}

func toQuery(key string) string {
	return fmt.Sprint(cmdQuery, ":", key)

}

func toHost() string {
	return cmdHost
}

func headerName(cmd string) (string, bool) {
	if b := strings.HasPrefix(cmd, "HEADER:"); b {
		return strings.TrimPrefix(cmd, "HEADER:"), true
	}
	return "", false

}

func queryName(cmd string) (string, bool) {
	if b := strings.HasPrefix(cmd, "QUERY:"); b {
		return strings.TrimPrefix(cmd, "QUERY:"), true
	}
	return "", false
}

func isLocation(cmd string) bool {
	return cmd == cmdLocation
}

func isHost(cmd string) bool {
	return cmd == cmdHost
}

func isMethod(cmd string) bool {
	return cmd == cmdMethod
}
