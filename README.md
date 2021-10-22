# Geocode Proxy Server

Simple proxy server which allows to geocode addresses using Google API.

## Installation
### Using Go
Requires  Go 1.15+ [www.golang.org](https://www.golang.org)

Run 
```
go get github.com/kpawlik/geocode_proxy
go install github.com/kpawlik/geocode_proxy
```

## Configuration
Create configuration JSON file.
```
{
    "geocoder": {
        "apiKey": string,
        "clientId": string,
        "clientSecret": string,
        "channel": string
    },
    "server": {
        "port": int,
        "logLevel": string
    },
    "quota": int,
    "quotaTimeInMinutes": int,
    "workersNumber": int
}
```

- `apiKey` - 
- `clientId` - 
- `clientSecret` - 
- `channel` - 
- `port` - (default `9998`)
- `logLevel` - (default `info`)
- `quota` - (default `0`)
- `quotaTimeInMinutes` - (default `0` - unlimited)
- `workersNumber` - (default `10`  - unlimited)
## How to start

Create config file in the same location as executable.
Execute:

Windows
```
.\geocode_proxy.exe
```
Linux
```
./geocode_proxy
```
Create config file in any location.
Execute:

Windows
```
.\geocode_proxy.exe -config [PATH_TO_JSON_CONFIG_FILE]
```
Linux
```
./geocode_proxy -config [PATH_TO_JSON_CONFIG_FILE]
```

Geocode service will be aviable on `http://localhost:8080/geocode`

## How to use
Send HTTP POST request.

### JSON Request
```
{
    "addresses": [
        {
            "address": string,
            "id": string
        }
    ]
}
```
Example using jQuery:

```js
const data = {"addresses":[
    {
        "address": "Address string",
        "id": "ID string"
    }
]};
let result = await $.post("/geocode", JSON.stringify(data));
```

### Response
```
{
    "addresses": [
        {
            "address": string,
            "id": string,
            "lat": float,
            "lng": float,
            "error": string
        },
        ...
    ],
    "error": string
}
```
### Errors
- `UNABLE_TO_GEOCODE` - Google API was not able to geocode this address
- `GOOGLE_OVER_QUERY_LIMIT` - You exceeded your limit of API requests on Google
- `SERVER_OVER_QUERY_LIMIT` - You exceeded your limit of API requests on the server
