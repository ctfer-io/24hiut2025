package main

import (
	"fmt"
	"os"
	"slices"

	"github.com/ctfer-io/chall-manager/sdk"
	"github.com/ctfer-io/challenges/active-directory/mort-a-popacola/infra/scenario/utils"
	"github.com/go-viper/mapstructure/v2"
	"github.com/muhlba91/pulumi-proxmoxve/sdk/v6/go/proxmoxve"
	"github.com/muhlba91/pulumi-proxmoxve/sdk/v6/go/proxmoxve/storage"
	"github.com/muhlba91/pulumi-proxmoxve/sdk/v6/go/proxmoxve/vm"
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-random/sdk/v4/go/random"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type VM struct {
	Name       string
	TemplateID int
	Type       string
	Node       string
	Datastore  string
	CloudImgID string
	MacAddress string
}

const (
	default_front_bridge = "admin" // TODO change with front bridge
	default_back_bridge  = "admin" // TODO change with back bridge

	default_pve_endpoint = "https://10.17.10.101:8006/api2/json"
)

type Additional struct {
	ProxmoxAPIToken    string `json:"proxmox_api_token"`
	ProxmoxEndpoint    string `json:"proxmox_endpoint"`
	ProxmoxSSHPassword string `json:"proxmox_ssh_password"`

	FrontBridge string `json:"front_bridge"`
	FrontVlan   int    `json:"front_vlan"`
	BackBridge  string `json:"back_bridge"`
}

var (
	additional Additional

	vms = []VM{
		{
			Name:       "adds",
			Type:       "Windows",
			TemplateID: 4100,
			Node:       "pve02", // TODO change to the real one
			Datastore:  "raid0",
		},
		{
			Name:       "adsrv",
			Type:       "Windows",
			TemplateID: 4101,
			Node:       "pve02", // TODO change to the real one
			Datastore:  "raid0",
		},
		{
			Name:       "client",
			Type:       "Windows",
			TemplateID: 4102,
			Node:       "pve03", // TODO change to the real one
			Datastore:  "raid0",
		},
		{
			Name:       "hacker",
			Type:       "Hacker",
			CloudImgID: "local:iso/noble-server-cloudimg-amd64.img", // download it before
			Node:       "pve",                                       // TODO change to the real one
			Datastore:  "raid5",
		},
	}
)

