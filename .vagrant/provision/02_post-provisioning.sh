#!/usr/bin/env bash

VAGRANT_CORE_FOLDER="/vagrant"
VAGRANT_USERNAME="vagrant"
VAGRANT_USERHOME="/home/${VAGRANT_USERNAME}";

echo "Copyng dot files to ${VAGRANT_USERHOME}"
rsync -av "${VAGRANT_CORE_FOLDER}/files/dot/" "${VAGRANT_USERHOME}/" >/dev/null

echo "Updating npm to last version"
sudo npm install -g --unsafe-perm npm >/dev/null

echo "Installing necessary global node modules"
sudo npm install -g --unsafe-perm bower grunt-cli node-inspector gulp >/dev/null

echo "Installing dependency tool for go"
go get github.com/tools/godep

echo "Restoring homedir owner"
chown -R vagrant:vagrant "${VAGRANT_USERHOME}" &>/dev/null

sudo service nginx restart
sudo service mongod restart

# perform overall clean up
sudo apt-get -y update > /dev/null
sudo apt-get -y autoremove
sudo apt-get -y autoclean
