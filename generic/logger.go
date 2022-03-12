package generic

import (
	"fmt"
	"log"
	"runtime"
)

func Error(err error, args ...string) {
	_, file, line, _ := runtime.Caller(1)
	log.Println(fmt.Sprintf(`%s - %s:%d [%s]`, err.Error(), file, line, args))
}

func Fatal(err error) {
	_, file, line, _ := runtime.Caller(1)
	log.Fatalln(fmt.Sprintf(`%s - %s:%d`, err.Error(), file, line))
}
