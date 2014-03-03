picolog
=======

Tiny levelled logging framework for go.

Documentation
=============

http://godoc.org/github.com/anchor/picolog

Example
=======


    package picolog_test
    
    import (
    	"github.com/anchor/picolog"
    	"os"
    	"fmt"
    )
    
    func Example() {
    	level, err := picolog.ParseLogLevel("info")
    	if err != nil {
    		fmt.Printf("This can't happen: %v", err)
    	}
    	// Log messages will be prefixed by "[example]" and a timestamp.
    	logger := picolog.NewLogger(level, "example", os.Stdout)
    	logger.Infof("Printing a log message!")
    	logger.Debugf("Not printing this message at INFO log level.")
    }
    
    func ExampleSubLogger() {
    	level, err := picolog.ParseLogLevel("info")
    	if err != nil {
    		fmt.Printf("This can't happen: %v", err)
    	}
    	logger := picolog.NewLogger(level, "things", os.Stdout)
    	logger.Infof("Things!")
    	subLogger := logger.NewSubLogger("related-things")
    	subLogger.Infof("This will be prefixed by [things][related-things] <timestamp>.")
    }
