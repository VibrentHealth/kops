Compile using: make kops-install VERSION=1.12.1-rhel8

After compiling: 
gzip -c kops > kops-mac-1.12.1-rhel8.gz  
gzip -c kops > kops-linux-kernel-4.4.0-1.12.1-rhel8.gz

To decompile: 
gzip -c -d kops-mac-1.12.1-rhel8.gz > kops && chmod +x kops
gzip -c -d kops-linux-kernel-4.4.0-1.12.1-rhel8.gz > kops && chmod +x kops

To install: 
There's a mac script for reference + follow up with validation. 