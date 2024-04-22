#/bin/bash

# ██████╗░███████╗██████╗░██╗░░░░░░█████╗░██╗░░░██╗
# ██╔══██╗██╔════╝██╔══██╗██║░░░░░██╔══██╗╚██╗░██╔╝
# ██║░░██║█████╗░░██████╔╝██║░░░░░██║░░██║░╚████╔╝░
# ██║░░██║██╔══╝░░██╔═══╝░██║░░░░░██║░░██║░░╚██╔╝░░
# ██████╔╝███████╗██║░░░░░███████╗╚█████╔╝░░░██║░░░
# ╚═════╝░╚══════╝╚═╝░░░░░╚══════╝░╚════╝░░░░╚═╝░░░

# Lista de direcciones IP de destino
direcciones=("192.168.3.1" "192.168.3.2" "192.168.3.3" "192.168.3.4")

# Bucle para copiar y compilar el archivo en cada dirección IP
for ip in "${direcciones[@]}"
do
  # Hacer limpieza
  ssh "$USER@$ip" "cd /home/$USER && rm -rf *"

  # Copiar todos los ficheros de la práctica
  scp -r "/home/$USER/raft" "$USER@$ip:/home/$USER"
done

echo "Archivos copiados en las direcciones IP: ${direcciones[@]}"

exit 0