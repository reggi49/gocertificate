package pdf

import (
	"testing"

	"certificate-generator/internal/domain"
)

func benchVars() map[string]string {
	return map[string]string{
		"recipient_name": "Maharani Syahputra",
		"issue_date":     "08 DESEMBER 2026",
		"event_location": "MASJID ISTIQLAL",
		"event_address":  "JL. TAMAN WIJAYA KUSUMA • PASAR BARU",
		"event_city":     "KECAMATAN SAWAH BESAR • JAKARTA PUSAT",
	}
}

// BenchmarkRenderDefaultJanaiz measures CPU time, allocations and output size of a single certificate render.
func BenchmarkRenderDefaultJanaiz(b *testing.B) {
	r := NewGoPDFRenderer("")
	doc := domain.DefaultJanaizDocument()
	vars := benchVars()
	b.ReportAllocs()
	b.ResetTimer()
	var size int
	for i := 0; i < b.N; i++ {
		out, err := r.Render(doc, vars, b.TempDir())
		if err != nil {
			b.Fatal(err)
		}
		size = len(out)
	}
	b.ReportMetric(float64(size)/1024, "KB/pdf")
}
