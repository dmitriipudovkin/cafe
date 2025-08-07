# cafe

## Run

```
cd ./back

brew services start redis
go run ./main.go
```

## Test 

```
go test -v ./... -run TestAuthService
```