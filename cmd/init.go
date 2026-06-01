package cmd

import "log/slog"

func runInit(args []string) int {
	dataDir := "./data"
	if len(args) > 0 {
		dataDir = args[0]
	}

	slog.Info("initializing application", "data_dir", dataDir)
	if code := runMigrateUp(nil); code != 0 {
		return code
	}
	if code := runIngest([]string{dataDir}); code != 0 {
		return code
	}

	slog.Info("application initialization complete")
	return 0
}
