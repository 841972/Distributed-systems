# Usa la imagen oficial de Golang como base
FROM golang:latest

# Establece el directorio de trabajo dentro del contenedor
WORKDIR /go/src/app

# Copia el código fuente al contenedor
COPY . .

# Compila el código
RUN CGO_ENABLED=0 go build -o servidor cmd/srvraft/main.go

# Exponer el puerto que usa tu servidor
EXPOSE 8080

# Comando para ejecutar el servidor al iniciar el contenedor
CMD ["./servidor"]
