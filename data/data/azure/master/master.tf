locals{
  // The name of the masters' ipconfiguration is hardcoded to "pipconfig". It needs to match cluster-api
  // https://github.com/openshift/cluster-api-provider-azure/blob/master/pkg/cloud/azure/services/networkinterfaces/networkinterfaces.go#L131
  ip_configuration_name = "pipConfig"
}

resource "azurerm_network_interface" "master" {
  count               = "${var.instance_count}"
  name                = "${var.cluster_id}-master-nic-${count.index}"
  location            = "${var.region}"
  resource_group_name = "${var.resource_group_name}"

  ip_configuration {
    subnet_id                     = "${var.subnet_id}"
    name                          = "${local.ip_configuration_name}"
    private_ip_address_allocation = "Static"
    private_ip_address            = "${cidrhost(var.master_subnet_cidr, 5 + count.index)}" # azure reserves first 3 ip, 4th is for bootstrap VM, so we start at 5
  }
}

resource "azurerm_network_interface_nat_rule_association" "master_ssh" {
  count                 = "${var.instance_count}"
  network_interface_id  = "${element(azurerm_network_interface.master.*.id, count.index)}"
  ip_configuration_name = "${local.ip_configuration_name}"
  nat_rule_id           = "${element(var.ssh_nat_rule_ids, count.index)}"
}

resource "azurerm_network_interface_backend_address_pool_association" "master" {
  count                   = "${var.instance_count}"
  network_interface_id    = "${element(azurerm_network_interface.master.*.id, count.index)}"
  backend_address_pool_id = "${var.elb_backend_pool_id}"
  ip_configuration_name   = "${local.ip_configuration_name}"                      #must be the same as nic's ip configuration name.
}

resource "azurerm_network_interface_backend_address_pool_association" "master_internal" {
  count                   = "${var.instance_count}"
  network_interface_id    = "${element(azurerm_network_interface.master.*.id, count.index)}"
  backend_address_pool_id = "${var.ilb_backend_pool_id}"
  ip_configuration_name   = "${local.ip_configuration_name}"                       #must be the same as nic's ip configuration name.
}

#TODO : make FD/UD configurable
resource "azurerm_availability_set" "master" {
  name                         = "controlplane-as"
  location                     = "${var.region}"
  resource_group_name          = "${var.resource_group_name}"
  managed                      = true
  platform_update_domain_count = 5
  platform_fault_domain_count  = 3                            # the available fault domain number depends on the region, so this needs to be configurable or dynamic
}

data "azurerm_subscription" "current" {}

resource "azurerm_virtual_machine" "master" {
  count                 = "${var.instance_count}"
  name                  = "${var.cluster_id}-master-${count.index}"
  location              = "${var.region}"
  resource_group_name   = "${var.resource_group_name}"
  network_interface_ids = ["${element(azurerm_network_interface.master.*.id, count.index)}"]
  vm_size               = "${var.vm_size}"
  availability_set_id   = "${azurerm_availability_set.master.id}"

  delete_os_disk_on_termination = true

  identity {
    type         = "UserAssigned"
    identity_ids = ["${var.identity}"]
  }

  storage_os_disk {
    name              = "${var.cluster_id}-master-${count.index}_OSDisk" # os disk name needs to match cluster-api convention
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Premium_LRS"
    disk_size_gb      = "${var.os_volume_size}"
  }

  storage_image_reference {
    id = "${data.azurerm_subscription.current.id}${var.vm_image}"
  }

  //we don't provide a ssh key, because it is set with ignition. 
  //it is required to provide at least 1 auth method to deploy a linux vm
  os_profile {
    computer_name  = "${var.cluster_id}-master-${count.index}"
    admin_username = "core"
    admin_password = "P@ssword1234!"
    custom_data    = "${var.ignition}"
  }

  os_profile_linux_config {
    disable_password_authentication = false
  }

  boot_diagnostics {
    enabled     = true
    storage_uri = "${var.boot_diag_blob_endpoint}"
  }
}
