package gobfd

type Logger interface {
	// Printf must have the same semantics as log.Printf.
	Printf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	//Info(format string)
	//Infof(format string, args ...interface{})
	//Debugf(format string, args ...interface{})
	//Error(format string)
	//Errorf(format string, args ...interface{})
}
