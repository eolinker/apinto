package params_check

import (
	"testing"

	"github.com/ohler55/ojg/oj"

	"github.com/eolinker/apinto/checker"
)

func TestParamCheck(t *testing.T) {
	data := `{
"code":"HWJXH_DEPTCOMPLEX_CSGLJ_HWJXHXT_XLJXHZYL_1.0",
"isPage":true,
"index":1,
"size":10,
"apiType":"deptCOMPLEX",
"userName":"APIFXJCXT",
"psd":"5190649892064f8c9bf387c3d30ce021",
"apiId":"d1d3f080f18448fcb5354f67b93e54e9",
"search":[{"param":"F_SECTIONNAME","type":"String","val":""}]
}`

	c, err := checker.Parse("**")
	if err != nil {
		t.Fatal(err)
	}
	checkers := []*paramChecker{
		{
			name:    "$.search[0].val",
			Checker: c,
		},
	}
	o, err := oj.ParseString(data)
	if err != nil {
		t.Fatal(err)
	}
	for _, ck := range checkers {

		err = jsonChecker(o, ck)
		if err != nil {
			t.Fatal(err)
		}
	}

}
