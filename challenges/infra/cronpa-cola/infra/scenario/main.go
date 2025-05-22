package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"

	"github.com/ctfer-io/chall-manager/sdk"
	"github.com/go-playground/form/v4"
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

type Config struct {
	ProxmoxAPIToken    string `form:"proxmox_api_token"`
	ProxmoxEndpoint    string `form:"proxmox_endpoint"`
	ProxmoxSSHPassword string `form:"proxmox_ssh_password"`

	SnippetsDatastore string `form:"snippets_datastore"`

	FrontBridge string `form:"front_bridge"`
	FrontVlan   int    `form:"front_vlan"`
	BackBridge  string `form:"back_bridge"`
}

var (
	vms = []VM{
		{
			Name:       "adds",
			Type:       "Windows",
			TemplateID: 9900,
			Node:       "pve05", // TODO change to the real one
			Datastore:  "raid0",
		},
		{
			Name:       "adsrv",
			Type:       "Windows",
			TemplateID: 9899,
			Node:       "pve05",
			Datastore:  "raid0",
		},
		{
			Name:       "client",
			Type:       "Windows",
			TemplateID: 9901,
			Node:       "pve07",
			Datastore:  "raid0",
		},
		{
			Name:       "hacker",
			Type:       "Hacker",
			CloudImgID: "local:iso/noble-server-cloudimg-amd64.img", // download it before
			Node:       "pve04",
			Datastore:  "local-lvm",
		},
	}
)

func main() {
	sdk.Run(func(req *sdk.Request, resp *sdk.Response, opts ...pulumi.ResourceOption) error {

		// 1. Load config
		conf, err := loadConfig(req.Config.Additional)
		if err != nil {
			return err
		}

		provider, err := proxmoxve.NewProvider(req.Ctx, "pve", &proxmoxve.ProviderArgs{
			ApiToken: pulumi.String(conf.ProxmoxAPIToken),
			Endpoint: pulumi.String(conf.ProxmoxEndpoint),
			Insecure: pulumi.Bool(true), // if you do not trust the x509 PVE Cert
			Ssh: proxmoxve.ProviderSshArgs{ // only for the nodes with Hacker for File upload
				Username: pulumi.String("root"),
				Password: pulumi.String(conf.ProxmoxSSHPassword),

				Nodes: proxmoxve.ProviderSshNodeArray{
					proxmoxve.ProviderSshNodeArgs{
						Address: pulumi.String("10.17.10.104"), // only on the one with the hacker vm
						Name:    pulumi.String("pve04"),
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
		vlan_id, err := GetAvailableId(100, 199)
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

		var bastion vm.VirtualMachineOutput

		for _, item := range vms {

			vm_id, err := GetAvailableId(50000, 59999)
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
						Bridge:     pulumi.String(conf.BackBridge),
						Enabled:    pulumi.Bool(true),
						Model:      pulumi.String("virtio"),
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
					VmId: pulumi.Int(item.TemplateID),
					Full: pulumi.Bool(false),
				}
			}

			if item.Type == "Hacker" {

				qemuEnabled = true

				networks = vm.VirtualMachineNetworkDeviceArray{
					vm.VirtualMachineNetworkDeviceArgs{
						Bridge:  pulumi.String(conf.FrontBridge),
						Enabled: pulumi.Bool(true),
						VlanId:  pulumi.Int(conf.FrontVlan),
					},
					vm.VirtualMachineNetworkDeviceArgs{
						Bridge:  pulumi.String(conf.BackBridge),
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
					DatastoreId: pulumi.String(conf.SnippetsDatastore),
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
					DatastoreId: pulumi.String(conf.SnippetsDatastore),
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
					DatastoreId: pulumi.String(conf.SnippetsDatastore),
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
				VmId:        pulumi.Int(vm_id),
				Description: pulumi.Sprintf("%s VM for %s, password for user account on Hacker VM is %s", item.Name, req.Config.Identity, password.Result),
				NodeName:    pulumi.String(item.Node), // adapt as your need
				Tags: pulumi.ToStringArray([]string{
					"instance",
					"lab-ad1",
					req.Config.Identity,
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

func loadConfig(additionals map[string]string) (*Config, error) {
	// Default conf
	conf := &Config{
		ProxmoxEndpoint:   "https://10.17.10.101:8006/api2/json",
		FrontBridge:       "front",
		FrontVlan:         24,
		BackBridge:        "back",
		SnippetsDatastore: "local",
	}

	// Override with additionals
	dec := form.NewDecoder()
	if err := dec.Decode(conf, toValues(additionals)); err != nil {
		return nil, err
	}

	// checks
	if conf.ProxmoxAPIToken == "" || conf.ProxmoxSSHPassword == "" {
		return nil, errors.New("missing proxmox_api_token or proxmox_ssh_password")
	}
	return conf, nil
}

func toValues(additionals map[string]string) url.Values {
	vals := make(url.Values, len(additionals))
	for k, v := range additionals {
		vals[k] = []string{v}
	}
	return vals
}

type RangeRequest struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

type RangeResponse struct {
	Value int `json:"value"`
}

func GetAvailableId(min, max int) (int, error) {
	url := "http://tools.ctfer-io.lab:8080/next"

	// Customize the range here
	requestBody := RangeRequest{
		Min: min,
		Max: max,
	}

	// Encode request to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return 0, err
	}

	// Send POST request
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error making request:", err)
		return 0, err
	}
	defer resp.Body.Close()

	// Handle non-200 responses
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Server error: %s\n", resp.Status)
		return 0, err
	}

	// Decode response
	var response RangeResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		fmt.Println("Error decoding response:", err)
		return 0, err
	}

	return response.Value, nil
}
