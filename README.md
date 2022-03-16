# _db-writer_

db-writer is a golang implementation to deploy datasets in csv format to databases.

###Supported databases

- [Redis](https://redis.io/)
- [Neo4j](https://neo4j.com/)
- [Elasticsearch](https://www.elastic.co/) [Concurrency is removed in db-writer to cope with resource utilization]

### Executable command

`./writer [-PARAMS]`

- db: database type [redis / neo4j / elasticsearch]
- host: host address of the database
- csv: file path of the csv
- uname: user of the database [OPTIONAL]
- pw: password of the database user [OPTIONAL]
- pwhide: true if the password is sensitive [OPTIONAL]
- ca: path of the CA certificate file [only for elasticsearch]
- limit: number of data items to be stored [If omitted, all data in csv file will be stored]

eg: `./writer -host=https://localhost:9200 -db=elasticsearch -csv=./github.com/YasiruR/db-writer/data/people.csv -uname=test-user -pw=1234 -ca=./http_ca.crt -limit=10`

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