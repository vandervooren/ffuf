package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"math/rand"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/buffer"
	"github.com/dop251/goja_nodejs/errors"
	"github.com/dop251/goja_nodejs/require"
)

const ModuleName = "crypto"

type Crypto struct {
	runtime *goja.Runtime
	buffer  *buffer.Buffer
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

	return c.buffer.WrapBytes(dataHmac)
}

func (c *Crypto) randomBytes(call goja.FunctionCall) goja.Value {
	size := call.Arguments[0].ToInteger()

	buf := make([]byte, size)
	_, err := rand.Read(buf)

	if len(call.Arguments) == 2 {
		callback, ok := goja.AssertFunction(call.Arguments[1])

		if !ok {
			panic("crypto.randomBytes() callback is not a function")
		}

		if err != nil {
			callback(goja.Undefined(), errors.NewTypeError(c.runtime, "ERR_CRYPTO", err), goja.Undefined())
		} else {
			callback(goja.Undefined(), goja.Undefined(), c.buffer.WrapBytes(buf))
		}

		return goja.Undefined()
	}

	if err != nil {
		panic(err)
	}

	return c.buffer.WrapBytes(buf)
}

func (c *Crypto) bufferReadInt32LE(call goja.FunctionCall) goja.Value {
	buf := buffer.Bytes(c.runtime, call.This)

	start := 0

	if len(call.Arguments) == 1 {
		start = int(call.Arguments[0].ToInteger())
	}

	result := int32(binary.LittleEndian.Uint32(buf[start:4]))

	return c.runtime.ToValue(result)
}

func Require(runtime *goja.Runtime, module *goja.Object) {
	c := &Crypto{
		runtime: runtime,
	}

	c.buffer = buffer.GetApi(c.runtime)

	// Patching Buffer module to provide readInt32LE() used by crypto-js

	bufferModule := require.Require(runtime, buffer.ModuleName).(*goja.Object)
	bufferObject := bufferModule.Get("Buffer").(*goja.Object)
	bufferProto := bufferObject.Get("prototype").(*goja.Object)
	bufferProto.Set("readInt32LE", c.bufferReadInt32LE)

	// Declaring crypto.randomBytes() used by crypto-js

	o := module.Get("exports").(*goja.Object)
	o.Set("randomBytes", c.randomBytes)

	// Adding native implementation of hmacSha256 (not present in node 'crypto')

	o.Set("hmacSha256", c.hmacSha256)
}

func Enable(runtime *goja.Runtime) {
	runtime.Set("crypto", require.Require(runtime, ModuleName))
}

func init() {
	require.RegisterCoreModule(ModuleName, Require)
}
