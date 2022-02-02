package pkg

import (
	"context"
	"github.com/go-redis/redis/v8"
)

func Run(enableCluster bool, redisAddress string, cmdNum, concurrentNum int, keyPrefix string, keySpaceLen int) error {
	var client redis.Cmdable
	if enableCluster {
		client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs: []string{redisAddress},
		})
	} else {
		client = redis.NewClient(&redis.Options{
			Addr: redisAddress,
		})
	}

	executor := newCmdExecutor(client, cmdNum, concurrentNum, keyPrefix, keySpaceLen)
	result, err := executor.run(context.Background())
	if err != nil {
		return err
	}

	printer := resultPrinter{}
	return printer.output(result)
}
