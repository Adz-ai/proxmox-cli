Feature: Node Management
  As a Proxmox administrator
  I want to manage cluster nodes using the CLI
  So that I can monitor and control node operations

  Scenario: List all nodes in the cluster
    Given the CLI is configured and authenticated
    And a Proxmox cluster with multiple nodes
    When I run the command "./proxmox-cli nodes get"
    Then I should see a list of all nodes with their status

  Scenario: Describe a specific node
    Given the CLI is configured and authenticated
    And a node named "pve" exists in the cluster
    When I run the command "./proxmox-cli nodes describe -n pve"
    Then I should see detailed information about the node
    And I should see CPU usage information
    And I should see memory usage information
    And I should see disk usage information

  @skip
  Scenario: List storage on a node
    Given the CLI is configured and authenticated
    And a node named "pve" exists with configured storage
    When I run the command "./proxmox-cli nodes storage -n pve"
    Then I should see a list of all storage on the node
    And I should see storage types and usage

  @skip
  Scenario: List tasks on a node
    Given the CLI is configured and authenticated
    And a node named "pve" has running tasks
    When I run the command "./proxmox-cli nodes tasks -n pve"
    Then I should see a list of tasks on the node
    And I should see task status and timestamps

  @skip
  Scenario: List only running tasks
    Given the CLI is configured and authenticated
    And a node named "pve" has both running and completed tasks
    When I run the command "./proxmox-cli nodes tasks -n pve -r"
    Then I should see only running tasks

  @skip
  Scenario: List services on a node
    Given the CLI is configured and authenticated
    And a node named "pve" exists in the cluster
    When I run the command "./proxmox-cli nodes services list -n pve"
    Then I should see a list of all services on the node
    And I should see service states and descriptions

  @skip
  Scenario: Restart a service on a node
    Given the CLI is configured and authenticated
    And a service named "pveproxy" exists on node "pve"
    When I run the command "./proxmox-cli nodes services restart -n pve -s pveproxy"
    Then the service should be restarted successfully