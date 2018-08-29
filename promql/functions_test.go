// Copyright 2015 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package promql

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/pkg/timestamp"
	"github.com/prometheus/prometheus/util/testutil"
)

func TestDeriv(t *testing.T) {
	// https://github.com/prometheus/prometheus/issues/2674#issuecomment-315439393
	// This requires more precision than the usual test system offers,
	// so we test it by hand.
	storage := testutil.NewStorage(t)
	defer storage.Close()
	engine := NewEngine(nil, nil, 10, 10*time.Second)

	a, err := storage.Appender()
	testutil.Ok(t, err)

	metric := labels.FromStrings("__name__", "foo")
	a.Add(metric, 1493712816939, 1.0)
	a.Add(metric, 1493712846939, 1.0)

	err = a.Commit()
	testutil.Ok(t, err)

	query, err := engine.NewInstantQuery(storage, "deriv(foo[30m])", timestamp.Time(1493712846939))
	testutil.Ok(t, err)

	result := query.Exec(context.Background())
	testutil.Ok(t, result.Err)

	vec, _ := result.Vector()
	testutil.Assert(t, len(vec) == 1, "Expected 1 result, got %d", len(vec))
	testutil.Assert(t, vec[0].V == 0.0, "Expected 0.0 as value, got %f", vec[0].V)
}

func TestSumNameChange(t *testing.T) {
	// https://github.com/prometheus/prometheus/issues/4562
	// This requires more precision than the usual test system offers,
	// so we test it by hand.
	storage := testutil.NewStorage(t)
	defer storage.Close()
	engine := NewEngine(nil, nil, 10, 10*time.Second)

	a, err := storage.Appender()
	testutil.Ok(t, err)

	metric1 := labels.FromStrings("__name__", "node_network_receive_drop", "iface", "eth0")
	a.Add(metric1, 1535541100395, 107315171.0)
	a.Add(metric1, 1535541130395, 107315172.0)
	a.Add(metric1, 1535541160395, 107315217.0)

	metric2 := labels.FromStrings("__name__", "node_network_receive_drop_total", "iface", "eth0")
	a.Add(metric2, 1535541190396, 107315254.0)
	a.Add(metric2, 1535541220395, 107315281.0)
	a.Add(metric2, 1535541250395, 107315334.0)
	a.Add(metric2, 1535541280395, 107315405.0)
	a.Add(metric2, 1535541310395, 107315457.0)
	a.Add(metric2, 1535541340395, 107315529.0)
	a.Add(metric2, 1535541370395, 107315598.0)
	a.Add(metric2, 1535541400395, 107315649.0)
	a.Add(metric2, 1535541430395, 107315697.0)

	err = a.Commit()
	testutil.Ok(t, err)

	query, err := engine.NewRangeQuery(storage, "sum(rate({__name__=~\"node_network_receive_drop_total|node_network_receive_drop\"}[5m]) > 0)", timestamp.Time(1535541100395), timestamp.Time(1535541430395), time.Second)
	testutil.Ok(t, err)

	result := query.Exec(context.Background())
	testutil.Ok(t, result.Err)

	vec, _ := result.Matrix()
	testutil.Assert(t, len(vec) == 1, "Expected 1 result, got %d", len(vec))
	testutil.Assert(t, len(vec[0].Points) == 301, "Expected 301 result, got %d", len(vec[0].Points))
}
