#!/bin/bash

sudo docker-compose down 
sudo docker-compose rm -f $(sudo docker-compose ps -aq)
sudo docker volume prune -f
sudo docker network prune -f
sudo docker-compose up --build -d  
