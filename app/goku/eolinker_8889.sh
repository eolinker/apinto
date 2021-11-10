#!/bin/bash
curl -X POST http://127.0.0.1:9400/api/service --data '{

    "name": "eolinker",

    "driver": "http-service",

    "desc": "匿名服务，直接转发，不需要配负载，没有鉴权",

    "timeout": 3000,

    "upstream": "www.eolinker.com",

    "retry": 3,

    "rewrite_url": "/",

    "scheme": "https"

}' -H "Content-type: application/json"
curl -X POST http://127.0.0.1:9400/api/router --data '{

    "name": "eolinker_8889",

    "driver": "http-service",

    "desc": "http-service",

    "listen": 8889,

    "rules": [{

        "location": "/"

    }],

    "target": "eolinker@service"

}' -H "Content-type: application/json"
