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
adds 4302 -> 4302 (vzdump-qemu-4302-2025_05_07-12_10_48.vma.zst, on pve)

## Networks
For the 24HIUT event, we add a tools exposed on http://tools.ctfer-io.lab:8080/next, that return an availabe id between ranges: 
- VLANs: for this challenge, 2000-2999. 
- VMIDs: for this challenge, 60000-69999.

All this vlans must be configured on the `back` bridges and on the switch on the interfaces for this bridge.