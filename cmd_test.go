package glick_test

import (
	"strings"
	"testing"
	"time"

	"github.com/documize/glick"
)

func cmdSwitchTest(t *testing.T, useJSON bool) {
	l, errN := glick.New(nil)
	if errN != nil {
		t.Error(errN)
	}
	var proto string
	outProto := func() interface{} { var s string; return interface{}(&s) }
	if err := l.RegAPI("string/&string", proto, outProto, 10*time.Second); err != nil {
		t.Error(err)
		return
	}
	if err := l.RegPlugin("string/&string", "pwdBad",
		glick.PluginCmd("pwdBad", nil, &proto)); err == nil {
		t.Error("incorrect cmd plugin did not fail")
		return
	}
	if _, err := l.Run(nil, "string/&string", "pwdBad", proto); err == nil {
		t.Error("bad simple command did not fail")
		return
	}
	if err := l.RegPlugin("string/&string", "pwd",
		glick.PluginCmd("pwd", nil, &proto)); err != nil {
		t.Error(err)
		return
	}
	if proto, err := l.Run(nil, "string/&string", "pwd", proto); err != nil {
		t.Error(err)
	} else {
		p := *(proto.(*string))
		if !strings.HasSuffix(p, "/glick\n") {
			t.Error("wrong output from pwd: " + p)
		}
	}
	if err := l.RegPlugin("string/&string", "exit1",
		glick.PluginCmd("bash", []string{"./_test/exit1.sh"}, &proto)); err != nil {
		t.Error(err)
		return
	}
	if _, err := l.Run(nil, "string/&string", "exit1", proto); err == nil {
		t.Error("exit1.sh does not give an error")
	} else {
		if err.Error() != "exit status 1" {
			t.Error("exit1.sh does not give correct error")
		}
	}
	var ip int
	ipProto := func() interface{} { var i int; return interface{}(&i) }
	if err := l.RegAPI("int/&int", ip, ipProto, 3*time.Second); err != nil {
		t.Error(err)
		return
	}
	if err := l.RegPlugin("int/&int", "pwd",
		glick.PluginCmd("pwd", nil, &proto)); err != nil {
		t.Error(err)
		return
	}
	if _, err := l.Run(nil, "int/&int", "pwd", 11); err == nil {
		t.Error("does not error on non string input value")
		return
	}
	if err := l.RegAPI("string/&int", proto, ipProto, 3*time.Second); err != nil {
		t.Error(err)
		return
	}
	if err := l.RegPlugin("string/&int", "pwd",
		glick.PluginCmd("pwd", nil, &proto)); err != nil {
		t.Error(err)
		return
	}
	if _, err := l.Run(nil, "string/&int", "pwd", "foo"); err == nil {
		t.Error("does not error on non *string ouput value")
		return
	}
	if err := l.RegAPI("alwaysTimeout", proto, outProto, 1*time.Second); err != nil {
		t.Error(err)
	}
	if err := l.RegPlugin("alwaysTimeout", "sleep",
		glick.PluginCmd("sleep", []string{"10"}, &proto)); err != nil {
		t.Error(err)
		return
	}
	if _, err := l.Run(nil, "alwaysTimeout", "sleep", "foo"); err == nil {
		t.Error("does not timeout when it should")
	}
}

func TestCmd(t *testing.T) {
	cmdSwitchTest(t, true)
	cmdSwitchTest(t, false)
}