func main() {
	sdk.Run(func(req *sdk.Request, resp *sdk.Response, opts ...pulumi.ResourceOption) error {

		// 1. Load config
		err := mapstructure.Decode(req.Config.Additional, &additional)
		if err != nil {
			return err
		}

		// Default configuration
		err = loadDefaults()
		if err != nil {
			return err
		}

		provider, err := proxmoxve.NewProvider(req.Ctx, "pve", &proxmoxve.ProviderArgs{
			ApiToken: pulumi.String(additional.ProxmoxAPIToken),
			Endpoint: pulumi.String(additional.ProxmoxEndpoint),
			Insecure: pulumi.Bool(true), // if you do not trust the x509 PVE Cert
			Ssh: proxmoxve.ProviderSshArgs{ // only for the nodes with Hacker for File upload
				Username: pulumi.String("root"),
				Password: pulumi.String(additional.ProxmoxSSHPassword),

				Nodes: proxmoxve.ProviderSshNodeArray{
					proxmoxve.ProviderSshNodeArgs{
						Address: pulumi.String("10.17.10.101"), // only on the one with the hacker vm
						Name:    pulumi.String("pve"),
						Port:    pulumi.Int(22),
					},
				},
			},
		})
		if err != nil {
			err = errors.Wrap(err, "creating PVE provider")
			return err
		}

		opts = append(opts, pulumi.Provider(provider))

		// vlan_id to use for the lab instance
		vlan_id, err := utils.GetAvailableId(2000, 2999)
		if err != nil {
			return err
		}

		password, err := random.NewRandomPassword(req.Ctx, "hacker-access", &random.RandomPasswordArgs{
			Length:  pulumi.Int(12), // change me
			Special: pulumi.BoolPtr(false),
		}, opts...)
		if err != nil {
			return err
		}

		// debug: remove before PR
		password.Result.ApplyT(func(pass string) error {
			fmt.Printf("password: %s\n", pass)
			return nil
		})

		var bastion vm.VirtualMachineOutput

		for _, item := range vms {

			vm_id, err := utils.GetAvailableId(50000, 59999)
			if err != nil {
				return err
			}
			// 2. Create resources
			networks := vm.VirtualMachineNetworkDeviceArray{}
			disks := vm.VirtualMachineDiskArray{}
			bios := pulumi.String("seabios")
			qemuEnabled := false

			var efidisk vm.VirtualMachineEfiDiskPtrInput = nil
			var clone vm.VirtualMachineClonePtrInput = nil
			var initialization vm.VirtualMachineInitializationPtrInput = nil

			if item.Type == "Windows" {
				networks = vm.VirtualMachineNetworkDeviceArray{
					vm.VirtualMachineNetworkDeviceArgs{
						Bridge:     pulumi.String(additional.BackBridge),
						Enabled:    pulumi.Bool(true),
						Model:      pulumi.String("virtio"), //e1000 mandatory for Windows without virtio supports
						VlanId:     pulumi.Int(vlan_id),
						MacAddress: pulumi.String(item.MacAddress),
					},
				}
				efidisk = vm.VirtualMachineEfiDiskArgs{
					DatastoreId:     pulumi.String(item.Datastore),
					FileFormat:      pulumi.String("qcow2"),
					PreEnrolledKeys: pulumi.Bool(true),
					Type:            pulumi.String("4m"),
				}

				bios = pulumi.String("ovmf")
				clone = vm.VirtualMachineCloneArgs{
					//DatastoreId: pulumi.String(item.Datastore),
					VmId: pulumi.Int(item.TemplateID),
					Full: pulumi.Bool(false),
				}
			}

			if item.Type == "Hacker" {

				qemuEnabled = true

				networks = vm.VirtualMachineNetworkDeviceArray{
					vm.VirtualMachineNetworkDeviceArgs{
						Bridge:  pulumi.String(additional.FrontBridge),
						Enabled: pulumi.Bool(true),
						VlanId:  pulumi.Int(additional.FrontVlan), // TODO make a var
					},
					vm.VirtualMachineNetworkDeviceArgs{
						Bridge:  pulumi.String(additional.BackBridge),
						Enabled: pulumi.Bool(true),
						VlanId:  pulumi.Int(vlan_id),
					},
				}

				disks = vm.VirtualMachineDiskArray{
					vm.VirtualMachineDiskArgs{
						DatastoreId: pulumi.String(item.Datastore),
						FileId:      pulumi.String(item.CloudImgID),
						Size:        pulumi.Int(20),
						Interface:   pulumi.String("scsi0"),
					},
				}

				netconf, err := storage.NewFile(req.Ctx, "hacker-netconf", &storage.FileArgs{
					ContentType: pulumi.String("snippets"),
					DatastoreId: pulumi.String(item.Datastore),
					NodeName:    pulumi.String(item.Node),
					SourceRaw: storage.FileSourceRawArgs{
						FileName: pulumi.Sprintf("%s-%s-netconf.yaml", req.Config.Identity, item.Name),
						Data: pulumi.String(`version: 1
config:
  - type: physical
    name: ens18
    subnets:
    - type: dhcp
  - type: physical
    name: ens19
    subnets:
    - type: static
      address: 25.0.1.254
      netmask: 255.255.255.0
`),
					},
				}, opts...)
				if err != nil {
					err = errors.Wrap(err, "creating VM")
					return err
				}

				user_data, err := storage.NewFile(req.Ctx, "proxmox-hack-userdata", &storage.FileArgs{
					ContentType: pulumi.String("snippets"),
					DatastoreId: pulumi.String(item.Datastore),
					NodeName:    pulumi.String(item.Node),
					SourceRaw: storage.FileSourceRawArgs{
						FileName: pulumi.Sprintf("%s-%s-userdata.yaml", req.Config.Identity, item.Name),
						Data: password.BcryptHash.ApplyT(func(pass string) string {
							return fmt.Sprintf(`#cloud-config

hostname: %s
timezone: Europe/Paris
users:
  - name: user
    lock_passwd: false
    passwd: %s
    sudo: ALL=(ALL) NOPASSWD:ALL
    shell: /bin/bash

package_update: true
packages:
  - qemu-guest-agent
runcmd:
  - cd /home/user; wget -O openvpn.sh https://get.vpnsetup.net/ovpn
  - cd /home/user; sudo -u user sudo bash openvpn.sh --auto --listenaddr $(hostname -I | awk '{print $1}')  --serveraddr $(hostname -I | awk '{print $1}')
  - systemctl enable qemu-guest-agent
  - systemctl start qemu-guest-agent

ssh_pwauth: true

`, req.Config.Identity, pass)
						}).(pulumi.StringOutput),
					},
				}, opts...)
				if err != nil {
					err = errors.Wrap(err, "creating VM")
					return err
				}

				meta_data, err := storage.NewFile(req.Ctx, "proxmox-hack-metadata", &storage.FileArgs{
					ContentType: pulumi.String("snippets"),
					DatastoreId: pulumi.String(item.Datastore),
					NodeName:    pulumi.String(item.Node),
					SourceRaw: storage.FileSourceRawArgs{
						FileName: pulumi.Sprintf("%s-%s-metadata.yaml", req.Config.Identity, item.Name),
						Data: pulumi.Sprintf(`#cloud-config

instance-id: %s
local-hostname: %s
`, req.Config.Identity, req.Config.Identity),
					},
				}, opts...)
				if err != nil {
					err = errors.Wrap(err, "creating VM")
					return err
				}

				initialization = vm.VirtualMachineInitializationArgs{
					UserDataFileId:    user_data.ID(),
					NetworkDataFileId: netconf.ID(),
					MetaDataFileId:    meta_data.ID(),
					DatastoreId:       pulumi.String(item.Datastore),
				}

			}

			res, err := vm.NewVirtualMachine(req.Ctx, fmt.Sprintf("proxmox-vm-%s", item.Name), &vm.VirtualMachineArgs{
				VmId: pulumi.Int(vm_id),
				// PoolId:      pulumi.String("chall-manager-vm"), // TODO
				Description: pulumi.Sprintf("%s VM for %s, password for user account on Hacker VM is %s", item.Name, req.Config.Identity, password.Result),
				NodeName:    pulumi.String(item.Node), // adapt as your need
				Tags: pulumi.ToStringArray([]string{
					"instance",
					"lab-ad1",
				}),
				Agent: vm.VirtualMachineAgentArgs{
					Enabled: pulumi.Bool(qemuEnabled),
				},
				Name:  pulumi.Sprintf("%s-%s", req.Config.Identity, item.Name),
				Clone: clone,
				Cpu: vm.VirtualMachineCpuArgs{
					Cores: pulumi.Int(1),
					Type:  pulumi.String("x86-64-v2-AES"),
				},
				Memory: vm.VirtualMachineMemoryArgs{
					Dedicated: pulumi.Int(2048),
				},
				NetworkDevices: networks,
				Bios:           bios,
				EfiDisk:        efidisk,
				Disks:          disks,
				Initialization: initialization,
				OnBoot:         pulumi.Bool(false), // do not start on boot
				StopOnDestroy:  pulumi.Bool(true),
			}, opts...)
			if err != nil {
				err = errors.Wrap(err, "creating VM")
				return err
			}

			if item.Type == "Hacker" {
				bastion = res.ToVirtualMachineOutput()
			}

		}

		// // Retreive IP from the dhcp
		// address := attacker.VmId.ApplyT(func(vmid int) string {
		// 	return findAddressbyName(additional.ProxmoxEndpoint, additional.ProxmoxAPIToken, vmid, "nic0")
		// }).(pulumi.StringOutput)
		// if err != nil {
		// 	err = errors.Wrap(err, "retrive IP")
		// 	return err
		// }
		// quel enfer cette syntaxe de con
		resp.ConnectionInfo = pulumi.All(bastion.Ipv4Addresses(), password.Result).ApplyT(func(all []any) string {

			address := all[0].([][]string)
			password := all[1].(string)

			for _, v := range address {
				if !slices.Contains(v, "127.0.0.1") {
					return fmt.Sprintf("ssh -l user %s, password = %s", v[0], password)
				}
			}

			return ""

		}).(pulumi.StringOutput)

		return nil

	})
}

