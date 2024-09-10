#!/bin/bash

# Initialize an empty string for the log paths
log_paths=""

# Get all container IDs
container_ids=$(docker ps -q)

# Loop through each container
for container_id in $container_ids
do
    # Get container name
    container_name=$(docker inspect --format='{{.Name}}' $container_id | sed 's/\///')

    # Get the log path
    log_path=$(docker inspect --format='{{.LogPath}}' $container_id)

    # Append to the log_paths string
    if [ -z "$log_paths" ]; then
        log_paths="${container_name}:${log_path}"
    else
        log_paths="${log_paths},${container_name}:${log_path}"
    fi
done

if [ -n "$log_paths" ]; then
    # Export the log paths as an environment variable
    export CONTAINER_LOG_PATHS="$log_paths"
    echo "Container log paths have been exported as CONTAINER_LOG_PATHS environment variable"
    echo "CONTAINER_LOG_PATHS=$CONTAINER_LOG_PATHS"
else
    echo "No container log paths were found"
fi
