// Copyright The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0.

package index

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/tsdb/chunks"
)

func BenchmarkIndexFormats(b *testing.B) {
	ctx := context.Background()
	for _, workload := range []string{"sequential", "random", "urls"} {
		b.Run("workload="+workload, func(b *testing.B) {
			values, series := indexBenchmarkSeries(workload)
			for _, version := range []int{FormatV2, FormatV4} {
				b.Run("version="+strconv.Itoa(version), func(b *testing.B) {
					path := writeFormatFixture(b, version, series)
					info, err := os.Stat(path)
					require.NoError(b, err)
					r, err := NewFileReader(path, DecodePostingsRaw)
					require.NoError(b, err)
					defer r.Close()
					for _, op := range []string{"exact", "missing", "batch16", "enumerate", "open", "open-scan", "write"} {
						b.Run("op="+op, func(b *testing.B) {
							_, err := r.LabelValues(ctx, "instance", nil)
							require.NoError(b, err)
							sorted := slices.Clone(series)
							slices.SortFunc(sorted, labels.Compare)
							symbols := indexBenchmarkSymbols(sorted)
							target := filepath.Join(b.TempDir(), "index")
							batch := make([]string, 16)
							b.ReportAllocs()
							b.ResetTimer()
							for i := 0; i < b.N; i++ {
								switch op {
								case "write":
									w, err := NewWriterWithOptions(ctx, target, WriterOptions{FormatVersion: version})
									if err != nil {
										b.Fatal(err)
									}
									for _, symbol := range symbols {
										if err = w.AddSymbol(symbol); err != nil {
											b.Fatal(err)
										}
									}
									for j, ls := range sorted {
										if err = w.AddSeries(storage.SeriesRef(j), ls, chunks.Meta{Ref: chunks.ChunkRef(j + 1), MinTime: 1, MaxTime: 2}); err != nil {
											b.Fatal(err)
										}
									}
									if err = w.Close(); err != nil {
										b.Fatal(err)
									}
								case "open", "open-scan":
									opened, err := NewFileReader(path, DecodePostingsRaw)
									if err != nil {
										b.Fatal(err)
									}
									if op == "open-scan" {
										for _, name := range []string{"", "__name__", "instance", "job", "zone"} {
											if _, err = opened.LabelValues(ctx, name, nil); err != nil {
												b.Fatal(err)
											}
										}
									}
									if err = opened.Close(); err != nil {
										b.Fatal(err)
									}
								case "enumerate":
									got, err := r.LabelValues(ctx, "instance", nil)
									if err != nil || len(got) != len(values) {
										b.Fatalf("values %d: %v", len(got), err)
									}
								default:
									at := (i * 7919) % len(values)
									var p Postings
									var err error
									expected := 1
									switch op {
									case "exact":
										p, err = r.Postings(ctx, "instance", values[at])
									case "missing":
										p, err = r.Postings(ctx, "instance", values[at]+"~missing")
										expected = 0
									case "batch16":
										for j := range batch {
											batch[j] = values[(at+j*97)%len(values)]
										}
										p, err = r.Postings(ctx, "instance", batch...)
										expected = 16
									}
									if err != nil {
										b.Fatal(err)
									}
									n := 0
									for p.Next() {
										n++
									}
									if p.Err() != nil || n != expected {
										b.Fatalf("postings %d: %v", n, p.Err())
									}
								}
							}
							b.StopTimer()
							if op == "write" {
								written, err := os.Stat(target)
								require.NoError(b, err)
								require.Equal(b, info.Size(), written.Size())
							}
							b.ReportMetric(float64(info.Size()), "index-bytes")
							b.ReportMetric(float64(r.b.Len()-indexTOCLen-int(r.toc.PostingsTable)), "postings-table-bytes")
						})
					}
				})
			}
		})
	}
}

func indexBenchmarkSeries(workload string) ([]string, []labels.Labels) {
	values := make([]string, 32768)
	series := make([]labels.Labels, len(values))
	for i := range values {
		value := fmt.Sprintf("pod-%08d", i)
		switch workload {
		case "random":
			value = fmt.Sprintf("%x", sha256.Sum256([]byte(value)))
		case "urls":
			value = "https://api.example.net/api/v1/organizations/org-00000001/resources/" + value + "?region=eu-west-1&format=json"
		}
		values[i] = value
		series[i] = labels.FromStrings("__name__", "m", "instance", value, "job", "api", "zone", strconv.Itoa(i%16))
	}
	return values, series
}

func indexBenchmarkSymbols(series []labels.Labels) []string {
	symbols := map[string]struct{}{}
	for _, ls := range series {
		ls.Range(func(l labels.Label) {
			symbols[l.Name] = struct{}{}
			symbols[l.Value] = struct{}{}
		})
	}
	return slices.Sorted(maps.Keys(symbols))
}

