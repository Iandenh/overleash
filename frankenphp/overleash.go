package frankenphp

// #include <stdlib.h>
// #include "overleash.h"
import "C"
import (
	"fmt"
	"time"
	"unsafe"

	"github.com/Iandenh/overleash/overleash"
	"github.com/Iandenh/overleash/unleashengine"
	frank "github.com/dunglas/frankenphp"
)

func init() {
	frank.RegisterExtension(unsafe.Pointer(&C.overleash_module_entry))
}

//export go_is_enabled
func go_is_enabled(phpFeatureFlag *C.zend_string, ctx *C.zend_array) (bool, *C.char) {
	flag := zendStringToGoString(phpFeatureFlag)

	con := &unleashengine.Context{}
	if ctx != nil {
		array, err := frank.GoMap[any](unsafe.Pointer(ctx))

		if err != nil {
			return false, C.CString(fmt.Sprintf("error on context: %s", err))
		}

		if len(array) > 0 {
			con = toContext(array)
		}
	}

	if hub == nil {
		return false, C.CString("Overleash not correctly loaded")
	}

	env := hub.overleash.ActiveFeatureEnvironment()

	e, err := env.Engine().Resolve(con, flag)

	if err != nil {
		return false, C.CString(fmt.Sprintf("error on context: %s", err))
	}

	collectMetric(e.Enabled, env, flag)

	return e.Enabled, nil
}

func collectMetric(e bool, env *overleash.FeatureEnvironment, flagName string) {
	t := time.Now()

	yes := 0
	no := 0
	if e == true {
		yes = 1
		no = 1
	}

	hub.overleash.AddMetric(&overleash.MetricsData{
		Environment:  env.Name(),
		ConnectVia:   nil,
		AppName:      "Overleash",
		InstanceID:   "random",
		ConnectionId: "",
		Bucket: overleash.Bucket{
			Start: t,
			Stop:  t,
			Toggles: map[string]overleash.ToggleCount{
				flagName: {
					Yes:      int32(yes),
					No:       int32(no),
					Variants: nil,
				},
			},
		},
	})
}

func zendStringToGoString(zendStr *C.zend_string) string {
	if zendStr == nil {
		return ""
	}

	return C.GoStringN((*C.char)(unsafe.Pointer(&zendStr.val)), C.int(zendStr.len))
}

func toContext(input map[string]any) *unleashengine.Context {
	ctx := &unleashengine.Context{
		Properties: make(map[string]string),
	}

	for k, v := range input {
		if k == "properties" {
			properties, ok := v.(map[string]string)

			if ok {
				ctx.Properties = properties
			}
			continue
		}

		strVal, ok := v.(string)
		if !ok {
			strVal = fmt.Sprint(v)
		}

		switch k {
		case "userId", "UserId":
			ctx.UserId = &strVal
		case "sessionId", "SessionId":
			ctx.SessionId = &strVal
		case "remoteAddress", "RemoteAddress":
			ctx.RemoteAddress = &strVal
		case "environment", "Environment":
			ctx.Environment = &strVal
		case "appName", "AppName":
			ctx.AppName = &strVal
		case "currentTime", "CurrentTime":
			ctx.CurrentTime = &strVal
		default:
			ctx.Properties[k] = strVal
		}
	}

	return ctx
}
