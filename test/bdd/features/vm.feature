Feature: Virtual Machine Management
  As a Proxmox administrator
  I want to manage virtual machines using the CLI
  So that I can automate VM operations

  @skip
  Scenario: Successfully create a virtual machine
    Given the CLI is configured and authenticated
    And a valid YAML spec file with the following content:
      """
      name: "test-vm"
      memory: 2048
      cores: 2
      sockets: 1
      scsi0: "local-lvm:32"
      ostype: "l26"
      net0: "virtio,bridge=vmbr0"
      """
    When I run the command "./proxmox-cli vm create -n pve -s test-vm.yaml -i 104"
    Then the virtual machine should be created successfully

