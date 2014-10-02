package logger

import (
	"bytes"
	"testing"
)

const (
	namet = name + ".Test"
)

func TestGetLevel(t *testing.T) {
	n := New("logger.Test.GetLevel")

	n.Info(n, "Starting")
	m := make(map[Logger]Priority)
	m[""] = DefaultPriority
	m["."] = DefaultPriority
	m["Test"] = DefaultPriority
	m[".Test"] = DefaultPriority

	SetLevel("Test2", Emergency)
	m["Test2"] = Emergency
	m["Test2.Test"] = Emergency
	m["Test2.Test.Test"] = Emergency
	m["Test2.Test.Test.Test"] = Emergency
	m["Test2.Test.Test.Test.Test"] = Emergency
	m["Test2.Test.Test.Test.Test.Test"] = Emergency

	for k, v := range m {
		o := GetLevel(k)
		if o != v {
			n.Error(n, "GOT: '", o, "', EXPECED: '", v, "'", ", KEY: '", k, "'")
			t.Fail()
		}
		n.Debug(n, "GOT: '", o, "', EXPECED: '", v, "'", ", KEY: '", k, "'")
	}
	n.Info(n, "Finished")
}

func TestLoggerGetLevel(t *testing.T) {
	l := New("TestLoggetGetLevel")
	l.SetLevel(Notice)

	level := l.GetLevel()

	if level != Notice {
		t.Fail()
	}
}

func TestLoggerSetNoColor(t *testing.T) {
	l := New("TestLoggetSetNoColor")
	l.SetLevel(Notice)

	l.SetNoColor(true)

	log := list.GetLogger(l)
	if log.NoColor != true {
		t.Fail()
	}
}

func TestSetLevelFail(t *testing.T) {
	l := New(namet + ".SetLevel.Fail")

	m := Disable + 1
	v := "priority does not exist"

	n := New("Test")
	o := n.SetLevel(m)

	if v != o.Error() {
		l.Critical("GOT: '", o, "', EXPECED: '", v, "'")
		t.Fail()
	}
}

func TestSetTimeFormat(t *testing.T) {
	l := New(namet + ".SetTimeFormat")

	m := "2005"
	v := ""

	n := New("Test")
	o := n.SetTimeFormat(m)

	if o == nil {
		return
	}

	if v != o.Error() {
		l.Critical("GOT: '", o, "', EXPECED: '", v, "'")
		t.Fail()
	}
}

func TestGetParent(t *testing.T) {
	n := New("logger.Test.getParent")

	n.Info(n, "Starting")
	m := [][]Logger{
		{"", "."},
		{".Test", "."},
		{".", "."},
		{"Test", "."},
		{"Test.Test", "Test"},
		{"Test.Test.Test", "Test.Test"},
		{"Test.Test.Test.Test", "Test.Test.Test"},
	}

	for i := range m {
		a := m[i]

		k := a[0]
		v := a[1]

		o := getParent(k)
		if o != v {
			n.Error(n, "GOT: '", o, "', EXPECED: '", v, "'", ", KEY: '", k, "'")
			t.Fail()
		}
		n.Debug(n, "GOT: '", o, "', EXPECED: '", v, "'", ", KEY: '", k, "'")
	}
	n.Info(n, "Finished")
}

func TestGetParentOutputSame(t *testing.T) {
	l := New(namet + ".GetParent.Output.Same")

	p := Logger("Test")
	p.SetFormat("{{.Message}}")

	c := Logger("Test.Test")
	l.Info("Parent: '", getParent(c), "'")

	var b bytes.Buffer
	p.SetOutput(&b)

	p.Notice("Test Parent,")
	c.Notice("Test Child")

	o := b.String()
	v := "Test Parent,Test Child"

	l.Debug("GOT: ", o, ", EXPECTED: ", v)
	if o != v {
		l.Critical("GOT: ", o, ", EXPECTED: ", v)
		t.Fail()
	}
}

func TestGetParentOutputDifferent(t *testing.T) {
	l := New(namet + ".GetParent.Output.Different")

	p := Logger("Test")
	p.SetFormat("{{.Message}}")

	c := Logger("Test.Test")
	l.Info("Parent: '", getParent(c), "'")

	var b bytes.Buffer
	c.SetOutput(&b)

	p.Notice("Test Parent,")
	c.Notice("Test Child")

	o := b.String()
	v := "Test Child"

	l.Debug("GOT: ", o, ", EXPECTED: ", v)
	if o != v {
		l.Critical("GOT: ", o, ", EXPECTED: ", v)
		t.Fail()
	}
}

func TestgetParentOutputInheritance(t *testing.T) {
	l := New(namet + ".GetParent.Output.Inheritance")

	p := Logger("Test")
	p.SetFormat("{{.Message}}")

	c := Logger("Test.Test")
	c.SetLevel(Debug)
	l.Info("Parent: '", getParent(c), "'")

	var b bytes.Buffer
	p.SetOutput(&b)

	p.Notice("Test Parent,")
	c.Notice("Test Child")

	o := b.String()
	v := "TestParent,Test Child"

	l.Debug("GOT: ", o, ", EXPECTED: ", v)
	if o != v {
		l.Critical("GOT: ", o, ", EXPECTED: ", v)
		t.Fail()
	}
}

