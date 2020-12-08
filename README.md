# Lango Quick

## Intro

Lango Quick is a microservice to serve your i18n json file on the fly.
It caches the whole internalization json in to memory.

## Links

1. [Firebase Translations](https://firebase.google.com/docs/ml-kit/translation)

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

## Pushing to production

No prod yet


## Firebase function example

```typescript
import { firebase, functions } from '../lib/firebase'
import express, { Request, Response, NextFunction } from 'express'
import cors from 'cors'

const app = express()

app.use(cors({ origin: true }))
app.use(express.json())
app.set('trust proxy', 1)

// Checks if user is authorized to use rsx api
const checkIfAuthorized = async (req: Request, res: Response, next: NextFunction) => {
  const email: string = req.body.email
  const token: string = req.body.token
  if (!email?.length || !token?.length) {
    res.sendStatus(400).end()
    return
  }
  try {
    const user = await firebase.auth().getUserByEmail(email)
    const decodedToken = await firebase.auth().verifyIdToken(token)
    const customClaims = <{ admin: boolean; rsxDeveloper: boolean; rsxService?: boolean }>user.customClaims || null
    const isAdmin =
      (decodedToken.admin && customClaims.admin) ||
      (decodedToken.rsxDeveloper && customClaims.rsxDeveloper) ||
      (decodedToken.rsxService && customClaims?.rsxService)
    if (!isAdmin) {
      res.sendStatus(401).end()
      return
    }
    next()
  } catch (e) {
    console.error(e)
    res.sendStatus(401).end()
  }
}

// Serves translations
app.post('/translations', checkIfAuthorized, async (_: Request, res: Response) => {
  try {
    const results: any[] = []
    const snapshot = await db.collection(collections.translations).get()
    snapshot.forEach((r: FirebaseFirestore.QueryDocumentSnapshot<FirebaseFirestore.DocumentData>) => {
      const result = r.data()
      results.push(result)
    })
    res.json(results).end()
  } catch (e) {
    console.error(e)
    res.sendStatus(401).end()
  }
  return
})

export const myApi = functions.https.onRequest(app)
```

## Testing

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
