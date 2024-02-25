# stepin

A simple certificate management with step-cli for internal network use


# Build
```shell
docker build --build-arg https_proxy=http://host.docker.internal:1080 -t allape/stepin .
```


# Run
```shell
docker compose -f compose.stepin.yaml up -d
```
