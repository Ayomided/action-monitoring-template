import pandas as pd

df = pd.read_csv('/Users/davidayomide/Downloads/Dev/FINALPROJ/action-monitoring-template/modelling/MINE/all_steps.csv')
df['StartedAt'] = pd.to_datetime(df['StartedAt'], utc=True)
df['CompletedAt'] = pd.to_datetime(df['CompletedAt'], utc=True)

print(df.head(5))
