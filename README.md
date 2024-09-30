# eth-indexer

![GitHub Tag](https://img.shields.io/github/v/tag/grassrootseconomics/eth-indexer)

A lightweight Postgres chain indexer designed to couple with [eth-tracker](https://github.com/grassrootseconomics/eth-tracker) to index all relevant GE related blockchain data on any EVM chain.

## Getting Started

### Prerequisites

* Git
* Docker
* Postgres server
* Access to a `eth-tracker` instance

See [docker-compose.yaml](dev/docker-compose.yaml) for an example on how to run and deploy a single instance.

### 1. Build the Docker image

We provide pre-built images for `linux/amd64`. See the packages tab on Github.

If you are on any other platform:

```bash
git clone https://github.com/grassrootseconomics/eth-indexer.git
cd eth-indexer
docker buildx build --build-arg BUILD=$(git rev-parse --short HEAD) --tag eth-indexer:$(git rev-parse --short HEAD) --tag eth-indexer:latest .
docker images
```

### 2. Run Postgres

For an example, see `dev/docker-compose.postgres.yaml`.

### 3. Update config values

See `.env.example` on how to override default values defined in `config.toml` using env variables. Alternatively, mount your own config.toml either during build time or Docker runtime.

```bash
# Override only specific config values
nano .env.example
mv .env.example .env
```

Special env variables:

* DEV=*

Refer to [`config.toml`](config.toml) to understand different config value settings.


### 4. Run the indexer

```bash
cd dev
docker compose up
```

## License

[AGPL-3.0](LICENSE).