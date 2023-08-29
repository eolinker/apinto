package response_rewrite_v2

import (
	"log"
	"regexp"
	"testing"
)

func TestRewrite(t *testing.T) {
	re := regexp.MustCompile(`\{([^}]+)\}`)
	log.Println(re.ReplaceAllString("abc{a}def{b}ghi", "%s"))
}
