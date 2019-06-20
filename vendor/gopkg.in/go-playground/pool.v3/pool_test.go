package pool

import (
	"os"
	"testing"
)

// NOTES:
// - Run "go test" to run tests
// - Run "gocov test | gocov report" to report on test converage by file
// - Run "gocov test | gocov annotate -" to report on all code and functions, those ,marked with "MISS" were never called
//
// or
//
// -- may be a good idea to change to output path to somewherelike /tmp
// go test -coverprofile cover.out && go tool cover -html=cover.out -o cover.html
//

// global pool for testing long running pool
var limitedGpool Pool

var unlimitedGpool Pool

func TestMain(m *testing.M) {

	// setup
	limitedGpool = NewLimited(4)
	defer limitedGpool.Close()

	unlimitedGpool = New()
	defer unlimitedGpool.Close()

	os.Exit(m.Run())

	// teardown
}
