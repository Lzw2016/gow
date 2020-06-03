@echo off
chcp 65001
choice /C wl /M "请选择操作系统: windows,linux"
if errorlevel 2 goto linux
if errorlevel 1 goto windows
:windows
echo build windows amd64
set CGO_ENABLED=0
set GOOS=windows
set GOARCH=amd64
go build -ldflags "-s -w" -o app.exe  main.go
goto end

:linux
echo build linux amd64
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64
go build -ldflags "-s -w" -o app  main.go
goto end

:end
echo done~