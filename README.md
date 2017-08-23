## quickstart

### start and configure api container

```
mkdir ~/WIOT-DEMO/
cd ~/WIOT-DEMO/

 # (this assumes you have logged in to the horizon registry)

docker run --name hznpolicyapi -d -p 8091:8091 -v `pwd`/policy.d/:/policy.d/ -v `pwd`/config/:/config/ summit.hovitos.engineering/x86/hzn-policy-api:latest

 # (container will exit since there is no config file)

cat <<\EOF | sudo tee config/config.toml
ListenAddr="0.0.0.0:8091"
PolicyDir="/policy.d"
## Assign the secret key to something random
SecretToken="0c8d0be08f1c414aa0fc772e38e3fd2c"
ServerKeyPath="server.key"
ServerCertPath="server.crt"
NoSec=true
EOF

docker start hznpolicyapi

docker ps 

 # (container should stay up since i'ts configured) 
```

### try some api calls

```

cat <<\EOF | sudo tee policy.d/test.policy
{
   "header": {
      "name": "test"
    }
}
EOF
cat <<\EOF | sudo tee policy.d/test2.policy
{
   "header": {
      "name": "test2"
    }
}
EOF

 # ( GET /policy/{id} )

curl -v -XGET http://localhost:8091/policy/test; echo

 # ( POST /policy/{id} )

cat <<\EOF | curl -v -H 'Content-Type: application/json' -d@- -XPOST http://localhost:8091/policy/test; echo
{
   "header": {
      "name": "test",
      "version": "2.0"
    }
}
EOF

 # ( DELETE /policy/{id} )

curl -v -XDELETE http://localhost:8091/policy/gettest

 # ( GET /policies )

curl -v -XGET http://localhost:8091/policies

 # ( POST /policies )

sudo rm policy/test.policy policy/test2.policy

cat <<\EOF | curl -v -H 'Content-Type: application/json' -d@- -XPOST http://localhost:8091/policies
{"batch1": {"header":{"name":"batch1"}}, "batch2": {"header":{"name":"batch2"}}}
EOF

ls policy.d/ -la

 # ( GET /policies/names )

curl -v -XGET http://localhost:8091/policies/names; echo

 # ( GET /status )

curl -v -XGET http://localhost:8091/status; echo

```
