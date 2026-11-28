// Copyright The Prometheus Authors
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

package index

import (
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/tsdb/chunks"
)

func TestSeriesRefForOffset(t *testing.T) {
	const last = uint64(math.MaxUint32) * seriesByteAlign
	ref, err := seriesRefForOffset(last)
	require.NoError(t, err)
	require.Equal(t, uint32(math.MaxUint32), ref)

	_, err = seriesRefForOffset(last + seriesByteAlign)
	require.ErrorContains(t, err, "exceeds maximum representable offset")
	_, err = seriesRefForOffset(last - 1)
	require.ErrorContains(t, err, "not 16-byte aligned")
}

func TestWriter_WritePostings_V2FileLimit(t *testing.T) {
	dir := t.TempDir()
	f, err := NewFileWriter(filepath.Join(dir, "index"))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, f.Close()) })
	postings, err := NewFileWriter(filepath.Join(dir, "postings"))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, postings.Close()) })
	require.NoError(t, postings.Write(make([]byte, 8)))

	// Only four bytes remain. The direct temporary-file copy must enforce the
	// same limit as an ordinary Write, before the postings table is appended.
	f.pos = uint64(math.MaxUint32)*seriesByteAlign - 4
	w := &Writer{f: f, fP: postings}
	require.ErrorIs(t, w.writePostings(), ErrIndexExceeds64GiB)
	require.NoError(t, f.Write(make([]byte, 4)))
	require.ErrorIs(t, f.Write([]byte{0}), ErrIndexExceeds64GiB)
}

func TestNewWriterWithOptions_V4TemporaryFileLimits(t *testing.T) {
	if strconv.IntSize < 64 {
		t.Skip("requires a 64-bit address space")
	}
	w, err := NewWriterWithOptions(t.Context(), filepath.Join(t.TempDir(), "index"), WriterOptions{FormatVersion: FormatV4})
	require.NoError(t, err)
	// Either temporary file can exceed the legacy ceiling independently while
	// collecting postings lists or the keys and offsets needed to build pages.
	const size = uint64(1) << 36
	require.NoError(t, w.fP.checkSize(size))
	require.NoError(t, w.fPO.checkSize(size))
	require.NoError(t, w.Close())
}

func TestWriter_AddSeries_V4SeriesReach(t *testing.T) {
	path := filepath.Join(t.TempDir(), "index")
	w, err := NewWriterWithOptions(t.Context(), path, WriterOptions{FormatVersion: FormatV4})
	require.NoError(t, err)
	require.NoError(t, w.AddSymbol("a"))
	require.NoError(t, w.ensureStage(idxStageSeries))
	pos := w.f.pos
	// Rounding even one byte beyond the last valid start would wrap the ref to zero.
	w.f.pos = uint64(math.MaxUint32)*seriesByteAlign + 1
	ls := labels.FromStrings("a", "a")
	require.ErrorContains(t, w.AddSeries(0, ls), "exceeds maximum representable offset")
	require.Equal(t, uint64(math.MaxUint32)*seriesByteAlign+1, w.f.pos)
	w.f.pos = pos
	require.NoError(t, w.Close())
}

func TestWriter_V4LargeIndexRoundTrip(t *testing.T) {
	if os.Getenv("PROMETHEUS_TEST_V4_LARGE") != "1" || strconv.IntSize < 64 {
		t.Skip("requires PROMETHEUS_TEST_V4_LARGE=1, a sparse filesystem and a 64-bit address space")
	}
	ls := labels.FromStrings("a", "b")
	// Measure the symbol framing on a small block so a sparse gap before the
	// symbols moves the first series to exactly the last representable start.
	small, err := os.ReadFile(writeFormatFixture(t, FormatV4, []labels.Labels{ls}))
	require.NoError(t, err)
	toc, err := NewTOCFromByteSlice(realByteSlice(small))
	require.NoError(t, err)
	const lastStart = uint64(math.MaxUint32) * seriesByteAlign
	gap := lastStart - ((toc.Series + seriesByteAlign - 1) &^ uint64(seriesByteAlign-1))

	path := filepath.Join(t.TempDir(), "index")
	w, err := NewWriterWithOptions(t.Context(), path, WriterOptions{FormatVersion: FormatV4})
	require.NoError(t, err)
	require.NoError(t, w.f.Flush())
	pos, err := w.f.f.Seek(int64(gap), io.SeekCurrent)
	require.NoError(t, err)
	w.f.pos = uint64(pos)
	require.NoError(t, w.AddSymbol("a"))
	require.NoError(t, w.AddSymbol("b"))
	// A wide chunk time range makes the record itself extend across 64 GiB.
	require.NoError(t, w.AddSeries(0, ls, chunks.Meta{Ref: 1, MinTime: 0, MaxTime: math.MaxInt64}))
	require.Greater(t, w.f.pos, uint64(1)<<36)
	require.NoError(t, w.Close())

	r, err := NewFileReader(path, DecodePostingsRaw)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, r.Close()) })
	require.Greater(t, r.toc.PostingsTable, uint64(1)<<36)
	for _, key := range []labels.Label{{}, {Name: "a", Value: "b"}} {
		p, err := r.Postings(t.Context(), key.Name, key.Value)
		require.NoError(t, err)
		refs, err := ExpandPostings(p)
		require.NoError(t, err)
		require.Equal(t, []storage.SeriesRef{math.MaxUint32}, refs)
		var got labels.ScratchBuilder
		require.NoError(t, r.Series(refs[0], &got, nil))
		require.True(t, labels.Equal(ls, got.Labels()))
	}
}