func TestPrintMessage(t *testing.T) {
	l := New(namet + ".PrintMessage")

	p := "\033[0m"
	b := "Test - " + p + p + "Debug" + p + " - "

	m := [][]string{
		{"", b},
		{"Test", b + "Test"},
		{"Test.Test", b + "Test.Test"},
		{"Test.Test.Test", b + "Test.Test.Test"},
	}

	r := list.GetLogger("Test")
	r.Format = "{{.Logger}} - {{.Priority}} - {{.Message}}"

	for _, d := range m {
		l.Info("Checking: ", d)

		k := d[0]
		v := d[1]

		var b bytes.Buffer
		r.Output = &b

		printMessage(r, Debug, k)
		o := b.String()

		l.Debug("GOT: '", o, "', EXPECED: '", v, "'", ", KEY: '", k, "'")
		if o != v {
			l.Critical("GOT: '", o, "', EXPECED: '", v, "'", ", KEY: '", k, "'")
			t.Fail()
		}
	}
}

func TestPrintMessageNoColor(t *testing.T) {
	l := New(namet + ".PrintMessage")

	m := [][]string{
		{"", "Test - Debug - "},
		{"Test", "Test - Debug - Test"},
		{"Test.Test", "Test - Debug - Test.Test"},
		{"Test.Test.Test", "Test - Debug - Test.Test.Test"},
	}

	r := list.GetLogger("Test")
	r.Format = "{{.Logger}} - {{.Priority}} - {{.Message}}"
	r.NoColor = true

	for _, d := range m {
		l.Info("Checking: ", d)

		k := d[0]
		v := d[1]

		var b bytes.Buffer
		r.Output = &b

		printMessage(r, Debug, k)
		o := b.String()

		l.Debug("GOT: '", o, "', EXPECED: '", v, "'", ", KEY: '", k, "'")
		if o != v {
			l.Critical("GOT: '", o, "', EXPECED: '", v, "'", ", KEY: '", k, "'")
			t.Fail()
		}
	}
}

func TestPrintColors(t *testing.T) {
	l := New("logger.Test.PrintColors")
	SetLevel("logger.Test.PrintColors", Disable)

	//TODO: Compare strings instead of printing.

	l.Log(Trace, "Trace")
	l.Trace("Trace")
	l.Debug("Debug")
	l.Info("Info")
	l.Notice("Notice")
	l.Warning("Warning")
	l.Error("Error")
	l.Critical("Critical")
	l.Alert("Alert")
	l.Emergency("Emergency")

	SetNoColor("logger.Test.PrintColors", true)
	l.Log(Trace, "Trace")
	l.Trace("Trace")
	l.Debug("NoColorDebug")
	l.Info("NoColorInfo")
	l.Notice("NoColorNotice")
	l.Warning("NoColorWarning")
	l.Error("NoColorError")
	l.Critical("NoColorCritical")
	l.Alert("NoColorAlert")
	l.Emergency("NoColorEmergency")
}

func TestCheckPriorityOK(t *testing.T) {
	l := New(namet + ".CheckPriority.OK")

	for k := range priorities {
		l.Info("Checking: ", k)

		e := checkPriority(k)
		l.Debug("Return of ", k, ": ", e)
		if e != nil {
			l.Critical(e)
			t.Fail()
		}
	}
}

func TestCheckPriorityFail(t *testing.T) {
	l := New(namet + ".CheckPriority.FAIL")

	k := Disable + 1

	l.Info("Checking: ", k)

	e := checkPriority(k)
	l.Debug("Return of ", k, ": ", e)
	if e == nil {
		l.Critical("Should not have succeeded")
		t.Fail()
		return
	}
}

func TestCheckPriorityFailDoesNotExist(t *testing.T) {
	l := New(namet + ".CheckPriority.FAIL.DoesNotExist")

	k := Disable + 1
	x := "priority does not exist"

	l.Info("Checking: ", k)

	e := checkPriority(k)
	l.Debug("Return of ", k, ": ", e)
	if e != nil {

		if e.Error() != x {
			l.Critical("Wrong error, EXPECTED: ", x, ", GOT: ", e.Error())
			t.Fail()
		}
	}
}

func TestGetPriorityFormat(t *testing.T) {
	l := New(namet + ".GetPriorityFormat")

	m := [][]int{
		{int(Debug), colornone, textnormal},
		{int(Notice), colorgreen, textnormal},
		{int(Info), colorblue, textnormal},
		{int(Warning), coloryellow, textnormal},
		{int(Error), coloryellow, textbold},
		{int(Critical), colorred, textnormal},
		{int(Alert), colorred, textbold},
		{int(Emergency), colorred, textblink},
	}

	for _, d := range m {
		p := Priority(d[0])
		n, e := NamePriority(p)
		if e != nil {
			l.Alert("Can not name priority: ", e)
			t.Fail()
		}

		c := d[1]
		f := d[2]

		a, b := getPriorityFormat(p)

		if c != a {
			l.Critical("Wrong color for ", n, ", EXPECTED: ", c, ", GOT: ", a)
			t.Fail()
		}

		if f != b {
			l.Critical("Wrong format for ", n, ", EXPECTED: ", c, ", GOT: ", b)
			t.Fail()
		}
	}
}

