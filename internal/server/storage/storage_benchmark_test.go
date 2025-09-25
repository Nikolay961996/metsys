package storage

import (
	"testing"
)

func BenchmarkMemStorageDBGet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := NewMemStorage()
		s.SetGauge("111", 111.1)
		s.SetGauge("112", 111.1)
		s.SetGauge("113", 111.1)
		s.SetGauge("114", 111.1)
		s.SetGauge("115", 111.1)
		s.SetGauge("116", 111.1)
		s.SetGauge("117", 111.1)
		s.SetGauge("118", 111.1)
		s.SetGauge("119", 111.1)
		_, _ = s.GetGauge("111")
		_, _ = s.GetGauge("11x")

		s.AddCounter("221", 222)
		s.AddCounter("222", 222)
		s.AddCounter("223", 222)
		s.AddCounter("224", 222)
		s.AddCounter("225", 222)
		s.AddCounter("226", 222)
		s.AddCounter("227", 222)
		s.AddCounter("228", 222)
		s.AddCounter("229", 222)
		_, _ = s.GetCounter("222")
		_, _ = s.GetCounter("22x")

		_ = s.GetAll()
	}
}
