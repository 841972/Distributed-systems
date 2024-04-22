#!/bin/bash

# ██████╗░███████╗██████╗░██╗░░░░░░█████╗░██╗░░░██╗
# ██╔══██╗██╔════╝██╔══██╗██║░░░░░██╔══██╗╚██╗░██╔╝
# ██║░░██║█████╗░░██████╔╝██║░░░░░██║░░██║░╚████╔╝░
# ██║░░██║██╔══╝░░██╔═══╝░██║░░░░░██║░░██║░░╚██╔╝░░
# ██████╔╝███████╗██║░░░░░███████╗╚█████╔╝░░░██║░░░
# ╚═════╝░╚══════╝╚═╝░░░░░╚══════╝░╚════╝░░░░╚═╝░░░

# Copiar todos los ficheros de la práctica
scp -r "practica1" "$USER@192.168.3.1:/home/$USER"

# Lista de direcciones IP de destino
direcciones=("192.168.3.2" "192.168.3.3" "192.168.3.4")

# Bucle para copiar y compilar el archivo en cada dirección IP
for ip in "${direcciones[@]}"
do
  # Copiar todos los ficheros de la práctica
  scp -r "practica1" "$USER@$ip:/home/$USER"
  
  # Utiliza ssh para conectarte al servidor remoto y compilar el archivo
  ssh "$USER@$ip" "cd /home/$USER/practica1/cmd/worker && go build -o worker && mv worker /home/$USER"
done

echo "Archivos copiados y compilados en las direcciones IP: ${direcciones[@]}"
