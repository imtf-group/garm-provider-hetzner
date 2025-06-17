# Garm External Provider For Hetzner (HCloud)

The Hetzner external provider allows [garm](https://github.com/cloudbase/garm) to create Linux and Windows runners on top of AWS virtual machines.

## Build

Clone the repo:

```bash
git clone https://github.com/imtf-group/garm-provider-hetzner
```

Build the binary:

```bash
cd garm-provider-hetzner
go build .
```

Copy the binary on the same system where garm is running, and [point to it in the config](https://github.com/cloudbase/garm/blob/main/doc/providers.md#the-external-provider).

## Configuration

The config file for this external provider is a simple toml used to configure the Hetzner token it needs to handle virtual machines.

```toml
location = "nbg1"
token = "sample_token"
```

## Customization

This provider can be customized through extra specs you would add to your GARM pool.

This is an example :

```json
{
    "location":"nbg1",
    "ssh_keys": [
        1111111,
        2222222
    ],
    "datacenter": "nbg1-dc2",
    "placement_group": 101010,
    "networks": [
        123123
    ],
    "firewalls": [
        321321,
        543456
    ],
    "disable_ipv6": true,
    "disable_ipv4": true,
    "disable_updates": true,
    "enable_boot_debug": true,
    "extra_context": {
        "GolangDownloadURL": "https://go.dev/dl/go1.22.4.linux-amd64.tar.gz"
    },
    "extra_packages": [
        "apg",
        "tmux"
    ],
    "pre_install_scripts": {
        "01-script": "IyEvYmluL2Jhc2gKCgplY2hvICJIZWxsbyBmcm9tICQwIiA+PiAvMDEtc2NyaXB0LnR4dAo=",
        "02-script": "IyEvYmluL2Jhc2gKCgplY2hvICJIZWxsbyBmcm9tICQwIiA+PiAvMDItc2NyaXB0LnR4dAo="
    },
    "runner_install_template": "(...)"
}
```

In a nutshell, `ssh_keys`, `placement_group`, `networks` and `firewalls` uses Hetzner resource ID whereas the other values uses the resource name.

The extra-specs can be added to the pool with the following command:

```
garm-cli pool update <pool ID> --extra-specs='{"ssh_keys":[104506]}'
```
