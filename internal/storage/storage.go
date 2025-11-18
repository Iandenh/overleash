package storage

import (
	"context"
	"strings"

	"github.com/Iandenh/overleash/config"
	"github.com/charmbracelet/log"
)

type Store interface {
	Read(key string) ([]byte, error)
	Write(key string, data []byte) (writeErr error)
}

type EventStore interface {
	Store
	Subscribe(ctx context.Context, handler func(key string, data []byte)) error
}

type BroadcastStore interface {
	Store
	Broadcast(key string, data []byte) error
}

func NewStoreFromConfig(cfg *config.Config) Store {
	backend := cfg.Storage

	switch backend {
	case "file":
		return NewFileStore()

	case "redis":
		redisCfg := RedisConfig{
			Addr:        cfg.RedisAddr,
			Password:    cfg.RedisPassword,
			DB:          cfg.RedisDB,
			UseSentinel: cfg.RedisSentinel,
			MasterName:  cfg.RedisMaster,
			Channel:     cfg.RedisChannel,
		}
		if redisCfg.UseSentinel {
			sentinels := cfg.RedisSentinels
			if sentinels != "" {
				redisCfg.Sentinels = strings.Split(sentinels, ",")
			}
		}
		return NewRedisStore(redisCfg)

	case "null":
		return &NullStore{}

	default:
		log.Fatalf("invalid storage backend: %s", backend)
		return nil
	}
}
