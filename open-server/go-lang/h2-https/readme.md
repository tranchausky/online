# build https support h2

```
# Build
go build -o static-https.exe main.go

# Lần đầu chạy (tự sinh cert) – ví dụ port 8443, host=localhost:
./static-https.exe -port 8443 -root . -host localhost
# Cert sẽ nằm ở: ./.certs/localhost.crt và ./.certs/localhost.key

# Lần sau chạy lại: sẽ tái sử dụng cert cũ
./static-https.exe -port 8443 -root . -host localhost

# Ép sinh lại cert:
./static-https.exe -port 8443 -root . -host localhost -regen

# Đổi thư mục lưu cert:
./static-https.exe -certdir ./_certs -host localhost

```
