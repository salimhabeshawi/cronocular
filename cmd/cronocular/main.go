package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/salimhabeshawi/cronocular/internal/audio"
	"github.com/salimhabeshawi/cronocular/internal/background"
	"github.com/salimhabeshawi/cronocular/internal/notify"
	"github.com/salimhabeshawi/cronocular/internal/timer"
	"github.com/salimhabeshawi/cronocular/internal/tui"
)

const (
	defaultFocus = 20 * time.Minute
	defaultRest  = 20 * time.Second
)

func main() {
	var (
		backgroundMode = flag.Bool("background", false, "run detached in the background")
		daemonMode     = flag.Bool("d", false, "run detached in the background")
		headless       = flag.Bool("headless", false, "run without the TUI; intended for detached background use")
		focus          = flag.Duration("focus", defaultFocus, "focus interval before a rest reminder")
		rest           = flag.Duration("rest", defaultRest, "rest interval before returning to work")
		version        = flag.Bool("version", false, "print version information")
	)
	flag.Parse()

	if *version {
		fmt.Println("cronocular 1.0.0")
		return
	}

	cfg := timer.Config{Focus: *focus, Rest: *rest}
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "cronocular: %v\n", err)
		os.Exit(2)
	}

	if *backgroundMode || *daemonMode {
		if err := background.Start(os.Args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "cronocular: failed to start in background: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("cronocular is running in the background")
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	actions := timer.Actions{
		RestStarted: func() {
			notify.Send("cronocular", "Look away for 20 seconds!")
		},
		RestEnded: func() {
			notify.Send("cronocular", "Time's up! Return to work.")
			go audio.PlayCompletionBeep()
		},
	}

	controller := timer.NewController(cfg, actions)
	controller.Start(ctx)

	if *headless {
		<-ctx.Done()
		controller.Stop()
		return
	}

	if err := tui.Run(ctx, controller); err != nil {
		controller.Stop()
		fmt.Fprintf(os.Stderr, "cronocular: %v\n", err)
		os.Exit(1)
	}
	controller.Stop()
}
