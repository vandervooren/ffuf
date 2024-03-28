package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/errors"
	"github.com/dop251/goja_nodejs/require"
)

const ModuleName = "crypto"

type Crypto struct {
	runtime *goja.Runtime
}

func (c *Crypto) hmacSha256(call goja.FunctionCall) goja.Value {
	// Usage: hmacSha256(message, key)

	if len(call.Arguments) != 2 {
		panic(errors.NewTypeError(c.runtime, "ERR_INVALID_ARGUMENTS", "Usage: hmacSha256(message, key)"))
	}

	messageV := call.Arguments[0]
	keyV := call.Arguments[1]

	hmac := hmac.New(sha256.New, []byte(keyV.String()))

	hmac.Write([]byte(messageV.String()))

	dataHmac := hmac.Sum(nil)

	hmacHex := base64.StdEncoding.EncodeToString(dataHmac)

	return c.runtime.ToValue(hmacHex)
}

func Require(runtime *goja.Runtime, module *goja.Object) {
	c := &Crypto{
		runtime: runtime,
	}

	o := module.Get("exports").(*goja.Object)
	o.Set("hmacSha256", c.hmacSha256)
}

func Enable(runtime *goja.Runtime) {
	runtime.Set("crypto", require.Require(runtime, ModuleName))
}

func init() {
	require.RegisterCoreModule(ModuleName, Require)
}
