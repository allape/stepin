# Project-Stepin

A simple certificate management with step-cli for internal network use.

## Run

### Docker

```shell
docker run -d --name stepin -p 8080:8080 -v "$(pwd)/database:/app/database" \
  -e STEPIN_ROOT_CA_PASSWORD="$(openssl rand -base64 15)" \
  ghcr.io/allape/stepin:main
```

### Docker Compose

```shell
vim docker.compose.yaml # at least change one of the passwords
docker compose -f docker.compose.yaml up -d
```

## Dev

### Backend

```shell
go run .
```

### Frontend

```shell
cd ui
npm i
npm run dev
```

# Credits

- [favicon.png](asset/favicon.png): https://www.irasutoya.com/2017/07/blog-post_676.html
