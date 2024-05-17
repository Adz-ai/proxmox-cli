package main

import (
	"fmt"
	"github.com/cucumber/godog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

var (
	specFilePath string
	nodeName     string
	vmId         int
)

func aValidYAMLSpecFileWithTheFollowingContent(specContent *godog.DocString) error {
	dir, err := os.MkdirTemp("", "proxmox-cli-test")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	specFilePath = filepath.Join(dir, "vm_spec.yaml")
	return os.WriteFile(specFilePath, []byte(specContent.Content), 0644)
}

func iRunTheCommand(command string) error {
	var err error
	command = strings.Replace(command, "test-vm.yaml", specFilePath, -1)
	parts := strings.Split(command, " ")
	nodeName = parts[4]
	vmId, err = strconv.Atoi(parts[8])
	if err != nil {
		fmt.Println("Error during conversion")
		return err
	}
	cmd := exec.Command(parts[0], parts[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run vm create command: %w, output: %s", err, output)
	}

	return nil
}

func theVirtualMachineShouldBeCreatedSuccessfully() error {
	cmd := exec.Command("./proxmox-cli", "vm", "describe", "-n", "pve", "-i", strconv.Itoa(vmId))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run vm describe command: %w, output: %s", err, output)
	}

	if !strings.Contains(string(output), fmt.Sprintf("Cores: %d", 2)) {
		return fmt.Errorf("expected number of cores not found in the output")
	}

	if !strings.Contains(string(output), fmt.Sprintf("Sockets: %d", 1)) {
		return fmt.Errorf("expected number of sockets not found in the output")
	}

	if !strings.Contains(string(output), fmt.Sprintf("Name: %s", "test-vm")) {
		return fmt.Errorf("expected number of sockets not found in the output")
	}

	// Add other assertions as needed

	return nil

}

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		TestSuiteInitializer: InitializeTestSuite,
		ScenarioInitializer:  InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
			TestingT: t, // Testing instance that will run subtests.
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.AfterSuite(cleanup)
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Step(`^a valid YAML spec file with the following content:$`, aValidYAMLSpecFileWithTheFollowingContent)
	ctx.Step(`^I run the command "([^"]*)"$`, iRunTheCommand)
	ctx.Step(`^the virtual machine should be created successfully$`, theVirtualMachineShouldBeCreatedSuccessfully)
}

func cleanup() {
	err := exec.Command("./proxmox-cli", "vm", "delete", "-n", nodeName, "-i", fmt.Sprintf("%d", vmId)).Run()
	if err != nil {
		fmt.Errorf("failed to run vm delete command: %w", err)
	}
	fmt.Sprintf("Test Cleanup Complete, VM %d deleted", vmId)
}
