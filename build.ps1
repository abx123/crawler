$Env:GOOS = "linux"
$Env:CGO_ENABLED = "0"
$Env:GOARCH = "amd64"
go build -o main main.go
~\Go\Bin\build-lambda-zip.exe -output main.zip main