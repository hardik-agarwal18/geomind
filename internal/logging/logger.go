package logging

import (
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-colorable"

	"geomind/internal/config"
)

func New(cfg config.Config) *slog.Logger {
	var handler slog.Handler

	if cfg.AppEnv == "development" {
		// Colored, human-readable logs for local development.
		handler = tint.NewTextHandler(
			colorable.NewColorable(os.Stdout),
			&tint.Options{
				Level:     cfg.LogLevel,
				AddSource: true,
				NoColor:   os.Getenv("NO_COLOR") != "",
			},
		)
	} else {
		// Machine-readable JSON logs for production and testing.
		handler = slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{
				Level: cfg.LogLevel,
			},
		)
	}

	return slog.New(handler).With(
		slog.String("service", "geomind-api"),
		slog.String("environment", cfg.AppEnv),
	)
}
