# Civo Packer builder

This is a builder plugin for Packer which can be used to generate storage templates on Civo. It utilises the [Civo Go API](https://github.com/civo/civogo) to interface with the Civo API.

## Installation

### Pre-built binaries

You can download the pre-built binaries of the plugin from the [GitHub releases page](https://github.com/civo/civo-packer/releases). Just download the archive for your operating system and architecture, unpack it, and place the binary in the appropriate location, e.g. on Linux `~/.packer.d/plugins`. Make sure the file is executable, then install [Packer](https://www.packer.io/).

### Installing from source

#### Prerequisites

You will need to have the [Go](https://golang.org/) programming language and the [Packer](https://www.packer.io/) itself installed. You can find instructions to install each of the prerequisites at their documentation.

#### Building and installing

Run the following commands to download and install the plugin from the source.

```
git clone https://github.com/civo/civo-packer
cd civo-packer
go build
cp civo-packer ~/.packer.d/plugins/packer-builder-civo
```

## Usage

The builder will automatically generate a temporary SSH key pair for the `root` user which is used for provisioning. This means that if you do not provision a user during the process you will not be able to gain access to your server.

If you want to login to a server deployed with the template, you might want to include an SSH key to your `root` user by replacing the `<ssh-rsa_key>` in the below example with your public key.

Here is a sample template, which you can also find in the `examples/` directory. It reads your Civo API credentials from the environment variables and creates an Debian buster server in the `lon1` region.

```json
{
  "variables": {
    "CIVO_TOKEN": "{{ env `CIVO_TOKEN` }}",
  },
  "builders": [
    {
      "type": "civo",
      "instance_name": "my-debian-packer",
      "api_token": "{{ user `CIVO_TOKEN` }}",
      "template": "debian-buster",
      "region": "lon1",
      "size": "g2.small",
      "ssh_username": "root"
    }
  ],
  "provisioners": [
    {
      "type": "shell",
      "inline": [
        "apt-get update",
        "apt-get upgrade -y",
        "echo '<ssh-rsa_key>' | tee /root/.ssh/authorized_keys"
      ]
    }
  ]
}
```

Enter the API user credentials in your terminal with the following commands. Replace the `<API_TOKEN>` with your user details.

```json
export CIVO_TOKEN=<API_TOKEN>
```
Then run Packer using the example template with the command underneath.
```
packer build examples/basic_example.json
```
If everything goes according to plan, you should see something like the example output below.

```json
civo output will be in this color.

===> civo: Creating temporary ssh key for instance...
==> civo: Creating instance...
==> civo: Waiting for instance to become active...
==> civo: Using ssh communicator to connect: IP_ADDRESS
==> civo: Waiting for SSH to become available...
==> civo: Connected to SSH!
==> civo: Provisioning with shell script: /var/folders/rp/84fl6_5n20n_gj5vjtszr9vr0000gn/T/packer-shell498350455
	civo: Get:1 http://security.debian.org buster/updates InRelease [65.4 kB]
    civo: Get:2 http://deb.debian.org/debian buster InRelease [121 kB]
    civo: Get:3 http://deb.debian.org/debian buster-updates InRelease [51.9 kB]
    civo: Get:4 http://deb.debian.org/debian buster-backports InRelease [46.7 kB]
    civo: Get:5 http://security.debian.org buster/updates/main Sources [131 kB]
    civo: Get:6 http://security.debian.org buster/updates/main amd64 Packages [213 kB]
...
==> civo: Gracefully shutting down instance...
==> civo: Creating snapshot: civo-packer-1595884528
==> civo: Waiting for snapshot to complete...
==> civo: Destroying instance...
==> civo: Deleting temporary ssh key...
Build 'civo' finished.

==> Builds finished. The artifacts of successful builds are:
--> civo: A snapshot was created: 'civo-packer-1595884528' (ID: ae2f9013-3db4-410c-a4c8-22034c3d605f) in regions 'lon1'
```

## Configuration reference

This section describes the available configuration options for the builder. Please note that the purpose of the builder is to create a storage template that can be used as a source for deploying new servers, therefore the temporary server used for building the template is not configurable.

### Required values

* `api_token` (string) Civo API token.
* `region` (string) The zone in which the server and template should be created (e.g. `lon1`).
* `size` (string) The size of the server, `g2.small`.
* `template` (string) The Code of the template, example `debian-buster`.

### Optional values

* `private_networking` (string) Set to true to enable private networking for the instance being created. This defaults to true.
* `snapshot_name` (string) The name of the resulting snapshot that will appear in your account. Defaults to `packer-{{timestamp}}`
* `state_timeout` (string) The time to wait, as a duration string, for a instance to enter a desired state (such as "active") before timing out. The default state timeout is "6m".
* `snapshot_timeout` (string) How long to wait for an image to be published to the shared image gallery before timing out. If your Packer build is failing on the Publishing to Shared Image Gallery step with the error `Original Error: context deadline exceeded`, but the image is present when you check your Azure dashboard, then you probably need to increase this timeout from its default of "60m" (valid time units include `s` for seconds, `m` for minutes, and `h` for hours.)
* `instance_name` (string) The name assigned to the instance. Civo sets the hostname of the machine to this value.

## License

This project is distributed under the [MIT License](https://opensource.org/licenses/MIT), see LICENSE.txt for more information.