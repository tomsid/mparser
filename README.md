mparser
=======
 
###Description
 
 A message parser that supports emoticons, mentions and links.
 Implemented as a simple webserver that accepts requests on root handler.
 Supports both http and https protocols.
 The parser itself is implemented as a subpackage `github.com/tomsid/mparser/parser`.
 Dependencies are managed with glide. Run `glide install` to install dependencies.
 
###Usage
 
 Start with necessary options:
 
   `# LISTEN_PORT=9091 ./mparser`

 In another terminal send a request:
 
   `# curl -X POST -d 'https://www.facebook.com/ (as) @somemention' "http://localhost:9091/"`
 
 Start server with TLS(example cert and key are included):
 
   `# LISTEN_HOST=127.0.0.1 LISTEN_PORT=2323 SSL_ON=1 SSL_KEY_PATH=./key.rsa SSL_CERT_PATH=./cert.crt ./mparser`
 
###Environment variables

 - LISTEN_PORT: default is "localhost"
 - LISTEN_HOST: default is "9080"
 - SSL_ON: "1" - enabled, "0" - disabled(default).
 - SSL_CERT_PATH: path to SSL certificate. Mandatory if SSL_ON=1.
 - SSL_KEY_PATH: path to key. Mandatory if SSL_ON=1.