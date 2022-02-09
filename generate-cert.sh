#!/bin/bash
mkdir -p $HOME/.httpecho/
mkcert -key-file $HOME/.httpecho/server.key -cert-file $HOME/.httpecho/server.crt localhost 127.0.0.1 ::1
