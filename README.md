# Lango Quick

## Intro

Lango Quick is a microservice to serve your i18n json file on the fly.
It caches the whole internalization json in to memory.

## Links

1. Local endpoint for getting translation when firebase is run with emulator ```http://localhost:5001/foodie-translation/us-central1/translations```

## Environment

1. Runs on localhost:8080, to change it set proper environment variable in `.env` and map ports in `docker-compose.yml`

## Production environment details

No production yet

## What do I need to get started

1. [docker](https://docs.docker.com/get-docker/)
2. [docker-compose](https://docs.docker.com/compose/install/)
3. [Firebase](https://firebase.google.com/)

## How to setup the project

- install docker and docker-compose
- set up Firebase account and get credentials
- point to your firebase functions API

## Dev commands (pick one from below )

- ```go run .``` if you have go 1.15 or higher on your machine
- ```docker-compose up -d``` then ```docker exec -it lango_quick /bash/bin``` to develop in docker container

## Build commands (pick one from below)

- ```docker-compose up -d```
- ```go run .```
- ```docker build .```
  If in docker, please state Docker commands here

## Contribution

Clone the repo and send merge request with proposed feature, fix or change

## ENDPOINTS:

```localhost:8080/translations/{language_code}``` - to get translations for all inputs at once
