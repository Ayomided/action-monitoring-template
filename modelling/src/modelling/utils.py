import pandas as pd

def parse_time(ts: str) -> pd.Timestamp:
    return pd.to_datetime(ts.replace(" UTC", ""), format='%Y-%m-%d %H:%M:%S %z').tz_localize(None)
