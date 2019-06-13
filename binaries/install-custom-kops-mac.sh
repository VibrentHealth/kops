#!/bin/bash
kops version | grep rhel8
if [ $? -eq 0 ]  #true if grep finds value
then
    echo "custom kops binary with RHEL8 support is installed, please run installation-validation.sh to verify env var"
else #user prompt to install custom kops (read user input doesn't work when launched from sublime text)
    echo "You don't have custom kops with RHEL8 support installed"
    echo "it requires 'export KOPS_BASE_URL=https://manand-state-store.s3.amazonaws.com/kops/1.12.1-rhel8' added as perm env var to your .bashrc or .zshrc file"
    read -p "Do you want to download and install mac custom kops binary 1.12.1 (with RHEL8 support)? (y = yes)" -n 1 -r
    echo #move to a new line
    if [[ $REPLY =~ ^[Yy]$ ]]
    then
        #export currentloc=`which kops` #returns /usr/local/bin/kops #It's possible it's not already installed
        mv /usr/local/bin/kops /usr/local/bin/kops.backup  #If this command fails won't hurt
        gzip -c -d kops-mac-1.12.1-rhel8.gz > kops && chmod +x kops
        mv kops /usr/local/bin/kops
    fi
fi #end of if statement
