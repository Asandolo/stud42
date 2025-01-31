#! /bin/bash

cat << EOF
         {                  
      {   }                 
       }_{ __{              
    .-{   }   }-.           
   (   }     {   )          First installation started...
   |'-.._____..-'|          
   |             ;--.       Due to the installation of internet on your machine,
   |            (__  \      this task can be long, very long. But don't worry
   |             | )  )     it's executed only the first time the devcontainer
   |             |/  /      is build (or rebuild). 
   |             /  /       
   |            (  /        Take a break, take a coffee, browse the issues
   \             y'         
    '-.._____..-'           
EOF

git config --global --add safe.directory /workspace
git config pull.rebase true
git config core.hooksPath /workspace/githooks
git config commit.gpgsign true
git update-index --assume-unchanged .devcontainer/.env

# Disable git info in zsh promtp due to performence issue
# https://github.com/microsoft/vscode/issues/133215
git config oh-my-zsh.hide-info 1

make -f build/Makefile devcontainer-init