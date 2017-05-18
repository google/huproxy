# HUProxy

## Setup

### nginx

#### Create user

```
sudo htpasswd -c /etc/nginx/users.proxy thomas
```

#### Add config to nginx

```
map $http_upgrade $connection_upgrade {
    default upgrade;
         '' close;
}
location /proxy {
    auth_basic "Proxy";
    auth_basic_user_file /etc/nginx/users.proxy;
    proxy_pass http://127.0.0.1:8999;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    # proxy_set_header Connection "upgrade";
    proxy_set_header Connection $connection_upgrade;
}
```

Start proxy:

```
./huproxy
```

## Running

These commands assume that HTTPS is used. If not, then change "wss://"
to "ws://".

```
echo thomas:secretpassword > ~/.huproxy.pw
chmod 600 ~/.huproxy.pw
cat >> ~/.ssh/config << EOF
Host shell.example.com
    ProxyCommand /path/to/huproxyclient -auth=@$HOME/.huproxy.pw wss://proxy.example.com/proxy/%h/%p
EOF

ssh shell.example.com
```

Or manually with these commands:

```
ssh -o 'ProxyCommand=./huproxyclient -auth=thomas:secretpassword wss://proxy.example.com/proxy/%h/%p' shell.example.com
ssh -o 'ProxyCommand=./huproxyclient -auth=@<(echo thomas:secretpassword) wss://proxy.example.com/proxy/%h/%p' shell.example.com
ssh -o 'ProxyCommand=./huproxyclient -auth=@$HOME/.huproxy.pw wss://proxy.example.com/proxy/%h/%p' shell.example.com
```