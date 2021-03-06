package context

import (
	"context"
	"errors"
	"github.com/elvisNg/broccoli/mysql/zmysql"
)

type ctxMysqlMarker struct{}

type ctxMysql struct {
	cli zmysql.Mysql
}

var (
	ctxMysqlKey = &ctxMysqlMarker{}
)

// ExtractMysql takes the mysql from ctx.
func ExtractMysql(ctx context.Context) (c zmysql.Mysql, err error) {
	r, ok := ctx.Value(ctxMysqlKey).(*ctxMysql)
	if !ok || r == nil {
		return nil, errors.New("ctxMysql was not set or nil")
	}
	if r.cli == nil {
		return nil, errors.New("ctxMysql.cli was not set or nil")
	}

	c = r.cli
	return
}

// MysqlToContext adds the mysql to the context for extraction later.
// Returning the new context that has been created.
func MysqlToContext(ctx context.Context, c zmysql.Mysql) context.Context {
	r := &ctxMysql{
		cli: c,
	}
	return context.WithValue(ctx, ctxMysqlKey, r)
}
