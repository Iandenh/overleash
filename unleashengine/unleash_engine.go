package unleashengine

import (
	"encoding/json"
	"log"
	"unsafe"
)

// #cgo LDFLAGS: -L. -lyggdrasilffi
// #include "unleash_engine.h"
import "C"

type UnleashEngine struct {
	ptr unsafe.Pointer
}

func NewUnleashEngine() *UnleashEngine {
	ptr := unsafe.Pointer(C.new_engine())
	return &UnleashEngine{ptr: ptr}
}

func (e *UnleashEngine) TakeState(json string) {
	C.take_state(e.ptr, C.CString(json))
}

func (e *UnleashEngine) ResolveAll(context *Context) []byte {
	jsonContext, err := json.Marshal(context)

	if err != nil {
		log.Fatalf("Failed to serialize context: %v", err)
		return []byte{}
	}
	cjsonContext := C.CString(string(jsonContext))

	defer C.free(unsafe.Pointer(cjsonContext))

	cresolveAllDef := C.resolve_all(e.ptr, cjsonContext)
	jsonResolveAll := C.GoString(cresolveAllDef)

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
