Feature: Authentication and Setup
  As a Proxmox administrator
  I want to configure and authenticate with Proxmox servers
  So that I can manage my infrastructure securely

  Scenario: Check status when not configured
    Given the CLI is not configured
    When I run the command "./proxmox-cli status"
    Then I should see "Status: Not configured"
    And I should see "Run 'proxmox-cli init' to configure"

  Scenario: Initialize configuration
    Given the CLI is not configured
    When I run the command "./proxmox-cli init" with input:
      """
      https://192.168.1.100:8006
      """
    Then I should see "Welcome to Proxmox CLI"
    And I should see "Configuration saved"
    And the config file should contain server URL "https://192.168.1.100:8006"

  Scenario: Initialize with URL without protocol
    Given the CLI is not configured
    When I run the command "./proxmox-cli init" with input:
      """
      192.168.1.100:8006
      """
    Then the config file should contain server URL "https://192.168.1.100:8006"

  Scenario: Reconfigure with force flag
    Given the CLI is configured with server "https://old-server:8006"
    When I run the command "./proxmox-cli init --force" with input:
      """
      https://new-server:8006
      """
    Then I should see "Configuration saved"
    And the config file should contain server URL "https://new-server:8006"

  @skip
  Scenario: Login with unconfigured CLI
    Given the CLI is not configured
    When I run the command "./proxmox-cli auth login -u root@pam" with input:
      """
      https://192.168.1.100:8006
      testpass123
      """
    Then I should see "Proxmox server URL not configured"
    And I should see "Server URL saved to configuration"
    And I should see "Authentication successful"

  @skip
  Scenario: Login with configured CLI
    Given the CLI is configured with server "https://192.168.1.100:8006"
    When I run the command "./proxmox-cli auth login -u root@pam" with password "testpass123"
    Then I should see "Authenticating with Proxmox server"
    And I should see "Authentication successful"
    And I should see "You can now use commands like"

  Scenario: Check status when authenticated
    Given the CLI is configured and authenticated
    When I run the command "./proxmox-cli status"
    Then I should see "Server URL: https://192.168.1.100:8006"
    And I should see "Authentication: Logged in"

  Scenario: Check status verbose
    Given the CLI is configured and authenticated
    When I run the command "./proxmox-cli status --verbose"
    Then I should see "Testing connection"
    And I should see "Connection successful"
    And I should see "Proxmox VE Version"

  Scenario: Logout when authenticated
    Given the CLI is configured and authenticated
    When I run the command "./proxmox-cli auth logout"
    Then I should see "Logged out successfully"
    And I should see "Your authentication has been cleared"

  Scenario: Logout when not authenticated
    Given the CLI is configured but not authenticated
    When I run the command "./proxmox-cli auth logout"
    Then I should see "Not currently logged in"

  Scenario: Access protected command without auth
    Given the CLI is configured but not authenticated
    When I run the command "./proxmox-cli nodes get"
    Then I should see "❌ Not authenticated"
    And I should see "Please run 'proxmox-cli auth login -u <username>' to log in"

  Scenario: Access protected command without config
    Given the CLI is not configured
    When I run the command "./proxmox-cli vm get"
    Then I should see "❌ Not configured"
    And I should see "Please run 'proxmox-cli auth login -u <username>' to set up"