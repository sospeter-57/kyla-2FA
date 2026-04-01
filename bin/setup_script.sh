#!/usr/bash

sudo su
cd /opt/
mkdir kyla-2fa
cd kyla-2fa
wget https://github.com/sospeter-57/kyla-2FA/blob/main/bin/kyla-2fa
rm /usr/bin/kyla-2fa
ln -s /opt/kyla-2fa/kyla-2fa /usr/bin/kyla-2fa

echo "\n\nFinished Setting up kyla-2fa\n"
echo "run kyla-2fa on on your command line from any location\n"

