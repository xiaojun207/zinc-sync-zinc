# zinc-sync-zinc
Zinc synchronization tool

# env

```shell
PRIMARY_ZINC_HOST="localhost:4081"
PRIMARY_ZINC_USER="admin"
PRIMARY_ZINC_PASSWORD="Complexpass#123"

SECONDARY_ZINC_HOST="localhost:4080"
SECONDARY_ZINC_USER="admin"
SECONDARY_ZINC_PASSWORD="Complexpass#123"

IGNORE_INDEX_LIST=""
PAGE_SIZE=100
GOROUTINE_LIMIT=1000
DEBUG=false

```

# quick start
```shell
docker run -e PRIMARY_ZINC_HOST="192.168.3.100:4080" \
 -e PRIMARY_ZINC_USER="admin"\
 -e PRIMARY_ZINC_PASSWORD="Complexpass#123"\
 -e SECONDARY_ZINC_HOST="192.168.3.102:4080" \
 -e SECONDARY_ZINC_USER="admin" \
 -e SECONDARY_ZINC_PASSWORD="Complexpass#123" \
 -e IGNORE_INDEX_LIST="" \
 -e PAGE_SIZE=100 \
 -e GOROUTINE_LIMIT=1000 \
 -e DEBUG="false" \
 xiaojun207/zinc-sync-zinc:0.0.1
 
```
