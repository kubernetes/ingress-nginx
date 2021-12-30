#!/usr/bin/bash
counter=0
declare -A counts
for i in $(seq 1 100)
do
	        #below the IP of the LoadBalancer is given in curl command
		output=`curl -s http://172.18.0.251/ | grep Hostname | awk '{print $2}'`
		let "counts[$output] += 1"


done

for key in "${!counts[@]}"; do
	echo "Total number of requests routed to Pod $key are: ${counts[$key]}"
done

exit 0



