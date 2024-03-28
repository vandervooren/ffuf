package modifier

import (
	"os"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"
	"github.com/ffuf/ffuf/v2/pkg/ffuf"
	"github.com/ffuf/ffuf/v2/pkg/modifier/crypto"
)

type GojaModifierInstance struct {
	vm      *goja.Runtime
	require *require.RequireModule
	modify  func(req *ffuf.Request) int
}

type GojaModifier struct {
	config   *ffuf.Config
	registry *require.Registry
	script   string
	program  *goja.Program
}

func NewGojaModifier(conf *ffuf.Config) *GojaModifier {
	var modifier GojaModifier

	modifier.config = conf
	modifier.registry = new(require.Registry)

	if conf.ModifierScript != "" {
		script, err := os.ReadFile(conf.ModifierScript)

		if err != nil {
			panic(err)
		}

		modifier.script = string(script)
		modifier.program = goja.MustCompile("modifier", string(script), true)
	}

	return &modifier
}

func (gm *GojaModifier) Modify(req *ffuf.Request) *ffuf.Request {
	if gm.program == nil {
		return req
	}

	var modifier GojaModifierInstance

	modifier.vm = goja.New()

	modifier.require = gm.registry.Enable(modifier.vm)

	console.Enable(modifier.vm)
	crypto.Enable(modifier.vm)

	modifier.vm.RunProgram(gm.program)

	err := modifier.vm.ExportTo(modifier.vm.Get("modify"), &modifier.modify)

	if err != nil {
		panic(err)
	}

	modifier.modify(req)

	return req
}
