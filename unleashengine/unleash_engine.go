package unleashengine

import (
	"encoding/json"
	"github.com/charmbracelet/log"
	"unsafe"
)

// #cgo LDFLAGS: -L. -lyggdrasilffi
// #include "unleash_engine.h"
import "C"

type Engine interface {
	TakeState(json string)
	Resolve(context *Context, featureName string) []byte
	ResolveAll(context *Context) []byte
}

type UnleashEngine struct {
	ptr unsafe.Pointer
}

func NewUnleashEngine() *UnleashEngine {
	ptr := unsafe.Pointer(C.new_engine())
	return &UnleashEngine{ptr: ptr}
}

func (e *UnleashEngine) TakeState(json string) {
	cjson := C.CString(json)

	defer C.free(unsafe.Pointer(cjson))

	res := C.take_state(e.ptr, cjson)

	if log.GetLevel() == log.DebugLevel {
		resJson := C.GoString(res)

		log.Debugf("TakeState: %s", string(resJson))
	}

	C.free_response(res)
}

func (e *UnleashEngine) Resolve(context *Context, featureName string) []byte {
	jsonContext, err := json.Marshal(context)

	if err != nil {
		log.Fatalf("Failed to serialize context: %v", err)
		return nil
	}

	cfeatureName := C.CString(featureName)
	cjsonContext := C.CString(string(jsonContext))
	defer func() {
		C.free(unsafe.Pointer(cfeatureName))
		C.free(unsafe.Pointer(cjsonContext))
	}()

	cresolveDef := C.resolve(e.ptr, cfeatureName, cjsonContext)
	jsonResolve := C.GoString(cresolveDef)
	C.free_response(cresolveDef)

	return []byte(jsonResolve)
}

func (e *UnleashEngine) ResolveAll(context *Context) []byte {
	jsonContext, err := json.Marshal(context)

	if err != nil {
		log.Fatalf("Failed to serialize context: %v", err)
		return nil
	}
	cjsonContext := C.CString(string(jsonContext))

	defer C.free(unsafe.Pointer(cjsonContext))

	cresolveAllDef := C.resolve_all(e.ptr, cjsonContext)
	jsonResolveAll := C.GoString(cresolveAllDef)
	C.free_response(cresolveAllDef)

	return []byte(jsonResolveAll)
}

type Context struct {
	UserID        *string            `json:"userId,omitempty"`
	SessionID     *string            `json:"sessionId,omitempty"`
	Environment   *string            `json:"environment,omitempty"`
	AppName       *string            `json:"appName,omitempty"`
	CurrentTime   *string            `json:"currentTime,omitempty"`
	RemoteAddress *string            `json:"remoteAddress,omitempty"`
	Properties    *map[string]string `json:"properties,omitempty"`
}
