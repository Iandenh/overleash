package storage

import (
	"context"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/spf13/viper"
)

type Store interface {
	Read(key string) ([]byte, error)
	Write(key string, data []byte) (writeErr error)
}

type EventStore interface {
	Store
	Subscribe(ctx context.Context, handler func(key string, data []byte)) error
}

func NewStoreFromConfig() Store {
	backend := viper.GetString("storage")

	switch backend {
	case "file":
		return NewFileStore()

	case "redis":
		cfg := RedisConfig{
			Addr:        viper.GetString("redis_addr"),
			Password:    viper.GetString("redis_password"),
			DB:          viper.GetInt("redis_db"),
			UseSentinel: viper.GetBool("redis_sentinel"),
			MasterName:  viper.GetString("redis_master"),
			Channel:     viper.GetString("redis_channel"),
		}
		if cfg.UseSentinel {
			sentinels := viper.GetString("redis_sentinels")
			if sentinels != "" {
				cfg.Sentinels = strings.Split(sentinels, ",")
			}
		}
		return NewRedisStore(cfg)

	default:
		log.Fatalf("invalid storage backend: %s", backend)
		return nil
	}
}
