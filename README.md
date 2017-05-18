HUProxy

## nginx

```
map $http_upgrade $connection_upgrade {
    default upgrade;
         '' close;
}
location /proxy {
    proxy_pass http://127.0.0.1:8999;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    # proxy_set_header Connection "upgrade";
    proxy_set_header Connection $connection_upgrade;
}
```
