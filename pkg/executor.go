package pkg

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	wr "github.com/mroth/weightedrand"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
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
	if err != nil && err != redis.Nil {
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
	concurrentNum int

	keyGen *keyGenerator
	samples *sampleCollector
}

func newCmdExecutor(client redis.Cmdable, cmdNum, concurrentNum int, keyPrefix string, keySpaceLen int) *cmdExecutor {
	var ms uint64 = 1000000

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
		concurrentNum: concurrentNum,
		keyGen: newKeyGenerator(keyPrefix, keySpaceLen),
		samples: newSampleCollector(ms),
	}
}

type benchmarkResult struct {
	Histogram *Histogram
	Duration time.Duration
}

func (e *cmdExecutor) run(ctx context.Context) (*benchmarkResult, error) {
	group, ctx := errgroup.WithContext(ctx)

	sampleStopped := make(chan bool, 1)
	go func() {
		if err := e.samples.run(ctx); err != nil {
			log.Err(err).Msg("failed to collect samples")
		}
		sampleStopped <- true
	}()

	weightSum := 0
	for _, task := range e.tasks {
		weightSum += task.weight
	}

	start := time.Now()
	for i := 0; i != e.concurrentNum; i++ {
		n := e.cmdNum / e.concurrentNum

		group.Go(func() error {
			return e.runCmd(ctx, n)
		})
	}

	err := group.Wait()
	if err != nil {
		return nil, err
	}
	duration := time.Since(start)

	log.Info().Msg("stopping sample collector")
	e.samples.stop()
	<-sampleStopped
	log.Info().Msg("all stopped")

	result := &benchmarkResult{
		Histogram: e.samples.histogram,
		Duration: duration,
	}
	return result, nil
}

func (e *cmdExecutor) runCmd(ctx context.Context, cmdNum int) error {
	const batchSize int = 100
	ds := make([]uint64, 0, batchSize)

	choices := make([]wr.Choice, 0, len(e.tasks))
	for _, task := range e.tasks {
		choices = append(choices, wr.Choice{
			Item: task.cmd,
			Weight: uint(task.weight),
		})
	}
	chooser, err := wr.NewChooser(choices...)
	if err != nil {
		log.Err(err).Msg("failed to create chooser")
		return nil
	}

	for i := 0; i != cmdNum; i++ {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		cmd := chooser.Pick().(command)
		key := e.keyGen.genKey()

		start := time.Now()
		err := cmd.request(ctx, e.client, key)
		if err != nil {
			return err
		}
		d := time.Since(start)

		ds = append(ds, uint64(d.Nanoseconds()))
		if i % batchSize == batchSize - 1 {
			e.samples.add(ds)
			// Reference `ds` is moved to samples, need to create a new one.
			ds = make([]uint64, 0, batchSize)
		}
	}

	remaining := cmdNum % batchSize
	if remaining != 0 {
		e.samples.add(ds[:remaining])
	}

	return nil
}
