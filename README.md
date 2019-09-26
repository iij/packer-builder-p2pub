# packer-builder-p2pub
Packer builder plugin for IIJ GIO P2 Public Resource

## Installation

- Download plugin binaries from [releases](https://github.com/sishihara/packer-builder-p2pub/releases).
- Install the binary by following [official instruction](https://www.packer.io/docs/extending/plugins.html#installing-plugins)

## Usage

### Example

- Before running this sample, you must get API key (available at [here](https://help.api.iij.jp/access_keys))
- Save this JSON as ```sample.json```

```
{
    "builders": [{
        "type": "p2pub",
        "access_key_id": "{{ user `access_key_id` }}",
        "secret_access_key": "{{ user `secret_access_key` }}",
        "gis_service_code": "{{ user `gis_service_code` }}",
        "storage_type": "SX30GB_UBUNTU18_64",
        "label": "packer-builder-p2pub-sample"
    }],
    "provisioners": [{
        "type": "shell",
        "inline": "touch /var/tmp/packer-builder-p2pub-sample"
    }]
}
```

- Run Packer

```
$ packer build -var access_key_id=<YOUR-ACCESS-KEY> secret_access_key=<YOUR_SECRET_KEY> -var gis_service_code=<YOUR_GIS_SERVICE_CODE> sample.json
```

After running this sample, the image is saved in [Storage Archive](https://manual.iij.jp/p2/pub/e-5-1.html).

### Configuration

**Builder-specific parameters**

| key | value | required |
|-|-|-|
| ```access_key_id``` | IIJ API Access key ID | ◯ |
| ```secret_access_key``` | IIJ API Secret access key | ◯ |
| ```gis_service_code``` | Service Code of P2 contract to use for building and save the artifact in | ◯ |
| ```storage_type``` | Built image's [Storage Type (System Storage)](https://manual.iij.jp/p2/pubapi/59949023.html). | ◯ |
| ```server_type``` | [Server Type](https://manual.iij.jp/p2/pubapi/59949011.html) used while build. Default is ```VB0-1```. | |
| ```base_image``` | Base image the artifact is built from. | |
| ```base_image.gis_service_code``` | Base image's Service Code (P2 root contract) | |
| ```base_image.iar_service_code``` | Base image's Service Code (Storage Archive) | |
| ```base_image.image_id``` | Base image's Image ID | |
| ```root_ssh_key``` | SSH public key for root user, used by provisioners. If it is empty, P2 PUB builder will generate temporary SSH key pair. | |
| ```label``` | Label text for build images. | |
| ```disable_global_address``` | If this is set true, P2 PUB builder connects to VMs through [Standard Private Network](https://manual.iij.jp/p2/pub/b-5-1-1.html). Default is ```false``` (the builder allocate a global IP address and provisioning by using the Internet).| |

**Common parameters modified by P2 PUB builder**

| key | value | required |
|-|-|-|
| ```ssh_username```| [SSH Username (SSH communicator)](https://www.packer.io/docs/templates/communicator.html#ssh-communicator). If it is empty, P2 PUB builder overrides as root user. | |
| ```ssh_private_key_file```| [SSH Private key file path (SSH communicator)](https://www.packer.io/docs/templates/communicator.html#ssh-communicator). If it is empty, P2 PUB builder will generate temporary SSH key pair. | |

full example:

```
{
    ...
    "builders": [{
        "type": "p2pub",
        "access_key_id": "YOUR_ACCESS_KEY_ID",
        "secret_access_key": "YOUR_SECRET_ACCESS_KEY",
        "gis_service_code": "gis00000000",
        "storage_type": "SX30GB_UBUNTU18_64",
        "server_type": "VB12-24",
        "base_image": {
            "gis_service_code": "gis99999999",
            "iar_service_code": "iar99999999",
            "image_id": "123456"
        },
        "root_ssh_key": "ssh-rsa AAAAA.....",
        "label": "packer-builder-p2pub-sample",
        "disable_global_address": false
    }],
    ...
}
```

## Development

...

## References

- IIJ GIO P2 Public Resource API Reference : http://manual.iij.jp/p2/pubapi/index.html