package misc

import (
	"strconv"
	"strings"
	"testing"

	"github.com/AlexanderThaller/logger"
)

const (
	name = "misc.test"
)

func init() {
	logger.SetLevel(name, logger.Info)
}

func TestReplaceNth(te *testing.T) {
	l := logger.New(name + ".ReplaceNth")

	t := [][]string{
		{"", "", "", "0", ""},
		{"A", "A", "B", "1", "B"},
		{"AA", "A", "B", "2", "AB"},
		{"AAA", "A", "B", "3", "AAB"},
		{"AAA AAA AAA AAA AAA", "A", "B", "4", "AAA BAA ABA AAB AAA"},
		{"", "", "", "-1", ""},
		{"AAA", "", "B", "1", "AAA"},
		{"AAA", "A", "", "1", "AAA"},
	}

	for _, d := range t {
		l.Debug("Data: ", d)

		st := d[0]
		ol := d[1]
		ne := d[2]
		nt, _ := strconv.Atoi(d[3])

		o := ReplaceNth(st, ol, ne, nt)
		x := d[4]

		if o != x {
			l.Error("OUTPUT: ", o)
			l.Error("EXPECT: ", x)

			te.Fail()
		}
	}
}

func BenchmarkReplaceNth0(be *testing.B) {
	d := []string{"", "", "", "0"}

	st := d[0]
	ol := d[1]
	ne := d[2]
	nt, _ := strconv.Atoi(d[3])

	be.ResetTimer()
	for i := 0; i < be.N; i++ {
		ReplaceNth(st, ol, ne, nt)
	}
}

func BenchmarkReplaceNthAAA5(be *testing.B) {
	d := []string{"AAA AAA AAA AAA AAA", "A", "B", "4"}

	st := d[0]
	ol := d[1]
	ne := d[2]
	nt, _ := strconv.Atoi(d[3])

	be.ResetTimer()
	for i := 0; i < be.N; i++ {
		ReplaceNth(st, ol, ne, nt)
	}
}

func BenchmarkReplaceNthAAA10000(be *testing.B) {
	d := []string{"", "A", "B", "4"}

	st := strings.Repeat("AAA ", 10000)
	ol := d[1]
	ne := d[2]
	nt, _ := strconv.Atoi(d[3])

	be.ResetTimer()
	for i := 0; i < be.N; i++ {
		ReplaceNth(st, ol, ne, nt)
	}
}
