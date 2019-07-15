# Example on how to start hidden Redmine service under Fedora

## Generate self-signed certificate inside ./tls directory

```bash
openssl req -nodes -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365
```

## write docker-compose.yml

```docker-compose
ersion: '3.7'

services:

  redmine:
    image: redmine
    restart: always
    environment:
      REDMINE_DB_MYSQL: db
      REDMINE_DB_PASSWORD: example
    volumes:
      - "./remine_files:/usr/src/redmine/files:z"

  db:
    image: mariadb
    restart: always
    environment:
      MYSQL_ROOT_PASSWORD: example
      MYSQL_DATABASE: redmine
    volumes:
      - "./mariadb:/var/lib/mysql:z"
    command: mysqld --character-set-server=utf8mb4 --collation-server=utf8mb4_unicode_ci

  tor:
    image: jess/tor
    restart: always
    user: root
    volumes:
      - "./tor_volumes:/hs:z"
    command: tor --allow-missing-torrc --ignore-missing-torrc HiddenServiceDir /hs HiddenServicePort "443 tls:4333"

  tls:
    build: "github.com/AnimusPEXUS/simpletlsproxy.git"
    restart: always
    volumes:
      - "./tls:/tls:z"
    command: app redmine:3000 :4333

```
