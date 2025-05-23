Feature: LXC Container Management
  As a Proxmox administrator
  I want to manage LXC containers using the CLI
  So that I can automate container operations

  Scenario: Create an LXC container from a YAML spec
    Given the CLI is configured and authenticated
    And a valid LXC YAML spec file with the following content:
      """
      ostemplate: "local:vztmpl/ubuntu-20.04-standard_20.04-1_amd64.tar.gz"
      hostname: "test-container"
      memory: 1024
      swap: 512
      cores: 2
      rootfs: "local-lvm:8"
      net0: "name=eth0,bridge=vmbr0,ip=dhcp"
      password: "testpass123"
      start: 1
      unprivileged: 1
      """
    When I run the command "./proxmox-cli lxc create -n pve -i 200 -s test-lxc.yaml"
    Then the LXC container should be created successfully
    And the container should have 2 cores
    And the container should have 1024 MB memory

  Scenario: List all LXC containers
    Given the CLI is configured and authenticated
    And there are LXC containers on the cluster
    When I run the command "./proxmox-cli lxc get"
    Then I should see a list of all LXC containers

  Scenario: Start a stopped LXC container
    Given the CLI is configured and authenticated
    And an LXC container with ID 200 is stopped
    When I run the command "./proxmox-cli lxc start -n pve -i 200"
    Then the container should be started successfully

  Scenario: Stop a running LXC container
    Given the CLI is configured and authenticated
    And an LXC container with ID 200 is running
    When I run the command "./proxmox-cli lxc stop -n pve -i 200"
    Then the container should be stopped successfully

  Scenario: Delete an LXC container
    Given the CLI is configured and authenticated
    And an LXC container with ID 200 exists
    When I run the command "./proxmox-cli lxc delete -n pve -i 200"
    Then the container should be deleted successfully

  Scenario: Clone an LXC container
    Given the CLI is configured and authenticated
    And an LXC container with ID 200 exists
    When I run the command "./proxmox-cli lxc clone -n pve -s 200 -t 201 --name clone-test"
    Then a new container with ID 201 should be created
    And the new container should be named "clone-test"

  Scenario: Create a snapshot of an LXC container
    Given the CLI is configured and authenticated
    And an LXC container with ID 200 exists
    When I run the command "./proxmox-cli lxc snapshot create -n pve -i 200 --name test-snapshot"
    Then a snapshot named "test-snapshot" should be created

  Scenario: List snapshots of an LXC container
    Given the CLI is configured and authenticated
    And an LXC container with ID 200 has snapshots
    When I run the command "./proxmox-cli lxc snapshot list -n pve -i 200"
    Then I should see a list of all snapshots