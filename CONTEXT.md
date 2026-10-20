# Overleash

Overleash lets users override upstream Unleash feature flags without changing the upstream configuration.

## Language

**Usable feature state**:
A successfully loaded feature-flag configuration for an environment. It may contain no feature flags; an empty configuration is not the same as an environment that has never loaded a configuration.

**Backup**:
A saved copy of an environment's upstream feature-flag configuration, excluding local overrides. It provides recovery when no usable feature state is otherwise available.
