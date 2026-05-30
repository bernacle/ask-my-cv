package embedding

import "testing"

func TestVectorLiteral(t *testing.T) {
	got := VectorLiteral([]float64{0.1, -2, 3.25})
	want := "[0.1,-2,3.25]"
	if got != want {
		t.Fatalf("VectorLiteral() = %q, want %q", got, want)
	}
}
