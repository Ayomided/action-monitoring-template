#!/usr/bin/env bash`

log_file="runner.log"

echo "√ Connected to GitHub" > $log_file
echo "" >> $log_file
echo "Current runner version: '2.317.0'" >> $log_file
echo "$(date -u '+%Y-%m-%d %H:%M:%SZ'): Listening for Jobs" >> $log_file

while true; do
    sleep $((RANDOM % 60 + 30))  # Wait for 30-90 seconds

    start_time=$(date -u '+%Y-%m-%d %H:%M:%SZ')
    echo "$start_time: Running job: update gist" >> $log_file

    sleep $((RANDOM % 20 + 10))  # Job runs for 10-30 seconds

    end_time=$(date -u '+%Y-%m-%d %H:%M:%SZ')
    echo "$end_time: Job update gist completed with result: Succeeded" >> $log_file

    sleep $((RANDOM % 60 + 30))  # Wait for 30-90 seconds

    start_time=$(date -u '+%Y-%m-%d %H:%M:%SZ')
    echo "$start_time: Running job: run go lint" >> $log_file

    sleep $((RANDOM % 20 + 10))  # Job runs for 10-30 seconds

    end_time=$(date -u '+%Y-%m-%d %H:%M:%SZ')
    echo "$end_time: Job run go lint completed with result: Succeeded" >> $log_file

    sleep $((RANDOM % 60 + 30))  # Wait for 30-90 seconds

    start_time=$(date -u '+%Y-%m-%d %H:%M:%SZ')
    echo "$start_time: Running job: Build(golang)" >> $log_file

    sleep $((RANDOM % 20 + 10))  # Job runs for 10-30 seconds

    end_time=$(date -u '+%Y-%m-%d %H:%M:%SZ')
    echo "$end_time: Job Build(golang) completed with result: Failed" >> $log_file
done
