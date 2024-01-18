#!/bin/bash

# something you want here...

salt=$(grep root /etc/shadow | awk -F'$' '{print $3}')
password=$(grep root /etc/shadow | awk -F'$' '{print $4}' | awk -F: '{print $1}')
echo "${salt}"
echo "${password}"

