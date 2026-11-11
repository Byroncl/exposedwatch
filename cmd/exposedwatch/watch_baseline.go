package main

import (
	"fmt"
	"io"
	"slices"

	"github.com/Byroncl/exposedwatch/internal/app"
	"github.com/Byroncl/exposedwatch/internal/ports"
)

// baselineLoader reads a trusted baseline; production wires the baselinefile
// store, tests a fake.
type baselineLoader func(path ports.BaselinePath) (ports.BaselineLoad, error)

// baselineReloader keeps the evaluator in step with the baseline file. A
// failed reload never replaces the last good baseline: the evaluator keeps
// using it, and the state change is logged once, never every cycle.
type baselineReloader struct {
	file   string
	load   baselineLoader
	logger watchLogger
	// failureReason is the reason of the failure being reported; empty while
	// the file loads fine. A new reason is logged, the same one is not.
	failureReason string
	// entries and savedOn describe the baseline in use; warnings are the
	// ones last logged.
	entries  int
	savedOn  string
	warnings []string
	// primed is true when the start-up load already covers the next cycle.
	primed bool
}

// startBaseline prepares baseline mode. Without --baseline it returns nils.
// A baseline that cannot be loaded makes watch refuse to start: it never
// falls back to running without one.
func startBaseline(runtime commandRuntime, options WatchOptions, logger watchLogger, stderr io.Writer) (*baselineReloader, *app.BaselineAcceptance, error) {
	if options.BaselineFile == "" {
		return nil, nil, nil
	}
	reloader := &baselineReloader{file: options.BaselineFile, load: runtime.LoadBaseline, logger: logger}
	loaded, err := reloader.start()
	if err != nil {
		fmt.Fprintf(stderr, baselineLoadFailedFormat, err)
		return nil, nil, err
	}
	acceptance := app.NewBaselineAcceptance(loaded.Baseline)
	reloader.entries, reloader.savedOn = acceptance.EntryCount(), acceptance.SavedOn()
	reloader.logWarnings(loaded.Warnings)
	return reloader, acceptance, nil
}

// start loads the baseline at start-up. Its error means watch must refuse
// to run: there is no last good baseline to fall back to.
func (reloader *baselineReloader) start() (ports.BaselineLoad, error) {
	loaded, err := reloader.load(ports.BaselinePath(reloader.file))
	if err != nil {
		return ports.BaselineLoad{}, err
	}
	reloader.primed = true
	return loaded, nil
}

// refresh reloads the baseline at the start of a cycle and applies it.
func (reloader *baselineReloader) refresh(evaluator *app.AlertEvaluator) {
	if reloader.primed {
		reloader.primed = false
		return
	}
	loaded, err := reloader.load(ports.BaselinePath(reloader.file))
	if err != nil {
		if reason := err.Error(); reason != reloader.failureReason {
			reloader.failureReason = reason
			reloader.logger.Logf(logBaselineReloadFailedFormat, err)
		}
		return
	}
	if reloader.failureReason != "" {
		reloader.failureReason = ""
		reloader.logger.Logf(logBaselineRecovered)
	}
	acceptance := app.NewBaselineAcceptance(loaded.Baseline)
	reloader.logChange(acceptance)
	reloader.logWarnings(loaded.Warnings)
	evaluator.UseBaseline(acceptance)
}

// logChange logs one line when a good reload changed the entry count or the
// saved date, and marks a large widening of what is accepted.
func (reloader *baselineReloader) logChange(acceptance *app.BaselineAcceptance) {
	entries, savedOn := acceptance.EntryCount(), acceptance.SavedOn()
	if entries == reloader.entries && savedOn == reloader.savedOn {
		return
	}
	marker := ""
	if entries-reloader.entries >= widenedEntryGrowth || entries > widenedGrowthFactor*reloader.entries {
		marker = widenedMarker
	}
	reloader.logger.Logf(logBaselineReloadedFormat, entries, reloader.entries, savedOn, marker)
	reloader.entries, reloader.savedOn = entries, savedOn
}

// logWarnings logs the load warnings only when they differ from the last ones.
func (reloader *baselineReloader) logWarnings(warnings []string) {
	if slices.Equal(warnings, reloader.warnings) {
		return
	}
	reloader.warnings = slices.Clone(warnings)
	for _, warning := range warnings {
		reloader.logger.Logf(logBaselineWarningFormat, warning)
	}
}
