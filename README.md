# Build image

```
docker buildx build --platform linux/arm/v7 -t <user>/<repo>:<tag> --push .
``` 

# Usage

### Docker compose

```
...

  pull-repo:
    image: <user>/<repo>:<tag>
    restart: unless-stopped
    volumes: 
      - /path_to_repo:/repo
      - /path_to_config/prod.yaml:/app/config/prod.yaml
      - /path_to_ssh_key/id_rsa:/root/.ssh/id_rsa
    ports:
      - "8888:80"
    environment:
      - ENV=prod

...
```

in github webhook configuration set url to: 

`http://your_ip_or_domain:8888/pull`