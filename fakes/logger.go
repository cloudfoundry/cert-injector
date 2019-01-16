package fakes

type Logger struct {
	PrintlnCall struct {
		CallCount int
		Receives  []PrintlnCallReceive
	}
}
type PrintlnCallReceive struct {
	Args []interface{}
}

func (l *Logger) Println(v ...interface{}) {
	l.PrintlnCall.CallCount++
	l.PrintlnCall.Receives = append(l.PrintlnCall.Receives, PrintlnCallReceive{
		Args: v,
	})
}
