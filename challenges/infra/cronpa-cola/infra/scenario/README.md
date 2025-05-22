# Prepare proxmox nodes

## Restore Backup

On all host, download the virtio-win.iso here https://fedorapeople.org/groups/virt/virtio-win/direct-downloads/stable-virtio/virtio-win.iso


On webui and on each VM
- Do to Datastores > <id> > Backups > select 4302 then click on restore
- Select the raid0 datastore
- Change the ID with the on on scenario (or update the main.go)
- Finnaly, convert to template

For this challenges: 
Dumps -> TemplateID
adds 200 -> 4100 (vzdump-qemu-200-2025_04_09-22_00_11.vma, on pve2)
adsrv 201 -> 4101 (vzdump-qemu-201-2025_04_09-22_02_40.vma, on pve2)
client 202 -> 4102 (vzdump-qemu-202-2025_04_09-22_03_45.vma, on pve3)

## Networks
For the 24HIUT event, we add a tools exposed on http://tools.ctfer-io.lab:8080/next, that return an availabe id between ranges: 
- VLANs: for this challenge, 1000-1999. 
- VMIDs: for this challenge, 50000-59999.

All this vlans must be configured on the `back` bridges and on the switch on the interfaces for this bridge.