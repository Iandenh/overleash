package unleashengine

import (
	"encoding/json"
	"fmt"
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

	defer func() {
		C.free(unsafe.Pointer(cjsonContext))
	}()

	cresolveAllDef := C.resolve_all(e.ptr, cjsonContext)
	jsonResolveAll := C.GoString(cresolveAllDef)

	return []byte(jsonResolveAll)
}

type VariantDef struct {
	Name    string   `json:"name,omitempty"`
	Payload *Payload `json:"payload,omitempty"`
	Enabled bool     `json:"enabled,omitempty"`
}

type Payload struct {
	PayloadType string `json:"type,omitempty"`
	Value       string `json:"value,omitempty"`
}

func (e *UnleashEngine) GetVariant(toggleName string, context *Context) *VariantDef {
	ctoggleName := C.CString(toggleName)

	jsonContext, err := json.Marshal(context)
	if err != nil {
		fmt.Printf("Failed to serialize context: %v\n", err)
		return nil
	}
	cjsonContext := C.CString(string(jsonContext))

	defer func() {
		C.free(unsafe.Pointer(ctoggleName))
		C.free(unsafe.Pointer(cjsonContext))
	}()

	cvariantDef := C.check_variant(e.ptr, ctoggleName, cjsonContext)
	jsonVariant := C.GoString(cvariantDef)

	variantDef := &VariantDef{}
	err = json.Unmarshal([]byte(jsonVariant), variantDef)
	if err != nil {
		fmt.Printf("Failed to deserialize variantDef: %v\n", err)
		return nil
	}
	return variantDef
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

func NewContext(userID, sessionID, environment, appName, currentTime, remoteAddress *string, properties *map[string]string) *Context {
	return &Context{
		UserID:        userID,
		SessionID:     sessionID,
		Environment:   environment,
		AppName:       appName,
		CurrentTime:   currentTime,
		RemoteAddress: remoteAddress,
		Properties:    properties,
	}
}
