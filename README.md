# Skyproxy - Reverse tunnel HTTP proxy

Skyproxy creates a tunnel between any machine to a Server which acts as an HTTP
proxy.

## How does it work?

TODO

## How to use it?

Let's assume I have a web app running on port 1081:

    python -m SimpleHTTPServer 1081

I launch a proxy server on port 1080 (it could be a remote machine running on
a Cloud platform):

    ./skyproxy serve --address "0.0.0.0:1080"

Then, I need to run the proxy client, usually on the same machine than my web
app:

    ./skyproxy connect --server localhost:1080 --receiver localhost:1081 --http-host "localhost:1080"

I tell my proxy client to register the HTTP Host "localhost:1080" and redirect
the traffic to the Receiver (my web app) which lives at localhost:1081

Finally, I can run HTTP queries to the proxy server directly:

    curl localhost:1080

What happens is that the query goes to the proxy server which redirects it to
the proxy client socket. Then the proxy client sends the traffic to the local
web app and handles the traffic back.

## Future features

- HTTPS on both Server and Clients
- More examples to run on prod
