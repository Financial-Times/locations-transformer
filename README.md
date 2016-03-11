# locations-transformer

[![Circle CI](https://circleci.com/gh/Financial-Times/locations-transformer/tree/master.png?style=shield)](https://circleci.com/gh/Financial-Times/locations-transformer/tree/master)

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
