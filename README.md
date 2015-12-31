# Skyproxy - Reverse tunnel HTTP proxy

Skyproxy creates a tunnel between any machine to a Server which acts as an HTTP
proxy.

## How does it work?

TODO: write

## How to use it?

Let's assume I have a web app running on port 1081:

    python -m SimpleHTTPServer 1081

I launch a proxy server on port 1080 (it could be a remote machine running on
a Cloud platform), let's say I want to serve the traffic from the domain `public.domain.tld`:

    ./skyproxy serve --proxy-http --proxy-http :80 --clients-http :1080

Or by using docker:

    docker run -d --net=host samalba/skyproxy serve --proxy-http --proxy-http :80 --clients-http :1080

Then, I need to run the proxy client, usually on the same machine than my web
app:

    ./skyproxy connect --server public.domain.tld:1080 --receiver localhost:1081 --http-host "public.domain.tld"

I tell my proxy client to register the HTTP Host "public.domain.tld", which
tells the Proxy server to redirect the traffic for this domain to my Skyproxy
client which runs locally. The local client will then redirect the traffic to
my local web app (on localhost:1081)

Finally, I can run HTTP queries to the proxy server directly (with curl or any
web browser):

    curl public.domain.tld

What happens is that the query goes to the proxy server which redirects it to
the proxy client socket. Then the proxy client sends the traffic to the local
web app and handles the traffic back.

## Security and production

Skyproxy supports HTTPS for the server, and client-side certificates to
identify the Skyproxy clients. However it's EXPERIMENTAL for now.

## TODO

- More examples to run on prod, more docs, more tests
