#1/bin/bash

git pull 

soda migrate 

go build -o bookings cmd/web/*.go 

sudo supervisiorctl stop book 
sudo supervisiorctl start book

#shell script to restart application and update migrations if the server shutdowns or encounters an error or is updated
#cmod 777 update.sh to make it executable