JobID,RunID,Attempt,Status,Conclusion,JobName,StartedAt,CompletedAt,StepName
,Number,JobID,StepName,Status,Conclusion,StartedAt,CompletedAt

sed -i '1c\Number JobID StepName Status Conclusion StartedAt CompletedAt' all_steps.csv
