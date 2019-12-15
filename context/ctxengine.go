package context

import (
	"context"
	"errors"

	"github.com/elvisNg/broccoli/engine"
)

type ctxEngineMarker struct{}

// engine
type ctxEngine struct {
	ng engine.Engine
}

var (
	ctxEngineKey = &ctxEngineMarker{}
)

func ExtractEngine(ctx context.Context) (ng engine.Engine, err error) {
	c, ok := ctx.Value(ctxEngineKey).(*ctxEngine)
	if !ok || c == nil {
		return nil, errors.New("ctxEngine was not set or nil")
	}

	ng = c.ng
	return
}

func EngineToContext(ctx context.Context, ng engine.Engine) context.Context {
	c := &ctxEngine{
		ng,
	}
	return context.WithValue(ctx, ctxEngineKey, c)
}
