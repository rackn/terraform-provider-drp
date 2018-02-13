package drp

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"

	"github.com/digitalrebar/provision/api"
	"github.com/digitalrebar/provision/server"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	flags "github.com/jessevdk/go-flags"
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

func generateArgs(args []string) *server.ProgOpts {
	var c_opts server.ProgOpts

	parser := flags.NewParser(&c_opts, flags.Default)
	if _, err := parser.ParseArgs(args); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}

	return &c_opts
}

func TestMain(m *testing.M) {
	var err error
	tmpDir, err = ioutil.TempDir("", "tf-")
	if err != nil {
		log.Printf("Creating temp dir for file root failed: %v", err)
		os.Exit(1)
	}

	testArgs := []string{
		"--base-root", tmpDir,
		"--tls-key", tmpDir + "/server.key",
		"--tls-cert", tmpDir + "/server.crt",
		"--api-port", "10031",
		"--static-port", "10032",
		"--tftp-port", "10033",
		"--dhcp-port", "10034",
		"--binl-port", "10035",
		"--fake-pinger",
		"--drp-id", "Fred",
		"--backend", "memory:///",
		"--debug-frontend", "0",
		"--debug-renderer", "0",
		"--debug-plugins", "0",
		"--local-content", "",
		"--default-content", "",
	}

	err = os.MkdirAll(tmpDir+"/plugins", 0755)
	if err != nil {
		log.Printf("Error creating required directory %s: %v", tmpDir+"/plugins", err)
		os.Exit(1)
	}

	c_opts := generateArgs(testArgs)
	go server.Server(c_opts)

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
		os.RemoveAll(tmpDir)
		os.Exit(1)
	}
	if err != nil {
		log.Printf("Server failed to start in time allowed")
		os.RemoveAll(tmpDir)
		os.Exit(1)
	}
	ret := m.Run()
	os.RemoveAll(tmpDir)
	os.Exit(ret)
}
