package requests

import (
	"github.com/rolandhe/go-base/commons"
	"testing"
)

func Test1(t *testing.T) {
	ret := &commons.Result[string]{
		Code: 200,
	}
	var r any
	r = ret
	cr, ok := r.(commons.CodedResult)
	t.Log(ok)
	if ok {
		t.Log(cr.GetCode())
	}
}
