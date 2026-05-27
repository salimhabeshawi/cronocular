package timer

import (
	"context"
	"errors"
	"sync"
	"time"
)

type Phase string

const (
	PhaseFocus Phase = "focus"
	PhaseRest  Phase = "rest"
)

type Config struct {
	Focus time.Duration
	Rest  time.Duration
}

func (c Config) Validate() error {
	if c.Focus <= 0 {
		return errors.New("focus duration must be greater than zero")
	}
	if c.Rest <= 0 {
		return errors.New("rest duration must be greater than zero")
	}
	return nil
}

type Actions struct {
	RestStarted func()
	RestEnded   func()
}

type Snapshot struct {
	Phase     Phase
	Paused    bool
	Remaining time.Duration
	Focus     time.Duration
	Rest      time.Duration
	Cycle     int
	UpdatedAt time.Time
}

type Controller struct {
	cfg     Config
	actions Actions

	mu        sync.RWMutex
	phase     Phase
	paused    bool
	deadline  time.Time
	remaining time.Duration
	cycle     int

	stopOnce sync.Once
	stop     chan struct{}
}

func NewController(cfg Config, actions Actions) *Controller {
	now := time.Now()
	return &Controller{
		cfg:       cfg,
		actions:   actions,
		phase:     PhaseFocus,
		deadline:  now.Add(cfg.Focus),
		remaining: cfg.Focus,
		stop:      make(chan struct{}),
	}
}

func (c *Controller) Start(ctx context.Context) {
	go c.run(ctx)
}

func (c *Controller) Stop() {
	c.stopOnce.Do(func() { close(c.stop) })
}

func (c *Controller) Pause() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.paused {
		return
	}
	c.remaining = positiveDuration(time.Until(c.deadline))
	c.paused = true
}

func (c *Controller) Resume() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.paused {
		return
	}
	c.deadline = time.Now().Add(c.remaining)
	c.paused = false
}

func (c *Controller) TogglePause() {
	c.mu.RLock()
	paused := c.paused
	c.mu.RUnlock()
	if paused {
		c.Resume()
		return
	}
	c.Pause()
}

func (c *Controller) Snapshot() Snapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()

	remaining := c.remaining
	if !c.paused {
		remaining = positiveDuration(time.Until(c.deadline))
	}

	return Snapshot{
		Phase:     c.phase,
		Paused:    c.paused,
		Remaining: remaining,
		Focus:     c.cfg.Focus,
		Rest:      c.cfg.Rest,
		Cycle:     c.cycle,
		UpdatedAt: time.Now(),
	}
}

func (c *Controller) run(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stop:
			return
		case <-ticker.C:
			c.advance(time.Now())
		}
	}
}

func (c *Controller) advance(now time.Time) {
	var restStarted, restEnded bool

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.paused || now.Before(c.deadline) {
		return
	}

	switch c.phase {
	case PhaseFocus:
		c.phase = PhaseRest
		c.deadline = now.Add(c.cfg.Rest)
		c.remaining = c.cfg.Rest
		restStarted = true
	case PhaseRest:
		c.phase = PhaseFocus
		c.deadline = now.Add(c.cfg.Focus)
		c.remaining = c.cfg.Focus
		c.cycle++
		restEnded = true
	}

	go func() {
		if restStarted && c.actions.RestStarted != nil {
			c.actions.RestStarted()
		}
		if restEnded && c.actions.RestEnded != nil {
			c.actions.RestEnded()
		}
	}()
}

func positiveDuration(d time.Duration) time.Duration {
	if d < 0 {
		return 0
	}
	return d
}
