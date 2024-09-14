import matplotlib.pyplot as plt
import seaborn as sns
import pandas as pd
import joblib
import base64
import tensorflow as tf
from io import BytesIO
from utils import parse_time
from prophet.serialize import model_from_json

class WorkflowAnalysisReporter:
    def __init__(self, data_path, rf_model_path, rnn_model_path, prophet_model_path):
        self.data = pd.read_csv(data_path)
        self.model = joblib.load(rf_model_path)
        self.rnn_model = tf.keras.models.load_model(rnn_model_path)
        with open(prophet_model_path, 'r') as fin:
            self.prophet_model = model_from_json(fin.read())
        self.visualizations = []

    def preprocess_data(self):
        self.data['StartedAt'] = pd.to_datetime(self.data['StartedAt'])
        self.data['CompletedAt'] = pd.to_datetime(self.data['CompletedAt'])
        self.data['ExecutionTime'] = (self.data['CompletedAt'] - self.data['StartedAt']).dt.total_seconds()

    def generate_execution_time_boxplot(self):
        plt.figure(figsize=(10, 6))
        sns.boxplot(x='StepName', y='ExecutionTime', data=self.data)
        plt.title('Execution Time Distribution by Step')
        plt.ylabel('Execution Time (seconds)')
        self._save_plot()

    def generate_failure_trend(self):
        daily_failure_rate = self.data.groupby(self.data['StartedAt'].dt.date)['Conclusion'].apply(lambda x: (x != 'success').mean()).reset_index()
        plt.figure(figsize=(12, 6))
        sns.lineplot(x='StartedAt', y='Conclusion', data=daily_failure_rate)
        plt.title('Daily Failure Rate Trend')
        plt.xlabel('Date')
        plt.ylabel('Failure Rate')
        plt.xticks(rotation=45)
        self._save_plot()

    def generate_rnn_prediction_plot(self):
        # Prepare sequence data for RNN prediction
        sequence_length = 10
        recent_data = self.data.sort_values('StartedAt').tail(sequence_length)
        X = recent_data[['StatusEncoded', 'ExecutionTime']].values.reshape(1, sequence_length, 2)

        # Make prediction
        prediction = self.rnn_model.predict(X)[0][0]

        plt.figure(figsize=(10, 6))
        plt.bar(['Failure', 'Success'], [1-prediction, prediction])
        plt.title('RNN Prediction for Next Workflow Run')
        plt.ylabel('Probability')
        self._save_plot()

    def generate_time_series_forecast_plot(self):
        future = self.prophet_model.make_future_dataframe(periods=30)
        forecast = self.prophet_model.predict(future)

        plt.figure(figsize=(12, 6))
        self.prophet_model.plot(forecast, uncertainty=True)
        plt.title('Time Series Forecast of Workflow Runs')
        plt.xlabel('Date')
        plt.ylabel('Number of Runs')
        self._save_plot()

    def _save_plot(self):
        buf = BytesIO()
        plt.savefig(buf, format='png')
        buf.seek(0)
        img_str = base64.b64encode(buf.read()).decode()
        self.visualizations.append(img_str)
        plt.close()

    def generate_html_report(self, output_path:str):
        html_content: str = f"""
        <html>
        <head>
            <title>Advanced Workflow Analysis Report</title>
            <style>
                body {{ font-family: Arial, sans-serif; margin: 0 auto; max-width: 800px; }}
                h1, h2 {{ color: #333366; }}
                img {{ max-width: 100%; height: auto; }}
            </style>
        </head>
        <body>
            <h1>Advanced Workflow Analysis Report</h1>

            <h2>Execution Time Distribution by Job</h2>
            <img src="data:image/png;base64,{self.visualizations[0]}">
            <p>This box plot displays the distribution of execution times for each programming language.</p>

            <h2>Daily Failure Rate Trend</h2>
            <img src="data:image/png;base64,{self.visualizations[1]}">
            <p>This line graph shows the trend of daily failure rates over time.</p>

            <h2>RNN Prediction for Next Workflow Run</h2>
            <img src="data:image/png;base64,{self.visualizations[2]}">
            <p>This bar chart shows the predicted probability of success for the next workflow run based on recent history.</p>

            <h2>Time Series Forecast of Workflow Runs</h2>
            <img src="data:image/png;base64,{self.visualizations[3]}">
            <p>This plot shows the forecasted number of workflow runs for the next 30 days, including uncertainty intervals.</p>
        </body>
        </html>
        """

        with open(output_path, 'w') as hout:
            hout.write(html_content)

# Usage
reporter = WorkflowAnalysisReporter('workflow_data.csv', 'workflow_models_rf.joblib', 'workflow_models_rnn.keras', 'workflow_models_prophet.json')
reporter.preprocess_data()
reporter.generate_execution_time_boxplot()
reporter.generate_failure_trend()
reporter.generate_rnn_prediction_plot()
reporter.generate_time_series_forecast_plot()
reporter.generate_html_report('advanced_workflow_analysis_report.html')
