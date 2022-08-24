package service

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"testing"

	"github.com/cucumber/godog"
)

func TestMain(m *testing.M) {

	go http.ListenAndServe("localhost:6060", nil)
	fmt.Printf("starting godog...\n")

	opts := godog.Options{
		Format: "pretty",
		Paths:  []string{"features"},
		Tags:   "replication",
	}

	status := godog.TestSuite{
		Name:                "godog",
		ScenarioInitializer: FeatureContext,
		Options:             &opts,
	}.Run()

	fmt.Printf("godog finished\n")

	if st := m.Run(); st > status {
		fmt.Printf("godog.TestSuite status %d\n", status)
		fmt.Printf("m.Run status %d\n", st)
		status = st
	}

	fmt.Printf("status %d\n", status)

	os.Exit(status)
}