func BenchmarkV4PostingsPageSize(b *testing.B) {
	ctx := b.Context()
	for _, workload := range []string{"sequential", "random", "urls"} {
		values, series := indexBenchmarkSeries(workload)
		slices.SortFunc(series, labels.Compare)
		sortedSymbols := indexBenchmarkSymbols(series)
		sortedValues := slices.Clone(values)
		slices.Sort(sortedValues)
		for _, limit := range []int{32, 128} {
			b.Run(fmt.Sprintf("workload=%s/entries=%d", workload, limit), func(b *testing.B) {
				write := func(path string) (err error) {
					w, err := NewWriterWithOptions(ctx, path, WriterOptions{FormatVersion: FormatV4})
					if err != nil {
						return err
					}
					defer func() { err = errors.Join(err, w.Close()) }()
					w.postingsPageLimit = limit
					for _, sym := range sortedSymbols {
						if err := w.AddSymbol(sym); err != nil {
							return err
						}
					}
					for i, ls := range series {
						if err := w.AddSeries(storage.SeriesRef(i), ls, chunks.Meta{Ref: chunks.ChunkRef(i + 1), MinTime: 1, MaxTime: 2}); err != nil {
							return err
						}
					}
					return nil
				}
				path := filepath.Join(b.TempDir(), "index")
				require.NoError(b, write(path))
				open := func(b *testing.B) *Reader {
					r, err := NewFileReader(path, DecodePostingsRaw)
					if err != nil {
						b.Fatal(err)
					}
					return r
				}
				lookup := func(b *testing.B, r *Reader, value string) storage.SeriesRef {
					p, err := r.Postings(ctx, "instance", value)
					if err != nil || !p.Next() {
						b.Fatalf("missing %q: %v", value, err)
					}
					ref := p.At()
					if p.Next() || p.Err() != nil {
						b.Fatalf("expected a single posting for %q: %v", value, p.Err())
					}
					return ref
				}

				// Check both sides of every page boundary on a separate reader. Looking
				// up the same values on the timed reader would validate its pages.
				check := open(b)
				pages := check.paged.groups["instance"].pages
				require.Equal(b, limit, pages[0].count)
				at, pageCount := 0, 0
				for _, group := range check.paged.groups {
					pageCount += len(group.pages)
				}
				for _, page := range pages {
					for _, v := range []string{sortedValues[at], sortedValues[at+page.count-1]} {
						var got labels.ScratchBuilder
						require.NoError(b, check.Series(lookup(b, check, v), &got, nil))
						require.Equal(b, v, got.Labels().Get("instance"))
					}
					at += page.count
				}
				require.Equal(b, len(sortedValues), at)
				all, err := check.LabelValues(ctx, "instance", nil)
				require.NoError(b, err)
				require.Equal(b, sortedValues, all)
				dirSize := binary.BigEndian.Uint64(check.b.Range(int(check.toc.PostingsTable)+16, int(check.toc.PostingsTable)+24))
				indexSize := check.Size()
				require.NoError(b, check.Close())

				b.Run("op=build", func(b *testing.B) {
					target := filepath.Join(b.TempDir(), "index")
					for b.Loop() {
						if err := write(target); err != nil {
							b.Fatal(err)
						}
					}
					b.ReportMetric(float64(pageCount), "pages")
					b.ReportMetric(float64(dirSize), "directory-bytes")
					b.ReportMetric(float64(indexSize), "index-bytes")
				})
				b.Run("op=open", func(b *testing.B) {
					for b.Loop() {
						r := open(b)
						b.StopTimer()
						require.NoError(b, r.Close())
						b.StartTimer()
					}
				})
				for _, fresh := range []bool{true, false} {
					op := "warm-lookup"
					if fresh {
						op = "first-lookup"
					}
					b.Run("op="+op, func(b *testing.B) {
						var r *Reader
						if !fresh {
							r = open(b)
							defer r.Close()
							_, err := r.LabelValues(ctx, "instance", nil)
							require.NoError(b, err)
						}
						i := 0
						for b.Loop() {
							if fresh {
								b.StopTimer()
								r = open(b)
								b.StartTimer()
							}
							lookup(b, r, values[i])
							i = (i + 7919) % len(values)
							if fresh {
								b.StopTimer()
								require.NoError(b, r.Close())
								b.StartTimer()
							}
						}
					})
				}
				b.Run("op=enumerate", func(b *testing.B) {
					r := open(b)
					defer r.Close()
					_, err := r.LabelValues(ctx, "instance", nil)
					require.NoError(b, err)
					for b.Loop() {
						got, err := r.LabelValues(ctx, "instance", nil)
						if err != nil || len(got) != len(values) {
							b.Fatalf("enumerated %d values: %v", len(got), err)
						}
					}
				})
			})
		}
	}
}
