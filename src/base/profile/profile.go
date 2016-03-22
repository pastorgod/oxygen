package profile

/*
import(
	"logger"
	"time"
	"fmt"
)

var profile_logger *logger.Logger
var profile_recoder map[uint32]profileRecoder

type profileRecoder struct {
	counter int64	// call counter.
	time int64		// elapsed time.
}

func InitProfile( prefix, name string ) {
	if nil != profile_logger {
		panic( "init profile error" )
	}
	profile_logger = logger.NewLogger( ".", prefix + "_profile.log", name, "DEBUG" )
	profile_recoder = make( map[uint32]profileRecoder, 1024 )

	go func() {
		c := time.Tick( time.Second )
		for _ := range( c ) {
		}
	}()
}

func record_profile( code uint32, elapsed int64 ) {

}

func Profile( code uint32 ) func() {

	start := time.Now().UnixNano()

	return func() {
		record_profile( code, time.Now().UnixNano() - start )
	}
}
*/
