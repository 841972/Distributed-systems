#!/bin/bash

# ██████╗░███████╗██████╗░██╗░░░░░░█████╗░██╗░░░██╗
# ██╔══██╗██╔════╝██╔══██╗██║░░░░░██╔══██╗╚██╗░██╔╝
# ██║░░██║█████╗░░██████╔╝██║░░░░░██║░░██║░╚████╔╝░
# ██║░░██║██╔══╝░░██╔═══╝░██║░░░░░██║░░██║░░╚██╔╝░░
# ██████╔╝███████╗██║░░░░░███████╗╚█████╔╝░░░██║░░░
# ╚═════╝░╚══════╝╚═╝░░░░░╚══════╝░╚════╝░░░░╚═╝░░░

# Copiar todos los ficheros de la práctica
scp -r "/home/$USER/practica2/codigo" "$USER@192.168.3.1:/home/$USER"

# Lista de direcciones IP de destino
direcciones=("192.168.3.2" "192.168.3.3" "192.168.3.4")

# Bucle para copiar y compilar el archivo en cada dirección IP
for ip in "${direcciones[@]}"
do
  # Hacer limpieza
  ssh "$USER@$ip" "cd /home/$USER && rm -rf *"

  # Copiar todos los ficheros de la práctica
  scp -r "/home/$USER/practica2/codigo" "$USER@$ip:/home/$USER"
  
  # Utiliza ssh para conectarte al servidor remoto y compilar el archivo
  ssh "$USER@$ip" "cd /home/$USER/codigo/cmd/escritor/ && go build && mv escritor /home/$USER"
  ssh "$USER@$ip" "cd /home/$USER/codigo/cmd/lector/ && go build && mv lector /home/$USER"

  # Mover "users.txt" al directorio home
  ssh "$USER@$ip" "cd /home/$USER/codigo/ms/ && mv users.txt /home/$USER"
done

echo "Archivos copiados y compilados en las direcciones IP: ${direcciones[@]}"

exit 0