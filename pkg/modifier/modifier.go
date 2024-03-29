package modifier

import (
	"fmt"
	"os"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/buffer"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"
	"github.com/dop251/goja_nodejs/url"
	"github.com/ffuf/ffuf/v2/pkg/ffuf"
	"github.com/ffuf/ffuf/v2/pkg/modifier/crypto"
)

type GojaModifier struct {
	config   *ffuf.Config
	registry *require.Registry
	program  *goja.Program
}

func NewGojaModifier(conf *ffuf.Config) *GojaModifier {
	var modifier GojaModifier

	modifier.config = conf
	modifier.registry = require.NewRegistry(require.WithGlobalFolders("."), require.WithLoader(require.DefaultSourceLoader))

	if conf.ModifierScript != "" {
		script, err := os.ReadFile(conf.ModifierScript)

		if err != nil {
			panic(err)
		}

		modifier.program = goja.MustCompile("", string(script), true)
	}

	return &modifier
}

func (gm *GojaModifier) Modify(req *ffuf.Request) *ffuf.Request {
	if gm.program == nil {
		return req
	}

	vm := goja.New()

	gm.registry.Enable(vm)

	buffer.Enable(vm)
	console.Enable(vm)
	url.Enable(vm)

	crypto.Enable(vm)

	vm.RunProgram(gm.program)

	modifyFunc, isSuccess := goja.AssertFunction(vm.Get("modify"))

	if !isSuccess {
		panic(fmt.Errorf("cannot find a symbol with name 'modify' or declared symbol is not a function"))
	}

	_, err := modifyFunc(goja.Undefined(), vm.ToValue(req))

	if err != nil {
		panic(err)
	}

	return req
}
