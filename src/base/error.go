package base

import "os"
import "fmt"

/*
#include <stdio.h>


static void StderrRedirect() {
	freopen( "internal.stderr.log", "a+", stderr );
}

static void StdoutRedirect() {
	freopen( "internal.stdout.log", "a+", stdout );
}
*/
import "C"

func RedirectErrorToFile() {

	C.StderrRedirect()
	C.StdoutRedirect()

	tips := fmt.Sprintf("############### %s @ %s #################\n", AppName(), ServerTimeStr())

	fmt.Fprintf(os.Stdout, "%s", tips)
	fmt.Fprintf(os.Stderr, "%s", tips)
}
