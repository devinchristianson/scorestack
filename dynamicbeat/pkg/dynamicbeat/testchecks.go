package dynamicbeat

import (
	"os"
	"os/signal"
	"runtime"
	"sync"

	"github.com/scorestack/scorestack/dynamicbeat/pkg/check"
	"github.com/scorestack/scorestack/dynamicbeat/pkg/checksource"
	"github.com/scorestack/scorestack/dynamicbeat/pkg/run"
	"go.uber.org/zap"
)

// Run starts dynamicbeat.
func TestChecks(f *checksource.Filesystem) error {
	zap.S().Infof("dynamicbeat is running to test checks! Hit CTRL-C to stop it.")

	// Set up CTRL+C handler
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	// Get initial check definitions
	var defs []check.Config
	// TODO: Find a better way for looping until we can hit Elasticsearch
	zap.S().Infof("Getting initial check definitions...")
	defs, err := f.LoadAll()
	if len(defs) == 0 {
		zap.S().Fatal("No checks found")
	}
	if err != nil {
		zap.S().Fatalf("Failed to load checks, error was: %s", err)
	}

	// Start publisher goroutine
	results := make(chan check.Result)
	published := make(chan uint64)
	go publishTestEvents(results, published)

	var wg sync.WaitGroup

	// Making quit handles in goroutine because we don't need the for look anymore (when running one round)
	go func() {
		<-quit
		// Purposefully don't wait for checks.RunChecks goroutines to exit because that could take 30s
		//wg.Wait()

		// Close the event publishing queue so the publishEvents goroutine will exit
		close(results)

		// Wait for all events to be published
		<-published
		close(published)
		os.Exit(1)
	}()

	// The logic for running a round is moved here so that it can be executed immediately as well as per the ticker
	zap.S().Infof("Number of goroutines: %d", runtime.NumGoroutine())
	zap.S().Infof("Starting a series of %d checks", len(defs))
	// Start the goroutine
	started := make(chan bool)
	wg.Add(1)
	go func() {
		defer wg.Done()
		run.Round(defs, results, started)
	}()
	<-started
	zap.S().Infof("Started series of checks")
	wg.Wait()
	return nil
}

func publishTestEvents(results <-chan check.Result, out chan<- uint64) {
	published := uint64(0)
	for result := range results {
		passed := "failed"
		if result.Passed {
			passed = "passed"
		}
		zap.S().Infof("Check %s for team %s %s: %s", result.Metadata.Name, result.Metadata.Group, passed, result.Message)
		published++
	}
	out <- published
}