func TestNamePriorityFail(t *testing.T) {
	_, err := NamePriority(999)

	if err.Error() != "priority does not exist" {
		t.Fail()
	}
}

func TestImportLoggers(t *testing.T) {
	loggers := make(map[string]string)
	loggers["."] = "Notice"

	err := ImportLoggers(loggers)
	if err != nil {
		t.Fail()
	}
}

func TestImportLoggersFail(t *testing.T) {
	loggers := make(map[string]string)
	loggers["."] = "FAIL"

	err := ImportLoggers(loggers)
	if err.Error() != "can not parse priority: do not recognize FAIL" {
		t.Fail()
	}
}

func TestImportLoggersNil(t *testing.T) {
	err := ImportLoggers(nil)

	if err.Error() != "the loglevel map is nil" {
		t.Fail()
	}
}

func BenchmarkLogRootEmergency(b *testing.B) {
	for i := 0; i < b.N; i++ {
		logMessage(".", Emergency, "Test")
	}
}

func BenchmarkLogRootEmergencyNoColor(b *testing.B) {
	SetNoColor(".", true)

	for i := 0; i < b.N; i++ {
		logMessage(".", Emergency, "Test")
	}
}

func BenchmarkGetLogger(b *testing.B) {
	for i := 0; i < b.N; i++ {
		list.GetLogger("BenchmarkGetLogger")
	}
}

func BenchmarkGetLoggerNoSaving(b *testing.B) {
	SaveLoggerLevels = false
	for i := 0; i < b.N; i++ {
		list.GetLogger("BenchmarkGetLoggerNoSaving")
	}
	SaveLoggerLevels = true
}

func BenchmarkLogRoot(b *testing.B) {
	for i := 0; i < b.N; i++ {
		logMessage(".", Debug, "Test")
	}
}

func BenchmarkLogChild(b *testing.B) {
	for i := 0; i < b.N; i++ {
		logMessage("BenchLogChild", Debug, "Test")
	}
}

func BenchmarkLogChildChild(b *testing.B) {
	for i := 0; i < b.N; i++ {
		logMessage("BenchLogChildChild.Test", Debug, "Test")
	}
}

func BenchmarkLogChildChildChild(b *testing.B) {
	for i := 0; i < b.N; i++ {
		logMessage("BenchLogChildChildChild.Test.Test", Debug, "Test")
	}
}

func BenchmarkLogChildAllocated(b *testing.B) {
	SetLevel("BenchLogChildAllocated", Emergency)
	for i := 0; i < b.N; i++ {
		logMessage("BenchLogChildAllocated", Debug, "Test")
	}
}

func BenchmarkLogChildChildAllocated(b *testing.B) {
	SetLevel("BenchLogChildChildAllocated.Test", Emergency)
	for i := 0; i < b.N; i++ {
		logMessage("BenchLogChildChildAllocated.Test", Debug, "Test")
	}
}

func BenchmarkGetParentRoot(b *testing.B) {
	for i := 0; i < b.N; i++ {
		getParent(".")
	}
}

func BenchmarkGetParentChild(b *testing.B) {
	for i := 0; i < b.N; i++ {
		getParent("BenchgetParentChild")
	}
}

func BenchmarkGetParentChildChild(b *testing.B) {
	for i := 0; i < b.N; i++ {
		getParent("BenchgetParentChildChild.Test")
	}
}

func BenchmarkGetParentChildChildChild(b *testing.B) {
	for i := 0; i < b.N; i++ {
		getParent("BenchgetParentChildChild.Test.Test")
	}
}

func BenchmarkGetParentChildChildChildChild(b *testing.B) {
	for i := 0; i < b.N; i++ {
		getParent("BenchgetParentChildChildChild.Test.Test")
	}
}

func BenchmarkGetParentChildChildChildChildChild(b *testing.B) {
	for i := 0; i < b.N; i++ {
		getParent("BenchgetParentChildChildChildChild.Test.Test.Test")
	}
}

func BenchmarkPrintMessage(b *testing.B) {
	var a bytes.Buffer
	l := list.GetLogger("BenchprintMessage")
	l.Output = &a

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		printMessage(l, Debug, "Message")
	}
}

func BenchmarkFormatMessage(b *testing.B) {
	l := list.GetLogger("BenchformatMessage")

	m := new(message)
	m.Time = "Mo 30 Sep 2013 20:29:19 CEST"
	m.Logger = l.Logger
	m.Priority = "Debug"
	m.Message = "Test"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		formatMessage(m, l.Format)
	}
}
