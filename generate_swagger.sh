#!/bin/sh

echo "formatting swagger comments..."
swag fmt -d internal/app/controller

echo "generate swagger docs..."
swag init -g internal/app/controller/controller.go -d ./