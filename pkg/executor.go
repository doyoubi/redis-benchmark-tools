package pkg

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog/log"
)

type cmdTask struct {
	cmd command
	weight int
}

type command interface {
	request(ctx context.Context, client redis.Cmdable, key string) error
}

type setCommand struct {}

func (c setCommand) request(ctx context.Context, client redis.Cmdable, key string) error {
	_, err := client.Set(ctx, key, key, 0).Result()
	if err != nil {
		log.Err(err).Msg("failed to SET")
		return err
	}
	return nil
}

type getCommand struct {}

func (c getCommand) request(ctx context.Context, client redis.Cmdable, key string) error {
	_, err := client.Get(ctx, key).Result()
	if err != nil {
		log.Err(err).Msg("failed to GET")
		return err
	}
	return nil
}

type delCommand struct {}

func (c delCommand) request(ctx context.Context, client redis.Cmdable, key string) error {
	_, err := client.Del(ctx, key).Result()
	if err != nil {
		log.Err(err).Msg("failed to DEL")
		return err
	}
	return nil
}

var (
	_ command = setCommand{}
	_ command = getCommand{}
	_ command = delCommand{}
)

type cmdExecutor struct {
	client redis.Cmdable
	tasks []cmdTask
	cmdNum int
}

func newCmdExecutor(client redis.Cmdable, cmdNum int) *cmdExecutor {
	return &cmdExecutor{
		client: client,
		tasks: []cmdTask{
			{
				cmd: setCommand{},
				weight: 3,
			},
			{
				cmd: getCommand{},
				weight: 5,
			},
			{
				cmd: delCommand{},
				weight: 1,
			},
		},
		cmdNum: cmdNum,
	}
}

func (e *cmdExecutor) runCmd(ctx context.Context, cmd command, cmdNum int) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		err := cmd.request(ctx, e.client, "somekey")
		if err != nil {
			return err
		}
	}
}
