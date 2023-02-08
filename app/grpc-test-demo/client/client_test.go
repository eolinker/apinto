package main

import (
	"encoding/base64"
	"fmt"
	"testing"
)

var (
	users = []struct {
		username string
		password string
		success  bool
	}{
		{
			username: "admin",
			password: "123456",
			success:  true,
		},
		{
			username: "admin",
			password: "123456",
		},
		{
			username: "apinto",
			password: "123456",
			success:  false,
		},
	}
	md = map[string]string{
		"authorization-type": "jwt",
		"input":              "test",
	}
	names = []string{
		"test",
		"apinto",
	}
)

func TestJWTRequest(t *testing.T) {
	t.Log("begin test jwt request...")
	err := CurrentRequest(names, md, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2NTQ1ODg3NzMsInN1YiI6IjEyMzQ1Njc4OTAiLCJuYW1lIjoiSm9obiBEb2UiLCJuYmYiOjE1MTYyMzkwMjIsImlzcyI6ImFkbWluIiwidXNlciI6ImFkbWluIn0.dP2y7zdMZi-OZzH3inI3Uo31OYrbdEeoJFkNU6NIrUI")
	if err != nil {
		t.Errorf("request fail,err: %s\n", err)
	}

	t.Log("end jwt request...")
}

func TestBasicRequest(t *testing.T) {
	t.Log("begin test basic request...")
	for _, u := range users {
		var success = true
		err := CurrentRequest(names, md, fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", u.username, u.password)))))
		if err != nil {
			t.Errorf("request fail,err: %s,username: %s,password: %s\n", err, u.username, u.password)
			success = false
		}
		if success != u.success {
			t.Errorf("test result failed,username: %s,password: %s", u.username, u.password)
		}
	}
	t.Log("end basic request...")
}

func TestBasicStreamRequest(t *testing.T) {
	t.Log("begin test basic stream request...")
	for _, u := range users {
		var success = true
		err := StreamRequest(names, md, fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", u.username, u.password)))))
		if err != nil {
			t.Errorf("request fail,err: %s,username: %s,password: %s\n", err, u.username, u.password)
			success = false
		}
		if success != u.success {
			t.Errorf("test result failed,username: %s,password: %s", u.username, u.password)
		}
	}
	t.Log("end basic stream request...")
}

func TestBasicStreamResponse(t *testing.T) {
	t.Log("begin test basic stream request...")
	for _, u := range users {
		var success = true
		err := StreamResponse(names, md, fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", u.username, u.password)))))
		if err != nil {
			t.Errorf("request fail,err: %s,username: %s,password: %s\n", err, u.username, u.password)
			success = false
		}
		if success != u.success {
			t.Errorf("test result failed,username: %s,password: %s", u.username, u.password)
		}
	}
	t.Log("end basic stream request...")
}
