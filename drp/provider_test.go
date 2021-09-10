package drp

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"gitlab.com/rackn/provision/v4/api"
	"gitlab.com/rackn/provision/v4/test"
)

var testAccDrpProviders map[string]terraform.ResourceProvider
var testAccDrpProvider *schema.Provider

func init() {
	testAccDrpProvider = Provider().(*schema.Provider)
	testAccDrpProviders = map[string]terraform.ResourceProvider{
		"drp": testAccDrpProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ terraform.ResourceProvider = Provider()
}

func testAccDrpPreCheck(t *testing.T) {
	os.Setenv("RS_KEY", "rocketskates:r0cketsk8ts")
	os.Setenv("RS_ENDPOINT", "https://127.0.0.1:10031")
}

var tmpDir string
var session *api.Client

func TestMain(m *testing.M) {
	var err error
	tmpDir, err = ioutil.TempDir("", "tf-")
	if err != nil {
		log.Printf("Creating temp dir for file root failed: %v", err)
		os.Exit(1)
	}

	if err := test.StartServer(tmpDir, 10031); err != nil {
		log.Printf("Error starting dr-provision: %v", err)
		os.RemoveAll(tmpDir)
		os.Exit(1)
	}

	count := 0
	for count < 30 {
		session, err = api.UserSession("https://127.0.0.1:10031", "rocketskates", "r0cketsk8ts")
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
		count++
	}
	if session == nil {
		log.Printf("Failed to create UserSession: %v", err)
		test.StopServer()
		os.RemoveAll(tmpDir)
		os.Exit(1)
	}
	if err != nil {
		log.Printf("Server failed to start in time allowed")
		test.StopServer()
		os.RemoveAll(tmpDir)
		os.Exit(1)
	}
	ret := m.Run()
	test.StopServer()
	os.RemoveAll(tmpDir)
	os.Exit(ret)
}
