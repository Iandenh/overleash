package overleash

import "github.com/Iandenh/overleash/unleashengine"

// lockedEngine keeps callers from bypassing the state lock through Engine().
type lockedEngine struct{ environment *FeatureEnvironment }

func (e lockedEngine) TakeState(state string) error {
	e.environment.owner.mu.Lock()
	defer e.environment.owner.mu.Unlock()
	return e.environment.engine.TakeState(state)
}

func (e lockedEngine) Resolve(context *unleashengine.Context, name string) (*unleashengine.EvaluatedToggle, error) {
	e.environment.owner.mu.RLock()
	defer e.environment.owner.mu.RUnlock()
	return e.environment.engine.Resolve(context, name)
}

func (e lockedEngine) ResolveAll(context *unleashengine.Context, includeAll bool) (*unleashengine.EvaluatedToggleList, error) {
	e.environment.owner.mu.RLock()
	defer e.environment.owner.mu.RUnlock()
	return e.environment.engine.ResolveAll(context, includeAll)
}
