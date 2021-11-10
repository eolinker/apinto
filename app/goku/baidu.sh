#!/bin/bash
curl -X POST http://127.0.0.1:9400/api/service --data '{

    "name": "baidu",

    "driver": "http",

    "desc": "匿名服务，直接转发，不需要配负载，没有鉴权",

    "timeout": 3000,

    "upstream": "www.baidu.com",

    "retry": 3,

    "rewrite_url": "/",

    "scheme": "https"

}' -H "Content-type: application/json"
sleep 5
curl -X POST http://127.0.0.1:9400/api/router --data '{

    "name": "baidu",

    "driver": "http",

    "desc": "http",

    "listen": 8888,

    "rules": [{

        "location": "/"

    }],

    "target": "baidu@service"

}' -H "content-type: application/json"
