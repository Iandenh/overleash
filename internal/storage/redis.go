package storage

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/charmbracelet/log"
	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	client     *redis.Client
	pubsubCh   string
	instanceId string
}

type RedisConfig struct {
	Addr        string
	Password    string
	DB          int
	UseSentinel bool
	MasterName  string
	Sentinels   []string
	Channel     string
}

type message struct {
	Key        string `json:"key"`
	Data       []byte `json:"data"`
	InstanceID string `json:"instance_id"`
}

func NewRedisStore(cfg RedisConfig) *RedisStore {
	instanceID := strconv.FormatInt(time.Now().UnixNano(), 10)

	var rdb *redis.Client

	if cfg.UseSentinel {
		// Sentinel mode
		rdb = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:    cfg.MasterName,
			SentinelAddrs: cfg.Sentinels,
			Password:      cfg.Password,
			DB:            cfg.DB,
		})
	} else {
		// Normal standalone Redis
		rdb = redis.NewClient(&redis.Options{
			Addr:     cfg.Addr,
			Password: cfg.Password,
			DB:       cfg.DB,
		})
	}

	return &RedisStore{
		client:     rdb,
		pubsubCh:   cfg.Channel,
		instanceId: instanceID,
	}
}

func (r *RedisStore) Read(key string) ([]byte, error) {
	return r.client.Get(context.Background(), key).Bytes()
}

func (r *RedisStore) Write(key string, data []byte) error {
	ctx := context.Background()

	if err := r.client.Set(ctx, key, data, 0).Err(); err != nil {
		return err
	}
	msg := message{
		Key: key, Data: data, InstanceID: r.instanceId,
	}

	b, _ := json.Marshal(msg)
	return r.client.Publish(ctx, r.pubsubCh, b).Err()
}

// Subscribe listens for update notifications and calls handler
func (r *RedisStore) Subscribe(ctx context.Context, handler func(key string, data []byte)) error {
	sub := r.client.Subscribe(ctx, r.pubsubCh)
	ch := sub.Channel()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				var e message
				if err := json.Unmarshal([]byte(msg.Payload), &e); err != nil {
					log.Printf("invalid event payload: %s", msg.Payload)
					continue
				}
				if e.InstanceID == r.instanceId {
					continue
				}
				handler(e.Key, e.Data)
			}
		}
	}()
	return nil
}
