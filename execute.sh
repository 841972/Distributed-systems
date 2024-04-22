#!/bin/bash

user="a842236"

# Set the commands to execute
command1="ssh -t $user@central.cps.unizar.es \"cd /home/$user/practica1/cmd/server ; bash --login\""
command2="ssh -t $user@central.cps.unizar.es \"cd /home/$user/practica1/cmd/client ; bash --login\""

# Open new terminal windows and execute the commands
gnome-terminal -- bash -c "$command1; exec bash"
gnome-terminal -- bash -c "$command2; exec bash"
