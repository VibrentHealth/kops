#!/bin/bash
env | grep "KOPS_BASE_URL=https://manand-state-store.s3.amazonaws.com/kops/1.12.1-rhel8"
if [ $? -eq 0 ]  #true if grep finds value
then
    echo "Good: env var needed for custom kops with RHEL8 support is installed"
else 
	echo "Bad: You may need to add 'export KOPS_BASE_URL=https://manand-state-store.s3.amazonaws.com/kops/1.12.1-rhel8' to your .bashrc or .zshrc file"
	echo "or try running this script from a new instance of your terminal"
fi

kops version | grep rhel8
if [ $? -eq 0 ]  #true if grep finds value
then
    echo "Good: custom kops binary with RHEL8 support is installed"
else 
    echo "Bad: You may need to run install-custom-kops.sh"
fi