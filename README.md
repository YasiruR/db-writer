# _db-writer_

db-writer is a golang implementation which supports,
1. deploying datasets in csv format to databases
2. benchmarking databases with read and write operations.

### Supported databases

- [Redis](https://redis.io/)
- [Neo4j](https://neo4j.com/)
- [Elasticsearch](https://www.elastic.co/) [Concurrency is removed in db-writer to cope with resource utilization]
- [ArangoDB](https://www.arangodb.com/)

### Executable command

`./writer [-PARAMS]`

- db: database type [redis / neo4j / elasticsearch]
- host: host address of the database
- csv: file path of the csv
- uname: user of the database [OPTIONAL]
- pw: password of the database user [OPTIONAL]
- pwhide: true if the password is sensitive [OPTIONAL]
- ca: path of the CA certificate file [only for elasticsearch]
- table: name of the collection/table
  - arangoDB - required [if omitted, `my_collection` will be used]
  - elasticsearch - required [if omitted, `my_index` will be used]
- dname: name of the database
  - only required for arangodb [if omitted, `_system` will be used]
- unique: unique key of the database
  - redis - required
  - arangoDB - required
  - neo4j - not required
  - elasticsearch - required, but if omitted documents will be indexed by (1,n]

#### 1. Store mode

Additional parameter may be required if you are writing data to the database initially.

- limit: number of data items to be stored [if omitted, all data in csv file will be stored]

eg: `./writer -host=https://localhost:9200 -db=elasticsearch -csv=./github.com/YasiruR/db-writer/data/people.csv -uname=test-user -pw=1234 -ca=./http_ca.crt -limit=10`

#### 2. Test mode

Following parameters are required, if you are performing benchmark tests.

- benchmark: operation of the benchmark test [read / write / update (only for arangoDB)]
- load: number of entries to be used by the test as a burst of requests

eg: `./writer -host=https://localhost:9200 -db=elasticsearch -csv=./github.com/YasiruR/db-writer/data/people.csv -uname=test-user -pw=1234 -benchmark=read -load=100000`

## Elasticsearch Guide

### Deploy using docker

1. Download docker image `docker pull elasticsearch:X.X.X`
   - Replace required version
   - Note that SSL authentication is enabled by default from v8.0.0 upwards
2. Run the container `docker run -it --name elasticsearch -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" elasticsearch:X.X.X`
   - Create a docker network and tag it with above command if required
3. Download the CA certificate to a local directory `docker cp elasticsearch:/usr/share/elasticsearch/config/certs/http_ca.crt .`

### Sample API requests

- Get all documents from an index
  - `curl -H 'Content-Type: application/json' --cacert ./http_ca.crt -X GET https://localhost:9200/<index>/_search?pretty -u <username>:<password>`
- Get document by ID
  - `curl -H 'Content-Type: application/json' --cacert ./http_ca.crt -X GET https://localhost:9200/<index>/_doc/<id> -u <username>:<password>`
- Delete an entire index
  - `curl -X DELETE --cacert ./http_ca.crt https://localhost:9200/<index> -u <username>:<password>`

## Redis Guide

### Deploy using source code

1. Download the source code `wget https://download.redis.io/releases/redis-6.2.6.tar.gz`
2. Extract the file `tar xzf redis-6.2.6.tar.gz`
3. Compile the code `cd redis-6.2.6.tar.gz && make`
4. Create the redis configuration file, `redis.conf` as follows 
   ```
    port 7000 // only required for local deployments or port conflicts
    cluster-enabled yes
    cluster-config-file nodes.conf
    cluster-node-timeout 5000
    appendonly yes
    protected-mode no
   ```
5. Start each node `./redis-server redis.conf`
6. Create the cluster `redis-cli --cluster create 127.0.0.1:7000 127.0.0.2:7000
   --cluster-replicas 1`
    - NB: Use IP addresses since redis instance does not support hostnames

## ArangoDB Guide

### Deploy using package manager (Debian)

1. Add repository key 
   1. `curl -OL https://download.arangodb.com/arangodb39/DEBIAN/Release.key`
   2. `sudo apt-key add - < Release.key`
2. Install using package manager
   1. `echo 'deb https://download.arangodb.com/arangodb39/DEBIAN/ /' | sudo tee /etc/apt/sources.list.d/arangodb.list`
   2. `sudo apt-get install apt-transport-https`
   3. `sudo apt-get update`
   4. `sudo apt-get install arangodb3=3.9.0-1`
3. Start the cluster
   1. Locally: `arangodb --starter.local`
      1. Predefined cluster with 3 coordinators, 3 database servers and 3 agents
   2. Remote: `arangodb` and `arangodb --starter.join <coordinator ip>`
4. Check with Web UI `http://localhost:8529/`

