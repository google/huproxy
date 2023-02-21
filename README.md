# MOVED TO https://github.com/ThomasHabets/huproxy

# HUProxy

Copyright 2017 Google Inc.

This is not a Google product.

HTTP(S)-Upgrade Proxy — Tunnel anything (but primarily SSH) over HTTP
websockets.

## Why

The reason for not simply using a SOCKS proxy or similar is that they tend to
take up an entire port, while huproxy only takes up a single URL subdirectory.

There's also
[SSL/SSH multiplexers](http://www.rutschle.net/tech/sslh.shtml)
but they:

1. Take over the port and front both the web server and SSH, instead of letting
   the web server be the primary owner of port 443.
2. For SSH don't look like SSL for packet inspectors, because they're not.
3. Hide the original client address from the web server (without some
   [interesting iptables magic](https://github.com/yrutschle/sslh#transparent-proxy-support)).
4. Only allow connecting to the server itself, not treat it as a proxy jumpgate.

## Setup

### nginx

#### Create user

```bash
sudo htpasswd -c /etc/nginx/users.proxy thomas
```

#### Add config to nginx

```nginx
map $http_upgrade $connection_upgrade {
    default upgrade;
         '' close;
}

server {
    # ... other config
    location /proxy {
        auth_basic "Proxy";
        auth_basic_user_file /etc/nginx/users.proxy;
        proxy_pass http://127.0.0.1:8086;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        # proxy_set_header Connection "upgrade";
        proxy_set_header Connection $connection_upgrade;
    }
    # ... other config
}

```

Start proxy:

```bash
./huproxy
```

Start proxy with specific <IP:port>:

```bash
./huproxy -listen 10.1.2.3:8086
```

## Running

These commands assume that HTTPS is used. If not, then change "wss://"
to "ws://".

```bash
echo thomas:secretpassword > ~/.huproxy.pw
chmod 600 ~/.huproxy.pw
cat >> ~/.ssh/config << EOF
Host shell.example.com
    ProxyCommand /path/to/huproxyclient -auth=@$HOME/.huproxy.pw wss://proxy.example.com/proxy/%h/%p
EOF

ssh shell.example.com
```

Or manually with these commands:

```bash
ssh -o 'ProxyCommand=./huproxyclient -auth=thomas:secretpassword wss://proxy.example.com/proxy/%h/%p' shell.example.com
ssh -o 'ProxyCommand=./huproxyclient -auth=@<(echo thomas:secretpassword) wss://proxy.example.com/proxy/%h/%p' shell.example.com
ssh -o 'ProxyCommand=./huproxyclient -auth=@$HOME/.huproxy.pw wss://proxy.example.com/proxy/%h/%p' shell.example.com
```

If remote server uses self-signed or invalid certificate then use `-insecure_conn`, for example:

```bash
ssh -o 'ProxyCommand=./huproxyclient -insecure_conn wss://proxy.example.com/proxy/%h/%p' shell.example.com
```
