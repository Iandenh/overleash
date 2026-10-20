package overleash

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/log"
)

func (o *OverleashContext) initializeFeatureFiles() error {
	err := o.loadRemotes()

	o.mu.RLock()
	defer o.mu.RUnlock()
	var unavailable []string
	for _, environment := range o.featureEnvironments {
		// Compilation can publish an empty snapshot for an environment that has
		// never loaded state. Only a remote revision distinguishes that from a
		// successfully loaded configuration containing no feature flags.
		if environment.remoteRevision == 0 {
			unavailable = append(unavailable, environment.name)
		}
	}
	if len(unavailable) != 0 {
		return errors.Join(fmt.Errorf("no usable feature state for %s", strings.Join(unavailable, ", ")), err)
	}
	return nil
}

func (o *OverleashContext) loadRemotes() error {
	o.refreshMu.Lock()
	defer o.refreshMu.Unlock()

	// Fetch and recover outside the state lock. Capture revisions so neither a
	// late upstream response nor a slow backup can replace newer imported state.
	type update struct {
		file       *FeatureFile
		revision   uint64
		fromBackup bool
	}
	updates := make([]update, len(o.featureEnvironments))
	var refreshErr error
	for idx, environment := range o.featureEnvironments {
		o.mu.RLock()
		revision := environment.remoteRevision
		o.mu.RUnlock()
		file, err := o.client.getFeatures(environment.token)
		if err != nil {
			err = fmt.Errorf("fetch features for %s: %w", environment.name, err)
			log.Error(err)
			refreshErr = errors.Join(refreshErr, err)

			o.mu.RLock()
			hasState := environment.remoteRevision != 0
			o.mu.RUnlock()
			if hasState || !o.Config.Backup {
				continue
			}

			file, err = o.readFeatureBackup(environment.name)
			if err != nil {
				log.Error(err)
				refreshErr = errors.Join(refreshErr, err)
				continue
			}
			updates[idx] = update{file: file, revision: revision, fromBackup: true}
			continue
		}

		owned := cloneFeatureFile(*file)
		updates[idx] = update{file: &owned, revision: revision}
	}

	o.mu.Lock()
	hasRefreshed := false
	for idx, pending := range updates {
		environment := o.featureEnvironments[idx]
		if pending.file == nil || pending.revision != environment.remoteRevision {
			updates[idx].file = nil
			continue
		}
		environment.setRemoteFeatureFile(*pending.file)
		hasRefreshed = true
		if pending.fromBackup {
			log.Infof("Loaded backup for %s", environment.name)
		}
	}
	if hasRefreshed {
		o.compileFeatureFiles()
		o.lastSync = time.Now()
	}
	o.mu.Unlock()

	// Only accepted upstream state becomes a new backup. Keep writes outside the
	// state lock but inside refreshMu so refreshes cannot reorder backups.
	if o.Config.Backup {
		for idx, pending := range updates {
			if pending.file == nil || pending.fromBackup {
				continue
			}
			name := o.featureEnvironments[idx].name
			data, err := json.Marshal(pending.file)
			if err != nil {
				log.Warnf("Failed to encode backup for %s: %v", name, err)
				continue
			}
			if err := o.store.Write(name+"-backup.json", data); err != nil {
				log.Warnf("Failed to write backup for %s: %v", name, err)
			}
		}
	}
	return refreshErr
}

func (o *OverleashContext) readFeatureBackup(name string) (*FeatureFile, error) {
	data, err := o.store.Read(name + "-backup.json")
	if err != nil {
		return nil, fmt.Errorf("read backup for %s: %w", name, err)
	}
	var file FeatureFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("decode backup for %s: %w", name, err)
	}
	return &file, nil
}
