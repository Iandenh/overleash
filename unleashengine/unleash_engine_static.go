//go:build yggdrasil_static

package unleashengine

import (
	"errors"
	"fmt"
	"unsafe"

	"github.com/charmbracelet/log"
	"google.golang.org/protobuf/proto"
)

/*
#cgo CFLAGS: -g -Wall
#cgo LDFLAGS: ${SRCDIR}/libfrontendengine.a
#include "frontend_engine.h"
*/
import "C"

type Engine interface {
	TakeState(json string)
	Resolve(context *Context, featureName string) (*EvaluatedToggle, error)
	ResolveAll(context *Context, includeAll bool) (*EvaluatedToggleList, error)
	IsEnabled(context *Context, featureName string) bool
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

func (e *UnleashEngine) Resolve(context *Context, featureName string) (*EvaluatedToggle, error) {
	inputBytes, err := proto.Marshal(context)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize context: %v", err)
	}

	// 2. Prepare C arguments
	var inputPtr *C.uint8_t
	if len(inputBytes) > 0 {
		inputPtr = (*C.uint8_t)(unsafe.Pointer(&inputBytes[0]))
	}
	inputLen := C.size_t(len(inputBytes))

	cFeatureName := C.CString(featureName)

	var outLen C.size_t

	ptr := C.resolve(e.ptr, cFeatureName, inputPtr, inputLen, &outLen)

	defer C.free(unsafe.Pointer(cFeatureName))

	if ptr == nil {
		return nil, errors.New("resolution failed (check if toggle exists or engine is valid)")
	}

	defer C.free_rust_buffer((*C.uint8_t)(ptr), outLen)

	data := C.GoBytes(unsafe.Pointer(ptr), C.int(outLen))

	var toggle EvaluatedToggle
	if err := proto.Unmarshal(data, &toggle); err != nil {
		return nil, fmt.Errorf("failed to unmarshal EvaluatedToggle: %v", err)
	}

	return &toggle, nil
}

func (e *UnleashEngine) IsEnabled(context *Context, featureName string) bool {
	inputBytes, err := proto.Marshal(context)
	if err != nil {
		return false
	}

	// 2. Prepare C arguments
	var inputPtr *C.uint8_t
	if len(inputBytes) > 0 {
		inputPtr = (*C.uint8_t)(unsafe.Pointer(&inputBytes[0]))
	}
	inputLen := C.size_t(len(inputBytes))

	cFeatureName := C.CString(featureName)

	cResult := C.is_enabled(e.ptr, cFeatureName, inputPtr, inputLen)

	defer C.free(unsafe.Pointer(cFeatureName))

	return bool(cResult)
}

func (e *UnleashEngine) ResolveAll(context *Context, includeAll bool) (*EvaluatedToggleList, error) {
	inputBytes, err := proto.Marshal(context)

	var inputPtr *C.uint8_t
	if len(inputBytes) > 0 {
		inputPtr = (*C.uint8_t)(unsafe.Pointer(&inputBytes[0]))
	}
	inputLen := C.uintptr_t(len(inputBytes))

	if err != nil {
		return nil, errors.New("resolution failed or returned null")
	}

	cIncludeAll := C.bool(includeAll)

	var outLen C.size_t

	ptr := C.resolve_all(
		e.ptr,
		inputPtr,
		&cIncludeAll,
		inputLen,
		&outLen,
	)

	if ptr == nil {
		return nil, errors.New("resolution failed")
	}

	defer C.free_rust_buffer((*C.uint8_t)(ptr), outLen)

	// Convert C pointer to Go byte slice
	data := C.GoBytes(unsafe.Pointer(ptr), C.int(outLen))

	// Unmarshal into the generated Go struct
	var list EvaluatedToggleList
	err = proto.Unmarshal(data, &list)

	if err != nil {
		return nil, err
	}

	return &list, nil
}
