package helper

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"time"
)

func StartCPUTrace() func() {
	fileName := fmt.Sprintf("cpu-%d.pprof", time.Now().Unix())
	f, err := os.Create(fileName)
	if err != nil {
		log.Fatal().Msgf("could not create %s: %v", fileName, err)
	}

	log.Info().Msgf("creating CPU profile file %s", fileName)
	if err := pprof.StartCPUProfile(f); err != nil {
		log.Fatal().Msgf("could not start CPU profile: %v", err)
	}

	return func() {
		f.Close()
		pprof.StopCPUProfile()
	}
}

func StartMemoryProfile() func() {
	fileName := fmt.Sprintf("mem-%d.trace", time.Now().Unix())
	f, err := os.Create(fileName)
	if err != nil {
		log.Fatal().Msgf("could not create %s: %v", fileName, err)
	}
	log.Info().Msgf("creating memory profile file %s", fileName)

	runtime.GC() // get up-to-date statistics
	if err := trace.Start(f); err != nil {
		log.Fatal().Msgf("could not write memory profile: %v", err)
	}

	return func() {
		trace.Stop()
		f.Close()
	}
}
