# Sourcecraft Actions

This repository contains various actions for Yandex Cloud services.

## Docker Images

The repository includes multiple applications that can be built as Docker images:

- `apigw`: API Gateway action
- `coi`: COI action
- `container`: Container action
- `function`: Function action
- `obj-storage-upload`: Object Storage Upload action

### Building Docker Images

This repository uses Docker Buildx Bake to build all images. The configuration is defined in the `docker-bake.hcl` file.

#### Build all images

```bash
docker buildx bake
```

#### Build a specific image

```bash
docker buildx bake apigw
```

#### Build with custom registry and tag

```bash
docker buildx bake --set *.tags=myregistry.io/image:v1.0
```

Or using variables:

```bash
REGISTRY=myregistry.io/ TAG=v1.0 docker buildx bake
```

#### Build and push to registry

```bash
docker buildx bake --push
```

#### Build for multiple platforms

```bash
docker buildx bake --set *.platform=linux/amd64,linux/arm64
```

## Applications

### API Gateway (apigw)

API Gateway action for Yandex Cloud.

### COI (coi)

COI action for Yandex Cloud.

### Container (container)

Container action for Yandex Cloud.

### Function (function)

Function action for Yandex Cloud.

### Object Storage Upload (obj-storage-upload)

Object Storage Upload action for Yandex Cloud.