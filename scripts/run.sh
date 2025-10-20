#!/bin/bash

podman-compose -f podman-compose.yml build
podman-compose -f podman-compose.yml up -d

open frontend/index.html