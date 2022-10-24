listen:   # node listen port
  - 8099

admin:    # openAPI request info
  scheme: http # listen scheme
  listen: 9400 # listen port
  ip: 0.0.0.0 # listen ip
#ssl:
#  listen:
#    - port: 443       #https端口
#      certificate:    # 不配表示使用所有 cert_dir中的证书，默认pem文件后缀为pem，key后缀为key
#        - cert: cert.pem
#          key:  cert.key
#certificate:
#  dir: ./cert # 证书文件目录，不填则默认从cert目录下载