func loadDefaults() error {

	if additional.BackBridge == "" {
		additional.BackBridge = default_back_bridge
	}
	if additional.FrontBridge == "" {
		additional.FrontBridge = default_front_bridge
	}

	if additional.ProxmoxEndpoint == "" {
		additional.ProxmoxEndpoint = default_pve_endpoint
	}
	if additional.ProxmoxAPIToken == "" { // if not define in additional
		default_proxmox_api_token := os.Getenv("PROXMOX_VE_API_TOKEN") // varenv on chall-manager container
		if default_proxmox_api_token == "" {
			return fmt.Errorf("PROXMOX_VE_API_TOKEN must be provided")
		}
		additional.ProxmoxAPIToken = default_proxmox_api_token
	}

	if additional.ProxmoxSSHPassword == "" { // if not define in additional
		default_proxmox_ssh_password := os.Getenv("PROXMOX_VE_SSH_PASSWORD") // varenv on chall-manager container
		if default_proxmox_ssh_password == "" {
			return fmt.Errorf("PROXMOX_VE_SSH_PASSWORD must be provided")
		}
		additional.ProxmoxSSHPassword = default_proxmox_ssh_password
	}

	if additional.FrontVlan == 0 {
		additional.FrontVlan = 24 // front vlan id
	}
	return nil

}
