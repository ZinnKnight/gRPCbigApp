package EventActionMockOfOutbox

import "github.com/jackc/pgx/v5/pgxpool"

type TopicResolver func(e Event) string

var _ Emmiter = (*Writer)

type Writer struct {
	pool *pgxpool.Pool
}
