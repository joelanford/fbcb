# fbcb

`fbcb` (file-based config build) is a tool that builds composite catalogs for OLM

## Try it!

```console
git clone https://github.com/joelanford/fbcb
cd fbcb
go run ./ -c example/fbcb.yaml
docker images | grep fbcb-example
docker run -d --rm --name fbcb-example -p 50051:50051 quay.io/joelanford/fbcb-example:raw
grpcurl -plaintext localhost:50051 api.Registry/ListPackages
docker kill fbcb-example
```

## Configuration

There are two primary configuration files required by `fbcb`:
1. The base build configuration (typically called `fbcb.yaml`), which specifies the set of catalog images that are built by `fbcb` and to which the individual packages can contribute.
2. A per-package `config.yaml`, which specifies a set of package build configurations, each of which can contribute to one or more of the catalogs specified in the base build configuration file.

### Example `fbcb.yaml`

```yaml
packagesBaseDir: packages
catalogs:
  - name: basic-catalog1
    destination:
      baseImage: quay.io/operator-framework/opm:latest
      extraLabels:
        kind: basic
      outputImage: quay.io/joelanford/fbcb-example:basic1
  - name: basic-catalog2
    destination:
      baseImage: quay.io/operator-framework/opm:latest
      extraLabels:
        kind: basic
      outputImage: quay.io/joelanford/fbcb-example:basic2
  - name: semver-catalog
    destination:
      baseImage: quay.io/operator-framework/opm:latest
      extraLabels:
        kind: semver
      outputImage: quay.io/joelanford/fbcb-example:semver
  - name: custom-catalog
    destination:
      baseImage: quay.io/operator-framework/opm:latest
      extraLabels:
        kind: custom
      outputImage: quay.io/joelanford/fbcb-example:custom
  - name: raw-catalog
    destination:
      baseImage: quay.io/operator-framework/opm:latest
      extraLabels:
        kind: raw
      outputImage: quay.io/joelanford/fbcb-example:raw
```

### Example `config.yaml`

```yaml
catalogs:
  # Use opm basic veneer
  - buildConfigs:
    - basic-catalog1
    - basic-catalog2
    buildStrategy:
      name: opmBasicVeneer
      opmBasicVeneer:
        input: basic.yaml

  # Use opm semver veneer
  - buildConfigs:
    - semver-catalog
    buildStrategy:
      name: opmSemverVeneer
      opmSemverVeneer:
        input: semver.yaml

  # Use a custom veneer
  - buildConfigs:
    - custom-catalog
    buildStrategy:
      name: custom
      custom:
        command:
          - ./custom-build.sh

  # Use raw FBC
  - buildConfigs:
    - raw-catalog
    buildStrategy:
      name: raw
      raw:
        dir: raw
```
