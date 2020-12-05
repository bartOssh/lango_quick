# Lango Quick

# Intro

Lango Quick is a microservice to serve your i18n json file on the fly.
It caches the whole internalization json in to memory.

# Links

1. [Firebase Translations](https://firebase.google.com/docs/ml-kit/translation)

# Environment

1. Runs on localhost:8080, to change it set proper environment variable in `.env` and map ports in `docker-compose.yal`

## Production environment details

No production yet

# What do I need to get started

1. [docker](https://docs.docker.com/get-docker/)
2. [docker-compose](https://docs.docker.com/compose/install/)
3. [Firebase](https://firebase.google.com/)

# How to setup the project

- install docker and docker-compose
- set up Firebase account and get credentials
- point to your firebase functions API

# Dev commands

- ```go run .```
- ```docker-compose up -d``` then ```docker exec -it lango_quick /bash/bin```

# Build commands

- ```docker-compose up -d```
- ```go run .```
- ```docker build .```
  If in docker, please state Docker commands here

# Contribution

Clone the repo and send merge request with proposed feature, fix or change

# Pushing to production

No prod yet

# Testing

```javascript
const options = {
  host: "127.0.0.1",
  path: "/translations",
  port: "8080",
  method: "GET",
};

let numOfCalls = 1000;
const timeStart = Date.now();

(function loop() {
  if (numOfCalls > 0) {
    callback = (response, res) => {
      response.on("data", function (chunk) {});

      response.on("end", function () {
        numOfCalls--;
        loop();
      });
    };
    var req = http.request(options, callback);
    req.write("");
    req.end();
  } else {
    console.log(
      `time to get 1000 translations from GO server is: ${
        Date.now() - timeStart
      } ms`
    );
  }
})();
```

## ENDPOINTS:

This microservice has only one endpoint but two different endpoint calls are possible:

1. ```localhost:8080/translations``` - to get translations for all inputs at once
2. ```localhost:8080/translations?input=Bailout%20gases``` - to get translation for specific input
