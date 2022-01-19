# Golang client/server tools for Globitex/Nexpay API

## Client

Currenly it partially implements the following subset of API methods.

Nexpay methods:

- Create new payment
- Get Account Status
- Get Deposit Details
- Get Payment History
- Get Payment Status
- Get Payment Commission Amount

## Server

Server part provides middleware for validating request signature.


## Examples

```
GLOBITEX_TOOL_API_KEY=abc
GLOBITEX_TOOL_MESSAGE_SECRET=cde
GLOBITEX_TOOL_TRANSACTION_SECRET=fgh
GLOBITEX_TOOL_HOST=http://localhost:8090
go run main.go
```

See example of usage in main.go
