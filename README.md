*DECOMISSIONED*
See [Basic TME Transformer](https://github.com/Financial-Times/basic-tme-transformer) instead

# locations-transformer

[![CircleCI](https://circleci.com/gh/Financial-Times/locations-transformer.svg?style=svg)](https://circleci.com/gh/Financial-Times/locations-transformer) [![Go Report Card](https://goreportcard.com/badge/github.com/Financial-Times/locations-transformer)](https://goreportcard.com/report/github.com/Financial-Times/locations-transformer) [![Coverage Status](https://coveralls.io/repos/github/Financial-Times/locations-transformer/badge.svg?branch=master)](https://coveralls.io/github/Financial-Times/locations-transformer?branch=master)

Retrieves Locations taxonomy from TME and transforms the locations to the internal UP json model.
The service exposes endpoints for getting all the locations and for getting location by uuid.

# Usage
`go get github.com/Financial-Times/locations-transformer`

`$GOPATH/bin/locations-transformer --port=8080 --base-url="http://localhost:8080/transformers/locations/" --tme-base-url="https://tme.ft.com" --tme-username="user" --tme-password="pass" --token="token"`
```
export|set PORT=8080
export|set BASE_URL="http://localhost:8080/transformers/locations/"
export|set TME_BASE_URL="https://tme.ft.com"
export|set TME_USERNAME="user"
export|set TME_PASSWORD="pass"
export|set TOKEN="token"
$GOPATH/bin/locations-transformer
```

With Docker:

`docker build -t coco/locations-transformer .`

`docker run -ti --env BASE_URL=<base url> --env TME_BASE_URL=<structure service url> --env TME_USERNAME=<user> --env TME_PASSWORD=<pass> --env TOKEN=<token> coco/locations-transformer`
