#!/bin/bash
curl -X POST http://127.0.0.1:9400/api/service --data '{

    "name": "gokuapi",

    "driver": "http-service",

    "desc": "访问官网",

    "timeout": 3000,

    "upstream": "www.gokuapi.com",

    "retry": 3,

    "rewrite_url": "/",

    "scheme": "https"

}' -H "Content-type: application/json"

echo -e "\n"
curl -X POST http://127.0.0.1:9400/api/router --data '{

    "name": "gokuapi",

    "driver": "http-service",

    "desc": "http-service",

    "listen": 8888,

    "rules": [{

        "location": "/"

    }],

    "target": "gokuapi@service"

}' -H "Content-type: application/json"
echo -e "\n"
