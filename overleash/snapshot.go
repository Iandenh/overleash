package overleash

import "slices"

// Snapshots own their nested slices, maps and pointers. A shallow copy would
// still let dashboard sorting or callers mutate the state used by other readers.
func cloneFeatureFile(file FeatureFile) FeatureFile {
	file.Features = slices.Clone(file.Features)
	for i := range file.Features {
		f := &file.Features[i]
		f.Stale = clonePointer(f.Stale)
		f.CreatedAt = clonePointer(f.CreatedAt)
		f.LastSeenAt = clonePointer(f.LastSeenAt)
		f.Strategies = slices.Clone(f.Strategies)
		for j := range f.Strategies {
			s := &f.Strategies[j]
			s.SortOrder = clonePointer(s.SortOrder)
			s.Segments = slices.Clone(s.Segments)
			s.Constraints = cloneConstraints(s.Constraints)
			if s.Parameters != nil {
				parameters := make(ParameterMap, len(s.Parameters))
				for key, value := range s.Parameters {
					parameters[key] = cloneParameter(value)
				}
				s.Parameters = parameters
			}
			s.Variants = slices.Clone(s.Variants)
		}
		f.Variants = slices.Clone(f.Variants)
		for j := range f.Variants {
			v := &f.Variants[j]
			v.Overrides = slices.Clone(v.Overrides)
			for k := range v.Overrides {
				v.Overrides[k].Values = slices.Clone(v.Overrides[k].Values)
			}
		}
		if f.Dependencies != nil {
			dependencies := slices.Clone(*f.Dependencies)
			for j := range dependencies {
				d := &dependencies[j]
				d.Enabled = clonePointer(d.Enabled)
				if d.Variants != nil {
					variants := slices.Clone(*d.Variants)
					d.Variants = &variants
				}
			}
			f.Dependencies = &dependencies
		}
	}
	file.Segments = slices.Clone(file.Segments)
	for i := range file.Segments {
		file.Segments[i].Constraints = cloneConstraints(file.Segments[i].Constraints)
	}
	return file
}

func clonePointer[T any](p *T) *T {
	if p == nil {
		return nil
	}
	value := *p
	return &value
}

func cloneConstraints(constraints []Constraint) []Constraint {
	constraints = slices.Clone(constraints)
	for i := range constraints {
		constraints[i] = cloneConstraint(constraints[i])
	}
	return constraints
}

func cloneConstraint(constraint Constraint) Constraint {
	constraint.Values = slices.Clone(constraint.Values)
	constraint.Value = clonePointer(constraint.Value)
	return constraint
}

// Parameters received from JSON contain only scalar values, maps and slices.
func cloneParameter(value any) any {
	switch v := value.(type) {
	case map[string]any:
		result := make(map[string]any, len(v))
		for key, value := range v {
			result[key] = cloneParameter(value)
		}
		return result
	case []any:
		result := make([]any, len(v))
		for i, value := range v {
			result[i] = cloneParameter(value)
		}
		return result
	default:
		return value
	}
}

func cloneOverride(override *Override) *Override {
	if override == nil {
		return nil
	}
	result := *override
	result.Constraints = slices.Clone(override.Constraints)
	for i := range result.Constraints {
		result.Constraints[i].Constraint = cloneConstraint(result.Constraints[i].Constraint)
	}
	return &result
}

func cloneOverrides(overrides map[string]*Override) map[string]*Override {
	result := make(map[string]*Override, len(overrides))
	for key, override := range overrides {
		result[key] = cloneOverride(override)
	}
	return result
}
