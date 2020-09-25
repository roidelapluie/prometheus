package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/tsdb"
	"github.com/prometheus/prometheus/tsdb/chunks"
)

func main() {
	db, err := tsdb.OpenDBReadOnly(os.Args[1], log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout)))
	if err != nil {
		panic(err)
	}
	start := time.Now()
	for _, t := range [][2]int64{
		{1592719200000, 1593302400000},
		{1598551200000, 1599134400000},
	} {
		q, err := db.Querier(context.TODO(), t[0], t[1])
		if err != nil {
			panic(err)
		}
		sel, _, err := q.Select(false, nil, labels.MustNewMatcher(labels.MatchRegexp, "__name__", ".*"))
		if err != nil {
			panic(err)
		}
		for sel.Next() {
		}
		q.Close()
	}
	fmt.Printf("%v\n", time.Since(start))
	chunks.PrintStats()
}
