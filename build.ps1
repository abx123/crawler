$Env:GOOS = "linux"
$Env:CGO_ENABLED = "0"
$Env:GOARCH = "amd64"
go build -o main main.go
md -Force outputs
~\Go\Bin\build-lambda-zip.exe -output outputs\main.zip main