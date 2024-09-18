import matplotlib.pyplot as plt
import seaborn as sns
import pandas as pd
import numpy as np
import joblib
import base64
import tensorflow as tf
from io import BytesIO
from prophet.serialize import model_from_json

class WorkflowAnalysisReporter:
    def __init__(self, data_path, rf_model_path, regression_model_path, rnn_model_path, prophet_model_path):
        self.data = pd.read_csv(data_path)
        self.rf_model = joblib.load(rf_model_path)
        self.regression_model = joblib.load(regression_model_path)
        self.rnn_model = tf.keras.models.load_model(rnn_model_path)
        with open(prophet_model_path, 'r') as fin:
            self.prophet_model = model_from_json(fin.read())
        self.visualizations = []

    def preprocess_data(self):
        self.data['StartedAt'] = pd.to_datetime(self.data['StartedAt'])
        self.data['CompletedAt'] = pd.to_datetime(self.data['CompletedAt'])
        self.data['ExecutionTime'] = (self.data['CompletedAt'] - self.data['StartedAt']).dt.total_seconds()
        self.data['StatusEncoded'] = pd.Categorical(self.data['Status']).codes
        self.data['SuccessEncoded'] = (self.data['Conclusion'] == 'success').astype(int)

    def generate_execution_time_boxplot(self):
        plt.figure(figsize=(10, 6))
        sns.boxplot(x='Status', y='ExecutionTime', data=self.data)
        plt.title('Execution Time Distribution by Status')
        plt.ylabel('Execution Time (seconds)')
        plt.xticks(rotation=45)
        self._save_plot()

    def generate_failure_trend(self):
        daily_failure_rate = self.data.groupby(self.data['StartedAt'].dt.date)['SuccessEncoded'].apply(lambda x: 1 - x.mean()).reset_index()
        plt.figure(figsize=(12, 6))
        sns.lineplot(x='StartedAt', y='SuccessEncoded', data=daily_failure_rate)
        plt.title('Daily Failure Rate Trend')
        plt.xlabel('Date')
        plt.ylabel('Failure Rate')
        plt.xticks(rotation=45)
        self._save_plot()

    def generate_rf_feature_importance(self):
        feature_importance = self.rf_model.feature_importances_
        feature_names = ['StatusEncoded', 'ExecutionTime']
        plt.figure(figsize=(10, 6))
        sns.barplot(x=feature_importance, y=feature_names)
        plt.title('Random Forest Feature Importance')
        plt.xlabel('Importance')
        self._save_plot()

    def generate_regression_prediction_plot(self):
        X = self.data[['StatusEncoded', 'ExecutionTime']]
        y_pred = self.regression_model.predict(X)
        plt.figure(figsize=(10, 6))
        plt.scatter(self.data['ExecutionTime'], y_pred, alpha=0.5)
        plt.plot([self.data['ExecutionTime'].min(), self.data['ExecutionTime'].max()],
                 [self.data['ExecutionTime'].min(), self.data['ExecutionTime'].max()],
                 'r--', lw=2)
        plt.title('Regression Model: Predicted vs Actual Execution Time')
        plt.xlabel('Actual Execution Time')
        plt.ylabel('Predicted Execution Time')
        self._save_plot()

    def generate_rnn_prediction_plot(self):
        sequence_length = 10
        recent_data = self.data.sort_values('StartedAt').tail(sequence_length)
        X = recent_data[['StatusEncoded', 'ExecutionTime']].values.reshape(1, sequence_length, 2)
        X[:,:,1] = (X[:,:,1] - np.mean(X[:,:,1])) / np.std(X[:,:,1])  # Normalize ExecutionTime
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

    def generate_html_report(self, output_path):
        html_content = f"""
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

            <h2>Execution Time Distribution by Status</h2>
            <img src="data:image/png;base64,{self.visualizations[0]}">
            <p>This box plot displays the distribution of execution times for each workflow status.</p>

            <h2>Daily Failure Rate Trend</h2>
            <img src="data:image/png;base64,{self.visualizations[1]}">
            <p>This line graph shows the trend of daily failure rates over time.</p>

            <h2>Random Forest Feature Importance</h2>
            <img src="data:image/png;base64,{self.visualizations[2]}">
            <p>This bar plot shows the importance of each feature in the Random Forest model.</p>

            <h2>Regression Model: Predicted vs Actual Execution Time</h2>
            <img src="data:image/png;base64,{self.visualizations[3]}">
            <p>This scatter plot compares the predicted execution times from the regression model with the actual execution times.</p>

            <h2>RNN Prediction for Next Workflow Run</h2>
            <img src="data:image/png;base64,{self.visualizations[4]}">
            <p>This bar chart shows the probability of success and failure for the next workflow run, as predicted by the RNN model.</p>

            <h2>Time Series Forecast of Workflow Runs</h2>
            <img src="data:image/png;base64,{self.visualizations[5]}">
            <p>This plot shows the time series forecast of workflow runs using the Prophet model.</p>
        </body>
        </html>
        """

        with open(output_path, 'w') as hout:
            hout.write(html_content)

# Usage
reporter = WorkflowAnalysisReporter(
    'workflow_data.csv',
    'workflow_models_rf.joblib',
    'workflow_models_regression.joblib',
    'workflow_models_rnn.keras',
    'workflow_models_prophet.json'
)
reporter.preprocess_data()
reporter.generate_execution_time_boxplot()
reporter.generate_failure_trend()
reporter.generate_rf_feature_importance()
reporter.generate_regression_prediction_plot()
reporter.generate_rnn_prediction_plot()
reporter.generate_time_series_forecast_plot()
reporter.generate_html_report('advanced_workflow_analysis_report.html')
