package jwt

import "testing"

func TestDoJWTAuthentication(t *testing.T) {
	md := map[string][]string{
		":authority": {
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2NTQ1ODg3NzMsInN1YiI6IjEyMzQ1Njc4OTAiLCJuYW1lIjoiSm9obiBEb2UiLCJuYmYiOjE1MTYyMzkwMjIsImlzcyI6ImFkbWluIiwidXNlciI6ImFkbWluIn0.dP2y7zdMZi-OZzH3inI3Uo31OYrbdEeoJFkNU6NIrUI",
		},
	}
	check(md)
}
