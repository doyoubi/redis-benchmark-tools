package main

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/doyoubi/redis-benchmark-tools/pkg"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "benchmark",
		Short: "benchmark is a Redis benchmarking tool like redis-benchmark",
	}

	redisAddress := rootCmd.PersistentFlags().StringP("address", "a", "127.0.0.1:6379", "Server address (default 127.0.0.1:6379)")
	concurrentNum := rootCmd.PersistentFlags().IntP("concurrent-number", "c", 50, "Number of parallel clients")
	cmdNum := rootCmd.PersistentFlags().IntP("requests", "n", 100000, "Total number of requests (default 100000)")
	enableCluster := rootCmd.PersistentFlags().BoolP("cluster", "", false, "Enable cluster mode")
	keySpaceLen := rootCmd.PersistentFlags().IntP("keyspacelen", "r", 100000, "Use random keys for SET/GET/INCR")

	keyPrefix := rootCmd.PersistentFlags().StringP("key-prefix", "", "benchmark", "Key prefix (default benchmark)")

	rootCmd.Run = func(cmd *cobra.Command, args []string) {
		err := pkg.Run(*enableCluster, *redisAddress, *cmdNum, *concurrentNum, *keyPrefix, *keySpaceLen)
		if err != nil {
			log.Err(err).Msg("benchmark failed")
		}
	}

	err := rootCmd.Execute()
	if err != nil {
		log.Err(err).Msg("failed to parse command arguments")
		return
	}
}
