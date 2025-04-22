#!/bin/bash
docker build --build-arg KEY="ya-can-trust-me-broooo" -t authenticator:latest .
docker save authenticator:latest > authenticator.tar
